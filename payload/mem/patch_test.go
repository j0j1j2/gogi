package mem

import "testing"

func TestResolveAddress(t *testing.T) {
	mod := Module{Name: "libtarget.so", Base: 0x100000, End: 0x120000}
	got := ResolveAddress(mod, 0x1234)
	if got != 0x101234 {
		t.Fatalf("got %#x", got)
	}
}
