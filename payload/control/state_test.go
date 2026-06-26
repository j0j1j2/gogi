package control

import (
	"testing"

	"github.com/j0j1j2/gogi/payload/mem"
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

func TestRegistryToggleCallsApplier(t *testing.T) {
	applier := &fakeApplier{}
	reg := NewRegistry()
	reg.SetApplier(applier)
	reg.Register(mem.PatchSpec{ID: "god_mode", Library: "libtarget.so"})

	if err := reg.Toggle("god_mode", true); err != nil {
		t.Fatalf("Toggle enable returned error: %v", err)
	}
	if applier.appliedID != "god_mode" {
		t.Fatalf("appliedID = %q", applier.appliedID)
	}

	if err := reg.Toggle("god_mode", false); err != nil {
		t.Fatalf("Toggle disable returned error: %v", err)
	}
	if applier.restoredID != "god_mode" {
		t.Fatalf("restoredID = %q", applier.restoredID)
	}
}

type fakeApplier struct {
	appliedID  string
	restoredID string
}

func (f *fakeApplier) Apply(spec mem.PatchSpec) (mem.AppliedPatch, error) {
	f.appliedID = spec.ID
	return mem.AppliedPatch{Spec: spec, Original: []byte{0x07}, Length: 1}, nil
}

func (f *fakeApplier) Restore(applied mem.AppliedPatch) error {
	f.restoredID = applied.Spec.ID
	return nil
}
