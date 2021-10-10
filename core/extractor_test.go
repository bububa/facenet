package core

import (
	"image"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExtract(t *testing.T) {
	expected := map[string]int{
		"1.jpg":  2,
		"2.jpg":  1,
		"3.jpg":  1,
		"4.jpg":  1,
		"5.jpg":  1,
		"6.jpg":  1,
		"7.jpg":  0,
		"8.jpg":  0,
		"9.jpg":  0,
		"10.jpg": 0,
		"11.jpg": 0,
		"12.jpg": 1,
		"13.jpg": 0,
		"14.jpg": 0,
		"15.jpg": 0,
		"16.jpg": 1,
		"17.jpg": 2,
		"18.jpg": 2,
		"19.jpg": 0,
	}

	if err := filepath.Walk("../testdata", func(fileName string, info fs.FileInfo, err error) error {
		if info.IsDir() || filepath.Base(filepath.Dir(fileName)) != "testdata" {
			return nil
		}

		t.Run(fileName, func(t *testing.T) {
			baseName := filepath.Base(fileName)

			fn, err := os.Open(fileName)
			if err != nil {
				t.Fatal(err)
			}
			defer fn.Close()
			img, _, err := image.Decode(fn)

			faces, err := Extract(img, true, 20)

			if err != nil {
				t.Fatal(err)
			}

			t.Logf("found %d faces in '%s'", len(faces), baseName)

			if len(faces) > 0 {
				// t.Logf("results: %#v", faces)

				for i, f := range faces {
					t.Logf("marker[%d]: %#v %#v", i, f.CropArea(), f.Area)
					t.Logf("landmarks[%d]: %s", i, f.RelativeLandmarksJSON())
				}
			}

			if i, ok := expected[baseName]; ok {
				assert.Equal(t, i, faces.Count())

				if faces.Count() == 0 {
					assert.Equal(t, 100, faces.Uncertainty())
				} else {
					assert.Truef(t, faces.Uncertainty() >= 0 && faces.Uncertainty() <= 50, "uncertainty should be between 0 and 50")
				}
				t.Logf("uncertainty: %d", faces.Uncertainty())
			} else {
				t.Logf("unknown test result for %s", baseName)
			}
		})

		return nil
	}); err != nil {
		t.Fatal(err)
	}
}

func TestExtractOverlap(t *testing.T) {
	expected := map[string]int{
		"1.jpg": 2,
		"2.jpg": 2,
		"3.jpg": 2,
		"4.jpg": 1,
	}

	if err := filepath.Walk("../testdata/overlap", func(fileName string, info fs.FileInfo, err error) error {
		if info.IsDir() || filepath.Base(filepath.Dir(fileName)) != "overlap" {
			return nil
		}

		t.Run(fileName, func(t *testing.T) {
			baseName := filepath.Base(fileName)

			fn, err := os.Open(fileName)
			if err != nil {
				t.Fatal(err)
			}
			defer fn.Close()
			img, _, err := image.Decode(fn)

			faces, err := Extract(img, true, 20)

			if err != nil {
				t.Fatal(err)
			}

			t.Logf("found %d faces in '%s'", len(faces), baseName)

			if len(faces) > 0 {
				// t.Logf("results: %#v", faces)

				for i, f := range faces {
					t.Logf("marker[%d]: %#v %#v", i, f.CropArea(), f.Area)
					t.Logf("landmarks[%d]: %s", i, f.RelativeLandmarksJSON())
				}
			}

			if i, ok := expected[baseName]; ok {
				assert.Equal(t, i, faces.Count())

				if faces.Count() == 0 {
					assert.Equal(t, 100, faces.Uncertainty())
				} else {
					assert.Truef(t, faces.Uncertainty() >= 0 && faces.Uncertainty() <= 50, "uncertainty should be between 0 and 50")
				}
				t.Logf("uncertainty: %d", faces.Uncertainty())
			} else {
				t.Logf("unknown test result for %s", baseName)
			}
		})

		return nil
	}); err != nil {
		t.Fatal(err)
	}
}
