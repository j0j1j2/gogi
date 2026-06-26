package runtime

import "testing"

func TestDemoPatchSpecParsesRVA(t *testing.T) {
	original := demoTargetValueRVAHex
	defer func() { demoTargetValueRVAHex = original }()

	demoTargetValueRVAHex = "0x1234"

	spec := demoPatchSpec()
	if spec.RVA != 0x1234 {
		t.Fatalf("RVA = %#x", spec.RVA)
	}
	if spec.Library != "libtarget.so" {
		t.Fatalf("Library = %q", spec.Library)
	}
}
