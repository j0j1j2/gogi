package sdk

import "testing"

func TestContextRegistersPatch(t *testing.T) {
	ResetForTest()
	ctx := NewContext()

	ctx.RegisterPatch(Patch{
		ID:      "coins",
		Library: "libgame.so",
		RVA:     0x1234,
		Expect:  []byte{0x01},
		Replace: []byte{0x02},
	})

	patches := RegisteredPatches()
	if len(patches) != 1 {
		t.Fatalf("patch count = %d", len(patches))
	}
	if patches[0].ID != "coins" {
		t.Fatalf("patch id = %q", patches[0].ID)
	}
}

func TestContextRegistersAction(t *testing.T) {
	ResetForTest()
	ctx := NewContext()

	ctx.Action("give_coins", func(req ActionRequest) (any, error) {
		return map[string]any{"received": string(req.Payload)}, nil
	})

	actions := RegisteredActions()
	if len(actions) != 1 {
		t.Fatalf("action count = %d", len(actions))
	}
	if actions[0].ID != "give_coins" {
		t.Fatalf("action id = %q", actions[0].ID)
	}
	result, err := actions[0].Handler(ActionRequest{ID: "give_coins", Payload: []byte(`{"amount":10}`)})
	if err != nil {
		t.Fatal(err)
	}
	if result.(map[string]any)["received"] != `{"amount":10}` {
		t.Fatalf("result = %#v", result)
	}
}
