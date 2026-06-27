package devserver

import (
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestHandlerServesFrontendAndInjectsReloadScript(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "index.html"), "<html><body><main>menu</main></body></html>")
	writeFile(t, filepath.Join(dir, "style.css"), "body { color: red; }")

	handler := NewHandler(Options{FrontendDir: dir})

	html := getBody(t, handler, "/")
	if !strings.Contains(html, `class="gogi-phone"`) || !strings.Contains(html, `src="/gogi-dev/app/"`) {
		t.Fatalf("root response missing phone preview shell: %s", html)
	}
	if !strings.Contains(html, `class="gogi-log-panel"`) || !strings.Contains(html, `/gogi-dev/logs`) {
		t.Fatalf("root response missing activity log panel: %s", html)
	}
	if !strings.Contains(html, `id="gogi-toast"`) || !strings.Contains(html, `id="gogi-memory-list"`) {
		t.Fatalf("root response missing visible event and memory indicators: %s", html)
	}

	app := getBody(t, handler, "/gogi-dev/app/")
	if !strings.Contains(app, "<main>menu</main>") {
		t.Fatalf("app response missing frontend html: %s", app)
	}
	if !strings.Contains(app, `/gogi-dev/reload.js`) {
		t.Fatalf("app response missing reload script: %s", app)
	}

	css := getBody(t, handler, "/gogi-dev/app/style.css")
	if !strings.Contains(css, "color: red") {
		t.Fatalf("css response = %q", css)
	}
}

func TestHandlerServesMockAPIState(t *testing.T) {
	handler := NewHandler(Options{})

	body := getBody(t, handler, "/api/state")
	var state map[string]any
	if err := json.Unmarshal([]byte(body), &state); err != nil {
		t.Fatalf("decode state: %v, body=%s", err, body)
	}
	if _, ok := state["patches"].(map[string]any); !ok {
		t.Fatalf("state missing patches object: %#v", state)
	}
}

func TestHandlerTogglesMockPatch(t *testing.T) {
	handler := NewHandler(Options{})

	req := httptest.NewRequest(http.MethodPost, "/api/toggle/example?enabled=true", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("toggle status = %d, body=%s", rec.Code, rec.Body.String())
	}

	body := getBody(t, handler, "/api/state")
	if !strings.Contains(body, `"example"`) || !strings.Contains(body, `"Enabled":true`) {
		t.Fatalf("state did not preserve toggle: %s", body)
	}

	logs := getBody(t, handler, "/gogi-dev/logs")
	if !strings.Contains(logs, "mock memory") || !strings.Contains(logs, "example") {
		t.Fatalf("logs missing mock memory toggle event: %s", logs)
	}
	if !strings.Contains(logs, `"patches"`) || !strings.Contains(logs, `"Enabled":true`) {
		t.Fatalf("logs missing mock memory state: %s", logs)
	}
}

func TestHandlerServesClientScript(t *testing.T) {
	handler := NewHandler(Options{})

	body := getBody(t, handler, "/gogi.js")
	if !strings.Contains(body, "window.gogi") || !strings.Contains(body, "action:") {
		t.Fatalf("client script missing gogi client API: %s", body)
	}
}

func TestHandlerMocksAction(t *testing.T) {
	handler := NewHandler(Options{})

	req := httptest.NewRequest(http.MethodPost, "/api/action/give_coins", strings.NewReader(`{"amount":10}`))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("action status = %d, body=%s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"id":"give_coins"`) {
		t.Fatalf("action response = %s", rec.Body.String())
	}

	logs := getBody(t, handler, "/gogi-dev/logs")
	if !strings.Contains(logs, "mock action") || !strings.Contains(logs, "give_coins") {
		t.Fatalf("logs missing mock action event: %s", logs)
	}
}

func TestHandlerIgnoresFavicon(t *testing.T) {
	handler := NewHandler(Options{})

	req := httptest.NewRequest(http.MethodGet, "/favicon.ico", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("favicon status = %d, body=%s", rec.Code, rec.Body.String())
	}
}

func TestHandlerProxiesAPI(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/state" {
			t.Fatalf("proxied path = %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		_, _ = w.Write([]byte(`{"proxied":true}`))
	}))
	t.Cleanup(upstream.Close)

	handler := NewHandler(Options{Proxy: upstream.URL})
	req := httptest.NewRequest(http.MethodGet, "/api/state", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusAccepted {
		t.Fatalf("proxy status = %d, body=%s", rec.Code, rec.Body.String())
	}
	if rec.Body.String() != `{"proxied":true}` {
		t.Fatalf("proxy body = %q", rec.Body.String())
	}
}

func TestListenFallsBackWhenPortIsBusy(t *testing.T) {
	busy, nextAddr := reservePortWithFreeNext(t)
	t.Cleanup(func() { _ = busy.Close() })

	listener, actual, err := Listen(Options{Addr: busy.Addr().String(), PortSearchLimit: 2})
	if err != nil {
		t.Fatalf("Listen returned error: %v", err)
	}
	t.Cleanup(func() { _ = listener.Close() })

	if actual != nextAddr {
		t.Fatalf("actual addr = %q, want %q", actual, nextAddr)
	}
}

func getBody(t *testing.T, handler http.Handler, path string) string {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, path, nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET %s status = %d, body=%s", path, rec.Code, rec.Body.String())
	}
	return rec.Body.String()
}

func writeFile(t *testing.T, path string, data string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(data), 0o644); err != nil {
		t.Fatal(err)
	}
}

func reservePortWithFreeNext(t *testing.T) (net.Listener, string) {
	t.Helper()
	for i := 0; i < 100; i++ {
		busy, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			t.Fatal(err)
		}
		_, port, err := net.SplitHostPort(busy.Addr().String())
		if err != nil {
			_ = busy.Close()
			t.Fatal(err)
		}
		next, err := strconv.Atoi(port)
		if err != nil {
			_ = busy.Close()
			t.Fatal(err)
		}
		next++
		nextAddr := net.JoinHostPort("127.0.0.1", strconv.Itoa(next))
		probe, err := net.Listen("tcp", nextAddr)
		if err == nil {
			_ = probe.Close()
			return busy, nextAddr
		}
		_ = busy.Close()
	}
	t.Fatal("could not find adjacent free ports")
	return nil, ""
}
