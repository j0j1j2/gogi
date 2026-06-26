package control

import (
	"testing"

	"gogi/payload/mem"
)

func TestRegistryToggleState(t *testing.T) {
	reg := NewRegistry()
	reg.Register(mem.PatchSpec{ID: "god_mode", Library: "libtarget.so"})

	if err := reg.Toggle("god_mode", true); err != nil {
		t.Fatalf("Toggle returned error: %v", err)
	}

	state := reg.Snapshot()
	if !state.Patches["god_mode"].Enabled {
		t.Fatalf("expected god_mode enabled")
	}
}

func TestRegistryRejectsUnknownPatch(t *testing.T) {
	reg := NewRegistry()
	if err := reg.Toggle("missing", true); err == nil {
		t.Fatal("expected unknown patch error")
	}
}
