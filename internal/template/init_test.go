package template

import (
	"os"
	"path/filepath"
	"testing"
)

func TestInitProjectCreatesManifestAndPayload(t *testing.T) {
	root := t.TempDir()
	if err := InitProject(root, "sample"); err != nil {
		t.Fatalf("InitProject returned error: %v", err)
	}

	required := []string{
		"gogi.toml",
		"payload/main.go",
		"payload/menu/assets/menu.html",
		"payload/menu/assets/menu.css",
		"payload/menu/assets/menu.js",
	}
	for _, rel := range required {
		if _, err := os.Stat(filepath.Join(root, rel)); err != nil {
			t.Fatalf("expected %s to exist: %v", rel, err)
		}
	}
}
