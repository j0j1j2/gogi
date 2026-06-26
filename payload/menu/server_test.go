package menu

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"gogi/payload/control"
	"gogi/payload/mem"
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
