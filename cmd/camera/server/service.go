package server

import (
	"context"
	"log"
	"net"
	"net/http"

	"github.com/bububa/camera"
	"github.com/bububa/facenet"
	"github.com/bububa/facenet/cmd/camera/server/handlers"
)

// Server represents server
type Server struct {
	srv         *http.Server
	net         *facenet.Instance
	cam         *camera.Camera
	delay       int
	bind        string
	frameWidth  float64
	frameHeight float64
}

// New returns new Server for binding address, facenet instance and camera
func New(bind string, net *facenet.Instance, cam *camera.Camera) *Server {
	s := &Server{
		srv:         new(http.Server),
		net:         net,
		cam:         cam,
		bind:        bind,
		frameWidth:  FrameWidth,
		frameHeight: FrameHeight,
		delay:       Delay,
	}
	return s
}

// SetFrameSize set frame size for display
func (s *Server) SetFrameSize(width float64, height float64) {
	s.frameWidth = width
	s.frameHeight = height
}

// SetDelay set delay between two frames in milliseconds
func (s *Server) SetDelay(delay int) {
	s.delay = delay
}

// Start to start server
func (s *Server) Start() error {
	if err := s.cam.Start(); err != nil {
		return err
	}
	return s.ListenAndServe()
}

// Shutdown to shutdown server
func (s *Server) Shutdown(ctx context.Context) error {
	log.Println("camera: closing")
	if err := s.cam.Close(); err != nil {
		return err
	}
	log.Println("camera: closed")
	return s.srv.Shutdown(ctx)
}

// ListenAndServe listens on the TCP address and serves requests.
func (s *Server) ListenAndServe() error {
	http.Handle("/html/webgl", handlers.NewHTML(s.frameWidth, s.frameHeight, true))
	http.Handle("/html", handlers.NewHTML(s.frameWidth, s.frameHeight, false))
	http.Handle("/jpeg", handlers.NewJPEG(s.net, s.cam))
	http.Handle("/mjpeg", handlers.NewMJPEG(s.net, s.cam, s.delay))
	http.Handle("/socket", handlers.NewSocket(s.net, s.cam, s.delay))

	http.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	http.Handle("/", handlers.NewIndex())

	listener, err := net.Listen("tcp4", s.bind)
	if err != nil {
		return err
	}

	return s.srv.Serve(listener)
}
