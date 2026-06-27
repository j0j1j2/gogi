package devserver

import (
	"encoding/json"
	"errors"
	"fmt"
	"hash/fnv"
	"io"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	urlpath "path"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/j0j1j2/gogi/internal/webclient"
)

type Options struct {
	FrontendDir     string
	Addr            string
	Proxy           string
	Stdout          io.Writer
	PortSearchLimit int
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
	events      []devEvent
	nextEventID int
}

type devEvent struct {
	ID      int    `json:"id"`
	Time    string `json:"time"`
	Kind    string `json:"kind"`
	Title   string `json:"title"`
	Message string `json:"message"`
	Detail  string `json:"detail,omitempty"`
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
	listener, addr, err := Listen(opts)
	if err != nil {
		return err
	}
	if opts.Stdout != nil {
		fmt.Fprintf(opts.Stdout, "dev server listening on http://%s\n", addr)
	}
	return http.Serve(listener, NewHandler(opts))
}

func Listen(opts Options) (net.Listener, string, error) {
	addr := opts.Addr
	if addr == "" {
		addr = "127.0.0.1:17374"
	}
	limit := opts.PortSearchLimit
	if limit <= 0 {
		limit = 20
	}
	host, portText, err := net.SplitHostPort(addr)
	if err != nil {
		listener, listenErr := net.Listen("tcp", addr)
		return listener, addr, listenErr
	}
	port, err := strconv.Atoi(portText)
	if err != nil || port == 0 {
		listener, listenErr := net.Listen("tcp", addr)
		return listener, addr, listenErr
	}
	for offset := 0; offset < limit; offset++ {
		candidate := net.JoinHostPort(host, strconv.Itoa(port+offset))
		listener, err := net.Listen("tcp", candidate)
		if err == nil {
			return listener, listener.Addr().String(), nil
		}
		if !isAddrInUse(err) {
			return nil, "", err
		}
	}
	return nil, "", fmt.Errorf("no available port found from %s within %d attempts", addr, limit)
}

func isAddrInUse(err error) bool {
	return errors.Is(err, syscall.EADDRINUSE)
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
	case "/gogi-dev/logs":
		h.serveLogs(w)
	case "/gogi.js":
		h.serveClientScript(w)
	case "/favicon.ico":
		w.WriteHeader(http.StatusNoContent)
	default:
		if r.URL.Path == "/" {
			h.servePreviewShell(w)
			return
		}
		if strings.HasPrefix(r.URL.Path, "/gogi-dev/app/") {
			h.serveFrontend(w, r, "/gogi-dev/app/")
			return
		}
		http.NotFound(w, r)
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
	h.appendEventLocked(devEvent{
		Kind:    "memory",
		Title:   "mock memory",
		Message: fmt.Sprintf("patch %s set to %t", id, enabled),
		Detail:  fmt.Sprintf("mock://memory/patch/%s", id),
	})
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
	h.appendEvent(devEvent{
		Kind:    "action",
		Title:   "mock action",
		Message: fmt.Sprintf("action %s invoked", id),
		Detail:  string(payload),
	})
	writeJSON(w, map[string]any{
		"ok": true,
		"result": map[string]any{
			"id":      id,
			"payload": json.RawMessage(payload),
		},
	})
}

func (h *handler) appendEvent(event devEvent) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.appendEventLocked(event)
}

func (h *handler) appendEventLocked(event devEvent) {
	h.nextEventID++
	event.ID = h.nextEventID
	event.Time = time.Now().Format("15:04:05")
	h.events = append(h.events, event)
	if len(h.events) > 200 {
		h.events = h.events[len(h.events)-200:]
	}
}

func (h *handler) serveLogs(w http.ResponseWriter) {
	h.mu.Lock()
	defer h.mu.Unlock()
	events := make([]devEvent, len(h.events))
	copy(events, h.events)
	patches := make(map[string]patchState, len(h.patches))
	for id, patch := range h.patches {
		patches[id] = patch
	}
	writeJSON(w, map[string]any{"events": events, "patches": patches})
}

func (h *handler) serveClientScript(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
	_, _ = io.WriteString(w, webclient.Script)
}

