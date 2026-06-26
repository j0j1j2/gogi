package menu

import (
	"embed"
	"encoding/json"
	"net/http"
	"strings"

	"gogi/payload/control"
)

//go:embed assets/menu.html assets/menu.css assets/menu.js
var assets embed.FS

type Server struct {
	registry *control.Registry
}

func NewServer(registry *control.Registry) *Server {
	return &Server{registry: registry}
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/state", s.handleState)
	mux.HandleFunc("/api/toggle/", s.handleToggle)
	mux.HandleFunc("/", s.handleAsset)
	return mux
}

func (s *Server) handleState(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")
	_ = json.NewEncoder(w).Encode(s.registry.Snapshot())
}

func (s *Server) handleToggle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	id := strings.TrimPrefix(r.URL.Path, "/api/toggle/")
	enabled := r.URL.Query().Get("enabled") == "true"
	if err := s.registry.Toggle(id, enabled); err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	s.handleState(w, r)
}

func (s *Server) handleAsset(w http.ResponseWriter, r *http.Request) {
	name := "menu.html"
	if r.URL.Path == "/menu.css" {
		name = "menu.css"
		w.Header().Set("content-type", "text/css")
	}
	if r.URL.Path == "/menu.js" {
		name = "menu.js"
		w.Header().Set("content-type", "application/javascript")
	}
	data, err := assets.ReadFile("assets/" + name)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	_, _ = w.Write(data)
}
