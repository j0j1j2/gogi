package devserver

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestHandlerServesFrontendAndInjectsReloadScript(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "index.html"), "<html><body><main>menu</main></body></html>")
	writeFile(t, filepath.Join(dir, "style.css"), "body { color: red; }")

	handler := NewHandler(Options{FrontendDir: dir})

	html := getBody(t, handler, "/")
	if !strings.Contains(html, "<main>menu</main>") {
		t.Fatalf("index response missing frontend html: %s", html)
	}
	if !strings.Contains(html, `/gogi-dev/reload.js`) {
		t.Fatalf("index response missing reload script: %s", html)
	}

	css := getBody(t, handler, "/style.css")
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
