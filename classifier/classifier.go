package classifier

import (
	"io"

	"github.com/bububa/facenet/core"
)

// Classifier represents classifier interface
type Classifier interface {
	Identity() ClassifierIdentity
	Train(people *core.People, split float64, iterations int, verbosity int)
	BatchTrain(people *core.People, split float64, iterations int, verbosity int, batch int)
	Predict(input []float32) []float64
	Match(input []float32) (int, float64)
	Write(io.Writer) error
	Read(io.Reader) error
}

// ClassifierIdentity represents classifier type
type ClassifierIdentity int

const (
	// UnknownClassifier represents unknown classifier which is not defined
	UnknownClassifier ClassifierIdentity = iota
	// NeuralClassifier represents neural deep learning classifier
	NeuralClassifier
	// BayesClassifier represents bayes classifier
	BayesClassifier
)

// NewDefault returns a default Neural classifier
func NewDefault() Classifier {
	return new(Neural)
}

const (
	// NeuralMatchThreshold returns neural classifier match threshold
	NeuralMatchThreshold float64 = 0.75
)
