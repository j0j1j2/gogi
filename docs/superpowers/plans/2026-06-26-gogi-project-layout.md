# Gogi Project Layout Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Move `gogi` projects to a `frontend/` plus `backend/` model where runtime patch behavior is Go code, not TOML configuration.

**Architecture:** `gogi.toml` describes environment, build, overlay, frontend, and backend entry settings. User patch/menu behavior lives in generated Go backend code and is compiled into the shared object. The CLI gains `compile` for `.so` builds while `build` remains the APK/XAPK integration command surface.

**Tech Stack:** Go, BurntSushi TOML, Android Go `c-shared`, HTML/CSS/JS frontend assets.

## Global Constraints

- The payload logic is written in Go, not C++.
- Frontend files are HTML/CSS/JS.
- Memory patch definitions do not live in `gogi.toml`.
- `gogi init` creates a minimal compile-safe project.
- `gogi compile` builds a signed-independent `.so`; APK/XAPK signing belongs to `gogi build`.

---

### Task 1: Config Model Without Patch Definitions

**Files:**
- Modify: `internal/project/manifest.go`
- Modify: `internal/project/manifest_test.go`
- Modify: `examples/basic_patch/gogi.toml`

**Interfaces:**
- Produces: `project.Manifest` with `Build`, `Overlay`, `Frontend`, and `Backend` sections.
- Produces: `(*Manifest).Validate() error` that rejects missing environment settings but does not validate patch definitions.

- [x] Write tests showing a config with build/overlay/frontend/backend validates.
- [x] Write tests showing old `[[patches]]` config is not required.
- [x] Remove `Patch`, `MenuToggle`, and patch-reference validation from `Manifest`.
- [x] Update the example config to contain no patch definitions.
- [x] Run `go test ./internal/project`.

### Task 2: Init Template Generates Frontend and Backend

**Files:**
- Modify: `internal/template/init.go`
- Modify: `internal/template/init_test.go`

**Interfaces:**
- Produces: generated files `frontend/index.html`, `frontend/style.css`, `frontend/main.js`, `backend/main.go`, and `gogi.toml`.

- [x] Update the init test to expect `frontend/` and `backend/` files instead of `payload/`.
- [x] Implement a TOML template with only build/overlay/frontend/backend settings.
- [x] Implement minimal backend Go code that registers runtime behavior in code.
- [x] Implement minimal HTML/CSS/JS frontend files.
- [x] Run `go test ./internal/template`.

### Task 3: CLI Compile Command

**Files:**
- Modify: `internal/cli/cli.go`
- Modify: `internal/cli/cli_test.go`
- Modify: `README.md`

**Interfaces:**
- Produces: `gogi compile [--abi arm64-v8a] [--api 24]`.
- Keeps: `gogi build` as APK/XAPK integration entry point.

- [x] Add a CLI test for help text containing `gogi compile`.
- [x] Add a CLI test that `build` with no target APK reports APK/XAPK integration usage.
- [x] Move the existing shared-object command output from `build` to `compile`.
- [x] Update README workflow.
- [x] Run `go test ./internal/cli`.

### Task 4: Full Verification

**Files:**
- No new files.

**Interfaces:**
- Verifies all repository packages still pass.

- [x] Run `go test ./...`.
- [x] Run `go run ./cmd/gogi init tmp/layout-smoke`.
- [x] Run `go run ./cmd/gogi validate tmp/layout-smoke/gogi.toml`.
- [x] Run `find tmp/layout-smoke -maxdepth 2 -type f | sort` and confirm the expected layout.
