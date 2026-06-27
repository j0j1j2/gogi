package sdk

import "sync"

type Patch struct {
	ID      string
	Library string
	RVA     uintptr
	Expect  []byte
	Replace []byte
	Startup bool
}

var patchState = struct {
	sync.Mutex
	items []Patch
}{}

func (c *Context) RegisterPatch(patch Patch) {
	patchState.Lock()
	defer patchState.Unlock()
	patchState.items = append(patchState.items, patch)
}

func RegisteredPatches() []Patch {
	patchState.Lock()
	defer patchState.Unlock()
	out := make([]Patch, len(patchState.items))
	copy(out, patchState.items)
	return out
}

func ResetForTest() {
	patchState.Lock()
	defer patchState.Unlock()
	patchState.items = nil
	actionState.Lock()
	defer actionState.Unlock()
	actionState.items = nil
}
