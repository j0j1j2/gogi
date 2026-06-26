package buildenv

import (
	"os"
	"path/filepath"
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

func TestResolveAndroidFindsInstalledNDKFromSDK(t *testing.T) {
	dir := t.TempDir()
	clangPath := filepath.Join(dir, "ndk", "28.2.13676358", "toolchains", "llvm", "prebuilt", "darwin-arm64", "bin", "aarch64-linux-android24-clang")
	if err := os.MkdirAll(filepath.Dir(clangPath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(clangPath, []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatal(err)
	}

	cfg, err := ResolveAndroid(map[string]string{
		"ANDROID_HOME": dir,
	}, "arm64-v8a", 24, "darwin-arm64")
	if err != nil {
		t.Fatalf("ResolveAndroid returned error: %v", err)
	}
	if cfg.NDKHome != filepath.Join(dir, "ndk", "28.2.13676358") {
		t.Fatalf("NDKHome = %q", cfg.NDKHome)
	}
	if cfg.CC != clangPath {
		t.Fatalf("CC = %q, want %q", cfg.CC, clangPath)
	}
}

func TestResolveAndroidFallsBackToInstalledHostTag(t *testing.T) {
	dir := t.TempDir()
	clangPath := filepath.Join(dir, "toolchains", "llvm", "prebuilt", "darwin-x86_64", "bin", "aarch64-linux-android24-clang")
	if err := os.MkdirAll(filepath.Dir(clangPath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(clangPath, []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatal(err)
	}

	cfg, err := ResolveAndroid(map[string]string{
		"ANDROID_NDK_HOME": dir,
	}, "arm64-v8a", 24, "darwin-arm64")
	if err != nil {
		t.Fatalf("ResolveAndroid returned error: %v", err)
	}
	if cfg.HostTag != "darwin-x86_64" {
		t.Fatalf("HostTag = %q", cfg.HostTag)
	}
	if cfg.CC != clangPath {
		t.Fatalf("CC = %q, want %q", cfg.CC, clangPath)
	}
}
