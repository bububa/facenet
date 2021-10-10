package core

import (
	"image"
	"math"
	"path"
	"sync"

	"github.com/bububa/facenet/imageutil"

	tf "github.com/tensorflow/tensorflow/tensorflow/go"
	"github.com/tensorflow/tensorflow/tensorflow/go/op"
)

// Net is a wrapper for the TensorFlow Facenet model.
type Net struct {
	model     *tf.SavedModel
	modelPath string
	modelName string
	modelTags []string
	mutex     sync.Mutex
}

// NewNet returns a new TensorFlow Facenet instance.
func NewNet(modelPath string) *Net {
	return &Net{modelPath: modelPath, modelTags: []string{"serve"}}
}

// DetectMultiple detect multiple faces try to use different minSize
func (t *Net) DetectMultiple(img image.Image, minSize int) (faces Faces, err error) {
	src := imageutil.NormalizeImage(img, MaxImageSize)
	faces, err = TrySizeExtractMultiple(src, false, minSize)

	if err != nil {
		return faces, err
	}

	if err = t.loadModel(); err != nil {
		return faces, err
	}
	for i, face := range faces {
		thumb := imageutil.Thumb(src, face.CropArea(), CropSize)
		if embeddings, err := t.getEmbeddings(thumb); err == nil {
			faces[i].Embeddings = embeddings
		}
	}
	return faces, nil
}

// DetectSingle detect single face try to use different minSize
func (t *Net) DetectSingle(img image.Image, minSize int) (face Face, err error) {
	src := imageutil.NormalizeImage(img, MaxImageSize)
	face, err = TrySizeExtractSingle(src, false, minSize)

	if err != nil {
		return face, err
	}

	if err = t.loadModel(); err != nil {
		return face, err
	}
	thumb := imageutil.Thumb(src, face.CropArea(), CropSize)
	if embeddings, err := t.getEmbeddings(thumb); err == nil {
		face.Embeddings = embeddings
	}
	return face, nil
}

// Detect runs the detection and facenet algorithms over the provided source image.
func (t *Net) Detect(img image.Image, minSize int, expected int) (faces Faces, err error) {
	src := imageutil.NormalizeImage(img, MaxImageSize)
	faces, err = Extract(src, false, minSize)

	if err != nil {
		return faces, err
	}

	if c := len(faces); c == 0 || expected > 0 && c == expected {
		return faces, nil
	}

	if err = t.loadModel(); err != nil {
		return faces, err
	}

	for i, f := range faces {
		if f.Area.Col == 0 || f.Area.Row == 0 {
			continue
		}

		thumb := imageutil.Thumb(src, f.CropArea(), CropSize)
		if embeddings, err := t.getEmbeddings(thumb); err == nil {
			faces[i].Embeddings = embeddings
		}
	}

	return faces, nil
}

// Train train images with label defined
func (t *Net) Train(label string, images []image.Image, minSize int) (person Person, err error) {
	person.Name = label
	person.Embeddings = make([]*Person_Embedding, 0, len(images))
	for _, img := range images {
		face, err := t.DetectSingle(img, minSize)
		if err != nil {
			return person, err
		}
		person.Embeddings = append(person.Embeddings, &Person_Embedding{
			Value: face.Embeddings[0],
		})
	}
	return
}

// ModelLoaded tests if the TensorFlow model is loaded.
func (t *Net) ModelLoaded() bool {
	return t.model != nil
}

func (t *Net) loadModel() error {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	if t.ModelLoaded() {
		return nil
	}

	modelPath := path.Join(t.modelPath)

	// log.Printf("faces: loading %s\n", filepath.Base(modelPath))

	// Load model
	model, err := tf.LoadSavedModel(modelPath, t.modelTags, nil)

	if err != nil {
		return err
	}

	t.model = model

	return nil
}

