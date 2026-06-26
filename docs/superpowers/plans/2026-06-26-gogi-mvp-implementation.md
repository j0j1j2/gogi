# gogi MVP Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build the first usable `gogi` MVP: a Go CLI that creates and builds Go-based Android `c-shared` payload projects with manifest-driven memory patches and an HTML/CSS WebView menu foundation.

**Architecture:** The host-side `cmd/gogi` CLI manages project generation, TOML manifest validation, and Android Go/cgo build orchestration. The generated payload is a Go `package main` built with `-buildmode=c-shared`, exposing `JNI_OnLoad` and `ModInit`, with reusable packages for module resolution, patch state, and menu serving.

**Tech Stack:** Go 1.25+, Android NDK clang, `go build -buildmode=c-shared`, `github.com/BurntSushi/toml`, standard-library `net/http`, `embed`, `syscall`, and `testing`.

## Global Constraints

- Tool name is `gogi`.
- Payload logic stays in Go.
- Core MVP target ABI is `arm64-v8a`.
- Android payload build mode is `go build -buildmode=c-shared`.
- Primary entries are `JNI_OnLoad` and `ModInit`.
- Manifest format is TOML.
- Internal menu frontend uses embedded HTML/CSS/JS.
- WebView attachment must require an explicit Android `Context`/`Activity`; without one, return `context_required`.
- Current workspace is not a git repository. Commit steps are written for the execution phase after `git init` or after moving this plan into a repository.

---

## File Structure

Create this structure:

```text
cmd/gogi/main.go
internal/cli/cli.go
internal/project/manifest.go
internal/project/manifest_test.go
internal/project/hexbytes.go
internal/project/hexbytes_test.go
internal/buildenv/android.go
internal/buildenv/android_test.go
internal/template/init.go
internal/template/init_test.go
payload/main.go
payload/runtime/entry.go
payload/runtime/log_android.go
payload/mem/maps.go
payload/mem/maps_test.go
payload/mem/module.go
payload/mem/module_test.go
payload/mem/patch.go
payload/mem/patch_test.go
payload/mem/applier.go
payload/mem/applier_test.go
payload/mem/applier_android.go
payload/mem/applier_host.go
payload/control/state.go
payload/control/state_test.go
payload/menu/model.go
payload/menu/server.go
payload/menu/server_test.go
payload/menu/webview_backend.go
payload/menu/assets/menu.html
payload/menu/assets/menu.css
payload/menu/assets/menu.js
examples/basic_patch/gogi.toml
```

Responsibility map:

- `cmd/gogi/main.go`: tiny binary entrypoint.
- `internal/cli`: command routing for `init`, `validate`, and `build`.
- `internal/project`: TOML manifest model, hex byte parsing, validation.
- `internal/buildenv`: Android NDK/ABI/API compiler resolution.
- `internal/template`: project skeleton generation for `gogi init`.
- `payload/runtime`: exported entries and startup lifecycle.
- `payload/mem`: maps parsing, module resolution, patch calculations and writes.
- `payload/control`: patch registry and toggle state.
- `payload/menu`: menu model, local HTTP API, WebView backend boundary.
- `payload/menu/assets`: embedded HTML/CSS/JS menu frontend.

---

### Task 1: Initialize Go Module and CLI Skeleton

**Files:**
- Create: `go.mod`
- Create: `cmd/gogi/main.go`
- Create: `internal/cli/cli.go`
- Test: `internal/cli/cli_test.go`

**Interfaces:**
- Produces: `cli.Run(args []string, stdout io.Writer, stderr io.Writer) int`
- Produces: CLI commands `init`, `validate`, `build`, `help`

- [ ] **Step 1: Write the failing CLI tests**

Create `internal/cli/cli_test.go`:

```go
package cli

import (
	"bytes"
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
```

- [ ] **Step 2: Run tests and verify they fail**

Run: `go test ./internal/cli -run TestRun -v`

Expected: fail because package/files do not exist.

- [ ] **Step 3: Add module and minimal CLI**

Create `go.mod`:

```go
module gogi

go 1.25

require github.com/BurntSushi/toml v1.5.0
```

Create `internal/cli/cli.go`:

```go
package cli

import (
	"fmt"
	"io"
)

func Run(args []string, stdout io.Writer, stderr io.Writer) int {
	if len(args) == 0 {
		printHelp(stdout)
		return 0
	}

	switch args[0] {
	case "help", "-h", "--help":
		printHelp(stdout)
		return 0
	case "init":
		fmt.Fprintln(stderr, "init command is unavailable until project templates are added")
		return 1
	case "validate":
		fmt.Fprintln(stderr, "validate command is unavailable until manifest support is added")
		return 1
	case "build":
		fmt.Fprintln(stderr, "build command is unavailable until android build support is added")
		return 1
	default:
		fmt.Fprintf(stderr, "unknown command %q\n", args[0])
		printHelp(stderr)
		return 2
	}
}

func printHelp(w io.Writer) {
	fmt.Fprintln(w, "gogi - Go-based Android injectable .so builder")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Usage:")
	fmt.Fprintln(w, "  gogi init <name>")
	fmt.Fprintln(w, "  gogi validate [manifest]")
	fmt.Fprintln(w, "  gogi build [--abi arm64-v8a] [--api 24] [--menu webview]")
}
```

Create `cmd/gogi/main.go`:

```go
package main

import (
	"os"

	"gogi/internal/cli"
)

func main() {
	os.Exit(cli.Run(os.Args[1:], os.Stdout, os.Stderr))
}
```

- [ ] **Step 4: Run tests and verify they pass**

Run: `go test ./internal/cli -run TestRun -v`

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add go.mod cmd/gogi/main.go internal/cli/cli.go internal/cli/cli_test.go
git commit -m "feat: add gogi cli skeleton"
```

---

### Task 2: Manifest Model, Hex Parser, and Validation

**Files:**
- Create: `internal/project/manifest.go`
- Create: `internal/project/manifest_test.go`
- Create: `internal/project/hexbytes.go`
- Create: `internal/project/hexbytes_test.go`
- Modify: `internal/cli/cli.go`

**Interfaces:**
- Consumes: `cli.Run(args []string, stdout io.Writer, stderr io.Writer) int`
- Produces: `project.LoadManifest(path string) (*Manifest, error)`
- Produces: `project.ParseHexBytes(input string) ([]byte, error)`
- Produces: `(*Manifest).Validate() error`

- [ ] **Step 1: Write failing hex parser tests**

Create `internal/project/hexbytes_test.go`:

```go
package project

