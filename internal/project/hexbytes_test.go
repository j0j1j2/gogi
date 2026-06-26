package project

import "testing"

func TestParseHexBytes(t *testing.T) {
	got, err := ParseHexBytes("00 01 aa FF")
	if err != nil {
		t.Fatalf("ParseHexBytes returned error: %v", err)
	}
	want := []byte{0x00, 0x01, 0xaa, 0xff}
	if string(got) != string(want) {
		t.Fatalf("got % x, want % x", got, want)
	}
}

func TestParseHexBytesRejectsOddToken(t *testing.T) {
	_, err := ParseHexBytes("0")
	if err == nil {
		t.Fatal("expected error for odd-length token")
	}
}
