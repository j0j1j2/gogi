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