import "testing"

func TestParseHexBytes(t *testing.T) {
	got, err := ParseHexBytes("00 01 aa FF")
	if err != nil {
		t.Fatalf("ParseHexBytes returned error: %v", err)
	}
	want := []byte{0x00, 0x01, 0xaa, 0xff}
	if string(got) != string(want) {
		t.Fatalf("got % x, want % x", got, want)
	}
}

func TestParseHexBytesRejectsOddToken(t *testing.T) {
	_, err := ParseHexBytes("0")
	if err == nil {
		t.Fatal("expected error for odd-length token")
	}
}
```

- [ ] **Step 2: Write failing manifest tests**

Create `internal/project/manifest_test.go`:

```go
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
```

- [ ] **Step 3: Run tests and verify they fail**

Run: `go test ./internal/project -v`

Expected: fail because functions and types are undefined.

- [ ] **Step 4: Implement hex parser**

Create `internal/project/hexbytes.go`:

```go
package project

import (
	"encoding/hex"
	"fmt"
	"strings"
)

func ParseHexBytes(input string) ([]byte, error) {
	clean := strings.ReplaceAll(strings.TrimSpace(input), " ", "")
	if clean == "" {
		return nil, nil
	}
	if len(clean)%2 != 0 {
		return nil, fmt.Errorf("hex byte string has odd length")
	}
	out, err := hex.DecodeString(clean)
	if err != nil {
		return nil, fmt.Errorf("decode hex bytes: %w", err)
	}
	return out, nil
}
```

- [ ] **Step 5: Implement manifest model and validation**

Create `internal/project/manifest.go`:

```go
package project

import (
	"fmt"
	"strings"

	"github.com/BurntSushi/toml"
)

type Manifest struct {
	Name        string       `toml:"name"`
	Package     string       `toml:"package"`
	ABIs        []string     `toml:"abis"`
	API         int          `toml:"api"`
	Entry       []string     `toml:"entry"`
	MenuBackend string       `toml:"menu_backend"`
	Patches     []Patch      `toml:"patches"`
	Menu        MenuConfig   `toml:"menu"`
}

type MenuConfig struct {
	Toggles []MenuToggle `toml:"toggles"`
}

type Patch struct {
	ID      string `toml:"id"`
	Library string `toml:"library"`
	RVA     string `toml:"rva"`
	Expect  string `toml:"expect"`
	Replace string `toml:"replace"`
	Startup bool   `toml:"startup"`
}

type MenuToggle struct {
	ID      string `toml:"id"`
	Label   string `toml:"label"`
	Patch   string `toml:"patch"`
	Initial bool   `toml:"initial"`
}

func LoadManifest(path string) (*Manifest, error) {
	var manifest Manifest
	if _, err := toml.DecodeFile(path, &manifest); err != nil {
		return nil, fmt.Errorf("load manifest %s: %w", path, err)
	}
	return &manifest, nil
}

func (m *Manifest) Validate() error {
	if strings.TrimSpace(m.Name) == "" {
		return fmt.Errorf("name is required")
	}
	if strings.TrimSpace(m.Package) == "" {
		return fmt.Errorf("package is required")
	}
	if len(m.ABIs) == 0 {
		return fmt.Errorf("at least one ABI is required")
	}
	if m.API == 0 {
		return fmt.Errorf("api is required")
	}

	patchIDs := map[string]bool{}
	for _, patch := range m.Patches {
		if patch.ID == "" {
			return fmt.Errorf("patch id is required")
		}
		if patchIDs[patch.ID] {
			return fmt.Errorf("duplicate patch id %q", patch.ID)
		}
		patchIDs[patch.ID] = true
		if patch.Library == "" {
			return fmt.Errorf("patch %q library is required", patch.ID)
		}
		if patch.RVA == "" {
			return fmt.Errorf("patch %q rva is required", patch.ID)
		}
		expect, err := ParseHexBytes(patch.Expect)
		if err != nil {
			return fmt.Errorf("patch %q expect: %w", patch.ID, err)
		}
		replace, err := ParseHexBytes(patch.Replace)
		if err != nil {
			return fmt.Errorf("patch %q replace: %w", patch.ID, err)
		}
		if len(expect) != 0 && len(replace) != 0 && len(expect) != len(replace) {
			return fmt.Errorf("patch %q expect and replace lengths differ", patch.ID)
		}
	}

	for _, toggle := range m.Menu.Toggles {
		if toggle.ID == "" {
			return fmt.Errorf("menu toggle id is required")
		}
		if toggle.Label == "" {
			return fmt.Errorf("menu toggle %q label is required", toggle.ID)
		}
		if !patchIDs[toggle.Patch] {
			return fmt.Errorf("menu toggle %q references unknown patch %q", toggle.ID, toggle.Patch)
		}
	}

	return nil
}
```

- [ ] **Step 6: Wire `gogi validate`**

Modify `internal/cli/cli.go` so `validate` loads `gogi.toml` by default:

```go
case "validate":
	path := "gogi.toml"
	if len(args) > 1 {
		path = args[1]
	}
	manifest, err := project.LoadManifest(path)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	if err := manifest.Validate(); err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	fmt.Fprintf(stdout, "%s is valid\n", path)
	return 0
```

Also add import:

```go
import "gogi/internal/project"
```

- [ ] **Step 7: Run tests and verify they pass**

Run: `go test ./internal/project ./internal/cli -v`

Expected: PASS.

- [ ] **Step 8: Commit**

```bash
git add go.mod go.sum internal/project internal/cli
git commit -m "feat: add manifest validation"
```

---

### Task 3: Android Build Environment Resolver

**Files:**
- Create: `internal/buildenv/android.go`
- Create: `internal/buildenv/android_test.go`

**Interfaces:**
- Produces: `buildenv.ResolveAndroid(env map[string]string, abi string, api int, hostTag string) (AndroidConfig, error)`
- Produces: `AndroidConfig.GoArch string`
- Produces: `AndroidConfig.CC string`

- [ ] **Step 1: Write failing resolver tests**

Create `internal/buildenv/android_test.go`:

```go
package buildenv

