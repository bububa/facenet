package core

import (
	"image"

	"github.com/bububa/facenet/imageutil"
)

// FaceMarkers image detected with markers
type FaceMarkers struct {
	img     image.Image
	markers []FaceMarker
}

// NewFaceMarkers init face markers
func NewFaceMarkers(img image.Image) *FaceMarkers {
	return &FaceMarkers{
		img: img,
	}
}

// Markers get markers
func (fm *FaceMarkers) Markers() []FaceMarker {
	return fm.markers
}

// Append append FaceMarker to FaceMarkers
func (fm *FaceMarkers) Append(m FaceMarker) {
	fm.markers = append(fm.markers, m)
}

// Draw draw face markers on image
func (fm FaceMarkers) Draw(font *imageutil.Font, txtColor string, successColor string, failedColor string, strokeWidth float64, succeedOnly bool) image.Image {
	img := fm.img
	if img == nil {
		return nil
	}
	var scales float64 = 1
	w := img.Bounds().Max.X
	h := img.Bounds().Max.Y
	if w > h {
		scales = float64(w) / float64(MaxImageSize)
	} else if w < h {
		scales = float64(h) / float64(MaxImageSize)
	}
	i := imageutil.ImageToRGBA(img)
	for _, m := range fm.markers {
		area := m.Bounds(i)
		color := successColor
		if m.Error() != nil {
			color = failedColor
			if succeedOnly {
				continue
			}
		}
		imageutil.DrawRectangle(i, area, color, "", strokeWidth)
		if m.label != "" && font != nil {
			imageutil.DrawLabel(i, font, m.label, image.Pt(area.Min.X, area.Max.Y), txtColor, color, scales)
		}
	}
	return i
}

// FaceImages get face images from face markers
func (fm *FaceMarkers) FaceImages(img image.Image) []image.Image {
	imgs := make([]image.Image, 0, len(fm.markers))
	for _, m := range fm.markers {
		imgs = append(imgs, m.Thumb(img))
	}
	return imgs
}
