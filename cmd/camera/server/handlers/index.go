package handlers

import (
	"embed"
	"html/template"
	"log"
	"net/http"
)

//go:embed templates
var tplFS embed.FS

var tpl *template.Template

func init() {
	var err error
	tpl, err = template.ParseFS(tplFS, "templates/*.tpl")
	if err != nil {
		log.Fatalln(err)
	}
}

// Index handler.
type Index struct {
}

// NewIndex returns new Index handler.
func NewIndex() *Index {
	return &Index{}
}

// ServeHTTP handles requests on incoming connections.
func (s *Index) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" && r.Method != "HEAD" {
		http.Error(w, "405 Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	tpl.ExecuteTemplate(w, "index.tpl", nil)
}