import (
	"strings"
	"testing"
)

func TestResolveAndroidArm64(t *testing.T) {
	cfg, err := ResolveAndroid(map[string]string{
		"ANDROID_NDK_HOME": "/ndk",
	}, "arm64-v8a", 24, "darwin-arm64")
	if err != nil {
		t.Fatalf("ResolveAndroid returned error: %v", err)
	}
	if cfg.GoArch != "arm64" {
		t.Fatalf("GoArch = %q", cfg.GoArch)
	}
	if !strings.Contains(cfg.CC, "aarch64-linux-android24-clang") {
		t.Fatalf("CC path missing clang name: %q", cfg.CC)
	}
}

func TestResolveAndroidRequiresNDK(t *testing.T) {
	_, err := ResolveAndroid(map[string]string{}, "arm64-v8a", 24, "darwin-arm64")
	if err == nil {
		t.Fatal("expected missing NDK error")
	}
}
```

- [ ] **Step 2: Run tests and verify they fail**

Run: `go test ./internal/buildenv -v`

Expected: fail because package/files do not exist.

- [ ] **Step 3: Implement resolver**

Create `internal/buildenv/android.go`:

```go
package buildenv

import (
	"fmt"
	"path/filepath"
)

type AndroidConfig struct {
	NDKHome string
	ABI     string
	API     int
	HostTag string
	GoOS    string
	GoArch  string
	CC      string
}

func ResolveAndroid(env map[string]string, abi string, api int, hostTag string) (AndroidConfig, error) {
	ndk := env["ANDROID_NDK_HOME"]
	if ndk == "" {
		ndk = env["ANDROID_NDK_ROOT"]
	}
	if ndk == "" {
		return AndroidConfig{}, fmt.Errorf("ANDROID_NDK_HOME or ANDROID_NDK_ROOT is required")
	}
	if api <= 0 {
		return AndroidConfig{}, fmt.Errorf("android api must be positive")
	}

	var goarch string
	var clang string
	switch abi {
	case "arm64-v8a":
		goarch = "arm64"
		clang = fmt.Sprintf("aarch64-linux-android%d-clang", api)
	default:
		return AndroidConfig{}, fmt.Errorf("unsupported abi %q", abi)
	}

	cc := filepath.Join(ndk, "toolchains", "llvm", "prebuilt", hostTag, "bin", clang)
	return AndroidConfig{
		NDKHome: ndk,
		ABI:     abi,
		API:     api,
		HostTag: hostTag,
		GoOS:    "android",
		GoArch:  goarch,
		CC:      cc,
	}, nil
}
```

- [ ] **Step 4: Run tests and verify they pass**

Run: `go test ./internal/buildenv -v`

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/buildenv
git commit -m "feat: resolve android build environment"
```

---

### Task 4: Project Template Generation

**Files:**
- Create: `internal/template/init.go`
- Create: `internal/template/init_test.go`
- Modify: `internal/cli/cli.go`
- Create: `examples/basic_patch/gogi.toml`

**Interfaces:**
- Consumes: `project.Manifest`
- Produces: `template.InitProject(root string, name string) error`
- CLI: `gogi init <name>`

- [ ] **Step 1: Write failing template test**

Create `internal/template/init_test.go`:

```go
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
```

- [ ] **Step 2: Run tests and verify they fail**

Run: `go test ./internal/template -v`

Expected: fail because `InitProject` is undefined.

- [ ] **Step 3: Implement template generation**

Create `internal/template/init.go`:

```go
package template

import (
	"fmt"
	"os"
	"path/filepath"
)

func InitProject(root string, name string) error {
	files := map[string]string{
		"gogi.toml": manifestTemplate(name),
		"payload/main.go": payloadMainTemplate(),
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
```

- [ ] **Step 4: Wire `gogi init <name>`**

Modify the `init` case in `internal/cli/cli.go`:

```go
case "init":
	if len(args) != 2 {
		fmt.Fprintln(stderr, "usage: gogi init <name>")
		return 2
	}
	if err := gogitemplate.InitProject(args[1], args[1]); err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	fmt.Fprintf(stdout, "created %s\n", args[1])
	return 0
```

Add import:

```go
import gogitemplate "gogi/internal/template"
```

Use `gogitemplate.InitProject` in code to avoid naming conflict with the Go keyword-adjacent concept of templates.

- [ ] **Step 5: Add example manifest**

Create `examples/basic_patch/gogi.toml` using the same content as `manifestTemplate("basic_patch")`.

- [ ] **Step 6: Run tests**

Run: `go test ./internal/template ./internal/cli -v`

Expected: PASS.

- [ ] **Step 7: Commit**

```bash
git add internal/template internal/cli examples/basic_patch/gogi.toml
git commit -m "feat: generate gogi project templates"
```

---

### Task 5: Payload Runtime Entries and Logging Boundary

**Files:**
- Create: `payload/main.go`
- Create: `payload/runtime/entry.go`
- Create: `payload/runtime/log_android.go`

**Interfaces:**
- Produces: exported `ModInit()`
- Produces: exported `JNI_OnLoad(vm unsafe.Pointer, reserved unsafe.Pointer) C.jint`
- Produces: `runtime.Start(vm unsafe.Pointer)`
- Produces: `runtime.Logf(format string, args ...any)`

- [ ] **Step 1: Add payload runtime files**

Create `payload/runtime/entry.go`:

```go
package runtime

import (
	"sync"
	"unsafe"
)

var startOnce sync.Once
var capturedVM unsafe.Pointer

func Start(vm unsafe.Pointer) {
	startOnce.Do(func() {
		capturedVM = vm
		Logf("gogi runtime started")
		go func() {
			Logf("gogi worker started")
		}()
	})
}

func CapturedVM() unsafe.Pointer {
	return capturedVM
}
```

Create `payload/runtime/log_android.go`:

