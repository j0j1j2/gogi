package cli

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"
)

func TestRunHelp(t *testing.T) {
	var out bytes.Buffer
	var errOut bytes.Buffer

	code := Run([]string{"help"}, &out, &errOut)

	if code != 0 {
		t.Fatalf("expected exit code 0, got %d, stderr=%q", code, errOut.String())
	}
	if !bytes.Contains(out.Bytes(), []byte("gogi init <name>")) {
		t.Fatalf("help output missing init usage: %q", out.String())
	}
	if !bytes.Contains(out.Bytes(), []byte("gogi compile")) {
		t.Fatalf("help output missing compile usage: %q", out.String())
	}
	if !bytes.Contains(out.Bytes(), []byte("gogi version")) {
		t.Fatalf("help output missing version usage: %q", out.String())
	}
}

func TestRunVersion(t *testing.T) {
	var out bytes.Buffer
	var errOut bytes.Buffer

	code := Run([]string{"version"}, &out, &errOut)

	if code != 0 {
		t.Fatalf("expected exit code 0, got %d, stderr=%q", code, errOut.String())
	}
	if !bytes.Contains(out.Bytes(), []byte("gogi ")) {
		t.Fatalf("version output missing gogi version: %q", out.String())
	}
	if !bytes.Contains(out.Bytes(), []byte("commit ")) {
		t.Fatalf("version output missing commit: %q", out.String())
	}
}

func TestRunUnknownCommand(t *testing.T) {
	var out bytes.Buffer
	var errOut bytes.Buffer

	code := Run([]string{"missing"}, &out, &errOut)

	if code != 2 {
		t.Fatalf("expected exit code 2, got %d", code)
	}
	if !bytes.Contains(errOut.Bytes(), []byte("unknown command")) {
		t.Fatalf("stderr missing unknown command message: %q", errOut.String())
	}
}

func TestRunBuildRequiresTargetBundle(t *testing.T) {
	var out bytes.Buffer
	var errOut bytes.Buffer

	code := Run([]string{"build"}, &out, &errOut)

	if code != 2 {
		t.Fatalf("expected exit code 2, got %d", code)
	}
	if !bytes.Contains(errOut.Bytes(), []byte("usage: gogi build --apk <path>|--xapk <path>")) {
		t.Fatalf("stderr missing build usage: %q", errOut.String())
	}
}

func TestRunInitResolvesSDKDependency(t *testing.T) {
	dir := t.TempDir()
	projectDir := filepath.Join(dir, "sample")
	oldResolver := dependencyResolver
	var gotRoot string
	dependencyResolver = func(root string, stdout io.Writer, stderr io.Writer) error {
		gotRoot = root
		return nil
	}
	t.Cleanup(func() { dependencyResolver = oldResolver })

	var out bytes.Buffer
	var errOut bytes.Buffer
	code := Run([]string{"init", projectDir}, &out, &errOut)

	if code != 0 {
		t.Fatalf("expected code 0, got %d, stderr=%q", code, errOut.String())
	}
	if gotRoot != projectDir {
		t.Fatalf("dependency root = %q, want %q", gotRoot, projectDir)
	}
	backend, err := os.ReadFile(filepath.Join(projectDir, "backend", "main.go"))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(backend, []byte("github.com/j0j1j2/gogi/sdk")) {
		t.Fatalf("backend missing sdk import: %s", backend)
	}
}
