package imageutil

import (
	"image"

	"github.com/disintegration/imaging"
)

// Resample downscales an image and returns it.
func Resample(img image.Image, width, height int, opts ...ResampleOption) image.Image {
	var resImg image.Image

	method, filter := ResampleOptions(opts...)

	if method == ResampleFit {
		resImg = imaging.Fit(img, width, height, filter)
	} else if method == ResampleFillCenter {
		resImg = imaging.Fill(img, width, height, imaging.Center, filter)
	} else if method == ResampleFillTopLeft {
		resImg = imaging.Fill(img, width, height, imaging.TopLeft, filter)
	} else if method == ResampleFillBottomRight {
		resImg = imaging.Fill(img, width, height, imaging.BottomRight, filter)
	} else if method == ResampleResize {
		resImg = imaging.Resize(img, width, height, filter)
	}

	return resImg
}
