package main

import (
	"bytes"
	"flag"
	"image"
	"image/jpeg"
	"io/fs"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"sync"

	"github.com/llgcode/draw2d"

	"github.com/bububa/facenet"
	"github.com/bububa/facenet/core"
)

// Request request options
type Request struct {
	// Model facenet model path
	Model string
	// DB path
	DB string
	// Train train images folder path
	Train string
	// Output output file/folder path
	Output string
	// Font font fold path
	Font string
}

var (
	request      Request
	infoAction   bool
	updateAction string
	deleteAction string
	detectAction string
)

func init() {
	flag.StringVar(&request.Train, "train", "", "train file path")
	flag.StringVar(&request.Output, "output", "", "output file path for detect result checking")
	flag.StringVar(&request.Model, "model", "", "facenet model file path")
	flag.StringVar(&request.DB, "db", "", "db file")
	flag.StringVar(&request.Font, "font", "", "font path")
	flag.StringVar(&deleteAction, "delete", "", "delete person, multiple names are separated by comma")
	flag.StringVar(&updateAction, "update", "", "delete person, multiple names are separated by comma")
	flag.StringVar(&detectAction, "detect", "", "detect faces in image file")
	flag.BoolVar(&infoAction, "info", false, "people model info")
}

func main() {
	flag.Parse()
	wd, err := os.Getwd()
	if err != nil {
		log.Fatalln(err)
	}
	var opts []facenet.Option
	if request.DB == "" {
		log.Fatalln("[ERR] missing db file path")
	}
	request.DB = cleanPath(wd, request.DB)
	opts = append(opts, facenet.WithDB(request.DB))
	if request.Model == "" && !infoAction {
		log.Fatalln("[ERR] missing facenet model file path")
	} else {
		request.Model = cleanPath(wd, request.Model)
		opts = append(opts, facenet.WithModel(request.Model))
	}
	if request.Font != "" {
		request.Font = cleanPath(wd, request.Font)
		opts = append(opts, facenet.WithFontPath(request.Font))
	}
	instance, err := facenet.New(opts...)
	if err != nil {
		log.Fatalln(err)
	}
	err = instance.SetFont(&draw2d.FontData{
		Name: "NotoSansCJKsc",
		//Name:   "Roboto",
		Family: draw2d.FontFamilySans,
		Style:  draw2d.FontStyleNormal,
	}, 9)
	if err != nil {
		log.Fatalln(err)
	}
	log.Printf("[INFO] loaded %d people\n", len(instance.People().GetList()))
	if infoAction {
		for _, people := range instance.People().GetList() {
			log.Printf("[INFO] people:%s, embeddings:%d\n", people.GetName(), len(people.GetEmbeddings()))
		}
		return
	}
	if deleteAction != "" {
		labels := strings.Split(deleteAction, ",")
		for _, label := range labels {
			if deleted := instance.DeletePerson(label); deleted {
				log.Printf("[INFO] person: %s deleted\n", label)
				continue
			}
			log.Printf("[WRN] person: %s not found\n", label)
		}
		err := instance.SaveDB(request.DB)
		if err != nil {
			log.Fatalln(err)
		}
		return
	}

	if request.Output != "" {
		request.Output = cleanPath(wd, request.Output)
	}

	if detectAction != "" {
		detectFilePath := cleanPath(wd, detectAction)
		img, err := loadImage(detectFilePath)
		if err != nil {
			log.Fatalln(err)
		}
		markers, err := instance.DetectFaces(img, 20)
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
		if request.Output != "" {
			markerImg := instance.DrawMarkers(markers, "#FFF", "#4CAF50", "#F44336", 2, false)
			if err := saveImage(markerImg, request.Output); err != nil {
				log.Fatalln(err)
			}
		}
		return
	}

	if request.Train == "" {
		log.Fatalln("[ERR] missing train file path")
	}
	request.Train = cleanPath(wd, request.Train)
	trainPathBase := filepath.Base(request.Train)

	updateLabels := make(map[string]struct{})
	if updateAction != "" {
		labels := strings.Split(updateAction, ",")
		for _, label := range labels {
			updateLabels[strings.TrimSpace(label)] = struct{}{}
		}
	}

	var labelPathes []string
	if err := filepath.Walk(request.Train, func(labelPath string, info fs.FileInfo, err error) error {
		label := filepath.Base(labelPath)
		if !info.IsDir() || trainPathBase == label {
			return nil
		}
		labelPathes = append(labelPathes, labelPath)
		return nil
	}); err != nil {
		log.Fatalln(err)
	}
	wg := new(sync.WaitGroup)
	for _, labelPath := range labelPathes {
		label := filepath.Base(labelPath)
		label = strings.TrimSpace(label)
		if _, found := updateLabels[label]; !found && updateAction != "" {
			continue
		}
		wg.Add(1)
		go func(ins *facenet.Estimator, label string, labelPath string, output string) {
			defer wg.Done()
			extractPersonInFolder(ins, label, labelPath, output)
		}(instance, label, labelPath, request.Output)
	}
	wg.Wait()

	instance.BatchTrain(0.75, 1000, 20, 4)
	if err := instance.SaveDB(request.DB); err != nil {
		log.Fatalln(err)
	}
}

