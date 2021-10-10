package facenet

import (
	"image"
	"os"

	"github.com/llgcode/draw2d"

	"github.com/bububa/facenet/core"
	"github.com/bububa/facenet/imageutil"
)

// Instance facenet Instance struct
type Instance struct {
	net    *core.Net
	people *core.People
	font   *imageutil.Font
}

// New init a new facenet instance
func New(opts ...Option) (*Instance, error) {
	instance := &Instance{}
	for _, opt := range opts {
		if err := opt.apply(instance); err != nil {
			return nil, err
		}
	}
	return instance, nil
}

// SetNet set face net model
func (ins *Instance) SetNet(net *core.Net) {
	ins.net = net
}

// SetPeople set people model
func (ins *Instance) SetPeople(people *core.People) {
	ins.people = people
}

// LoadPeople load people with people model filepath
func (ins *Instance) LoadPeople(filePath string) error {
	if ins.people == nil {
		ins.people = new(core.People)
	}
	fn, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer fn.Close()
	return core.LoadPeople(fn, ins.people)
}

// SaveModel save people model to model file
func (ins *Instance) SaveModel(modelPath string) error {
	fn, err := os.OpenFile(modelPath, os.O_RDWR|os.O_CREATE, 0600)
	if err != nil {
		return err
	}
	if err := fn.Truncate(0); err != nil {
		return err
	}
	fn.Seek(0, 0)
	return ins.people.Save(fn)
}

// SetFont set font
func (ins *Instance) SetFont(data *draw2d.FontData, size float64) error {
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
func (ins *Instance) SetFontSize(size float64) {
	if ins.font == nil {
		ins.font = new(imageutil.Font)
	}
	ins.font.Size = size
}

// SetFontCache set font cache
func (ins *Instance) SetFontCache(cache draw2d.FontCache) error {
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
func (ins *Instance) SetFontPath(cachePath string) error {
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
func (ins *Instance) People() *core.People {
	return ins.people
}

// AddPerson add person to people
func (ins *Instance) AddPerson(items ...*core.Person) {
	if ins.people == nil {
		ins.people = new(core.People)
	}
	ins.people.Append(items...)
}

// DeletePerson delete a person by name
func (ins *Instance) DeletePerson(name string) bool {
	if ins.people == nil {
		return false
	}
	return ins.people.Delete(name)
}

// Reload reload person model
func (ins *Instance) Reload() {
	if ins.people == nil {
		return
	}
	ins.People().Setup()
}

// Match match a person with embedding
func (ins *Instance) Match(embedding []float32) (*core.Person, float64, error) {
	return ins.People().Match(embedding)
}

// ExtractFace extract face for a person from image
func (ins *Instance) ExtractFace(person *core.Person, img image.Image, minSize int) (*core.FaceMarker, error) {
	face, err := ins.net.DetectSingle(img, minSize)
	if err != nil {
		return nil, err
	}
	person.Embeddings = append(person.Embeddings, &core.Person_Embedding{
		Value: face.Embeddings[0],
	})
	return core.NewFaceMarker(face, person.GetName(), 1), nil
}

// DetectFaces detect face markers from image
func (ins *Instance) DetectFaces(img image.Image, minSize int) (*core.FaceMarkers, error) {
	faces, err := ins.net.DetectMultiple(img, minSize)
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

// DrawMarkers draw face markers on image
func (ins *Instance) DrawMarkers(markers *core.FaceMarkers, txtColor string, successColor string, failedColor string, strokeWidth float64, succeedOnly bool) image.Image {
	return markers.Draw(ins.font, txtColor, successColor, failedColor, strokeWidth, succeedOnly)
}