func (t *Net) getEmbeddings(img image.Image) ([][]float32, error) {
	tensor, err := imageToTensor(img, CropSize.Width, CropSize.Height, true)

	if err != nil {
		// log.Printf("faces: failed to convert image to tensor: %v\n", err)
		return nil, err
	}

	trainPhaseBoolTensor, err := tf.NewTensor(false)
	if err != nil {
		return nil, err
	}

	output, err := t.model.Session.Run(
		map[tf.Output]*tf.Tensor{
			t.model.Graph.Operation("input").Output(0):       tensor,
			t.model.Graph.Operation("phase_train").Output(0): trainPhaseBoolTensor,
		},
		[]tf.Output{
			t.model.Graph.Operation("embeddings").Output(0),
		},
		nil)

	if err != nil {
		// log.Printf("faces: %s\n", err)
		return nil, err
	}

	if len(output) < 1 {
		return nil, NewError(InferenceFailedErr, "inference failed, no output")
	}
	return output[0].Value().([][]float32), nil
}

func imageToTensor(img image.Image, imageHeight, imageWidth int, preWhiten bool) (tfTensor *tf.Tensor, err error) {
	if imageHeight <= 0 || imageWidth <= 0 {
		return tfTensor, NewError(ImageToTensorSizeErr, "image width and height must be > 0")
	}

	var tfImage [1][][][3]float32

	for j := 0; j < imageHeight; j++ {
		tfImage[0] = append(tfImage[0], make([][3]float32, imageWidth))
	}

	for i := 0; i < imageWidth; i++ {
		for j := 0; j < imageHeight; j++ {
			r, g, b, _ := img.At(i, j).RGBA()
			tfImage[0][j][i][0] = convertValue(r)
			tfImage[0][j][i][1] = convertValue(g)
			tfImage[0][j][i][2] = convertValue(b)
		}
	}
	if !preWhiten {
		return tf.NewTensor(tfImage)
	}
	// pre whiten image
	mean, std := meanStd(tfImage[0])
	tensor, err := tf.NewTensor(tfImage)
	if err != nil {
		return nil, err
	}
	return preWhitenImage(tensor, mean, std)
}

func preWhitenImage(img *tf.Tensor, mean, std float32) (*tf.Tensor, error) {
	s := op.NewScope()
	pimg := op.Placeholder(s, tf.Float, op.PlaceholderShape(tf.MakeShape(1, -1, -1, 3)))

	out := op.Mul(s, op.Sub(s, pimg, op.Const(s.SubScope("mean"), mean)),
		op.Const(s.SubScope("scale"), float32(1.0)/std))
	outs, err := runScope(s, map[tf.Output]*tf.Tensor{pimg: img}, []tf.Output{out})
	if err != nil {
		return nil, err
	}

	return outs[0], nil
}

func runScope(s *op.Scope, inputs map[tf.Output]*tf.Tensor, outputs []tf.Output) ([]*tf.Tensor, error) {
	graph, err := s.Finalize()
	if err != nil {
		return nil, err
	}

	session, err := tf.NewSession(graph, nil)
	if err != nil {
		return nil, err
	}
	defer session.Close()
	return session.Run(inputs, outputs, nil)
}

func convertValue(value uint32) float32 {
	return (float32(value>>8) - float32(127.5)) / float32(127.5)
}

func meanStd(img [][][3]float32) (mean float32, std float32) {
	count := len(img) * len(img[0]) * len(img[0][0])
	for _, x := range img {
		for _, y := range x {
			for _, z := range y {
				mean += z
			}
		}
	}
	mean /= float32(count)

	for _, x := range img {
		for _, y := range x {
			for _, z := range y {
				std += (z - mean) * (z - mean)
			}
		}
	}

	xstd := math.Sqrt(float64(std) / float64(count-1))
	minstd := 1.0 / math.Sqrt(float64(count))
	if xstd < minstd {
		xstd = minstd
	}

	std = float32(xstd)
	return
}
