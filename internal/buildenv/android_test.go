package buildenv

import (
	"strings"
	"testing"
)

func TestResolveAndroidArm64(t *testing.T) {
	cfg, err := ResolveAndroid(map[string]string{
		"ANDROID_NDK_HOME": "/ndk",
	}, "arm64-v8a", 24, "darwin-arm64")
	if err != nil {
		t.Fatalf("ResolveAndroid returned error: %v", err)
	}
	if cfg.GoArch != "arm64" {
		t.Fatalf("GoArch = %q", cfg.GoArch)
	}
	if !strings.Contains(cfg.CC, "aarch64-linux-android24-clang") {
		t.Fatalf("CC path missing clang name: %q", cfg.CC)
	}
}

func TestResolveAndroidRequiresNDK(t *testing.T) {
	_, err := ResolveAndroid(map[string]string{}, "arm64-v8a", 24, "darwin-arm64")
	if err == nil {
		t.Fatal("expected missing NDK error")
	}
}
