# gogi

`gogi` (Go Game Injector) builds native Android game mod libraries in Go.

## Install

```bash
go install github.com/j0j1j2/gogi/cmd/gogi@latest
```

Make sure your Go bin directory is on `PATH`:

```bash
export PATH="$(go env GOPATH)/bin:$PATH"
```

## MVP workflow

```bash
gogi init sample
cd sample

# edit frontend/ and backend/
gogi validate
gogi compile --abi arm64-v8a --api 24
gogi build --apk victim.apk --out victim-gogi.apk
```

`gogi compile` builds:

```bash
dist/arm64-v8a/libgogi.so
```

Generated projects keep user code in `backend/` and UI assets in `frontend/`.
During compile, `frontend/` is embedded into the payload and `backend.Init(nil)` is linked into the generated shared library.
`gogi.toml` configures the build and overlay environment; memory patches and menu actions belong in backend Go code.
The WebView menu frontend is HTML/CSS/JS served by the Go payload over a local HTTP API.
