package mem

import "testing"

func TestFindModule(t *testing.T) {
	mods := []Module{{Name: "libtarget.so", Base: 0x1000, End: 0x2000}}
	mod, ok := FindModule(mods, "libtarget.so")
	if !ok {
		t.Fatal("expected module")
	}
	if !mod.Contains(0x1800) {
		t.Fatal("expected address to be contained")
	}
}
