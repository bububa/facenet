package facenet

import (
	"os"

	"github.com/bububa/facenet/core"
	"github.com/bububa/facenet/imageutil"
)

// Option face instance option interface
type Option interface {
	apply(*Instance) error
}

type optionFunc func(ins *Instance) error

func (fn optionFunc) apply(ins *Instance) error {
	return fn(ins)
}

// WithNet set net model with model path
func WithNet(modelPath string) Option {
	return optionFunc(func(ins *Instance) error {
		ins.net = core.NewNet(modelPath)
		return nil
	})
}

// WithPeople set people model with model path
func WithPeople(modelPath string) Option {
	return optionFunc(func(ins *Instance) error {
		if ins.people == nil {
			ins.people = new(core.People)
		}
		fn, err := os.Open(modelPath)
		if err != nil {
			return err
		}
		defer fn.Close()
		return core.LoadPeople(fn, ins.people)
	})
}

// WithFontPath set font with font path
func WithFontPath(fontPath string) Option {
	return optionFunc(func(ins *Instance) error {
		if ins.font == nil {
			ins.font = new(imageutil.Font)
		}
		ins.font.Cache = imageutil.NewFontCache(fontPath)
		return nil
	})
}
