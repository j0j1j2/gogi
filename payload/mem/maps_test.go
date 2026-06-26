package mem

import (
	"strings"
	"testing"
)

func TestParseMaps(t *testing.T) {
	input := `7a00000000-7a00021000 r-xp 00000000 fd:00 123 /data/app/lib/arm64/libtarget.so
7a00021000-7a00023000 rw-p 00021000 fd:00 123 /data/app/lib/arm64/libtarget.so
`
	mods, err := ParseMaps(strings.NewReader(input))
	if err != nil {
		t.Fatalf("ParseMaps returned error: %v", err)
	}
	if len(mods) != 2 {
		t.Fatalf("got %d modules", len(mods))
	}
	if mods[0].Base != 0x7a00000000 || mods[0].End != 0x7a00021000 {
		t.Fatalf("bad range: %#x-%#x", mods[0].Base, mods[0].End)
	}
	if mods[0].Name != "libtarget.so" {
		t.Fatalf("name = %q", mods[0].Name)
	}
}
