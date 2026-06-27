package cli

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/j0j1j2/gogi/internal/devserver"
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
	if !bytes.Contains(out.Bytes(), []byte("gogi dev")) {
		t.Fatalf("help output missing dev usage: %q", out.String())
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

func TestRunDevStartsServer(t *testing.T) {
	dir := t.TempDir()
	oldCwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(oldCwd); err != nil {
			t.Fatal(err)
		}
	})
	if err := os.WriteFile(filepath.Join(dir, "gogi.toml"), []byte(`name = "sample"

[build]
package = "com.example.target"
abis = ["arm64-v8a"]
min_sdk = 24

[overlay]
enabled = true
mode = "webview"
width = 333
height = 444
collapsed_size = 55
draggable = true

[frontend]
entry = "ui/index.html"

[backend]
entry = "backend"
`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}

	oldDevServer := devServer
	oldDevBackend := devBackendStarter
	var got devserver.Options
	var backendRoot string
	devServer = func(opts devserver.Options) error {
		got = opts
		return nil
	}
	devBackendStarter = func(manifestPath string, stdout io.Writer, stderr io.Writer) (string, func(), error) {
		backendRoot = manifestPath
		return "http://127.0.0.1:19090", func() {}, nil
	}
	t.Cleanup(func() { devServer = oldDevServer })
	t.Cleanup(func() { devBackendStarter = oldDevBackend })

	var out bytes.Buffer
	var errOut bytes.Buffer
	code := Run([]string{"dev", "--addr", "127.0.0.1:18080", "--proxy", "http://127.0.0.1:17373"}, &out, &errOut)

	if code != 0 {
		t.Fatalf("expected code 0, got %d, stderr=%q", code, errOut.String())
	}
	if got.Addr != "127.0.0.1:18080" {
		t.Fatalf("addr = %q", got.Addr)
	}
	if got.Proxy != "http://127.0.0.1:17373" {
		t.Fatalf("proxy = %q", got.Proxy)
	}
	if backendRoot != "" {
		t.Fatalf("backend starter should not run when --proxy is supplied, got %q", backendRoot)
	}
	if got.FrontendDir != "ui" {
		t.Fatalf("frontend dir = %q", got.FrontendDir)
	}
	if got.Overlay.Width != 333 || got.Overlay.Height != 444 || got.Overlay.CollapsedSize != 55 || !got.Overlay.Draggable {
		t.Fatalf("overlay options = %#v", got.Overlay)
	}
}

func TestRunDevStartsBackendRunnerByDefault(t *testing.T) {
	dir := t.TempDir()
	oldCwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(oldCwd); err != nil {
			t.Fatal(err)
		}
	})
	if err := os.WriteFile(filepath.Join(dir, "gogi.toml"), []byte(`name = "sample"

[build]
package = "com.example.target"
abis = ["arm64-v8a"]
min_sdk = 24

[overlay]
enabled = true
mode = "webview"

[frontend]
entry = "frontend/index.html"

[backend]
entry = "backend"
`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}

	oldDevServer := devServer
	oldDevBackend := devBackendStarter
	var got devserver.Options
	cleaned := false
	devServer = func(opts devserver.Options) error {
		got = opts
		return nil
	}
	devBackendStarter = func(manifestPath string, stdout io.Writer, stderr io.Writer) (string, func(), error) {
		if manifestPath != "gogi.toml" {
			t.Fatalf("manifest path = %q", manifestPath)
		}
		return "http://127.0.0.1:19090", func() { cleaned = true }, nil
	}
	t.Cleanup(func() { devServer = oldDevServer })
	t.Cleanup(func() { devBackendStarter = oldDevBackend })

	var out bytes.Buffer
	var errOut bytes.Buffer
	code := Run([]string{"dev"}, &out, &errOut)

	if code != 0 {
		t.Fatalf("expected code 0, got %d, stderr=%q", code, errOut.String())
	}
	if got.Proxy != "http://127.0.0.1:19090" {
		t.Fatalf("proxy = %q", got.Proxy)
	}
	if !cleaned {
		t.Fatal("backend cleanup was not called")
	}
	if !bytes.Contains(out.Bytes(), []byte("backend connected")) {
		t.Fatalf("stdout missing backend connection message: %q", out.String())
	}
}

func TestRunDevRequiresGogiProject(t *testing.T) {
	dir := t.TempDir()
	oldCwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(oldCwd); err != nil {
			t.Fatal(err)
		}
	})
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}

	oldDevServer := devServer
	devServer = func(opts devserver.Options) error {
		t.Fatal("dev server should not start outside a gogi project")
		return nil
	}
	t.Cleanup(func() { devServer = oldDevServer })

	var out bytes.Buffer
	var errOut bytes.Buffer
	code := Run([]string{"dev"}, &out, &errOut)

	if code != 1 {
		t.Fatalf("expected code 1, got %d", code)
	}
	if !bytes.Contains(errOut.Bytes(), []byte("gogi.toml not found")) {
		t.Fatalf("stderr missing project error: %q", errOut.String())
	}
}
