package menu

import (
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
