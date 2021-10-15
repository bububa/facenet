package facenet

import (
	"errors"
	"image"
	"sync"

	"github.com/llgcode/draw2d"

	"github.com/bububa/facenet/core"
	"github.com/bububa/facenet/imageutil"
)

// Estimator represents facenet estimator
type Estimator struct {
	model *core.Net
	db    *Storage
	font  *imageutil.Font
	lock  *sync.RWMutex
}

// New init a new facenet Estimator
func New(opts ...Option) (*Estimator, error) {
	instance := &Estimator{
		lock: new(sync.RWMutex),
	}
	for _, opt := range opts {
		if err := opt.apply(instance); err != nil {
			return nil, err
		}
	}
	return instance, nil
}

// SetModel set face net model
func (ins *Estimator) SetModel(net *core.Net) {
	ins.lock.Lock()
	defer ins.lock.Unlock()
	ins.model = net
}

// SetDB set db
func (ins *Estimator) SetDB(db *Storage) {
	ins.lock.Lock()
	defer ins.lock.Unlock()
	ins.db = db
}

// LoadDB load db file
func (ins *Estimator) LoadDB(fname string) error {
	ins.lock.Lock()
	defer ins.lock.Unlock()
	if ins.db == nil {
		ins.db = NewStorage(nil, nil)
	}
	return ins.db.Load(fname)
}

// SaveDB save db file
func (ins *Estimator) SaveDB(fname string) error {
	ins.lock.RLock()
	defer ins.lock.RUnlock()
	if ins.db == nil {
		return nil
	}
	return ins.db.Save(fname)
}

// SetFont set font
func (ins *Estimator) SetFont(data *draw2d.FontData, size float64) error {
	ins.lock.Lock()
	defer ins.lock.Unlock()
	if ins.font == nil {
		ins.font = new(imageutil.Font)
	}
	ins.font.Size = size
	ins.font.Data = data
	if ins.font.Cache != nil {
		return ins.font.Load(ins.font.Cache)
	}
	return nil
}

// SetFontSize set font size
func (ins *Estimator) SetFontSize(size float64) {
	ins.lock.Lock()
	defer ins.lock.Unlock()
	if ins.font == nil {
		ins.font = new(imageutil.Font)
	}
	ins.font.Size = size
}

// SetFontCache set font cache
func (ins *Estimator) SetFontCache(cache draw2d.FontCache) error {
	ins.lock.Lock()
	defer ins.lock.Unlock()
	if ins.font == nil {
		ins.font = new(imageutil.Font)
	}
	ins.font.Cache = cache
	if ins.font.Data != nil {
		return ins.font.Load(ins.font.Cache)
	}
	return nil
}

// SetFontPath set font cache with font cache path
func (ins *Estimator) SetFontPath(cachePath string) error {
	ins.lock.Lock()
	defer ins.lock.Unlock()
	if ins.font == nil {
		ins.font = new(imageutil.Font)
	}
	ins.font.Cache = imageutil.NewFontCache(cachePath)
	if ins.font.Data != nil {
		return ins.font.Load(ins.font.Cache)
	}
	return nil
}

// People get people
func (ins *Estimator) People() *core.People {
	if ins.db == nil {
		return nil
	}
	return ins.db.People()
}

// PeopleSafe get people (mutlthread safe)
func (ins *Estimator) PeopleSafe() *core.People {
	ins.lock.RLock()
	defer ins.lock.RUnlock()
	return ins.People()
}

// AddPerson add person to people
func (ins *Estimator) AddPerson(items ...*core.Person) {
	if ins.db == nil {
		return
	}
	ins.db.Add(items...)
}

// AddPersonSafe add person to people (multithread safe)
func (ins *Estimator) AddPersonSafe(items ...*core.Person) {
	ins.lock.Lock()
	defer ins.lock.Unlock()
	ins.AddPerson(items...)
}

// DeletePerson delete a person by name
func (ins *Estimator) DeletePerson(name string) bool {
	if ins.db == nil {
		return false
	}
	return ins.db.Delete(name)
}

// DeletePersonSafe delete a person by name (multithread safe)
func (ins *Estimator) DeletePersonSafe(name string) bool {
	ins.lock.Lock()
	defer ins.lock.Unlock()
	return ins.DeletePerson(name)
}

