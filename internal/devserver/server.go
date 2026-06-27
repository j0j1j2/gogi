package devserver

import (
	"encoding/json"
	"fmt"
	"hash/fnv"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	urlpath "path"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/j0j1j2/gogi/internal/webclient"
)

type Options struct {
	FrontendDir string
	Addr        string
	Proxy       string
	Stdout      io.Writer
}

type patchState struct {
	Enabled bool
	Spec    patchSpec
}

type patchSpec struct {
	ID string
}

type handler struct {
	frontendDir string
	api         http.Handler
	mu          sync.Mutex
	patches     map[string]patchState
}

func NewHandler(opts Options) http.Handler {
	h := &handler{
		frontendDir: opts.FrontendDir,
		patches:     map[string]patchState{},
	}
	if h.frontendDir == "" {
		h.frontendDir = "frontend"
	}
	if opts.Proxy != "" {
		if upstream, err := url.Parse(opts.Proxy); err == nil {
			h.api = httputil.NewSingleHostReverseProxy(upstream)
		}
	}
	return h
}

func Serve(opts Options) error {
	addr := opts.Addr
	if addr == "" {
		addr = "127.0.0.1:17374"
	}
	if opts.Stdout != nil {
		fmt.Fprintf(opts.Stdout, "dev server listening on http://%s\n", addr)
	}
	return http.ListenAndServe(addr, NewHandler(opts))
}

func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if strings.HasPrefix(r.URL.Path, "/api/") {
		h.serveAPI(w, r)
		return
	}
	switch r.URL.Path {
	case "/gogi-dev/reload.js":
		h.serveReloadJS(w)
	case "/gogi-dev/version":
		h.serveVersion(w)
	case "/gogi.js":
		h.serveClientScript(w)
	default:
		h.serveFrontend(w, r)
	}
}

func (h *handler) serveAPI(w http.ResponseWriter, r *http.Request) {
	if h.api != nil {
		h.api.ServeHTTP(w, r)
		return
	}
	switch {
	case r.Method == http.MethodGet && r.URL.Path == "/api/state":
		h.serveState(w)
	case r.Method == http.MethodPost && strings.HasPrefix(r.URL.Path, "/api/toggle/"):
		h.serveToggle(w, r)
	case r.Method == http.MethodPost && strings.HasPrefix(r.URL.Path, "/api/action/"):
		h.serveAction(w, r)
	default:
		http.NotFound(w, r)
	}
}

func (h *handler) serveState(w http.ResponseWriter) {
	h.mu.Lock()
	defer h.mu.Unlock()
	writeJSON(w, map[string]any{"patches": h.patches, "actions": map[string]any{}})
}

func (h *handler) serveToggle(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/toggle/")
	if id == "" {
		http.Error(w, "patch id is required", http.StatusBadRequest)
		return
	}
	enabled, err := strconv.ParseBool(r.URL.Query().Get("enabled"))
	if err != nil {
		http.Error(w, "enabled must be true or false", http.StatusBadRequest)
		return
	}
	h.mu.Lock()
	h.patches[id] = patchState{Enabled: enabled, Spec: patchSpec{ID: id}}
	h.mu.Unlock()
	writeJSON(w, map[string]any{"ok": true})
}

func (h *handler) serveAction(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/action/")
	if id == "" {
		http.Error(w, "action id is required", http.StatusBadRequest)
		return
	}
	payload, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeJSON(w, map[string]any{
		"ok": true,
		"result": map[string]any{
			"id":      id,
			"payload": json.RawMessage(payload),
		},
	})
}

func (h *handler) serveClientScript(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
	_, _ = io.WriteString(w, webclient.Script)
}

func (h *handler) serveReloadJS(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
	_, _ = io.WriteString(w, `(function(){
let last="";
async function tick(){
  try {
    const response = await fetch("/gogi-dev/version", {cache:"no-store"});
    const next = await response.text();
    if (last && next !== last) location.reload();
    last = next;
  } catch (_) {}
}
setInterval(tick, 600);
tick();
})();`)
}

func (h *handler) serveVersion(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	fmt.Fprint(w, h.frontendVersion())
}

func (h *handler) frontendVersion() string {
	hash := fnv.New64a()
	_ = filepath.WalkDir(h.frontendDir, func(path string, entry os.DirEntry, err error) error {
		if err != nil || entry.IsDir() {
			return nil
		}
		info, err := entry.Info()
		if err != nil {
			return nil
		}
		_, _ = fmt.Fprintf(hash, "%s:%d:%d\n", path, info.Size(), info.ModTime().UnixNano())
		return nil
	})
	return fmt.Sprintf("%x", hash.Sum64())
}

func (h *handler) serveFrontend(w http.ResponseWriter, r *http.Request) {
	assetPath := strings.TrimPrefix(urlpath.Clean("/"+r.URL.Path), "/")
	if assetPath == "." || assetPath == "" {
		assetPath = "index.html"
	}
	target := filepath.Join(h.frontendDir, filepath.FromSlash(assetPath))
	data, err := os.ReadFile(target)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	if filepath.Base(target) == "index.html" || strings.HasSuffix(target, ".html") {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		data = injectReloadScript(data)
	} else if strings.HasSuffix(target, ".css") {
		w.Header().Set("Content-Type", "text/css; charset=utf-8")
	} else if strings.HasSuffix(target, ".js") {
		w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
	}
	http.ServeContent(w, r, target, time.Now(), strings.NewReader(string(data)))
}

func injectReloadScript(data []byte) []byte {
	script := []byte(`<script src="/gogi-dev/reload.js"></script>`)
	bodyEnd := []byte("</body>")
	if idx := strings.LastIndex(strings.ToLower(string(data)), string(bodyEnd)); idx >= 0 {
		out := make([]byte, 0, len(data)+len(script))
		out = append(out, data[:idx]...)
		out = append(out, script...)
		out = append(out, data[idx:]...)
		return out
	}
	return append(data, script...)
}

func writeJSON(w http.ResponseWriter, value any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(value)
}
