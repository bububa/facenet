package handlers

import (
	"fmt"
	"net/http"
)

// HTML handler.
type HTML struct {
	Width    float64
	Height   float64
	UseWebGL bool
}

// NewHTML returns new HTML handler.
func NewHTML(width, height float64, useWebGL bool) *HTML {
	return &HTML{
		Width:    width,
		Height:   height,
		UseWebGL: useWebGL,
	}
}

// ServeHTTP handles requests on incoming connections.
func (h *HTML) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" && r.Method != "HEAD" {
		msg := fmt.Sprintf("405 Method Not Allowed (%s)", r.Method)
		http.Error(w, msg, http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	tpl.ExecuteTemplate(w, "html.tpl", h)
}
