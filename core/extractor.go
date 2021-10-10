package core

import (
	// use embed
	_ "embed"

	// use jpeg image
	_ "image/jpeg"

	"fmt"
	"image"
	"sort"
	"sync"

	pigo "github.com/esimov/pigo/core"
)

//go:embed cascade/facefinder
var cascadeFile []byte

//go:embed cascade/puploc
var puplocFile []byte

var (
	classifier *pigo.Pigo
	plc        *pigo.PuplocCascade
	flpcs      map[string][]*FlpCascade
)

func init() {
	var err error

	p := pigo.NewPigo()
	// Unpack the binary file. This will return the number of cascade trees,
	// the tree depth, the threshold and the prediction from tree's leaf nodes.
	classifier, err = p.Unpack(cascadeFile)

	if err != nil {
		panic(err)
	}

	pl := pigo.NewPuplocCascade()
	plc, err = pl.UnpackCascade(puplocFile)

	if err != nil {
		panic(err)
	}

	flpcs, err = ReadCascadeDir(pl, "cascade/lps")

	if err != nil {
		panic(err)
	}
}

var (
	eyeCascades   = []string{"lp46", "lp44", "lp42", "lp38", "lp312"}
	mouthCascades = []string{"lp93", "lp84", "lp82", "lp81"}
)

// Extractor struct contains Pigo face detector general settings.
type Extractor struct {
	minSize            int
	angle              float64
	shiftFactor        float64
	scaleFactor        float64
	iouThreshold       float64
	scoreThreshold     float32
	perturb            int
	landmarkCoordsPool *sync.Pool
}

// TrySizeExtractMultiple extract multiple faces with different minSize
func TrySizeExtractMultiple(img image.Image, findLandmarks bool, minSize int) (faces Faces, err error) {
	maxSize := MaxImageSize
	w := img.Bounds().Max.X
	h := img.Bounds().Max.Y
	if w > h {
		maxSize = MaxImageSize * h / w
	} else if w < h {
		maxSize = MaxImageSize * w / h
	}
	scales := maxSize / minSize
	for i := 1; i <= scales; i++ {
		size := minSize * i
		faces, err = Extract(img, findLandmarks, size)
		if err != nil || len(faces) == 0 {
			continue
		}
		break
	}
	if len(faces) == 0 {
		return nil, NewError(NoFaceErr, "no face detected")
	}
	return faces, err
}

// TrySizeExtractSingle extract single face with different minSize
func TrySizeExtractSingle(img image.Image, findLandmarks bool, minSize int) (face Face, err error) {
	maxSize := MaxImageSize
	w := img.Bounds().Max.X
	h := img.Bounds().Max.Y
	if w > h {
		maxSize = MaxImageSize * h / w
	} else if w < h {
		maxSize = MaxImageSize * w / h
	}
	scales := maxSize / minSize
	var list Faces
	for i := 1; i <= scales; i++ {
		size := minSize * i
		faces, err := Extract(img, findLandmarks, size)
		if err != nil || len(faces) != 1 {
			continue
		}
		list = append(list, faces[0])
	}
	if len(list) == 0 {
		return face, NewError(NoFaceErr, "no face detected")
	}
	sort.Slice(list, func(i, j int) bool {
		return list[i].Score > list[j].Score
	})
	face = list[0]
	if face.Area.Col == 0 || face.Area.Row == 0 {
		return face, NewError(NoFaceErr, "no face detected")
	}
	return list[0], nil
}

// Extract runs the detection algorithm over the provided source image.
func Extract(img image.Image, findLandmarks bool, minSize int) (faces Faces, err error) {

	if minSize < 20 {
		minSize = 20
	}

	extractor := &Extractor{
		minSize:        minSize,
		angle:          0.0,
		shiftFactor:    0.1,
		scaleFactor:    1.1,
		iouThreshold:   0.2,
		scoreThreshold: float32(ScoreThreshold),
		perturb:        63,
		landmarkCoordsPool: &sync.Pool{
			New: func() interface{} {
				return make([]Area, 0, len(flpcs))
			},
		},
	}

	det, params, err := extractor.Extract(img)

	if err != nil {
		return faces, err
	}

	if det == nil {
		return faces, NewError(NoFaceErr, "no face detected")
	}

	return extractor.Faces(det, params, findLandmarks)
}

// Extract runs the detection algorithm over the provided source image.
func (d *Extractor) Extract(img image.Image) (faces []pigo.Detection, params pigo.CascadeParams, err error) {
	src := pigo.ImgToNRGBA(img)

	pixels := pigo.RgbToGrayscale(src)
	cols, rows := src.Bounds().Max.X, src.Bounds().Max.Y

	var maxSize int

	if cols < 20 || rows < 20 || cols < d.minSize || rows < d.minSize {
		err = NewError(ExtractImageSizeTooSmallErr, fmt.Sprintf("image size %dx%d is too small", cols, rows))
		return faces, params, err
	} else if cols < rows {
		maxSize = cols - 4
	} else {
		maxSize = rows - 4
	}

	imageParams := &pigo.ImageParams{
		Pixels: pixels,
		Rows:   rows,
		Cols:   cols,
		Dim:    cols,
	}

	params = pigo.CascadeParams{
		MinSize:     d.minSize,
		MaxSize:     maxSize,
		ShiftFactor: d.shiftFactor,
		ScaleFactor: d.scaleFactor,
		ImageParams: *imageParams,
	}

	//log.Printf("faces: image size %dx%d, face size min %d, max %d\n", cols, rows, params.MinSize, params.MaxSize)

	// Run the classifier over the obtained leaf nodes and return the Face results.
	// The result contains quadruplets representing the row, column, scale and Face score.
	faces = classifier.RunCascade(params, d.angle)

	// Calculate the intersection over union (IoU) of two clusters.
	faces = classifier.ClusterDetections(faces, d.iouThreshold)

	return faces, params, nil
}

