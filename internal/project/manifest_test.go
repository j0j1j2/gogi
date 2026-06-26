package project

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadManifestAndValidate(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "gogi.toml")
	data := `
name = "sample"
package = "com.example.target"
abis = ["arm64-v8a"]
api = 24
entry = ["jni_onload", "modinit"]
menu_backend = "webview"

[[patches]]
id = "god_mode"
library = "libtarget.so"
rva = "0x1234"
expect = "00 00 80 52"
replace = "20 00 80 52"
startup = false

[[menu.toggles]]
id = "god_mode"
label = "God Mode"
patch = "god_mode"
initial = false
`
	if err := os.WriteFile(path, []byte(data), 0o644); err != nil {
		t.Fatal(err)
	}

	manifest, err := LoadManifest(path)
	if err != nil {
		t.Fatalf("LoadManifest returned error: %v", err)
	}
	if manifest.Name != "sample" {
		t.Fatalf("manifest name = %q", manifest.Name)
	}
	if err := manifest.Validate(); err != nil {
		t.Fatalf("Validate returned error: %v", err)
	}
}

func TestManifestRejectsMissingPatchReference(t *testing.T) {
	manifest := &Manifest{
		Name:    "sample",
		Package: "com.example.target",
		ABIs:    []string{"arm64-v8a"},
		API:     24,
		Menu: MenuConfig{
			Toggles: []MenuToggle{
				{ID: "missing", Label: "Missing", Patch: "no_patch"},
			},
		},
	}

	err := manifest.Validate()
	if err == nil {
		t.Fatal("expected validation error")
	}
	if !strings.Contains(err.Error(), "unknown patch") {
		t.Fatalf("error should mention unknown patch, got %v", err)
	}
}
