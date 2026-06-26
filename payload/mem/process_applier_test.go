package mem

import (
	"strings"
	"syscall"
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

func TestProcessApplierUsesTargetMappingPermissions(t *testing.T) {
	maps := `10000000-10001000 r--p 00000000 fd:00 1 /data/app/lib/arm64/libtarget.so
10001000-10002000 r-xp 00001000 fd:00 1 /data/app/lib/arm64/libtarget.so
10008000-10009000 rw-p 00008000 fd:00 1 /data/app/lib/arm64/libtarget.so
`
	var gotModule Module
	applier := NewProcessApplierWith(ProcessApplierHooks{
		MapsReader: strings.NewReader(maps),
		ApplyFunc: func(module Module, spec PatchSpec) (AppliedPatch, error) {
			gotModule = module
			return AppliedPatch{Spec: spec, Address: module.Base + spec.RVA, Original: []byte{0x07}, Length: 1}, nil
		},
	})

	_, err := applier.Apply(PatchSpec{ID: "target_value", Library: "libtarget.so", RVA: 0x8008})
	if err != nil {
		t.Fatalf("Apply returned error: %v", err)
	}
	if gotModule.Base != 0x10000000 {
		t.Fatalf("base module changed to %#x", gotModule.Base)
	}
	if gotModule.Perms != "rw-p" {
		t.Fatalf("target mapping perms = %q", gotModule.Perms)
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

func TestProtFromPermsAddsWriteWithoutAddingExec(t *testing.T) {
	prot, err := protFromPerms("r--p", true)
	if err != nil {
		t.Fatalf("protFromPerms returned error: %v", err)
	}
	if prot != syscall.PROT_READ|syscall.PROT_WRITE {
		t.Fatalf("prot = %#x", prot)
	}

	prot, err = protFromPerms("rw-p", false)
	if err != nil {
		t.Fatalf("protFromPerms returned error: %v", err)
	}
	if prot != syscall.PROT_READ|syscall.PROT_WRITE {
		t.Fatalf("restore prot = %#x", prot)
	}
}
