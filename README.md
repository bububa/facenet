# Golang lib for detect/recognize by tensorflow facenet

[![Go Reference](https://pkg.go.dev/badge/github.com/bububa/facenet.svg)](https://pkg.go.dev/github.com/bububa/facenet)
[![Go](https://github.com/bububa/facenet/actions/workflows/go.yml/badge.svg)](https://github.com/bububa/facenet/actions/workflows/go.yml)
[![goreleaser](https://github.com/bububa/facenet/actions/workflows/goreleaser.yml/badge.svg)](https://github.com/bububa/facenet/actions/workflows/goreleaser.yml)
[![GitHub go.mod Go version of a Go module](https://img.shields.io/github/go-mod/go-version/bububa/facenet.svg)](https://github.com/bububa/facenet)
[![GoReportCard](https://goreportcard.com/badge/github.com/bububa/facenet)](https://goreportcard.com/report/github.com/bububa/facenet)
[![GitHub license](https://img.shields.io/github/license/bububa/facenet.svg)](https://github.com/bububa/facenet/blob/master/LICENSE)
[![GitHub release](https://img.shields.io/github/release/bububa/facenet.svg)](https://GitHub.com/bububa/facenet/releases/)

## Prerequest

1. libtensorfow 1.x
   Follow the instruction [Install TensorFlow for C](https://www.tensorflow.org/install/lang_c#macos)
2. facenet tenorflow saved_model [Google Drive](https://drive.google.com/drive/folders/1SV59OmZRrYBC1n-5r52rb0H4BtoRNDZ3?usp=sharing)
3. build the executable
4. download font(optional) [Google Drive](https://drive.google.com/drive/folders/1h1ezExfKkZuHQqdAurZTvYSxQeef1I7m?usp=sharing)

```bash
# generated to ./bin/facenet
make facenet
```

## Demo

![demo screen capture](https://github.com/bububa/facenet/blob/main/cmd/camera/demo.gif?raw=true)

## Install

go get -u github.com/bububa/facenet

## Usage

### Train faces

```bash
./bin/facenet -model=./models/facenet -db=./models/people.db -train={image folder for training} -output={fold path for output thumbs(optional)}
```

the train folder include folders which name is the label with images inside

### Update distinct labels

```bash
./bin/facenet -model=./models/facenet -db=./models/people.db -update={labels for update seperated by comma} -output={fold path for output thumbs(optional)}
```

### Delete distinct labels from people model

```bash
./bin/facenet -model=./models/facenet -db=./models/people.db -delete={labels for delete seperated by comma} -output={fold path for output thumbs(optional)}
```

### Detect faces for image

```bash
./bin/facenet -model=./models/facenet -db=./models/people.db -detect={the image file path for detecting} -font={font folder for output image(optional)} -output={fold path for output thumbs(optional)}
```

## Camera & Server

### Requirements

- [libjpeg-turbo](https://www.libjpeg-turbo.org/) (use `-tags jpeg` to build without `CGo`)
- On Linux/RPi native Go [V4L](https://github.com/korandiz/v4l) implementation is used to capture images.

### Use Opencv4

```bash
make cvcamera
```

### On linux/Pi

```bash
# use native Go V4L implementation is used to capture images
make linux_camera
```

### Use image/jpeg instead of libjpeg-turbo

use jpeg build tag to build with native Go `image/jpeg` instead of `libjpeg-turbo`

```bash
go build -o=./bin/cvcamera -tags=cv4,jpeg ./cmd/camera
```

### Usage as Server

```
Usage of camera:
  -bind string
	Bind address (default ":56000")
  -delay int
	Delay between frames, in milliseconds (default 10)
  -width float
	Frame width (default 640)
  -height float
	Frame height (default 480)
  -index int
	Camera index
  -model string
    saved_mode path
  -db string
    classifier db
```

## User as lib

```golang
import (
    "log"

	"github.com/llgcode/draw2d"

    "github.com/bububa/facenet"
)

func main() {
    estimator, err := facenet.New(
        facenet.WithModel("./models/facenet"),
        facenet.WithDB("./models/people.db"),
        facenet.WithFontPath("./font"),
    )
    if err != nil {
       log.Fatalln(err)
    }
	err = estimator.SetFont(&draw2d.FontData{
		Name: "NotoSansCJKsc",
		//Name:   "Roboto",
		Family: draw2d.FontFamilySans,
		Style:  draw2d.FontStyleNormal,
	}, 9)
	if err != nil {
		log.Fatalln(err)
	}

    // Delete labels
    {
        labels := []string{"xxx", "yyy"}
        for _, label := range labels {
            if deleted := estimator.DeletePerson(label); deleted {
                log.Printf("[INFO] person: %s deleted\n", label)
                continue
            }
            log.Printf("[WRN] person: %s not found\n", label)
        }
        err := estimator.SaveDB("./models/people.db")
        if err != nil {
            log.Fatalln(err)
        }
    }

    // Detect faces
    {
        img, _ := loadImage(imgPath)
        minSize := 20
		markers, err := instance.DetectFaces(img, minSize)
		if err != nil {
			log.Fatalln(err)
		}
		for _, marker := range markers.Markers() {
			if marker.Error() != nil {
				log.Printf("label: %s, %v\n", marker.Label(), marker.Error())
			} else {
				log.Printf("label: %s, distance:%f\n", marker.Label(), marker.Distance())
			}
		}
		if outputPath != "" {
            txtColor := "#FFF"
            successColor := "#4CAF50"
            failedColor := "#F44336"
            strokeWidth := 2
            successMarkerOnly := false
			markerImg := estimator.DrawMarkers(markers, txtColor, successColor, failedColor, 2, successMarkerOnly)
			if err := saveImage(markerImg, outputPath); err != nil {
				log.Fatalln(err)
			}
		}
    }

    // Training
    // check cmd/facenet
}
```
