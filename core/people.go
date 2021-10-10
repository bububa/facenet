package core

import (
	"fmt"
	"io"
	"strings"

	"github.com/montanaflynn/stats"
	"google.golang.org/protobuf/proto"
)

// LoadPeople load people model
func LoadPeople(r io.Reader, people *People) error {
	buf, err := io.ReadAll(r)
	if err != nil {
		return err
	}
	if err = proto.Unmarshal(buf, people); err != nil {
		return err
	}
	people.Setup()
	return nil
}

// Save save people to a model file
func (people *People) Save(w io.Writer) error {
	people.Setup()
	buf, err := proto.Marshal(people)
	if err != nil {
		return err
	}
	_, err = w.Write(buf)
	return err
}

// Setup recalculate people's center and collisions
func (people *People) Setup() {
	for _, person := range people.GetList() {
		person.ReCenter()
	}
	people.ResolveCollisions()
}

// Delete delete a person from people
func (people *People) Delete(name string) bool {
	var deleted bool
	name = strings.TrimSpace(name)
	list := people.GetList()
	l := len(list)
	for i := 0; i < l; i++ {
		if list[i].GetName() == name {
			people.List = append(people.List[:i], people.List[i+1:]...)
			deleted = true
			break
		}
	}
	return deleted
}

// Append append person to people and update when duplicate
func (people *People) Append(items ...*Person) {
	list := people.GetList()
	exists := make(map[string]struct{}, len(list))
	for _, p := range list {
		exists[p.GetName()] = struct{}{}
	}
	replaces := make(map[string]*Person, len(items))
	for _, item := range items {
		if _, found := exists[item.GetName()]; !found {
			list = append(list, item)
			exists[item.GetName()] = struct{}{}
		} else {
			replaces[item.GetName()] = item
		}
	}
	l := len(list)
	for i := 0; i < l; i++ {
		name := list[i].GetName()
		if person, found := replaces[name]; found {
			list[i] = person
		}
	}
	people.List = list
}

// Match match a person from people based on embedding
func (people *People) Match(embedding []float32) (*Person, float64, error) {
	person, dist := people.Nearest(embedding)
	// Any reasons embeddings do not match this face?
	switch {
	case dist < 0:
		// Should never happen.
		return person, dist, NewError(NegativeDistanceMatchErr, fmt.Sprintf("distance is too small, %f", dist))
	case dist > (person.GetRadius() + MatchDist):
		// Too far.
		return person, dist, NewError(TooFarMatchErr, fmt.Sprintf("distance is too far, %f", dist))
	case person.GetCollisionRadius() > 0.1 && dist > person.GetCollisionRadius():
		// log.Printf("person: %s, collision: %f, dist: %f\n", person.Name, collisionRadius, dist)
		// Within radius of reported collisions.
		return person, dist, NewError(CollisionMatchErr, fmt.Sprintf("distance(%f) is larger than collision radius(%f)", dist, person.GetCollisionRadius()))
	}
	return person, dist, nil
}

// Nearest returns nearest person in people
func (people *People) Nearest(embedding []float32) (*Person, float64) {
	var ret *Person
	dist := -1.0

	// Find the nearest person for this data point
	for _, person := range people.GetList() {
		// d := EuclideanDistance(person.GetCenter())
		for _, embeddingObj := range person.GetEmbeddings() {
			d := EuclideanDistance(embedding, embeddingObj.GetValue())
			if ret == nil || d < dist {
				dist = d
				ret = person
			}
		}
	}
	return ret, dist
}

// Neighbour returns neighbour person in people
func (people *People) Neighbour(embedding []float32, nearest *Person) (*Person, float64) {
	var ret *Person
	var dist float64

	// Find the nearest person for this data point
	for _, person := range people.GetList() {
		if person.Equal(nearest) {
			continue
		}
		d := person.AverageDistance(embedding)
		if ret == nil || d < dist {
			dist = d
			ret = person
		}
	}
	return ret, dist
}

// ResolveCollisions resolves collisions of different subject's faces.
func (people *People) ResolveCollisions() {
	list := people.GetList()
	for _, f1 := range list {
		for _, f2 := range list {
			if f1.Equal(f2) {
				continue
			}
			f1.ResolveCollision(f2)
		}
	}
}

// Equal check two person are equal
func (person *Person) Equal(another *Person) bool {
	return person.Name == another.Name
}

// Append append embedding to person
func (person *Person) Append(embedding []float32) {
	person.Embeddings = append(person.Embeddings, &Person_Embedding{Value: embedding})
}

// ReCenter recalculate person's center
func (person *Person) ReCenter() {
	person.Center, person.Radius, _ = person.CalcCenter()
}

// CalcCenter returns the center coordinates of a set of people
func (person *Person) CalcCenter() (result []float32, radius float64, count int) {
	embeddings := person.GetEmbeddings()
	count = len(embeddings)
	// No embeddings?
	if count == 0 {
		return result, radius, count
	} else if count == 1 {
		return embeddings[0].GetValue(), 0.0, count
	}
	dim := len(embeddings[0].GetValue())

	// No embedding values?
	if dim == 0 {
		return []float32{}, 0.0, count
	}

	result = make([]float32, dim)

	// The mean of a set of vectors is calculated component-wise.
	for i := 0; i < dim; i++ {
		values := make(stats.Float64Data, count)

		for j := 0; j < count; j++ {
			values[j] = float64(embeddings[j].GetValue()[i])
		}

		if m, err := stats.Mean(values); err != nil {
			continue
		} else {
			result[i] = float32(m)
		}
	}

	// Radius is the max embedding distance + 0.01 from result.
	for _, emb := range embeddings {
		if d := EuclideanDistance(result, emb.GetValue()); d > radius {
			radius = d + 0.01
		}
	}

	return result, radius, count
}

// AverageDistance returns the average distance between o and all people
func (person *Person) AverageDistance(embedding []float32) float64 {
	var d float64
	var l int

	for _, p := range person.GetEmbeddings() {
		dist := EuclideanDistance(p.GetValue(), embedding)
		if dist == 0 {
			continue
		}

		l++
		d += dist
	}

	if l == 0 {
		return 0
	}
	return d / float64(l)
}

// Match match embedding with a person
func (person *Person) Match(embedding []float32) (bool, float64) {
	personEmbeddings := person.GetEmbeddings()
	var dist float64 = -1

	if len(personEmbeddings) == 0 {
		// Np embeddings, no match.
		return false, dist
	}

	for _, personEmbedding := range personEmbeddings {
		// Calculate smallest distance to embeddings.
		if d := EuclideanDistance(embedding, personEmbedding.GetValue()); d < dist || dist < 0 {
			dist = d
		}
	}

	// Any reasons embeddings do not match this face?
	switch {
	case dist < 0:
		// Should never happen.
		return false, dist
	case dist > (person.Radius + MatchDist):
		// Too far.
		return false, dist
	case person.GetCollisionRadius() > 0.1 && dist > person.GetCollisionRadius():
		// Within radius of reported collisions.
		return false, dist
	}
	// If not, at least one of the embeddings match!
	return true, dist

}

// ResolveCollision calculate CollisionRadius for a person
func (person *Person) ResolveCollision(p2 *Person) {
	for _, embedding := range p2.GetEmbeddings() {
		if matched, dist := person.Match(embedding.GetValue()); matched && (person.CollisionRadius < 0.1 || person.CollisionRadius < dist-0.01) {
			person.CollisionRadius = dist - 0.01
		}
	}
}
