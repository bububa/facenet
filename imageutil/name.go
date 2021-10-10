package imageutil

import "strings"

// Name represents a crop size name.
type Name string

// Jpeg returns the crop name with a jpeg file extension suffix as string.
func (n Name) Jpeg() string {
	var builder strings.Builder
	builder.WriteString(string(n))
	builder.WriteString(".jpg")
	return builder.String()
}

// Names of standard crop sizes.
const (
	Tile50  Name = "tile_50"
	Tile100 Name = "tile_100"
	Tile160 Name = "tile_160"
	Tile224 Name = "tile_224"
	Tile320 Name = "tile_320"
	Tile500 Name = "tile_500"
	Fit720  Name = "fit_720"
	Fit1280 Name = "fit_1280"
	Fit1920 Name = "fit_1920"
	Fit2048 Name = "fit_2048"
	Fit2560 Name = "fit_2560"
	Fit3840 Name = "fit_3840"
	Fit4096 Name = "fit_4096"
	Fit7680 Name = "fit_7680"
)
