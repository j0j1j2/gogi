package menu

import (
	"embed"
	"encoding/json"
	"io"
	"io/fs"
	"net/http"
	"path"
	"strings"

	"github.com/j0j1j2/gogi/internal/webclient"
	"github.com/j0j1j2/gogi/payload/control"
)

//go:embed assets/menu.html assets/menu.css assets/menu.js
var assets embed.FS

type Server struct {
	registry *control.Registry
	assets   Assets
}

type Assets struct {
	FS    fs.FS
	Root  string
	Index string
	CSS   string
	JS    string
}

func NewServer(registry *control.Registry) *Server {
	return NewServerWithAssets(registry, Assets{
		FS:    assets,
		Root:  "assets",
		Index: "menu.html",
		CSS:   "menu.css",
		JS:    "menu.js",
	})
}

func NewServerWithAssets(registry *control.Registry, assets Assets) *Server {
	return &Server{registry: registry, assets: assets}
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/state", s.handleState)
	mux.HandleFunc("/api/toggle/", s.handleToggle)
	mux.HandleFunc("/api/action/", s.handleAction)
	mux.HandleFunc("/gogi.js", s.handleClientScript)
	mux.HandleFunc("/favicon.ico", s.handleFavicon)
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

func (s *Server) handleAction(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	id := strings.TrimPrefix(r.URL.Path, "/api/action/")
	payload, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	result, err := s.registry.DispatchAction(id, json.RawMessage(payload))
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	w.Header().Set("content-type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"ok": true, "result": result})
}

func (s *Server) handleClientScript(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/javascript")
	_, _ = io.WriteString(w, webclient.Script)
}

func (s *Server) handleFavicon(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleAsset(w http.ResponseWriter, r *http.Request) {
	name := s.assets.Index
	if r.URL.Path == "/menu.css" {
		name = s.assets.CSS
		w.Header().Set("content-type", "text/css")
	}
	if r.URL.Path == "/menu.js" {
		name = s.assets.JS
		w.Header().Set("content-type", "application/javascript")
	}
	data, err := fs.ReadFile(s.assets.FS, path.Join(s.assets.Root, name))
	if err != nil {
		http.NotFound(w, r)
		return
	}
	_, _ = w.Write(data)
}
