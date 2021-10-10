package imageutil

import (
	"github.com/disintegration/imaging"
)

// ResampleOption represents resample option
type ResampleOption int

const (
	// ResampleFillCenter .
	ResampleFillCenter ResampleOption = iota
	// ResampleFillTopLeft .
	ResampleFillTopLeft
	// ResampleFillBottomRight .
	ResampleFillBottomRight
	// ResampleFit .
	ResampleFit
	// ResampleResize .
	ResampleResize
	// ResampleNearestNeighbor .
	ResampleNearestNeighbor
	// ResampleDefault .
	ResampleDefault
	// ResamplePng .
	ResamplePng
)

var (
	// DefaultResampleOptions represents default resample options
	DefaultResampleOptions = []ResampleOption{ResampleFillCenter, ResampleDefault}
	// DefaultResampleFitOptions represents default resample fit options
	DefaultResampleFitOptions = []ResampleOption{ResampleFit, ResampleDefault}
	// Filter represents default resample filter
	Filter = ResampleLanczos
)

// ResampleOptions extracts filter, format, and method from resample options.
func ResampleOptions(opts ...ResampleOption) (method ResampleOption, filter imaging.ResampleFilter) {
	method = ResampleFit
	filter = imaging.Lanczos

	for _, option := range opts {
		switch option {
		case ResampleNearestNeighbor:
			filter = imaging.NearestNeighbor
		case ResampleDefault:
			filter = Filter.Imaging()
		case ResampleFillTopLeft:
			method = ResampleFillTopLeft
		case ResampleFillCenter:
			method = ResampleFillCenter
		case ResampleFillBottomRight:
			method = ResampleFillBottomRight
		case ResampleFit:
			method = ResampleFit
		case ResampleResize:
			method = ResampleResize
		}
	}

	return method, filter
}

// ResampleFilter represents resample filter
type ResampleFilter string

const (
	// ResampleBlackman .
	ResampleBlackman ResampleFilter = "blackman"
	// ResampleLanczos .
	ResampleLanczos ResampleFilter = "lanczos"
	// ResampleCubic .
	ResampleCubic ResampleFilter = "cubic"
	// ResampleLinear .
	ResampleLinear ResampleFilter = "linear"
)

// Imaging returns resample filter
func (a ResampleFilter) Imaging() imaging.ResampleFilter {
	switch a {
	case ResampleBlackman:
		return imaging.Blackman
	case ResampleLanczos:
		return imaging.Lanczos
	case ResampleCubic:
		return imaging.CatmullRom
	case ResampleLinear:
		return imaging.Linear
	default:
		return imaging.Lanczos
	}
}
