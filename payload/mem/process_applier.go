package mem

import (
	"fmt"
	"io"
	"os"
)

type ProcessApplier struct {
	hooks ProcessApplierHooks
}

type ProcessApplierHooks struct {
	MapsReader  io.Reader
	ApplyFunc   func(module Module, spec PatchSpec) (AppliedPatch, error)
	RestoreFunc func(applied AppliedPatch) error
}

func NewProcessApplier() *ProcessApplier {
	return NewProcessApplierWith(ProcessApplierHooks{})
}

func NewProcessApplierWith(hooks ProcessApplierHooks) *ProcessApplier {
	if hooks.ApplyFunc == nil {
		hooks.ApplyFunc = ApplyProcessPatch
	}
	if hooks.RestoreFunc == nil {
		hooks.RestoreFunc = RestoreProcessPatch
	}
	return &ProcessApplier{hooks: hooks}
}

func (a *ProcessApplier) Apply(spec PatchSpec) (AppliedPatch, error) {
	reader, closeReader, err := a.mapsReader()
	if err != nil {
		return AppliedPatch{}, err
	}
	if closeReader != nil {
		defer closeReader()
	}

	modules, err := ParseMaps(reader)
	if err != nil {
		return AppliedPatch{}, err
	}
	module, ok := FindModule(modules, spec.Library)
	if !ok {
		return AppliedPatch{}, fmt.Errorf("module %q not found", spec.Library)
	}
	return a.hooks.ApplyFunc(module, spec)
}

func (a *ProcessApplier) Restore(applied AppliedPatch) error {
	return a.hooks.RestoreFunc(applied)
}

func (a *ProcessApplier) mapsReader() (io.Reader, func(), error) {
	if a.hooks.MapsReader != nil {
		return a.hooks.MapsReader, nil, nil
	}
	file, err := os.Open("/proc/self/maps")
	if err != nil {
		return nil, nil, fmt.Errorf("open maps: %w", err)
	}
	return file, func() { _ = file.Close() }, nil
}
