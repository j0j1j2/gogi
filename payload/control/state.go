package control

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/j0j1j2/gogi/payload/mem"
)

type Registry struct {
	mu      sync.RWMutex
	patches map[string]PatchRecord
	actions map[string]ActionSpec
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
	Actions map[string]ActionInfo  `json:"actions"`
}

func NewRegistry() *Registry {
	return &Registry{patches: map[string]PatchRecord{}, actions: map[string]ActionSpec{}}
}

type ActionRequest struct {
	ID      string
	Payload json.RawMessage
}

type ActionHandler func(req ActionRequest) (any, error)

type ActionSpec struct {
	ID      string
	Handler ActionHandler `json:"-"`
}

type ActionInfo struct {
	ID string `json:"id"`
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

func (r *Registry) RegisterAction(spec ActionSpec) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.actions[spec.ID] = spec
}

func (r *Registry) DispatchAction(id string, payload json.RawMessage) (any, error) {
	r.mu.RLock()
	action, ok := r.actions[id]
	r.mu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("unknown action %q", id)
	}
	if action.Handler == nil {
		return nil, fmt.Errorf("action %q has no handler", id)
	}
	return action.Handler(ActionRequest{ID: id, Payload: payload})
}

func (r *Registry) Snapshot() State {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := State{Patches: map[string]PatchRecord{}, Actions: map[string]ActionInfo{}}
	for id, record := range r.patches {
		out.Patches[id] = record
	}
	for id := range r.actions {
		out.Actions[id] = ActionInfo{ID: id}
	}
	return out
}