```go
package runtime

/*
#include <android/log.h>
#include <stdlib.h>
static void gogi_log(const char* msg) {
	__android_log_write(3, "gogi", msg);
}
*/
import "C"

import (
	"fmt"
	"unsafe"
)

func Logf(format string, args ...any) {
	msg := C.CString(fmt.Sprintf(format, args...))
	defer C.free(unsafe.Pointer(msg))
	C.gogi_log(msg)
}
```

Create `payload/main.go`:

```go
package main

/*
#include <jni.h>
*/
import "C"

import (
	"unsafe"

	gogiruntime "gogi/payload/runtime"
)

//export ModInit
func ModInit() {
	gogiruntime.Start(nil)
}

//export JNI_OnLoad
func JNI_OnLoad(vm unsafe.Pointer, reserved unsafe.Pointer) C.jint {
	gogiruntime.Start(vm)
	return C.JNI_VERSION_1_6
}

func main() {}
```

- [ ] **Step 2: Verify host package compilation**

Run: `go test ./payload/runtime -v`

Expected: may fail on non-Android host because `android/log.h` is unavailable.

- [ ] **Step 3: Split Android logging behind build tags if host test fails**

Rename `payload/runtime/log_android.go` header to include:

```go
//go:build android
```

Create `payload/runtime/log_host.go`:

```go
//go:build !android

package runtime

import "log"

func Logf(format string, args ...any) {
	log.Printf(format, args...)
}
```

- [ ] **Step 4: Verify package compilation**

Run: `go test ./payload/runtime -v`

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add payload/main.go payload/runtime
git commit -m "feat: add go payload entries"
```

---

### Task 6: Maps Parser and Module Resolver

**Files:**
- Create: `payload/mem/maps.go`
- Create: `payload/mem/maps_test.go`
- Create: `payload/mem/module.go`
- Create: `payload/mem/module_test.go`

**Interfaces:**
- Produces: `mem.ParseMaps(r io.Reader) ([]Module, error)`
- Produces: `mem.FindModule(mods []Module, name string) (Module, bool)`
- Produces: `Module.Contains(addr uintptr) bool`

- [ ] **Step 1: Write failing maps tests**

Create `payload/mem/maps_test.go`:

```go
package mem

import (
	"strings"
	"testing"
)

func TestParseMaps(t *testing.T) {
	input := `7a00000000-7a00021000 r-xp 00000000 fd:00 123 /data/app/lib/arm64/libtarget.so
7a00021000-7a00023000 rw-p 00021000 fd:00 123 /data/app/lib/arm64/libtarget.so
`
	mods, err := ParseMaps(strings.NewReader(input))
	if err != nil {
		t.Fatalf("ParseMaps returned error: %v", err)
	}
	if len(mods) != 2 {
		t.Fatalf("got %d modules", len(mods))
	}
	if mods[0].Base != 0x7a00000000 || mods[0].End != 0x7a00021000 {
		t.Fatalf("bad range: %#x-%#x", mods[0].Base, mods[0].End)
	}
	if mods[0].Name != "libtarget.so" {
		t.Fatalf("name = %q", mods[0].Name)
	}
}
```

- [ ] **Step 2: Write failing module tests**

Create `payload/mem/module_test.go`:

```go
package mem

import "testing"

func TestFindModule(t *testing.T) {
	mods := []Module{{Name: "libtarget.so", Base: 0x1000, End: 0x2000}}
	mod, ok := FindModule(mods, "libtarget.so")
	if !ok {
		t.Fatal("expected module")
	}
	if !mod.Contains(0x1800) {
		t.Fatal("expected address to be contained")
	}
}
```

- [ ] **Step 3: Run tests and verify they fail**

Run: `go test ./payload/mem -run 'Test(ParseMaps|FindModule)' -v`

Expected: fail because package/files do not exist.

- [ ] **Step 4: Implement parser and resolver**

Create `payload/mem/maps.go`:

```go
package mem

import (
	"bufio"
	"fmt"
	"io"
	"path/filepath"
	"strconv"
	"strings"
)

type Module struct {
	Name  string
	Base  uintptr
	End   uintptr
	Path  string
	Perms string
}

func ParseMaps(r io.Reader) ([]Module, error) {
	var mods []Module
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) < 6 {
			continue
		}
		bounds := strings.Split(fields[0], "-")
		if len(bounds) != 2 {
			continue
		}
		base, err := strconv.ParseUint(bounds[0], 16, 64)
		if err != nil {
			return nil, fmt.Errorf("parse base %q: %w", bounds[0], err)
		}
		end, err := strconv.ParseUint(bounds[1], 16, 64)
		if err != nil {
			return nil, fmt.Errorf("parse end %q: %w", bounds[1], err)
		}
		path := fields[5]
		mods = append(mods, Module{
			Name:  filepath.Base(path),
			Base:  uintptr(base),
			End:   uintptr(end),
			Path:  path,
			Perms: fields[1],
		})
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan maps: %w", err)
	}
	return mods, nil
}
```

Create `payload/mem/module.go`:

```go
package mem

func FindModule(mods []Module, name string) (Module, bool) {
	for _, mod := range mods {
		if mod.Name == name || mod.Path == name {
			return mod, true
		}
	}
	return Module{}, false
}

func (m Module) Contains(addr uintptr) bool {
	return addr >= m.Base && addr < m.End
}
```

- [ ] **Step 5: Run tests and verify they pass**

Run: `go test ./payload/mem -v`

Expected: PASS.

- [ ] **Step 6: Commit**

```bash
git add payload/mem/maps.go payload/mem/maps_test.go payload/mem/module.go payload/mem/module_test.go
git commit -m "feat: parse android process maps"
```

---

### Task 7: Patch Runtime and Control State

**Files:**
- Create: `payload/mem/patch.go`
- Create: `payload/mem/patch_test.go`
- Create: `payload/control/state.go`
- Create: `payload/control/state_test.go`

**Interfaces:**
- Produces: `mem.PatchSpec`
- Produces: `mem.ResolveAddress(module Module, rva uintptr) uintptr`
- Produces: `control.Registry`
- Produces: `(*Registry).Register(spec PatchSpec)`
- Produces: `(*Registry).Toggle(id string, enabled bool) error`
- Produces: `(*Registry).Snapshot() State`

- [ ] **Step 1: Write failing patch tests**

Create `payload/mem/patch_test.go`:

```go
package mem

