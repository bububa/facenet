package core

import (
	"encoding/json"

	"github.com/bububa/facenet/imageutil"
)

// Face represents a face detected.
type Face struct {
	Rows       int         `json:"rows,omitempty"`
	Cols       int         `json:"cols,omitempty"`
	Score      int         `json:"score,omitempty"`
	Area       Area        `json:"face,omitempty"`
	Eyes       Areas       `json:"eyes,omitempty"`
	Landmarks  Areas       `json:"landmarks,omitempty"`
	Embeddings [][]float32 `json:"embeddings,omitempty"`
}

// Size returns the absolute face size in pixels.
func (f *Face) Size() int {
	return f.Area.Scale
}

// Dim returns the max number of rows and cols as float32 to calculate relative coordinates.
func (f *Face) Dim() float32 {
	if f.Cols > 0 {
		return float32(f.Cols)
	}

	return float32(1)
}

// CropArea returns the relative image area for cropping.
func (f *Face) CropArea() imageutil.Area {
	if f.Rows < 1 {
		f.Cols = 1
	}

	if f.Cols < 1 {
		f.Cols = 1
	}

	x := float32(f.Area.Col-f.Area.Scale/2) / float32(f.Cols)
	y := float32(f.Area.Row-f.Area.Scale/2) / float32(f.Rows)

	return imageutil.NewArea(
		f.Area.Name,
		x,
		y,
		float32(f.Area.Scale)/float32(f.Cols),
		float32(f.Area.Scale)/float32(f.Rows),
	)
}

// EyesMidpoint returns the point in between the eyes.
func (f *Face) EyesMidpoint() Area {
	if len(f.Eyes) != 2 {
		return Area{
			Name:  "midpoint",
			Row:   f.Area.Row,
			Col:   f.Area.Col,
			Scale: f.Area.Scale,
		}
	}

	return Area{
		Name:  "midpoint",
		Row:   (f.Eyes[0].Row + f.Eyes[1].Row) / 2,
		Col:   (f.Eyes[0].Col + f.Eyes[1].Col) / 2,
		Scale: (f.Eyes[0].Scale + f.Eyes[1].Scale) / 2,
	}
}

// RelativeLandmarks returns relative face areas.
func (f *Face) RelativeLandmarks() imageutil.Areas {
	p := f.EyesMidpoint()

	m := f.Landmarks.Relative(p, float32(f.Rows), float32(f.Cols))
	m = append(m, f.Eyes.Relative(p, float32(f.Rows), float32(f.Cols))...)

	return m
}

// RelativeLandmarksJSON returns relative face areas as JSON.
func (f *Face) RelativeLandmarksJSON() (b []byte) {
	var noResult = []byte("")

	l := f.RelativeLandmarks()

	if len(l) < 1 {
		return noResult
	}

	result, err := json.Marshal(l)
	if err != nil {
		return noResult
	}
	return result
}

// EmbeddingsJSON returns detected face embeddings as JSON array.
func (f *Face) EmbeddingsJSON() (b []byte) {
	var noResult = []byte("")

	result, err := json.Marshal(f.Embeddings)
	if err != nil {
		return noResult
	}
	return result
}
