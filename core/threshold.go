package core

import (
	"github.com/bububa/facenet/imageutil"
)

// CropSize default crop size for image
var CropSize = imageutil.Sizes[imageutil.Tile160]

// MatchDist default match distance
var MatchDist = 0.46

// ClusterDist default cluster distance
var ClusterDist = 0.64

// ClusterCore depreciated
var ClusterCore = 4

// ClusterMinScore depreciated
var ClusterMinScore = 15

// ClusterMinSize depreciated
var ClusterMinSize = 95

// SampleThreshold depreciated
var SampleThreshold = 2 * ClusterCore

// OverlapThreshold default face overlap threshold
var OverlapThreshold = 42

// OverlapThresholdFloor default face overlap threshold floor
var OverlapThresholdFloor = OverlapThreshold - 1

// ScoreThreshold default quality threshold score
var ScoreThreshold = 4.0

// MaxImageSize defines maxium image size for detection resize
var MaxImageSize = 640

// QualityThreshold returns the scale adjusted quality score threshold.
func QualityThreshold(scale int) (score float32) {
	score = float32(ScoreThreshold)

	// Smaller faces require higher quality.
	switch {
	case scale < 26:
		score += 26.0
	case scale < 32:
		score += 16.0
	case scale < 40:
		score += 11.0
	case scale < 50:
		score += 9.0
	case scale < 80:
		score += 6.0
	case scale < 110:
		score += 2.0
	}

	return score
}