import "testing"

func TestResolveAddress(t *testing.T) {
	mod := Module{Name: "libtarget.so", Base: 0x100000, End: 0x120000}
	got := ResolveAddress(mod, 0x1234)
	if got != 0x101234 {
		t.Fatalf("got %#x", got)
	}
}
```

- [ ] **Step 2: Write failing control tests**

Create `payload/control/state_test.go`:

```go
package control

import (
	"testing"

	"gogi/payload/mem"
)

func TestRegistryToggleState(t *testing.T) {
	reg := NewRegistry()
	reg.Register(mem.PatchSpec{ID: "god_mode", Library: "libtarget.so"})

	if err := reg.Toggle("god_mode", true); err != nil {
		t.Fatalf("Toggle returned error: %v", err)
	}

	state := reg.Snapshot()
	if !state.Patches["god_mode"].Enabled {
		t.Fatalf("expected god_mode enabled")
	}
}

func TestRegistryRejectsUnknownPatch(t *testing.T) {
	reg := NewRegistry()
	if err := reg.Toggle("missing", true); err == nil {
		t.Fatal("expected unknown patch error")
	}
}
```

- [ ] **Step 3: Run tests and verify they fail**

Run: `go test ./payload/mem ./payload/control -v`

Expected: fail because new types/functions are undefined.

- [ ] **Step 4: Implement patch spec and address calculation**

Create `payload/mem/patch.go`:

```go
package mem

type PatchSpec struct {
	ID      string
	Library string
	RVA     uintptr
	Expect  []byte
	Replace []byte
	Startup bool
}

func ResolveAddress(module Module, rva uintptr) uintptr {
	return module.Base + rva
}
```

- [ ] **Step 5: Implement control registry**

Create `payload/control/state.go`:

```go
package control

import (
	"fmt"
	"sync"

	"gogi/payload/mem"
)

type Registry struct {
	mu      sync.RWMutex
	patches map[string]PatchRecord
}

type PatchRecord struct {
	Spec    mem.PatchSpec
	Enabled bool
}

type State struct {
	Patches map[string]PatchRecord `json:"patches"`
}

func NewRegistry() *Registry {
	return &Registry{patches: map[string]PatchRecord{}}
}

func (r *Registry) Register(spec mem.PatchSpec) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.patches[spec.ID] = PatchRecord{Spec: spec}
}

func (r *Registry) Toggle(id string, enabled bool) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	record, ok := r.patches[id]
	if !ok {
		return fmt.Errorf("unknown patch %q", id)
	}
	record.Enabled = enabled
	r.patches[id] = record
	return nil
}

func (r *Registry) Snapshot() State {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := State{Patches: map[string]PatchRecord{}}
	for id, record := range r.patches {
		out.Patches[id] = record
	}
	return out
}
```

- [ ] **Step 6: Run tests and verify they pass**

Run: `go test ./payload/mem ./payload/control -v`

Expected: PASS.

- [ ] **Step 7: Commit**

```bash
git add payload/mem/patch.go payload/mem/patch_test.go payload/control
git commit -m "feat: add patch control state"
```

---

### Task 8: Patch Apply and Restore Runtime

**Files:**
- Create: `payload/mem/applier.go`
- Create: `payload/mem/applier_test.go`
- Create: `payload/mem/applier_android.go`
- Create: `payload/mem/applier_host.go`

**Interfaces:**
- Consumes: `mem.PatchSpec`
- Consumes: `mem.Module`
- Produces: `mem.AppliedPatch`
- Produces: `mem.ApplyToSlice(buf []byte, offset int, spec PatchSpec) (AppliedPatch, error)`
- Produces: `mem.RestoreSlice(buf []byte, applied AppliedPatch) error`
- Produces: `mem.ApplyProcessPatch(module Module, spec PatchSpec) (AppliedPatch, error)`
- Produces: `mem.RestoreProcessPatch(applied AppliedPatch) error`

- [ ] **Step 1: Write failing apply/restore tests**

Create `payload/mem/applier_test.go`:

```go
package mem

import "testing"

func TestApplyToSliceAndRestore(t *testing.T) {
	buf := []byte{0x00, 0x00, 0x80, 0x52}
	spec := PatchSpec{
		ID:      "sample",
		Expect:  []byte{0x00, 0x00, 0x80, 0x52},
		Replace: []byte{0x20, 0x00, 0x80, 0x52},
	}

	applied, err := ApplyToSlice(buf, 0, spec)
	if err != nil {
		t.Fatalf("ApplyToSlice returned error: %v", err)
	}
	if got := buf[0]; got != 0x20 {
		t.Fatalf("first byte = %#x", got)
	}
	if err := RestoreSlice(buf, applied); err != nil {
		t.Fatalf("RestoreSlice returned error: %v", err)
	}
	if got := buf[0]; got != 0x00 {
		t.Fatalf("restored first byte = %#x", got)
	}
}

func TestApplyToSliceRejectsExpectMismatch(t *testing.T) {
	buf := []byte{0xff}
	spec := PatchSpec{ID: "sample", Expect: []byte{0x00}, Replace: []byte{0x01}}
	if _, err := ApplyToSlice(buf, 0, spec); err == nil {
		t.Fatal("expected mismatch error")
	}
}
```

- [ ] **Step 2: Run tests and verify they fail**

Run: `go test ./payload/mem -run 'TestApplyToSlice|TestApplyToSliceRejectsExpectMismatch' -v`

Expected: fail because applier functions are undefined.

- [ ] **Step 3: Implement shared applier logic**

Create `payload/mem/applier.go`:

```go
package mem

import (
	"bytes"
	"fmt"
)

type AppliedPatch struct {
	Spec     PatchSpec
	Address  uintptr
	Original []byte
	Length   int
}

