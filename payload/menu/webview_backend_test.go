package menu

import (
	"errors"
	"testing"
)

func TestWebViewBackendRequiresContext(t *testing.T) {
	backend := NewWebViewBackend()
	err := backend.Attach(nil, "http://127.0.0.1:12345/")
	if !errors.Is(err, ErrContextRequired) {
		t.Fatalf("expected ErrContextRequired, got %v", err)
	}
}