// Faces adds landmark coordinates to detected faces and returns the results.
func (d *Extractor) Faces(det []pigo.Detection, params pigo.CascadeParams, findLandmarks bool) (results Faces, err error) {
	// Sort results by size.
	sort.Slice(det, func(i, j int) bool {
		return det[i].Scale > det[j].Scale
	})

	results = NewFaces(len(det))
	puplocPool := &sync.Pool{
		New: func() interface{} {
			return new(pigo.Puploc)
		},
	}
	eyesCoords := make([]Area, 0, 2)
	facePool := &sync.Pool{
		New: func() interface{} {
			return new(Face)
		},
	}
	for _, face := range det {
		// log.Printf("Q: %f, scale:%d, threshold:%f\n", face.Q, face.Scale, QualityThreshold(face.Scale))
		// Skip result if quality is too low.
		if face.Q < QualityThreshold(face.Scale) {
			continue
		}

		eyesCoords = eyesCoords[:0]
		landmarkCoords := d.landmarkCoordsPool.Get().([]Area)
		landmarkCoords = landmarkCoords[:0]
		puploc := puplocPool.Get().(*pigo.Puploc)

		faceCoord := NewArea(
			"face",
			face.Row,
			face.Col,
			face.Scale,
		)

		// Detect additional face landmarks?
		if face.Scale > 50 && findLandmarks {
			// Find left eye.
			puploc.Row = face.Row - int(0.075*float32(face.Scale))
			puploc.Col = face.Col - int(0.175*float32(face.Scale))
			puploc.Scale = float32(face.Scale) * 0.25
			puploc.Perturbs = d.perturb

			leftEye := plc.RunDetector(*puploc, params.ImageParams, d.angle, false)

			if leftEye.Row > 0 && leftEye.Col > 0 {
				eyesCoords = append(eyesCoords, NewArea(
					"eye_l",
					leftEye.Row,
					leftEye.Col,
					int(leftEye.Scale),
				))
			}

			// Find right eye.
			puploc.Row = face.Row - int(0.075*float32(face.Scale))
			puploc.Col = face.Col + int(0.185*float32(face.Scale))
			puploc.Scale = float32(face.Scale) * 0.25
			puploc.Perturbs = d.perturb

			rightEye := plc.RunDetector(*puploc, params.ImageParams, d.angle, false)

			if rightEye.Row > 0 && rightEye.Col > 0 {
				eyesCoords = append(eyesCoords, NewArea(
					"eye_r",
					rightEye.Row,
					rightEye.Col,
					int(rightEye.Scale),
				))
			}

			if leftEye != nil && rightEye != nil {
				for _, eye := range eyeCascades {
					for _, flpc := range flpcs[eye] {
						if flpc == nil {
							continue
						}

						flp := flpc.GetLandmarkPoint(leftEye, rightEye, params.ImageParams, d.perturb, false)
						if flp.Row > 0 && flp.Col > 0 {
							landmarkCoords = append(landmarkCoords, NewArea(
								eye,
								flp.Row,
								flp.Col,
								int(flp.Scale),
							))
						}

						flp = flpc.GetLandmarkPoint(leftEye, rightEye, params.ImageParams, d.perturb, true)
						if flp.Row > 0 && flp.Col > 0 {
							landmarkCoords = append(landmarkCoords, NewArea(
								eye+"_v",
								flp.Row,
								flp.Col,
								int(flp.Scale),
							))
						}
					}
				}
			}

			// Find mouth.
			for _, mouth := range mouthCascades {
				for _, flpc := range flpcs[mouth] {
					if flpc == nil {
						continue
					}

					flp := flpc.GetLandmarkPoint(leftEye, rightEye, params.ImageParams, d.perturb, false)
					if flp.Row > 0 && flp.Col > 0 {
						landmarkCoords = append(landmarkCoords, NewArea(
							"mouth_"+mouth,
							flp.Row,
							flp.Col,
							int(flp.Scale),
						))
					}
				}
			}

			flpc := flpcs["lp84"][0]

			if flpc != nil {
				flp := flpc.GetLandmarkPoint(leftEye, rightEye, params.ImageParams, d.perturb, true)
				if flp.Row > 0 && flp.Col > 0 {
					landmarkCoords = append(landmarkCoords, NewArea(
						"lp84",
						flp.Row,
						flp.Col,
						int(flp.Scale),
					))
				}
			}
		}
		puplocPool.Put(puploc)

		// Create face.
		fCache := facePool.Get().(*Face)
		f := *fCache
		f.Rows = params.ImageParams.Rows
		f.Cols = params.ImageParams.Cols
		f.Score = int(face.Q)
		f.Area = faceCoord
		f.Eyes = eyesCoords
		f.Landmarks = landmarkCoords
		facePool.Put(fCache)

		d.landmarkCoordsPool.Put(landmarkCoords)

		// Does the face significantly overlap with previous results?
		if results.Contains(f) {
			// Ignore face.
		} else {
			// Append face.
			results.Append(f)
		}
	}

	return results, nil
}