func ApplyToSlice(buf []byte, offset int, spec PatchSpec) (AppliedPatch, error) {
	if offset < 0 || offset+len(spec.Replace) > len(buf) {
		return AppliedPatch{}, fmt.Errorf("patch %q range is outside target buffer", spec.ID)
	}
	current := buf[offset : offset+len(spec.Replace)]
	if len(spec.Expect) > 0 && !bytes.Equal(current[:len(spec.Expect)], spec.Expect) {
		return AppliedPatch{}, fmt.Errorf("patch %q expected bytes mismatch", spec.ID)
	}
	original := append([]byte(nil), current...)
	copy(current, spec.Replace)
	return AppliedPatch{
		Spec:     spec,
		Original: original,
		Length:   len(spec.Replace),
	}, nil
}

func RestoreSlice(buf []byte, applied AppliedPatch) error {
	if len(applied.Original) == 0 {
		return fmt.Errorf("patch %q has no original bytes", applied.Spec.ID)
	}
	if len(buf) < len(applied.Original) {
		return fmt.Errorf("patch %q restore buffer too small", applied.Spec.ID)
	}
	copy(buf[:len(applied.Original)], applied.Original)
	return nil
}
```

- [ ] **Step 4: Implement host process applier boundary**

Create `payload/mem/applier_host.go`:

```go
//go:build !android

package mem

import "fmt"

func ApplyProcessPatch(module Module, spec PatchSpec) (AppliedPatch, error) {
	return AppliedPatch{}, fmt.Errorf("process patching requires android build")
}

func RestoreProcessPatch(applied AppliedPatch) error {
	return fmt.Errorf("process patch restore requires android build")
}
```

- [ ] **Step 5: Implement Android process applier**

Create `payload/mem/applier_android.go`:

```go
//go:build android

package mem

/*
#include <sys/mman.h>
static void gogi_clear_cache(char* start, char* end) {
	__builtin___clear_cache(start, end);
}
*/
import "C"

import (
	"bytes"
	"fmt"
	"os"
	"syscall"
	"unsafe"
)

func ApplyProcessPatch(module Module, spec PatchSpec) (AppliedPatch, error) {
	if len(spec.Replace) == 0 {
		return AppliedPatch{}, fmt.Errorf("patch %q replacement is empty", spec.ID)
	}
	addr := ResolveAddress(module, spec.RVA)
	target := unsafe.Slice((*byte)(unsafe.Pointer(addr)), len(spec.Replace))
	if len(spec.Expect) > 0 && !bytes.Equal(target[:len(spec.Expect)], spec.Expect) {
		return AppliedPatch{}, fmt.Errorf("patch %q expected bytes mismatch", spec.ID)
	}
	original := append([]byte(nil), target...)
	if err := withWritablePage(addr, len(spec.Replace), func() {
		copy(target, spec.Replace)
	}); err != nil {
		return AppliedPatch{}, err
	}
	C.gogi_clear_cache((*C.char)(unsafe.Pointer(addr)), (*C.char)(unsafe.Pointer(addr+uintptr(len(spec.Replace)))))
	return AppliedPatch{Spec: spec, Address: addr, Original: original, Length: len(spec.Replace)}, nil
}

func RestoreProcessPatch(applied AppliedPatch) error {
	if applied.Address == 0 {
		return fmt.Errorf("patch %q has no process address", applied.Spec.ID)
	}
	target := unsafe.Slice((*byte)(unsafe.Pointer(applied.Address)), len(applied.Original))
	return withWritablePage(applied.Address, len(applied.Original), func() {
		copy(target, applied.Original)
	})
}

func withWritablePage(addr uintptr, length int, fn func()) error {
	pageSize := uintptr(os.Getpagesize())
	pageStart := addr & ^(pageSize - 1)
	pageEnd := (addr + uintptr(length) + pageSize - 1) & ^(pageSize - 1)
	pageLen := pageEnd - pageStart
	if _, _, errno := syscall.RawSyscall(syscall.SYS_MPROTECT, pageStart, pageLen, syscall.PROT_READ|syscall.PROT_WRITE|syscall.PROT_EXEC); errno != 0 {
		return errno
	}
	fn()
	_, _, _ = syscall.RawSyscall(syscall.SYS_MPROTECT, pageStart, pageLen, syscall.PROT_READ|syscall.PROT_EXEC)
	return nil
}
```

- [ ] **Step 6: Run tests and verify they pass**

Run: `go test ./payload/mem -v`

Expected: PASS on host. Android-only process patching compiles during the Android payload build.

- [ ] **Step 7: Commit**

```bash
git add payload/mem/applier.go payload/mem/applier_test.go payload/mem/applier_android.go payload/mem/applier_host.go
git commit -m "feat: apply and restore memory patches"
```

---

### Task 9: Embedded HTML/CSS Menu Server

**Files:**
- Create: `payload/menu/model.go`
- Create: `payload/menu/server.go`
- Create: `payload/menu/server_test.go`
- Create: `payload/menu/assets/menu.html`
- Create: `payload/menu/assets/menu.css`
- Create: `payload/menu/assets/menu.js`

**Interfaces:**
- Consumes: `control.Registry`
- Produces: `menu.NewServer(registry *control.Registry) *Server`
- Produces: `(*Server).Handler() http.Handler`
- Produces HTTP `GET /api/state`
- Produces HTTP `POST /api/toggle/{id}`

- [ ] **Step 1: Write failing menu server tests**

Create `payload/menu/server_test.go`:

```go
package menu

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"gogi/payload/control"
	"gogi/payload/mem"
)

func TestStateEndpoint(t *testing.T) {
	reg := control.NewRegistry()
	reg.Register(mem.PatchSpec{ID: "god_mode", Library: "libtarget.so"})
	server := NewServer(reg)

	req := httptest.NewRequest(http.MethodGet, "/api/state", nil)
	rec := httptest.NewRecorder()
	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "god_mode") {
		t.Fatalf("body missing patch: %s", rec.Body.String())
	}
}

func TestToggleEndpoint(t *testing.T) {
	reg := control.NewRegistry()
	reg.Register(mem.PatchSpec{ID: "god_mode", Library: "libtarget.so"})
	server := NewServer(reg)

	req := httptest.NewRequest(http.MethodPost, "/api/toggle/god_mode?enabled=true", nil)
	rec := httptest.NewRecorder()
	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	if !reg.Snapshot().Patches["god_mode"].Enabled {
		t.Fatal("expected patch enabled")
	}
}
```

- [ ] **Step 2: Run tests and verify they fail**

Run: `go test ./payload/menu -v`

Expected: fail because menu package is undefined.

- [ ] **Step 3: Add assets**

Create `payload/menu/assets/menu.html`:

```html
<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <link rel="stylesheet" href="/menu.css">
  <title>gogi</title>
