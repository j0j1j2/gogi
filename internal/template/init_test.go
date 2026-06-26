package template

import (
	"os"
	"path/filepath"
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
}
