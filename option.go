package facenet

import (
	"github.com/bububa/facenet/core"
	"github.com/bububa/facenet/imageutil"
)

// Option face instance option interface
type Option interface {
	apply(*Estimator) error
}

type optionFunc func(ins *Estimator) error

func (fn optionFunc) apply(ins *Estimator) error {
	return fn(ins)
}

// WithModel set net model with model path
func WithModel(modelPath string) Option {
	return optionFunc(func(ins *Estimator) error {
		ins.model = core.NewNet(modelPath)
		return nil
	})
}

// WithDB set db with dbpath
func WithDB(dbPath string) Option {
	return optionFunc(func(ins *Estimator) error {
		if ins.db == nil {
			ins.db = NewStorage(nil, nil)
		}
		return ins.db.Load(dbPath)
	})
}

// WithFontPath set font with font path
func WithFontPath(fontPath string) Option {
	return optionFunc(func(ins *Estimator) error {
		if ins.font == nil {
			ins.font = new(imageutil.Font)
		}
		ins.font.Cache = imageutil.NewFontCache(fontPath)
		return nil
	})
}
