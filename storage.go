package facenet

import (
	"archive/zip"
	"io"
	"os"

	"google.golang.org/protobuf/proto"

	"github.com/bububa/facenet/classifier"
	"github.com/bububa/facenet/core"
)

const (
	// PeopleFilename represents people data filename in zip
	PeopleFilename = "people.pb"
	// ClassifierFilename represents classifier data filename in zip
	ClassifierFilename = "classifier.model"
)

// Storage represents db storage
type Storage struct {
	people     *core.People
	classifier classifier.Classifier
}

// NewStorage returns new Storage
func NewStorage(people *core.People, classifier classifier.Classifier) *Storage {
	return &Storage{
		people:     people,
		classifier: classifier,
	}
}

// Load load storage from file
func (s *Storage) Load(fname string) error {
	if s.people == nil {
		s.people = new(core.People)
	}
	zipFn, err := zip.OpenReader(fname)
	if err != nil {
		if os.IsNotExist(err) {
			s.classifier = classifier.NewDefault()
			return nil
		}
		return err
	}
	defer zipFn.Close()
	for _, f := range zipFn.File {
		info := f.FileInfo()
		if info.IsDir() {
			continue
		}
		switch info.Name() {
		case PeopleFilename:
			r, err := f.Open()
			if err != nil {
				return err
			}
			buf, err := io.ReadAll(r)
			if err != nil {
				return err
			}
			if err = proto.Unmarshal(buf, s.people); err != nil {
				return err
			}
			s.people.Setup()
		case ClassifierFilename:
			r, err := f.Open()
			if err != nil {
				return err
			}
			s.classifier = new(classifier.Neural)
			s.classifier.Read(r)
		}
	}
	return nil
}

// Save save storage to file
func (s *Storage) Save(fname string) error {
	fn, err := os.Create(fname)
	if err != nil {
		return err
	}
	defer fn.Close()
	// 创建zip
	zipWriter := zip.NewWriter(fn)
	defer zipWriter.Close()
	if s.people != nil {
		peopleFn, err := zipWriter.Create(PeopleFilename)
		if err != nil {
			return err
		}
		if err := s.people.Save(peopleFn); err != nil {
			return err
		}
	}
	if s.classifier != nil {
		classifierFn, err := zipWriter.Create(ClassifierFilename)
		if err != nil {
			return err
		}
		s.classifier.Write(classifierFn)
	}
	return nil
}

// SetClassifier set classifier
func (s *Storage) SetClassifier(c classifier.Classifier) {
	s.classifier = c
}

// People returns people
func (s *Storage) People() *core.People {
	return s.people
}

// Add add person to people
func (s *Storage) Add(items ...*core.Person) {
	if s.people == nil {
		s.people = new(core.People)
	}
	s.people.Append(items...)
}

// Delete delete a person by name
func (s *Storage) Delete(name string) bool {
	if s.people == nil {
		return false
	}
	return s.people.Delete(name)
}

// Predict returns predictation results
func (s *Storage) Predict(input []float32) ([]*core.Person, []float64, error) {
	scores := s.classifier.Predict(input)
	if len(scores) == 0 {
		return nil, nil, core.NewError(core.NothingMatchErr, "no match results")
	}
	ret := make([]*core.Person, 0, len(scores))
	people := s.people.GetList()
	for idx := range scores {
		ret = append(ret, people[idx])
	}
	return ret, scores, nil
}

// Match returns best match result
func (s *Storage) Match(input []float32) (*core.Person, float64, error) {
	if s.classifier == nil {
		return s.people.Match(input)
	}
	idx, score := s.classifier.Match(input)
	if idx < 0 {
		return nil, score, core.NewError(core.NothingMatchErr, "no match results")
	}
	people := s.People().GetList()
	return people[idx], score, nil
}

// Train for trainging classifier
func (s *Storage) Train(split float64, iterations int, verbosity int) {
	s.classifier.Train(s.people, split, iterations, verbosity)
}

// BatchTrain for trainging classifier
func (s *Storage) BatchTrain(split float64, iterations int, verbosity int, batch int) {
	s.classifier.BatchTrain(s.people, split, iterations, verbosity, batch)
}
