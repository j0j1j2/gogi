package cli

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCompileCommandRequiresNDK(t *testing.T) {
	t.Setenv("ANDROID_NDK_HOME", "")
	t.Setenv("ANDROID_NDK_ROOT", "")
	t.Setenv("ANDROID_HOME", "")
	t.Setenv("ANDROID_SDK_ROOT", "")
	var out bytes.Buffer
	var errOut bytes.Buffer

	code := Run([]string{"compile", "--abi", "arm64-v8a", "--api", "24"}, &out, &errOut)

	if code != 1 {
		t.Fatalf("expected code 1, got %d", code)
	}
	if !strings.Contains(errOut.String(), "ANDROID_NDK_HOME") {
		t.Fatalf("stderr should mention NDK, got %q", errOut.String())
	}
}

func TestCompileCommandRunsGoBuild(t *testing.T) {
	dir := t.TempDir()
	clang := filepath.Join(dir, "ndk", "toolchains", "llvm", "prebuilt", defaultHostTag(), "bin", "aarch64-linux-android24-clang")
	if err := os.MkdirAll(filepath.Dir(clang), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(clang, []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(oldWd) })

	t.Setenv("ANDROID_NDK_HOME", filepath.Join(dir, "ndk"))
	t.Setenv("ANDROID_NDK_ROOT", "")

	var gotName string
	var gotArgs []string
	var gotEnv map[string]string
	oldRunner := commandRunner
	commandRunner = func(name string, args []string, env map[string]string, stdout, stderr io.Writer) error {
		gotName = name
		gotArgs = append([]string(nil), args...)
		gotEnv = env
		return nil
	}
	t.Cleanup(func() { commandRunner = oldRunner })

	var out bytes.Buffer
	var errOut bytes.Buffer
	code := Run([]string{"compile", "--abi", "arm64-v8a", "--api", "24"}, &out, &errOut)

	if code != 0 {
		t.Fatalf("expected code 0, got %d, stderr=%q", code, errOut.String())
	}
	if gotName != "go" {
		t.Fatalf("command = %q", gotName)
	}
	joinedArgs := strings.Join(gotArgs, " ")
	if !strings.Contains(joinedArgs, "build -buildmode=c-shared -o dist/arm64-v8a/libgogi.so ./payload") {
		t.Fatalf("go args = %q", joinedArgs)
	}
	if gotEnv["GOOS"] != "android" || gotEnv["GOARCH"] != "arm64" || gotEnv["CGO_ENABLED"] != "1" {
		t.Fatalf("env = %#v", gotEnv)
	}
	if gotEnv["CC"] != clang {
		t.Fatalf("CC = %q, want %q", gotEnv["CC"], clang)
	}
	if !strings.Contains(out.String(), "built dist/arm64-v8a/libgogi.so") {
		t.Fatalf("stdout = %q", out.String())
	}
}