</head>
<body>
  <main class="menu">
    <header class="menu__header">
      <strong>gogi</strong>
      <span id="status">connecting</span>
    </header>
    <section id="toggles" class="menu__toggles"></section>
  </main>
  <script src="/menu.js"></script>
</body>
</html>
```

Create `payload/menu/assets/menu.css`:

```css
:root {
  color-scheme: dark;
  --bg: rgba(15, 17, 19, 0.82);
  --fg: #f7f1e4;
  --muted: #b7afa0;
  --accent: #e6bc5c;
}

* {
  box-sizing: border-box;
}

body {
  margin: 0;
  font-family: ui-sans-serif, system-ui, sans-serif;
  background: transparent;
  color: var(--fg);
}

.menu {
  width: min(320px, calc(100vw - 24px));
  margin: 12px;
  padding: 12px;
  background: var(--bg);
  border: 1px solid rgba(247, 241, 228, 0.18);
  border-radius: 8px;
}

.menu__header {
  display: flex;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 12px;
}

.menu__toggles {
  display: grid;
  gap: 8px;
}

button {
  min-height: 44px;
  border: 0;
  border-radius: 6px;
  background: rgba(247, 241, 228, 0.12);
  color: var(--fg);
}

button[aria-pressed="true"] {
  background: var(--accent);
  color: #1b1608;
}
```

Create `payload/menu/assets/menu.js`:

```js
async function refresh() {
  const response = await fetch('/api/state');
  const state = await response.json();
  document.getElementById('status').textContent = 'ready';
  const root = document.getElementById('toggles');
  root.innerHTML = '';
  Object.entries(state.patches || {}).forEach(([id, record]) => {
    const button = document.createElement('button');
    button.type = 'button';
    button.textContent = id;
    button.setAttribute('aria-pressed', record.Enabled ? 'true' : 'false');
    button.addEventListener('click', async () => {
      const next = !record.Enabled;
      await fetch(`/api/toggle/${encodeURIComponent(id)}?enabled=${next}`, { method: 'POST' });
      await refresh();
    });
    root.appendChild(button);
  });
}

refresh().catch(error => {
  document.getElementById('status').textContent = error.message;
});
```

- [ ] **Step 4: Implement server**

Create `payload/menu/model.go`:

```go
package menu

type Toggle struct {
	ID      string `json:"id"`
	Label   string `json:"label"`
	Initial bool   `json:"initial"`
}
```

Create `payload/menu/server.go`:

```go
package menu

import (
	"embed"
	"encoding/json"
	"net/http"
	"strings"

	"gogi/payload/control"
)

//go:embed assets/menu.html assets/menu.css assets/menu.js
var assets embed.FS

type Server struct {
	registry *control.Registry
}

func NewServer(registry *control.Registry) *Server {
	return &Server{registry: registry}
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/state", s.handleState)
	mux.HandleFunc("/api/toggle/", s.handleToggle)
	mux.HandleFunc("/", s.handleAsset)
	return mux
}

func (s *Server) handleState(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")
	_ = json.NewEncoder(w).Encode(s.registry.Snapshot())
}

func (s *Server) handleToggle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	id := strings.TrimPrefix(r.URL.Path, "/api/toggle/")
	enabled := r.URL.Query().Get("enabled") == "true"
	if err := s.registry.Toggle(id, enabled); err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	s.handleState(w, r)
}

func (s *Server) handleAsset(w http.ResponseWriter, r *http.Request) {
	name := "menu.html"
	if r.URL.Path == "/menu.css" {
		name = "menu.css"
		w.Header().Set("content-type", "text/css")
	}
	if r.URL.Path == "/menu.js" {
		name = "menu.js"
		w.Header().Set("content-type", "application/javascript")
	}
	data, err := assets.ReadFile("assets/" + name)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	_, _ = w.Write(data)
}
```

- [ ] **Step 5: Run tests and verify they pass**

Run: `go test ./payload/menu ./payload/control -v`

Expected: PASS.

- [ ] **Step 6: Commit**

```bash
git add payload/menu
git commit -m "feat: serve embedded web menu"
```

---

### Task 10: WebView Backend Boundary

**Files:**
- Create: `payload/menu/webview_backend.go`
- Create: `payload/menu/webview_backend_test.go`

**Interfaces:**
- Produces: `menu.ErrContextRequired`
- Produces: `menu.WebViewBackend`
- Produces: `(*WebViewBackend).Attach(ctx unsafe.Pointer, url string) error`

- [ ] **Step 1: Write failing backend boundary test**

Create `payload/menu/webview_backend_test.go`:

```go
package menu

import (
	"errors"
	"testing"
)

func TestWebViewBackendRequiresContext(t *testing.T) {
	backend := NewWebViewBackend()
	err := backend.Attach(nil, "http://127.0.0.1:12345/")
	if !errors.Is(err, ErrContextRequired) {
		t.Fatalf("expected ErrContextRequired, got %v", err)
	}
}
```

- [ ] **Step 2: Run tests and verify they fail**

Run: `go test ./payload/menu -run TestWebViewBackendRequiresContext -v`

Expected: fail because backend is undefined.

- [ ] **Step 3: Implement explicit context boundary**

Create `payload/menu/webview_backend.go`:

```go
package menu

import (
	"errors"
	"unsafe"
)

var ErrContextRequired = errors.New("context_required")

type WebViewBackend struct{}

func NewWebViewBackend() *WebViewBackend {
	return &WebViewBackend{}
}

func (b *WebViewBackend) Attach(ctx unsafe.Pointer, url string) error {
	if ctx == nil {
		return ErrContextRequired
	}
	return nil
}
```

- [ ] **Step 4: Run tests and verify they pass**

Run: `go test ./payload/menu -run TestWebViewBackendRequiresContext -v`

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add payload/menu/webview_backend.go payload/menu/webview_backend_test.go
git commit -m "feat: define webview backend boundary"
```

