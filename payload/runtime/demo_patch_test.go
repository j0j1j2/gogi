package runtime

import "testing"

func TestDemoPatchSpecParsesRVA(t *testing.T) {
	original := demoTargetValueRVAHex
	defer func() { demoTargetValueRVAHex = original }()

	demoTargetValueRVAHex = "0x1234"

	spec, ok := demoPatchSpec()
	if !ok {
		t.Fatal("expected demo patch")
	}
	if spec.RVA != 0x1234 {
		t.Fatalf("RVA = %#x", spec.RVA)
	}
	if spec.Library != "libtarget.so" {
		t.Fatalf("Library = %q", spec.Library)
	}
}

func TestDemoPatchSpecDisabledWithoutRVA(t *testing.T) {
	original := demoTargetValueRVAHex
	defer func() { demoTargetValueRVAHex = original }()

	demoTargetValueRVAHex = ""

	_, ok := demoPatchSpec()
	if ok {
		t.Fatal("expected no demo patch without injected RVA")
	}
}
