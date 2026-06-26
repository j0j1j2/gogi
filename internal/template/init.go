package template

import (
	"fmt"
	"os"
	"path/filepath"
)

func InitProject(root string, name string) error {
	files := map[string]string{
		"gogi.toml":                     manifestTemplate(name),
		"payload/main.go":               payloadMainTemplate(),
		"payload/menu/assets/menu.html": "<!doctype html><html><head><link rel=\"stylesheet\" href=\"/menu.css\"></head><body><main id=\"app\"></main><script src=\"/menu.js\"></script></body></html>\n",
		"payload/menu/assets/menu.css":  "body{margin:0;font-family:sans-serif;background:rgba(18,18,18,.78);color:#f5f1e8}button{min-height:44px}\n",
		"payload/menu/assets/menu.js":   "fetch('/api/state').then(r=>r.json()).then(s=>{document.getElementById('app').textContent=JSON.stringify(s)})\n",
	}

	for rel, content := range files {
		path := filepath.Join(root, rel)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return fmt.Errorf("create directory for %s: %w", rel, err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			return fmt.Errorf("write %s: %w", rel, err)
		}
	}
	return nil
}

func manifestTemplate(name string) string {
	return fmt.Sprintf(`name = %q
package = "com.example.target"
abis = ["arm64-v8a"]
api = 24
entry = ["jni_onload", "modinit"]
menu_backend = "webview"

[[patches]]
id = "sample_patch"
library = "libtarget.so"
rva = "0x0"
expect = "00"
replace = "00"
startup = false

[[menu.toggles]]
id = "sample_patch"
label = "Sample Patch"
patch = "sample_patch"
initial = false
`, name)
}

func payloadMainTemplate() string {
	return `package main

import "C"

//export ModInit
func ModInit() {}

//export JNI_OnLoad
func JNI_OnLoad(vm uintptr, reserved uintptr) int {
	return 0x00010006
}

func main() {}
`
}
