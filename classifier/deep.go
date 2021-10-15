package classifier

import (
	"encoding/json"
	"io"

	deep "github.com/patrikeh/go-deep"
	"github.com/patrikeh/go-deep/training"

	"github.com/bububa/facenet/core"
)

// Neural represents neural classifier
type Neural struct {
	ml        *deep.Neural
	threshold float64
}

// Name return sclassifier name
func (n *Neural) Identity() ClassifierIdentity {
	return NeuralClassifier
}

// Write implement Classifier interface
func (n *Neural) Write(w io.Writer) error {
	dump := n.ml.Dump()
	return json.NewEncoder(w).Encode(dump)
}

// Read implement Classifier interface
func (n *Neural) Read(r io.Reader) error {
	var dump deep.Dump
	if err := json.NewDecoder(r).Decode(&dump); err != nil {
		return err
	}
	n.ml = deep.FromDump(&dump)
	return nil
}

// SetThreadshold set Neural match threshold
func (n *Neural) SetThreadshold(threshold float64) {
	n.threshold = threshold
}

func (n *Neural) peopleToExamples(people *core.People, split float64) (training.Examples, training.Examples) {
	var data training.Examples
	var heldout training.Examples
	classes := len(people.GetList())
	for idx, person := range people.GetList() {
		var examples training.Examples
		embeddings := person.GetEmbeddings()
		for _, embedding := range embeddings {
			e := training.Example{
				Response: onehot(classes, idx),
				Input:    convInputs(embedding.GetValue()),
			}
			deep.Standardize(e.Input)
			examples = append(examples, e)
		}
		examples.Shuffle()
		t, h := examples.Split(split)
		data = append(data, t...)
		heldout = append(heldout, h...)
	}
	data.Shuffle()
	heldout.Shuffle()
	return data, heldout
}

func (n *Neural) initDeep(inputs int, layout []int, std float64, mean float64) {
	n.ml = deep.NewNeural(&deep.Config{
		Inputs: inputs,
		Layout: layout,
		// Activation: deep.ActivationTanh,
		// Activation: deep.ActivationSigmoid,
		Activation: deep.ActivationReLU,
		//Activation: deep.ActivationSoftmax,
		Mode:   deep.ModeMultiClass,
		Weight: deep.NewNormal(std, mean),
		Bias:   true,
	})
}

// Train implement Classifier interface
func (n *Neural) Train(people *core.People, split float64, iterations int, verbosity int) {
	n.initDeep(512, []int{64, 16, len(people.GetList())}, 0.5, 0)
	//trainer := training.NewTrainer(training.NewSGD(0.01, 0.5, 1e-6, true), 1)
	//trainer := training.NewTrainer(training.NewSGD(0.005, 0.5, 1e-6, true), 50)
	//trainer := training.NewBatchTrainer(training.NewSGD(0.005, 0.1, 0, true), 50, 300, 16)
	//trainer := training.NewTrainer(training.NewAdam(0.1, 0, 0, 0), 50)
	// solver := training.NewSGD(0.01, 0.5, 1e-6, true)
	solver := training.NewAdam(0.02, 0.9, 0.999, 1e-8)
	trainer := training.NewTrainer(solver, verbosity)
	data, heldout := n.peopleToExamples(people, split)
	trainer.Train(n.ml, data, heldout, iterations)
}

// BatchTrain implement Classifier interface
func (n *Neural) BatchTrain(people *core.People, split float64, iterations int, verbosity int, batch int) {
	n.initDeep(512, []int{64, 16, len(people.GetList())}, 0.5, 0)
	//solver := training.NewSGD(0.01, 0.5, 1e-6, true)
	solver := training.NewAdam(0.02, 0.9, 0.999, 1e-8)
	trainer := training.NewBatchTrainer(solver, verbosity, batch, 4)
	data, heldout := n.peopleToExamples(people, split)
	trainer.Train(n.ml, data, heldout, iterations)
}

// Predict implement Classifier interface
func (n *Neural) Predict(embedding []float32) []float64 {
	return n.ml.Predict(convInputs(embedding))
}

// Match returns best match result
func (n *Neural) Match(input []float32) (int, float64) {
	scores := n.Predict(input)
	var index = -1
	var maxScore float64
	threshold := n.threshold
	if threshold < 1e-15 {
		threshold = NeuralMatchThreshold
	}
	for idx, score := range scores {
		if score >= threshold && maxScore < score {
			maxScore = score
			index = idx
		}
	}
	return index, maxScore
}

func convInputs(embedding []float32) []float64 {
	ret := make([]float64, len(embedding))
	for i, v := range embedding {
		ret[i] = float64(v)
	}
	return ret
}

func onehot(classes int, val int) []float64 {
	res := make([]float64, classes)
	res[val] = 1
	return res
}
