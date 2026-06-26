# gogi

`gogi` (Go Game Injector) builds native Android game mod libraries in Go.

It generates a small mod project with:

```text
frontend/   HTML, CSS, and JavaScript for the in-app WebView menu
backend/    Go entry point linked into the native payload
gogi.toml   Android build and overlay configuration
```

## Requirements

- Go 1.25 or newer
- Android SDK with build-tools
- Android NDK
- `apktool`
- Android debug keystore at `~/.android/debug.keystore`

`gogi compile` can find the NDK from `ANDROID_NDK_HOME`, `ANDROID_NDK_ROOT`, or the latest NDK under `ANDROID_HOME` / `ANDROID_SDK_ROOT`.

## Install

```bash
go install github.com/j0j1j2/gogi/cmd/gogi@latest
```

Make sure your Go bin directory is on `PATH`:

```bash
export PATH="$(go env GOPATH)/bin:$PATH"
```

Check the CLI:

```bash
gogi help
```

## Quick Start

```bash
gogi init mymod
cd mymod

gogi validate
gogi compile
gogi build --apk target.apk --out target-gogi.apk
```

For XAPK bundles:

```bash
gogi build --xapk target.xapk --out target-gogi.xapk
```

The compiled library is written to:

```text
dist/arm64-v8a/libgogi.so
```

## Project Layout

After `gogi init mymod`:

```text
mymod/
  go.mod
  gogi.toml
  frontend/
    index.html
    style.css
    main.js
  backend/
    main.go
```

`gogi compile` generates an internal `.gogi/build` package, embeds `frontend/`, links `backend.Init(nil)`, and builds an Android `c-shared` library.

## Configuration

Default `gogi.toml`:

```toml
name = "mymod"

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
```

`width`, `height`, and `collapsed_size` are interpreted as Android dp values by the injected overlay helper.

## Frontend

The frontend is served inside the target process through a local HTTP server.

Available endpoints:

```text
GET  /api/state
POST /api/toggle/<id>?enabled=true
POST /api/toggle/<id>?enabled=false
```

The generated `frontend/main.js` shows the basic pattern:

```js
async function refresh() {
  const response = await fetch("/api/state");
  const state = await response.json();
  document.getElementById("actions").textContent = JSON.stringify(state, null, 2);
}

refresh();
```

You can replace `frontend/index.html`, `frontend/style.css`, and `frontend/main.js` with your own menu UI.

## Backend

The backend is Go code compiled into the payload.

Generated `backend/main.go`:

```go
package backend

type Context interface{}

func Init(ctx Context) {
}
```

`Init` is called when the generated payload starts. The current SDK surface is intentionally small; patches and richer menu actions are being moved into backend code instead of `gogi.toml`.

## Build Integration

`gogi build --apk` performs:

1. `gogi compile`
2. APK decode with `apktool`
3. `libgogi.so` insertion under `lib/<abi>/`
4. Android overlay helper smali insertion
5. `System.loadLibrary("gogi")` insertion into the app entry path
6. APK rebuild
7. `zipalign`
8. signing with the Android debug keystore

`gogi build --xapk` replaces the base APK inside the XAPK bundle with a rebuilt signed APK and preserves the other bundle files.

## Notes

- APK/XAPK build currently targets `arm64-v8a`.
- Output APKs are signed with the debug keystore.
- Apps with unusual startup flows or heavily customized manifests may need additional adapter support.
- `gogi.toml` is for build and overlay settings. Memory edits and menu behavior belong in Go backend code.
