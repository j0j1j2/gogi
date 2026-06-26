# gogi Design

Date: 2026-06-26

## Purpose

`gogi` is a Go-based builder for Android injectable shared libraries used in CTF and controlled research environments. The payload is written primarily in Go and built as an Android `.so` with `go build -buildmode=c-shared`.

The first-class use case is an in-process Android native payload that can:

- locate loaded modules in the current process
- apply and restore memory patches
- expose patch state through a small control layer
- provide an optional internal mod menu overlay
- build repeatable ABI-specific artifacts from a manifest

## Design Principles

- Keep payload logic in Go.
- Keep C/C++ out of the core payload path.
- Treat Android-specific UI and loader details as replaceable backends.
- Make patch behavior explicit and reproducible through manifests.
- Start with `arm64-v8a`; add other ABIs after the runtime shape is stable.
- Prefer small, testable packages over one large payload file.

## Product Name

The tool name is `gogi`.

Terminology:

- `gogi builder`: the host-side Go CLI
- `gogi payload`: the Android `.so` built from Go
- `gogi runtime`: the payload packages that run inside the target process
- `gogi menu`: optional internal overlay menu extension

## High-Level Architecture

```text
gogi/
  cmd/gogi/                  # CLI entrypoint
  internal/project/          # project creation and manifest validation
  internal/build/            # Android Go/cgo/NDK build orchestration
  internal/template/         # payload project templates
  internal/package/          # dist layout and metadata

  payload/
    main.go                  # c-shared package main, exported entries
    runtime/
      entry.go               # JNI_OnLoad, ModInit, startup lifecycle
      scheduler.go           # delayed startup and module wait loop
      log_android.go         # logcat wrapper
    mem/
      maps.go                # /proc/self/maps parser
      module.go              # module resolver and base/range model
      patch.go               # mprotect/write/restore/verify
      pattern.go             # AOB pattern scanner
    control/
      state.go               # patch registry and active state
      command.go             # apply/restore/toggle commands
    menu/
      model.go               # menu model: toggles, buttons, values
      backend.go             # menu backend interface
      server.go              # embedded HTML/CSS/JS and local API
      webview_backend.go     # Android WebView overlay backend
      assets/
        menu.html
        menu.css
        menu.js

  examples/
    basic_patch/
    webview_menu/

  dist/
    arm64-v8a/
      libgogi.so
      libgogi.h
    build-info.json
    patches.resolved.json
```

## Build Model

`gogi` builds the payload with Go's `c-shared` build mode.

```bash
GOOS=android \
GOARCH=arm64 \
CGO_ENABLED=1 \
CC="$ANDROID_NDK_HOME/toolchains/llvm/prebuilt/<host>/bin/aarch64-linux-android24-clang" \
go build -buildmode=c-shared -o dist/arm64-v8a/libgogi.so ./payload
```

The builder is responsible for:

- resolving Android NDK paths
- selecting the correct clang for ABI/API level
- setting `GOOS`, `GOARCH`, `CGO_ENABLED`, and `CC`
- embedding manifest/menu assets into the payload
- placing output artifacts in `dist/`
- emitting build metadata for reproducibility

## Payload Entry Points

The payload exposes two primary entries:

```go
//export ModInit
func ModInit()

//export JNI_OnLoad
func JNI_OnLoad(vm unsafe.Pointer, reserved unsafe.Pointer) C.jint
```

`JNI_OnLoad` supports normal Android library loading paths such as `System.loadLibrary`.

`ModInit` supports explicit `dlopen + dlsym + call` style loaders in CTF setups.

The runtime must guard startup with `sync.Once` so both entries can be present without double-initializing the payload.

## Runtime Startup Flow

```text
libgogi.so loaded
  -> JNI_OnLoad or ModInit
  -> runtime.Start()
  -> capture JavaVM if available
  -> start background goroutine
  -> initialize patch registry
  -> optionally start menu backend
  -> wait for configured target libraries
  -> apply enabled startup patches
```

The startup path must not block the Android UI thread. Long-running module waits, scans, and patch operations run on a goroutine.

## Memory Patch Runtime

The memory runtime provides a small API around process-local patching.

Core types:

```go
type Module struct {
    Name  string
    Base  uintptr
    End   uintptr
    Path  string
    Perms string
}

type Patch struct {
    ID      string
    Library string
    RVA     uintptr
    Pattern string
    Expect  []byte
    Replace []byte
    Once    bool
}
```

Patch application flow:

