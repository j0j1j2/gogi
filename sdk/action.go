package sdk

import "sync"

type ActionRequest struct {
	ID      string
	Payload []byte
}

type ActionHandler func(req ActionRequest) (any, error)

type ActionSpec struct {
	ID      string
	Handler ActionHandler
}

var actionState = struct {
	sync.Mutex
	items []ActionSpec
}{}

func (c *Context) Action(id string, handler ActionHandler) {
	actionState.Lock()
	defer actionState.Unlock()
	actionState.items = append(actionState.items, ActionSpec{ID: id, Handler: handler})
}

func RegisteredActions() []ActionSpec {
	actionState.Lock()
	defer actionState.Unlock()
	out := make([]ActionSpec, len(actionState.items))
	copy(out, actionState.items)
	return out
}
