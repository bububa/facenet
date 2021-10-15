package handlers

import (
	"bytes"
	"context"
	"log"
	"net/http"
	"time"

	"nhooyr.io/websocket"

	"github.com/bububa/camera"
	"github.com/bububa/camera/image"
	"github.com/bububa/facenet"
)

// Socket handler.
type Socket struct {
	e     *facenet.Estimator
	cam   *camera.Camera
	delay int
}

// NewSocket returns new socket handler.
func NewSocket(e *facenet.Estimator, cam *camera.Camera, delay int) *Socket {
	return &Socket{e, cam, delay}
}

// ServeHTTP handles requests on incoming connections.
func (s *Socket) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	conn, err := websocket.Accept(w, r, nil)
	if err != nil {
		log.Printf("socket: accept: %v", err)
		return
	}

	ctx := context.Background()

	for {
		img, err := s.cam.Read()
		if err != nil {
			log.Printf("jpeg: read: %v", err)
			return
		}
		if s.e != nil {
			if markers, err := s.e.DetectFaces(img, DetectMinSize); err == nil {
				img = s.e.DrawMarkers(markers, TextColor, SuccessColor, FailedColor, StrokeWidth, false)
			}
		}

		w := new(bytes.Buffer)

		err = image.NewEncoder(w).Encode(img)
		if err != nil {
			log.Printf("socket: encode: %v", err)
			continue
		}

		b64 := image.EncodeToString(w.Bytes())

		err = conn.Write(ctx, websocket.MessageText, []byte(b64))
		if err != nil {
			break
		}

		time.Sleep(time.Duration(s.delay) * time.Millisecond)
	}

	conn.Close(websocket.StatusNormalClosure, "")
}
