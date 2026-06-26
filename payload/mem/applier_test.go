package mem

import "testing"

func TestApplyToSliceAndRestore(t *testing.T) {
	buf := []byte{0x00, 0x00, 0x80, 0x52}
	spec := PatchSpec{
		ID:      "sample",
		Expect:  []byte{0x00, 0x00, 0x80, 0x52},
		Replace: []byte{0x20, 0x00, 0x80, 0x52},
	}

	applied, err := ApplyToSlice(buf, 0, spec)
	if err != nil {
		t.Fatalf("ApplyToSlice returned error: %v", err)
	}
	if got := buf[0]; got != 0x20 {
		t.Fatalf("first byte = %#x", got)
	}
	if err := RestoreSlice(buf, applied); err != nil {
		t.Fatalf("RestoreSlice returned error: %v", err)
	}
	if got := buf[0]; got != 0x00 {
		t.Fatalf("restored first byte = %#x", got)
	}
}

func TestApplyToSliceRejectsExpectMismatch(t *testing.T) {
	buf := []byte{0xff}
	spec := PatchSpec{ID: "sample", Expect: []byte{0x00}, Replace: []byte{0x01}}
	if _, err := ApplyToSlice(buf, 0, spec); err == nil {
		t.Fatal("expected mismatch error")
	}
}

func TestRestoreSliceUsesOriginalOffset(t *testing.T) {
	buf := []byte{0xaa, 0x00, 0xbb}
	spec := PatchSpec{ID: "sample", Expect: []byte{0x00}, Replace: []byte{0x01}}

	applied, err := ApplyToSlice(buf, 1, spec)
	if err != nil {
		t.Fatalf("ApplyToSlice returned error: %v", err)
	}
	if err := RestoreSlice(buf, applied); err != nil {
		t.Fatalf("RestoreSlice returned error: %v", err)
	}
	if got := buf; string(got) != string([]byte{0xaa, 0x00, 0xbb}) {
		t.Fatalf("restored buffer = % x", got)
	}
}
