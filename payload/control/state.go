package control

import (
	"fmt"
	"sync"

	"gogi/payload/mem"
)

type Registry struct {
	mu      sync.RWMutex
	patches map[string]PatchRecord
}

type PatchRecord struct {
	Spec    mem.PatchSpec
	Enabled bool
}

type State struct {
	Patches map[string]PatchRecord `json:"patches"`
}

func NewRegistry() *Registry {
	return &Registry{patches: map[string]PatchRecord{}}
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