---

### Task 11: Build Command

**Files:**
- Modify: `internal/cli/cli.go`
- Create: `internal/cli/build_test.go`

**Interfaces:**
- Consumes: `project.LoadManifest`
- Consumes: `buildenv.ResolveAndroid`
- Produces: `gogi build --abi arm64-v8a --api 24 --menu webview`

- [ ] **Step 1: Write failing build command test**

Create `internal/cli/build_test.go`:

```go
package cli

import (
	"bytes"
	"strings"
	"testing"
)

func TestBuildCommandRequiresNDK(t *testing.T) {
	var out bytes.Buffer
	var errOut bytes.Buffer

	code := Run([]string{"build", "--abi", "arm64-v8a", "--api", "24"}, &out, &errOut)

	if code != 1 {
		t.Fatalf("expected code 1, got %d", code)
	}
	if !strings.Contains(errOut.String(), "ANDROID_NDK_HOME") {
		t.Fatalf("stderr should mention NDK, got %q", errOut.String())
	}
}
```

- [ ] **Step 2: Run test and verify it fails**

Run: `go test ./internal/cli -run TestBuildCommandRequiresNDK -v`

Expected: fail because build command still returns the old unavailable-command message.

- [ ] **Step 3: Implement argument parsing and NDK validation**

Modify the `build` case in `internal/cli/cli.go`:

```go
case "build":
	abi := "arm64-v8a"
	api := 24
	for i := 1; i < len(args); i++ {
		switch args[i] {
		case "--abi":
			i++
			if i >= len(args) {
				fmt.Fprintln(stderr, "--abi requires a value")
				return 2
			}
			abi = args[i]
		case "--api":
			i++
			if i >= len(args) {
				fmt.Fprintln(stderr, "--api requires a value")
				return 2
			}
			parsed, err := strconv.Atoi(args[i])
			if err != nil {
				fmt.Fprintf(stderr, "invalid --api %q\n", args[i])
				return 2
			}
			api = parsed
		case "--menu":
			i++
			if i >= len(args) {
				fmt.Fprintln(stderr, "--menu requires a value")
				return 2
			}
		default:
			fmt.Fprintf(stderr, "unknown build flag %q\n", args[i])
			return 2
		}
	}
	env := map[string]string{
		"ANDROID_NDK_HOME": os.Getenv("ANDROID_NDK_HOME"),
		"ANDROID_NDK_ROOT": os.Getenv("ANDROID_NDK_ROOT"),
	}
	cfg, err := buildenv.ResolveAndroid(env, abi, api, defaultHostTag())
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	fmt.Fprintf(stdout, "GOOS=%s GOARCH=%s CGO_ENABLED=1 CC=%s go build -buildmode=c-shared -o dist/%s/libgogi.so ./payload\n", cfg.GoOS, cfg.GoArch, cfg.CC, cfg.ABI)
	return 0
```

Add imports:

```go
import (
	"os"
	"runtime"
	"strconv"

	"gogi/internal/buildenv"
)
```

Add helper in `internal/cli/cli.go`:

```go
func defaultHostTag() string {
	switch runtime.GOOS {
	case "darwin":
		if runtime.GOARCH == "arm64" {
			return "darwin-arm64"
		}
		return "darwin-x86_64"
	case "linux":
		return "linux-x86_64"
	default:
		return runtime.GOOS + "-" + runtime.GOARCH
	}
}
```

- [ ] **Step 4: Run CLI tests**

Run: `go test ./internal/cli ./internal/buildenv -v`

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/cli internal/buildenv
git commit -m "feat: add android build command"
```

---

### Task 12: End-to-End Verification

**Files:**
- Modify: `README.md`
- Test: all packages

**Interfaces:**
- Consumes: every package from previous tasks.
- Produces: documented MVP workflow.

- [ ] **Step 1: Create README with exact MVP commands**

Create `README.md`:

```markdown
# gogi

`gogi` builds Go-based Android injectable shared libraries for game mod tooling.

## MVP workflow

```bash
gogi init sample
cd sample
gogi validate
gogi build --abi arm64-v8a --api 24 --menu webview
```

The payload is built with:

```bash
GOOS=android GOARCH=arm64 CGO_ENABLED=1 CC=<ndk-clang> go build -buildmode=c-shared -o dist/arm64-v8a/libgogi.so ./payload
```

The WebView menu frontend is HTML/CSS/JS served by the Go payload over a local HTTP API.
```

- [ ] **Step 2: Run full test suite**

Run: `go test ./...`

Expected: PASS.

- [ ] **Step 3: Run CLI help**

Run: `go run ./cmd/gogi help`

Expected output includes:

```text
gogi init <name>
gogi validate [manifest]
gogi build [--abi arm64-v8a] [--api 24] [--menu webview]
```

- [ ] **Step 4: Run manifest validation against example**

Run: `go run ./cmd/gogi validate examples/basic_patch/gogi.toml`

Expected:

```text
examples/basic_patch/gogi.toml is valid
```

- [ ] **Step 5: Commit**

```bash
git add README.md
git commit -m "docs: document gogi mvp workflow"
```

---

## Self-Review

Spec coverage:

- `gogi` name: covered in Task 1 and README.
- Go `c-shared` payload: covered in Task 5 and Task 11.
- `JNI_OnLoad` and `ModInit`: covered in Task 5.
- TOML manifest: covered in Task 2.
- `arm64-v8a` build config: covered in Task 3 and Task 11.
- `/proc/self/maps` parser and module resolver: covered in Task 6.
- RVA patch address model: covered in Task 7.
- RVA patch apply/restore runtime: covered in Task 8.
- Control layer: covered in Task 7.
- Embedded HTML/CSS/JS menu: covered in Task 9.
- WebView explicit context boundary: covered in Task 10.
- End-to-end verification: covered in Task 12.

Deferred from MVP:

- Actual Android WebView object construction is not implemented in this first plan; Task 10 establishes the explicit `context_required` boundary required by the design. A second plan should add JNI calls for Activity-provided WebView attachment.
- Pattern scanning is not implemented in this first plan.
