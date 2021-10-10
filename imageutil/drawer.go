package imageutil

import (
	"image"

	"github.com/llgcode/draw2d/draw2dimg"
	"github.com/llgcode/draw2d/draw2dkit"
)

// ImageToRGBA conver image.Image to *image.RGBA
func ImageToRGBA(img image.Image) *image.RGBA {
	i := image.NewRGBA(img.Bounds())
	gc := draw2dimg.NewGraphicContext(i)
	gc.DrawImage(img)
	return i
}

// DrawRectangle draw rectangle on image
func DrawRectangle(img *image.RGBA, bounds image.Rectangle, borderColor string, bgColor string, strokeWidth float64) {
	gc := draw2dimg.NewGraphicContext(img)
	gc.SetStrokeColor(ColorFromHex(borderColor))
	if bgColor != "" {
		gc.SetFillColor(ColorFromHex(bgColor))
	}
	gc.SetLineWidth(strokeWidth)
	draw2dkit.Rectangle(gc, float64(bounds.Min.X), float64(bounds.Min.Y), float64(bounds.Max.X), float64(bounds.Max.Y))
	gc.Stroke()
	if bgColor != "" {
		gc.Fill()
	}
}

// DrawLabel draw label text to image
func DrawLabel(img *image.RGBA, font *Font, label string, pt image.Point, txtColor string, bgColor string, scale float64) {
	if font.Cache == nil || font.Data == nil {
		return
	}
	gc := draw2dimg.NewGraphicContext(img)
	// gc.SetStrokeColor(ColorFromHex(txtColor))
	gc.FontCache = font.Cache
	gc.SetFontData(*font.Data)
	gc.SetFontSize(font.Size * scale)
	var (
		x       = float64(pt.X)
		y       = float64(pt.Y)
		padding = 2.0 * scale
	)
	left, top, right, bottom := gc.GetStringBounds(label)
	height := bottom - top
	width := right - left
	if bgColor != "" {
		gc.SetFillColor(ColorFromHex(bgColor))
		draw2dkit.Rectangle(gc, x, y, x+width+padding*2, y+height+padding*2)
		gc.Fill()
	}
	gc.SetFillColor(ColorFromHex(txtColor))
	gc.FillStringAt(label, x-left+padding, y-top+padding)
}
