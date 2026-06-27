package template

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInitProjectCreatesFrontendBackendAndConfig(t *testing.T) {
	root := t.TempDir()
	if err := InitProject(root, "sample"); err != nil {
		t.Fatalf("InitProject returned error: %v", err)
	}

	required := []string{
		"go.mod",
		"gogi.toml",
		"frontend/index.html",
		"frontend/style.css",
		"frontend/main.js",
		"backend/main.go",
	}
	for _, rel := range required {
		if _, err := os.Stat(filepath.Join(root, rel)); err != nil {
			t.Fatalf("expected %s to exist: %v", rel, err)
		}
	}

	if _, err := os.Stat(filepath.Join(root, "payload")); !os.IsNotExist(err) {
		t.Fatalf("init must not create user-editable payload directory, stat err=%v", err)
	}

	backend, err := os.ReadFile(filepath.Join(root, "backend", "main.go"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(backend), `"github.com/j0j1j2/gogi/sdk"`) {
		t.Fatalf("backend should import sdk for editor completion: %s", backend)
	}
	if !strings.Contains(string(backend), "func Init(ctx *sdk.Context)") {
		t.Fatalf("backend should use *sdk.Context: %s", backend)
	}
	if !strings.Contains(string(backend), `ctx.Action("give_coins"`) {
		t.Fatalf("backend should include action sample: %s", backend)
	}

	frontend, err := os.ReadFile(filepath.Join(root, "frontend", "main.js"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(frontend), "gogi.state()") || !strings.Contains(string(frontend), `gogi.action("give_coins"`) {
		t.Fatalf("frontend should use gogi client API: %s", frontend)
	}
}
