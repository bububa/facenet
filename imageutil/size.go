package imageutil

// Size .
type Size struct {
	// Name size name
	Name Name `json:"name"`
	// Source size source
	Source Name `json:"-"`
	// Use size usage
	Use string `json:"use"`
	// Width size width
	Width int `json:"w"`
	// Height size height
	Height int `json:"h"`
	// Options resample options
	Options []ResampleOption `json:"-"`
}

// SizeMap size map
type SizeMap map[Name]Size

// Sizes contains the properties of all thumbnail sizes.
var Sizes = SizeMap{
	Tile50:  {Tile50, Tile320, "Lists", 50, 50, DefaultResampleOptions},
	Tile100: {Tile100, Tile320, "Maps", 100, 100, DefaultResampleOptions},
	Tile160: {Tile160, Tile320, "FaceNet", 160, 160, DefaultResampleOptions},
	Tile224: {Tile224, Tile320, "TensorFlow, Mosaic", 224, 224, DefaultResampleOptions},
	Tile320: {Tile320, "", "UI", 320, 320, DefaultResampleOptions},
	Tile500: {Tile500, "", "FaceNet", 500, 500, DefaultResampleOptions},
	Fit720:  {Fit720, "", "Mobile, TV", 720, 720, DefaultResampleFitOptions},
	Fit1280: {Fit1280, Fit2048, "Mobile, HD Ready TV", 1280, 1024, DefaultResampleFitOptions},
	Fit1920: {Fit1920, Fit2048, "Mobile, Full HD TV", 1920, 1200, DefaultResampleFitOptions},
	Fit2048: {Fit2048, "", "Tablets, Cinema 2K", 2048, 2048, DefaultResampleFitOptions},
	Fit2560: {Fit2560, "", "Quad HD, Retina Display", 2560, 1600, DefaultResampleFitOptions},
	Fit3840: {Fit3840, "", "Ultra HD", 3840, 2400, DefaultResampleFitOptions}, // Deprecated in favor of fit_4096
	Fit4096: {Fit4096, "", "Ultra HD, Retina 4K", 4096, 4096, DefaultResampleFitOptions},
	Fit7680: {Fit7680, "", "8K Ultra HD 2, Retina 6K", 7680, 4320, DefaultResampleFitOptions},
}
