package control

import (
	"fmt"
	"sync"

	"github.com/j0j1j2/gogi/payload/mem"
)

type Registry struct {
	mu      sync.RWMutex
	patches map[string]PatchRecord
	applier PatchApplier
}

type PatchApplier interface {
	Apply(spec mem.PatchSpec) (mem.AppliedPatch, error)
	Restore(applied mem.AppliedPatch) error
}

type PatchRecord struct {
	Spec    mem.PatchSpec
	Enabled bool
	Applied mem.AppliedPatch `json:"-"`
}

type State struct {
	Patches map[string]PatchRecord `json:"patches"`
}

func NewRegistry() *Registry {
	return &Registry{patches: map[string]PatchRecord{}}
}

func (r *Registry) SetApplier(applier PatchApplier) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.applier = applier
}

func (r *Registry) Register(spec mem.PatchSpec) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.patches[spec.ID] = PatchRecord{Spec: spec}
}

func (r *Registry) Toggle(id string, enabled bool) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	record, ok := r.patches[id]
	if !ok {
		return fmt.Errorf("unknown patch %q", id)
	}
	if r.applier != nil && record.Enabled != enabled {
		if enabled {
			applied, err := r.applier.Apply(record.Spec)
			if err != nil {
				return err
			}
			record.Applied = applied
		} else {
			if err := r.applier.Restore(record.Applied); err != nil {
				return err
			}
			record.Applied = mem.AppliedPatch{}
		}
	}
	record.Enabled = enabled
	r.patches[id] = record
	return nil
}

func (r *Registry) Snapshot() State {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := State{Patches: map[string]PatchRecord{}}
	for id, record := range r.patches {
		out.Patches[id] = record
	}
	return out
}