func extractPersonInFolder(ins *facenet.Estimator, label string, labelPath string, output string) error {
	var filenames []string
	if err := filepath.Walk(labelPath, func(filename string, info fs.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		filenames = append(filenames, filename)
		return nil
	}); err != nil {
		return err
	}
	if len(filenames) == 0 {
		return nil
	}
	person := core.Person{
		Name: label,
	}
	wg := new(sync.WaitGroup)
	locker := new(sync.Mutex)
	for _, filename := range filenames {
		fname := filename
		wg.Add(1)
		go func(ins *facenet.Estimator, person *core.Person) {
			defer wg.Done()
			locker.Lock()
			defer locker.Unlock()
			extractPerson(ins, fname, person, output)
		}(ins, &person)
	}
	wg.Wait()
	log.Printf("[INFO] person: %s, embeddings: %d\n", person.GetName(), len(person.Embeddings))
	if len(person.GetEmbeddings()) > 0 {
		ins.AddPersonSafe(&person)
	}
	return nil
}

func extractPerson(ins *facenet.Estimator, filename string, person *core.Person, thumbPath string) error {
	label := person.GetName()
	baseName := filepath.Base(filename)
	img, err := loadImage(filename)
	if err != nil {
		log.Printf("[ERR] loadimage label:%s, file:%s, %v\n", label, baseName, err)
		return nil
	}
	marker, err := ins.ExtractFaceSafe(person, img, 20)
	if err != nil {
		log.Printf("[ERR] label:%s, file:%s, %v\n", label, baseName, err)
		return nil
	}
	if thumbPath != "" {
		folder := filepath.Join(thumbPath, label)
		if _, err := os.Stat(folder); err != nil && os.IsNotExist(err) {
			err = os.MkdirAll(folder, os.ModePerm)
			if err != nil {
				log.Fatalf("[ERR] create folder failed, %v\n", err)
			}
		}
		filePath := filepath.Join(folder, baseName)
		if err := saveImage(marker.Thumb(img), filePath); err != nil {
			log.Fatalf("[ERR] save image failed, %v\n", err)
		}
	}
	log.Printf("[SUCCESS] label:%s, file:%s, face detected\n", label, baseName)
	return nil
}

func loadImage(filePath string) (image.Image, error) {
	fn, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer fn.Close()
	img, _, err := image.Decode(fn)
	if err != nil {
		return nil, err
	}
	return img, nil
}

func saveImage(img image.Image, filePath string) error {
	buf := new(bytes.Buffer)
	if err := jpeg.Encode(buf, img, nil); err != nil {
		return err
	}
	fn, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer fn.Close()
	fn.Write(buf.Bytes())
	return nil
}

func cleanPath(wd string, path string) string {
	usr, _ := user.Current()
	dir := usr.HomeDir
	if path == "~" {
		return dir
	} else if strings.HasPrefix(path, "~/") {
		return filepath.Join(dir, path[2:])
	}
	return filepath.Join(wd, path)
}