// Match match a person with embedding
func (ins *Estimator) Match(embedding []float32) (*core.Person, float64, error) {
	if ins.db == nil {
		return nil, 0, errors.New("no db inited")
	}
	return ins.db.Match(embedding)
}

// MatchSafe match a person with embedding (multithread safe)
func (ins *Estimator) MatchSafe(embedding []float32) (*core.Person, float64, error) {
	ins.lock.RLock()
	defer ins.lock.RUnlock()
	return ins.Match(embedding)
}

// Predict returns embedding predicted results
func (ins *Estimator) Predict(embedding []float32) ([]*core.Person, []float64, error) {
	if ins.db == nil {
		return nil, nil, errors.New("no db inited")
	}
	return ins.db.Predict(embedding)
}

// PredictSafe returns embedding predicted results (multithread safe)
func (ins *Estimator) PredictSafe(embedding []float32) ([]*core.Person, []float64, error) {
	ins.lock.RLock()
	defer ins.lock.RUnlock()
	return ins.Predict(embedding)
}

// ExtractFace extract face for a person from image
func (ins *Estimator) ExtractFace(person *core.Person, img image.Image, minSize int) (*core.FaceMarker, error) {
	if ins.model == nil {
		return nil, errors.New("model not inited")
	}
	face, err := ins.model.DetectSingle(img, minSize)
	if err != nil {
		return nil, err
	}
	person.Embeddings = append(person.Embeddings, &core.Person_Embedding{
		Value: face.Embeddings[0],
	})
	return core.NewFaceMarker(face, person.GetName(), 1), nil
}

// ExtractFaceSafe extract face for a person from image (multithread safe)
func (ins *Estimator) ExtractFaceSafe(person *core.Person, img image.Image, minSize int) (*core.FaceMarker, error) {
	ins.lock.RLock()
	defer ins.lock.RUnlock()
	return ins.ExtractFace(person, img, minSize)
}

// DetectFaces detect face markers from image
func (ins *Estimator) DetectFaces(img image.Image, minSize int) (*core.FaceMarkers, error) {
	if ins.model == nil {
		return nil, errors.New("model not inited")
	}
	faces, err := ins.model.DetectMultiple(img, minSize)
	if err != nil {
		return nil, err
	}
	markers := core.NewFaceMarkers(img)
	for _, face := range faces {
		person, distance, err := ins.Match(face.Embeddings[0])
		marker := core.NewFaceMarker(face, person.GetName(), distance)
		if err != nil {
			marker.SetError(err)
		}
		markers.Append(*marker)
	}
	return markers, nil
}

// DetectFacesSafe detect face markers from image (multithread safe)
func (ins *Estimator) DetectFacesSafe(img image.Image, minSize int) (*core.FaceMarkers, error) {
	ins.lock.RLock()
	defer ins.lock.RUnlock()
	return ins.DetectFaces(img, minSize)
}

// Train for trainging classifier
func (ins *Estimator) Train(split float64, iterations int, verbosity int) {
	if ins.db == nil {
		return
	}
	ins.db.Train(split, iterations, verbosity)
}

// TrainSafe for trainging classifier (multithread safe)
func (ins *Estimator) TrainSafe(split float64, iterations int, verbosity int) {
	ins.lock.RLock()
	defer ins.lock.RUnlock()
	ins.Train(split, iterations, verbosity)
}

// BatchTrain for trainging classifier
func (ins *Estimator) BatchTrain(split float64, iterations int, verbosity int, batch int) {
	if ins.db == nil {
		return
	}
	ins.db.BatchTrain(split, iterations, verbosity, batch)
}

// BatchTrainSafe for trainging classifier (multithread safe)
func (ins *Estimator) BatchTrainSafe(split float64, iterations int, verbosity int, batch int) {
	ins.lock.RLock()
	defer ins.lock.RUnlock()
	ins.BatchTrain(split, iterations, verbosity, batch)
}

// DrawMarkers draw face markers on image
func (ins *Estimator) DrawMarkers(markers *core.FaceMarkers, txtColor string, successColor string, failedColor string, strokeWidth float64, succeedOnly bool) image.Image {
	return markers.Draw(ins.font, txtColor, successColor, failedColor, strokeWidth, succeedOnly)
}
