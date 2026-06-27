package menu

import (
	"encoding/json"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/j0j1j2/gogi/payload/control"
	"github.com/j0j1j2/gogi/payload/mem"
)

func TestStateEndpoint(t *testing.T) {
	reg := control.NewRegistry()
	reg.Register(mem.PatchSpec{ID: "god_mode", Library: "libtarget.so"})
	server := NewServer(reg)

	req := httptest.NewRequest(http.MethodGet, "/api/state", nil)
	rec := httptest.NewRecorder()
	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "god_mode") {
		t.Fatalf("body missing patch: %s", rec.Body.String())
	}
}

func TestServerServesCustomFrontendAssets(t *testing.T) {
	registry := control.NewRegistry()
	server := NewServerWithAssets(registry, Assets{
		FS: fstest.MapFS{
			"frontend/index.html": &fstest.MapFile{Data: []byte("<main>custom</main>")},
			"frontend/style.css":  &fstest.MapFile{Data: []byte("body{}")},
			"frontend/main.js":    &fstest.MapFile{Data: []byte("console.log('custom')")},
		},
		Root:  "frontend",
		Index: "index.html",
		CSS:   "style.css",
		JS:    "main.js",
	})

	for _, tc := range []struct {
		path string
		want string
	}{
		{path: "/", want: "custom"},
		{path: "/menu.css", want: "body{}"},
		{path: "/menu.js", want: "console.log('custom')"},
	} {
		req := httptest.NewRequest(http.MethodGet, tc.path, nil)
		rec := httptest.NewRecorder()
		server.Handler().ServeHTTP(rec, req)
		if !strings.Contains(rec.Body.String(), tc.want) {
			t.Fatalf("%s body = %q", tc.path, rec.Body.String())
		}
	}
}

var _ fs.FS = fstest.MapFS{}

func TestToggleEndpoint(t *testing.T) {
	reg := control.NewRegistry()
	reg.Register(mem.PatchSpec{ID: "god_mode", Library: "libtarget.so"})
	server := NewServer(reg)

	req := httptest.NewRequest(http.MethodPost, "/api/toggle/god_mode?enabled=true", nil)
	rec := httptest.NewRecorder()
	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	if !reg.Snapshot().Patches["god_mode"].Enabled {
		t.Fatal("expected patch enabled")
	}
}

func TestActionEndpoint(t *testing.T) {
	reg := control.NewRegistry()
	reg.RegisterAction(control.ActionSpec{
		ID: "give_coins",
		Handler: func(req control.ActionRequest) (any, error) {
			var payload map[string]int
			if err := json.Unmarshal(req.Payload, &payload); err != nil {
				return nil, err
			}
			return map[string]int{"amount": payload["amount"]}, nil
		},
	})
	server := NewServer(reg)

	req := httptest.NewRequest(http.MethodPost, "/api/action/give_coins", strings.NewReader(`{"amount":10}`))
	rec := httptest.NewRecorder()
	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"amount":10`) {
		t.Fatalf("body missing action result: %s", rec.Body.String())
	}
}

func TestClientScriptEndpoint(t *testing.T) {
	reg := control.NewRegistry()
	server := NewServer(reg)

	req := httptest.NewRequest(http.MethodGet, "/gogi.js", nil)
	rec := httptest.NewRecorder()
	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "window.gogi") {
		t.Fatalf("client script missing window.gogi: %s", rec.Body.String())
	}
}

func TestFaviconEndpointDoesNot404(t *testing.T) {
	reg := control.NewRegistry()
	server := NewServer(reg)

	req := httptest.NewRequest(http.MethodGet, "/favicon.ico", nil)
	rec := httptest.NewRecorder()
	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
}
