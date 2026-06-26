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
	if err := os.Mkdir("payload", 0o755); err != nil {
		t.Fatal(err)
	}

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

func TestCompileCommandUsesModulePayloadOutsideRepo(t *testing.T) {
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

	var gotArgs []string
	oldRunner := commandRunner
	commandRunner = func(name string, args []string, env map[string]string, stdout, stderr io.Writer) error {
		gotArgs = append([]string(nil), args...)
		return nil
	}
	t.Cleanup(func() { commandRunner = oldRunner })

	var out bytes.Buffer
	var errOut bytes.Buffer
	code := Run([]string{"compile"}, &out, &errOut)

	if code != 0 {
		t.Fatalf("expected code 0, got %d, stderr=%q", code, errOut.String())
	}
	if gotArgs[len(gotArgs)-1] != "github.com/j0j1j2/gogi/payload" {
		t.Fatalf("payload package = %q", gotArgs[len(gotArgs)-1])
	}
}

func TestCompileCommandUsesManifestBuildDefaults(t *testing.T) {
	dir := t.TempDir()
	clang := filepath.Join(dir, "ndk", "toolchains", "llvm", "prebuilt", defaultHostTag(), "bin", "aarch64-linux-android26-clang")
	if err := os.MkdirAll(filepath.Dir(clang), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(clang, []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	config := `name = "sample"

[build]
package = "com.example.target"
abis = ["arm64-v8a"]
min_sdk = 26

[overlay]
enabled = true
mode = "webview"
width = 320
height = 420
collapsed_size = 56
draggable = true

[frontend]
entry = "frontend/index.html"

[backend]
entry = "backend"
`
	if err := os.WriteFile(filepath.Join(dir, "gogi.toml"), []byte(config), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module sample\n\ngo 1.25\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(dir, "frontend"), 0o755); err != nil {
		t.Fatal(err)
	}
	for name, body := range map[string]string{
		"index.html": "<main>sample</main>",
		"style.css":  "body{}",
		"main.js":    "console.log('sample')",
	} {
		if err := os.WriteFile(filepath.Join(dir, "frontend", name), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
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

	var gotEnv map[string]string
	var gotArgs []string
	oldRunner := commandRunner
	commandRunner = func(name string, args []string, env map[string]string, stdout, stderr io.Writer) error {
		if len(args) > 0 && args[0] == "build" {
			gotEnv = env
			gotArgs = append([]string(nil), args...)
		}
		return nil
	}
	t.Cleanup(func() { commandRunner = oldRunner })

	var out bytes.Buffer
	var errOut bytes.Buffer
	code := Run([]string{"compile"}, &out, &errOut)

	if code != 0 {
		t.Fatalf("expected code 0, got %d, stderr=%q", code, errOut.String())
	}
	if gotEnv["CC"] != clang {
		t.Fatalf("CC = %q, want %q", gotEnv["CC"], clang)
	}
	joinedArgs := strings.Join(gotArgs, " ")
	if !strings.Contains(joinedArgs, "github.com/j0j1j2/gogi/payload/runtime.overlayWidth=320") {
		t.Fatalf("ldflags missing overlay width: %q", joinedArgs)
	}
	if !strings.Contains(joinedArgs, "github.com/j0j1j2/gogi/payload/runtime.overlayDraggable=true") {
		t.Fatalf("ldflags missing overlay draggable: %q", joinedArgs)
	}
}
