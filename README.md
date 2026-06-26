# gogi

`gogi` is short for Go Game Injector.

`gogi` builds Go-based Android injectable shared libraries for game mod tooling.

## MVP workflow

```bash
gogi init sample
cd sample

# edit frontend/ and backend/
gogi validate
gogi compile --abi arm64-v8a --api 24
gogi build --apk victim.apk --out victim-gogi.apk
```

The payload is built with:

```bash
GOOS=android GOARCH=arm64 CGO_ENABLED=1 CC=<ndk-clang> go build -buildmode=c-shared -o dist/arm64-v8a/libgogi.so ./payload
```

Generated projects keep user code in `backend/` and UI assets in `frontend/`.
`gogi.toml` configures the build and overlay environment; memory patches and menu actions belong in backend Go code.
The WebView menu frontend is HTML/CSS/JS served by the Go payload over a local HTTP API.
