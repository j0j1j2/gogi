package mem

import (
	"strings"
	"testing"
)

func TestProcessApplierApplyResolvesModule(t *testing.T) {
	maps := `70000000-70001000 rw-p 00000000 fd:00 1 /data/app/lib/arm64/libtarget.so
`
	var gotModule Module
	var gotSpec PatchSpec
	applier := NewProcessApplierWith(ProcessApplierHooks{
		MapsReader: strings.NewReader(maps),
		ApplyFunc: func(module Module, spec PatchSpec) (AppliedPatch, error) {
			gotModule = module
			gotSpec = spec
			return AppliedPatch{Spec: spec, Address: module.Base + spec.RVA, Original: []byte{0x07}, Length: 1}, nil
		},
	})

	spec := PatchSpec{ID: "target_value", Library: "libtarget.so", RVA: 0x20}
	applied, err := applier.Apply(spec)
	if err != nil {
		t.Fatalf("Apply returned error: %v", err)
	}
	if gotModule.Name != "libtarget.so" {
		t.Fatalf("module = %q", gotModule.Name)
	}
	if gotSpec.ID != "target_value" {
		t.Fatalf("spec = %q", gotSpec.ID)
	}
	if applied.Address != 0x70000020 {
		t.Fatalf("applied address = %#x", applied.Address)
	}
}

func TestProcessApplierRejectsMissingModule(t *testing.T) {
	applier := NewProcessApplierWith(ProcessApplierHooks{
		MapsReader: strings.NewReader(""),
	})
	_, err := applier.Apply(PatchSpec{ID: "target_value", Library: "libtarget.so"})
	if err == nil {
		t.Fatal("expected missing module error")
	}
}
