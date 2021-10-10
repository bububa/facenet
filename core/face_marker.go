package core

import (
	"image"

	"github.com/bububa/facenet/imageutil"
)

// FaceMarker detected face
type FaceMarker struct {
	face     Face
	label    string
	distance float64
	err      error
}

// NewFaceMarker init a FaceMarker
func NewFaceMarker(face Face, label string, distance float64) *FaceMarker {
	return &FaceMarker{
		face:     face,
		label:    label,
		distance: distance,
	}
}

// Label get marker label
func (f FaceMarker) Label() string {
	return f.label
}

// Distance get marker distance
func (f FaceMarker) Distance() float64 {
	return f.distance
}

// Face get marker face
func (f FaceMarker) Face() Face {
	return f.face
}

// SetError set face maker match failed
func (f *FaceMarker) SetError(err error) {
	f.err = err
}

// Error identify the successful of matching
func (f FaceMarker) Error() error {
	return f.err
}

// Bounds FackeMarker image bounds
func (f FaceMarker) Bounds(img image.Image) image.Rectangle {
	cropArea := f.face.CropArea()
	min, max, _ := cropArea.Bounds(img)
	return image.Rect(min.X, min.Y, max.X, max.Y)
}

// Thumb generate thumb image of a face marker
func (f FaceMarker) Thumb(img image.Image) image.Image {
	return imageutil.Thumb(img, f.face.CropArea(), CropSize)
}