```text
resolve module
  -> calculate address from RVA or pattern
  -> read current bytes
  -> verify expected bytes when configured
  -> page-align target range
  -> mprotect writable
  -> write replacement bytes
  -> flush instruction cache when code is patched
  -> restore page permissions
  -> store original bytes for restore
```

Initial MVP supports `Library + RVA + Expect + Replace`.

Pattern scanning is part of the architecture but can follow after the RVA patch path is reliable.

## Manifest

The project manifest is the source of truth for build and patch configuration.

Example:

```toml
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
```

The builder validates:

- patch IDs are unique
- menu items reference existing patches
- byte strings are valid hex
- replacement length matches expected length when both are present
- configured ABI/API can be built with the discovered NDK

## Control Layer

The control layer owns patch state independently from UI.

Responsibilities:

- register patches from embedded config
- apply, restore, and toggle patches by ID
- expose state snapshots for menu rendering
- serialize command results into simple success/error responses

The menu must call the control layer rather than directly calling memory APIs.

## HTML/CSS Mod Menu

The internal overlay menu is implemented as a WebView frontend backed by Go.

Architecture:

```text
Android WebView overlay
  -> loads http://127.0.0.1:<port>/

Embedded HTML/CSS/JS
  -> fetch("/api/state")
  -> fetch("/api/toggle/<id>", { method: "POST" })

Go menu server
  -> control.Toggle(id)
  -> returns JSON state
```

The menu frontend is stored as embedded assets:

```text
payload/menu/assets/menu.html
payload/menu/assets/menu.css
payload/menu/assets/menu.js
```

The UI should be compact, touch-friendly, and readable over game content. It should avoid heavy layout work and unnecessary animation because it runs inside another app's process.

Expected controls:

- toggle
- button
- slider
- value label
- group section

## WebView Overlay Backend

The `webview` backend attaches an Android WebView overlay inside the current process.

Responsibilities:

- obtain or receive a usable Android `Context`/`Activity`
- create a WebView on the UI thread
- load the local menu URL
- attach the WebView to the window
- support show/hide
- forward lifecycle cleanup to the Go menu backend

Because injectable contexts vary, backend initialization supports multiple modes:

1. `JNI_OnLoad` captures `JavaVM`.
2. Explicit exported entry can receive an Activity/Context when a loader can provide one.
3. App-specific adapters can be added later for CTF targets that expose a known Activity path.

The first implementation should make the context dependency explicit. Automatic Activity discovery can be a later enhancement.

## CLI

Initial commands:

```bash
gogi init <name>
gogi add-patch <id> --lib <library> --rva <hex> --expect <hex> --replace <hex>
gogi add-toggle <id> --label <label> --patch <patch-id>
gogi build --abi arm64-v8a --api 24 --menu webview
```

`gogi init` creates a payload project with a valid manifest and minimal Go payload.

`gogi build` produces:

```text
dist/
  arm64-v8a/libgogi.so
  arm64-v8a/libgogi.h
  build-info.json
  patches.resolved.json
```

## Testing Strategy

Host-side tests:

- manifest parsing and validation
- ABI/API build configuration resolution
- patch byte parsing
- template generation

Payload package tests where possible:

- `/proc/self/maps` parser with fixture input
- module range resolution
- patch address calculation
- pattern parser/scanner with byte slices
- control state transitions

Manual Android validation:

- build `arm64-v8a` payload
- load with `System.loadLibrary` in a test app
- load with `dlopen + ModInit` in a CTF harness
- verify logcat startup
- verify RVA patch apply/restore
- verify WebView menu show/hide
- verify toggle applies/restores patch

## MVP Scope

MVP includes:

- `gogi` CLI skeleton
- project init
- TOML manifest
- Android `arm64-v8a` c-shared build
- exported `JNI_OnLoad` and `ModInit`
- logcat logging
- `/proc/self/maps` parser
- module base resolver
- RVA patch apply/restore
- control layer
- embedded HTML/CSS/JS menu assets
- local Go HTTP menu server
- WebView menu backend interface

The first implementation must keep WebView attachment explicit. If an Activity/Context is not provided, `gogi` still starts the local menu server and serves the embedded frontend, but the in-process overlay reports a clear `context_required` error instead of pretending to attach.

## Later Extensions

- `armeabi-v7a`
- pattern-based patch targets
- patch groups and profiles
- persistent menu state
- richer WebView frontend components
- app-specific Activity/Context adapters
- optional render overlay backend
- symbol-assisted patch resolution
