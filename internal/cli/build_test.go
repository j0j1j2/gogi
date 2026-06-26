package cli

import (
	"bytes"
	"strings"
	"testing"
)

func TestBuildCommandRequiresNDK(t *testing.T) {
	t.Setenv("ANDROID_NDK_HOME", "")
	t.Setenv("ANDROID_NDK_ROOT", "")
	var out bytes.Buffer
	var errOut bytes.Buffer

	code := Run([]string{"build", "--abi", "arm64-v8a", "--api", "24"}, &out, &errOut)

	if code != 1 {
		t.Fatalf("expected code 1, got %d", code)
	}
	if !strings.Contains(errOut.String(), "ANDROID_NDK_HOME") {
		t.Fatalf("stderr should mention NDK, got %q", errOut.String())
	}
}
