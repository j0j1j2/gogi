# gogi

`gogi` builds Go-based Android injectable shared libraries for CTF and controlled research targets.

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
