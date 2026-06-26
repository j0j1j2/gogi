package project

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestLoadProjectConfigAndValidate(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "gogi.toml")
	data := `
name = "sample"

[build]
package = "com.example.target"
abis = ["arm64-v8a"]
min_sdk = 24

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
	if manifest.Build.Package != "com.example.target" {
		t.Fatalf("build package = %q", manifest.Build.Package)
	}
	if manifest.Frontend.Entry != "frontend/index.html" {
		t.Fatalf("frontend entry = %q", manifest.Frontend.Entry)
	}
	if err := manifest.Validate(); err != nil {
		t.Fatalf("Validate returned error: %v", err)
	}
}

func TestManifestSchemaDoesNotExposePatchConfig(t *testing.T) {
	manifestType := reflect.TypeOf(Manifest{})
	if _, ok := manifestType.FieldByName("Patches"); ok {
		t.Fatal("Manifest must not expose Patches; patch behavior belongs in backend Go code")
	}
	if _, ok := manifestType.FieldByName("Menu"); ok {
		t.Fatal("Manifest must not expose Menu toggles; UI actions belong in backend Go code")
	}
}

func TestManifestRejectsMissingBackendEntry(t *testing.T) {
	manifest := &Manifest{
		Name: "sample",
		Build: BuildConfig{
			Package: "com.example.target",
			ABIs:    []string{"arm64-v8a"},
			MinSDK:  24,
		},
		Overlay: OverlayConfig{
			Enabled: true,
			Mode:    "webview",
		},
		Frontend: FrontendConfig{
			Entry: "frontend/index.html",
		},
	}

	err := manifest.Validate()
	if err == nil {
		t.Fatal("expected validation error")
	}
	if err.Error() != "backend entry is required" {
		t.Fatalf("expected backend entry error, got %v", err)
	}
}
