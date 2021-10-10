package imageutil

import (
	"image"

	"github.com/disintegration/imaging"
)

// Thumb returns a cropped area from an existing thumbnail image.
func Thumb(img image.Image, area Area, size Size) image.Image {
	// Get absolute crop coordinates and dimension.
	min, max, _ := area.Bounds(img)

	/*
		if dim < size.Width {
			log.Printf("crop: image is too small, upscaling %dpx to %dpx", dim, size.Width)
		}
	*/

	// Crop area from image.
	thumb := imaging.Crop(img, image.Rect(min.X, min.Y, max.X, max.Y))

	// Resample crop area.
	return Resample(thumb, size.Width, size.Height, size.Options...)
}

// NormalizeImage resize image to 640x640
func NormalizeImage(img image.Image, maxSize int) image.Image {
	w := img.Bounds().Max.X
	h := img.Bounds().Max.Y
	if w > h && w > maxSize {
		h = int(float64(maxSize*h) / float64(w))
		w = maxSize
	} else if w < h && h > maxSize {
		w = int(float64(maxSize*w) / float64(h))
		h = maxSize
	} else if w == h && w < maxSize {
		w = maxSize
		h = maxSize
	}
	return Resample(img, w, h, ResampleResize)
}