func (h *handler) servePreviewShell(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = io.WriteString(w, `<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>gogi dev</title>
  <style>
    :root {
      color-scheme: dark;
      font-family: ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif;
      --gogi-preview-width: min(390px, calc(100vw - 32px));
      --gogi-preview-height: min(844px, calc(100vh - 86px));
      background: #171918;
      color: #edf2ef;
    }
    * { box-sizing: border-box; }
    html, body { min-height: 100%; margin: 0; }
    body {
      display: grid;
      place-items: center;
      padding: 28px;
      background:
        linear-gradient(135deg, rgba(255,255,255,0.05), transparent 32%),
        #171918;
    }
    .gogi-stage {
      display: grid;
      gap: 14px;
      justify-items: center;
    }
    .gogi-workbench {
      display: grid;
      grid-template-columns: minmax(280px, 390px) minmax(280px, 360px);
      gap: 22px;
      align-items: start;
    }
    .gogi-toolbar {
      width: var(--gogi-preview-width);
      display: flex;
      align-items: center;
      justify-content: space-between;
      color: #aeb8b2;
      font-size: 12px;
    }
    .gogi-phone {
      width: var(--gogi-preview-width);
      height: var(--gogi-preview-height);
      border: 10px solid #0a0d0c;
      border-radius: 34px;
      background: #101312;
      overflow: hidden;
      box-shadow: 0 24px 80px rgba(0,0,0,0.45), inset 0 0 0 1px rgba(255,255,255,0.08);
      position: relative;
    }
    .gogi-phone::before {
      content: "";
      position: absolute;
      top: 8px;
      left: 50%;
      width: 92px;
      height: 20px;
      transform: translateX(-50%);
      border-radius: 999px;
      background: #090b0a;
      z-index: 2;
    }
    .gogi-screen {
      width: 100%;
      height: 100%;
      border: 0;
      background: #101312;
    }
    .gogi-log-panel {
      width: min(360px, calc(100vw - 32px));
      height: var(--gogi-preview-height);
      padding: 16px;
      border: 1px solid rgba(255,255,255,0.08);
      border-radius: 18px;
      background: #202420;
      box-shadow: 0 18px 50px rgba(0,0,0,0.28);
      overflow: hidden;
      display: flex;
      flex-direction: column;
    }
    .gogi-live-status {
      display: grid;
      gap: 8px;
      margin-bottom: 12px;
      padding: 10px 12px;
      border-radius: 12px;
      background: #151916;
      border: 1px solid rgba(255,255,255,0.07);
    }
    .gogi-live-label {
      color: #8f9b93;
      font-size: 11px;
      text-transform: uppercase;
      letter-spacing: 0.08em;
    }
    .gogi-live-value {
      color: #eef5f0;
      font-size: 13px;
      line-height: 1.35;
      min-height: 18px;
      overflow-wrap: anywhere;
    }
    .gogi-live-status.is-hot {
      border-color: rgba(104, 202, 170, 0.75);
      box-shadow: 0 0 0 3px rgba(104, 202, 170, 0.13);
    }
    .gogi-toast {
      position: fixed;
      left: 50%;
      bottom: 22px;
      z-index: 5;
      transform: translate(-50%, 18px);
      opacity: 0;
      pointer-events: none;
      max-width: min(520px, calc(100vw - 32px));
      padding: 11px 14px;
      border-radius: 999px;
      background: #e8f5ee;
      color: #14201a;
      box-shadow: 0 18px 55px rgba(0,0,0,0.34);
      font-size: 13px;
      font-weight: 700;
      transition: opacity 160ms ease, transform 160ms ease;
    }
    .gogi-toast.is-visible {
      opacity: 1;
      transform: translate(-50%, 0);
    }
    .gogi-memory-card {
      display: grid;
      gap: 9px;
      margin-bottom: 12px;
      padding: 12px;
      border-radius: 12px;
      background: #151916;
      border: 1px solid rgba(255,255,255,0.07);
    }
    .gogi-memory-head {
      display: flex;
      justify-content: space-between;
      gap: 12px;
      color: #eef5f0;
      font-size: 13px;
      font-weight: 700;
    }
    .gogi-memory-list {
      display: grid;
      gap: 7px;
    }
    .gogi-memory-row {
      display: flex;
      justify-content: space-between;
      gap: 10px;
      align-items: center;
      color: #c8d2cc;
      font-size: 12px;
      overflow-wrap: anywhere;
    }
    .gogi-memory-badge {
      flex: 0 0 auto;
      min-width: 44px;
      border-radius: 999px;
      padding: 3px 8px;
      text-align: center;
      background: #333b35;
      color: #aeb8b2;
      font-size: 11px;
      font-weight: 700;
    }
    .gogi-memory-badge.is-on {
      background: #2f7d68;
      color: #f2fff9;
    }
    .gogi-log-head {
      display: flex;
      justify-content: space-between;
      gap: 12px;
      align-items: baseline;
      padding-bottom: 12px;
      border-bottom: 1px solid rgba(255,255,255,0.08);
    }
    .gogi-log-head strong {
      font-size: 15px;
      letter-spacing: 0;
    }
    .gogi-log-head span {
      color: #9ba79f;
      font-size: 12px;
    }
    .gogi-log-list {
      display: grid;
      gap: 10px;
      min-height: 0;
      margin: 0;
      padding: 14px 0 2px;
      overflow: auto;
      list-style: none;
    }
    .gogi-log-empty {
      color: #89958d;
      font-size: 13px;
      line-height: 1.45;
      padding: 14px 0;
    }
    .gogi-log-item {
      display: grid;
      gap: 6px;
      padding: 11px 12px;
      border-radius: 12px;
      background: #151916;
      border: 1px solid rgba(255,255,255,0.07);
    }
    .gogi-log-item.is-new {
      border-color: rgba(104, 202, 170, 0.75);
      box-shadow: 0 0 0 3px rgba(104, 202, 170, 0.11);
    }
    .gogi-log-line {
      display: flex;
      justify-content: space-between;
      gap: 12px;
      color: #f0f5f1;
      font-size: 13px;
      font-weight: 650;
    }
    .gogi-log-time {
      color: #8f9b93;
      font-size: 12px;
      font-weight: 500;
    }
    .gogi-log-message {
      color: #c8d2cc;
      font-size: 12px;
      line-height: 1.4;
      overflow-wrap: anywhere;
    }
    .gogi-log-detail {
      color: #88bfae;
      font-size: 11px;
      line-height: 1.35;
      overflow-wrap: anywhere;
    }
    @media (max-width: 880px) {
      body { align-items: start; }
      .gogi-workbench {
        grid-template-columns: 1fr;
        justify-items: center;
      }
      .gogi-log-panel {
        height: min(420px, calc(100vh - 72px));
      }
    }
    @media (max-width: 460px) {
      body { padding: 0; background: #101312; }
      .gogi-toolbar, .gogi-log-panel { display: none; }
      .gogi-toast { display: none; }
      .gogi-stage, .gogi-phone {
        width: 100vw;
        height: 100vh;
        max-height: none;
      }
      .gogi-phone {
        border: 0;
        border-radius: 0;
        box-shadow: none;
      }
      .gogi-phone::before { display: none; }
    }
  </style>
</head>
<body>
  <main class="gogi-stage">
    <div class="gogi-toolbar">
      <strong>gogi dev</strong>
      <span>390 x 844 preview</span>
    </div>
    <div class="gogi-workbench">
      <section class="gogi-phone" aria-label="Phone preview">
        <iframe class="gogi-screen" src="/gogi-dev/app/" title="gogi frontend preview"></iframe>
      </section>
      <aside class="gogi-debug-panel gogi-log-panel" aria-label="Debug panel">
        <div class="gogi-log-head">
          <strong>Debug panel</strong>
          <span id="gogi-log-count">0 events</span>
        </div>
        <div class="gogi-live-status" id="gogi-live-status">
          <span class="gogi-live-label">Latest event</span>
          <span class="gogi-live-value" id="gogi-live-value">No mock events yet</span>
        </div>
        <section class="gogi-memory-card" aria-label="Mock memory state">
          <div class="gogi-memory-head">
            <span>Mock memory</span>
            <span id="gogi-memory-count">0 patches</span>
          </div>
          <div class="gogi-memory-list" id="gogi-memory-list">
            <div class="gogi-log-empty">No mock memory edits yet.</div>
          </div>
        </section>
        <ul class="gogi-log-list" id="gogi-log-list">
          <li class="gogi-log-empty">Interact with the preview to see mock actions and memory edits.</li>
        </ul>
      </aside>
    </div>
  </main>
  <div class="gogi-toast" id="gogi-toast" role="status" aria-live="polite"></div>
  <script>
    const list = document.getElementById("gogi-log-list");
    const count = document.getElementById("gogi-log-count");
    const toast = document.getElementById("gogi-toast");
    const liveStatus = document.getElementById("gogi-live-status");
    const liveValue = document.getElementById("gogi-live-value");
    const memoryList = document.getElementById("gogi-memory-list");
    const memoryCount = document.getElementById("gogi-memory-count");
    let seenLatestEventID = 0;
    let toastTimer = 0;
    function escapeText(value) {
      return String(value == null ? "" : value).replace(/[&<>"']/g, character => ({
        "&": "&amp;",
        "<": "&lt;",
        ">": "&gt;",
        "\"": "&quot;",
        "'": "&#39;"
      })[character]);
    }
    async function refreshLogs() {
      try {
        const response = await fetch("/gogi-dev/logs", {cache: "no-store"});
        const state = await response.json();
        const events = state.events || [];
        const patches = state.patches || {};
        const latest = events[events.length - 1];
        count.textContent = events.length + (events.length === 1 ? " event" : " events");
        renderMemory(patches);
        if (latest) renderLatest(latest);
        if (events.length === 0) {
          list.innerHTML = '<li class="gogi-log-empty">Interact with the preview to see mock actions and memory edits.</li>';
          return;
        }
        list.innerHTML = events.slice().reverse().map(event => {
          const detail = event.detail ? '<div class="gogi-log-detail">' + escapeText(event.detail) + '</div>' : "";
          const newest = latest && event.id === latest.id ? " is-new" : "";
          return '<li class="gogi-log-item' + newest + '">' +
            '<div class="gogi-log-line">' +
            '<span>' + escapeText(event.title || event.kind) + '</span>' +
            '<span class="gogi-log-time">' + escapeText(event.time) + '</span>' +
            '</div>' +
            '<div class="gogi-log-message">' + escapeText(event.message) + '</div>' +
            detail +
            '</li>';
        }).join("");
      } catch (_) {}
    }
    function renderLatest(event) {
      const text = (event.title || event.kind) + ": " + event.message;
      liveValue.textContent = text;
      if (event.id > seenLatestEventID) {
        seenLatestEventID = event.id;
        liveStatus.classList.remove("is-hot");
        void liveStatus.offsetWidth;
        liveStatus.classList.add("is-hot");
        showToast(text);
      }
    }
    function renderMemory(patches) {
      const entries = Object.entries(patches);
      memoryCount.textContent = entries.length + (entries.length === 1 ? " patch" : " patches");
      if (entries.length === 0) {
        memoryList.innerHTML = '<div class="gogi-log-empty">No mock memory edits yet.</div>';
        return;
      }
      memoryList.innerHTML = entries.map(([id, record]) => {
        const enabled = Boolean(record.Enabled);
        return '<div class="gogi-memory-row">' +
          '<span>' + escapeText(id) + '</span>' +
          '<span class="gogi-memory-badge' + (enabled ? " is-on" : "") + '">' + (enabled ? "ON" : "OFF") + '</span>' +
          '</div>';
      }).join("");
    }
    function showToast(message) {
      toast.textContent = message;
      toast.classList.add("is-visible");
      clearTimeout(toastTimer);
      toastTimer = setTimeout(() => toast.classList.remove("is-visible"), 1800);
    }
    setInterval(refreshLogs, 250);
    refreshLogs();
  </script>
</body>
</html>
`)
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

func (h *handler) serveFrontend(w http.ResponseWriter, r *http.Request, prefix string) {
	requestPath := r.URL.Path
	if prefix != "" {
		requestPath = "/" + strings.TrimPrefix(r.URL.Path, prefix)
	}
	assetPath := strings.TrimPrefix(urlpath.Clean("/"+requestPath), "/")
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
