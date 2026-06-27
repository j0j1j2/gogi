package devbackend

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/j0j1j2/gogi/internal/project"
)

type Cleanup func()

func Start(manifestPath string, stdout io.Writer, stderr io.Writer) (string, Cleanup, error) {
	absManifest, err := filepath.Abs(manifestPath)
	if err != nil {
		return "", nil, err
	}
	root := filepath.Dir(absManifest)
	manifest, err := project.LoadManifest(absManifest)
	if err != nil {
		return "", nil, err
	}
	modulePath, err := readModulePath(filepath.Join(root, "go.mod"))
	if err != nil {
		return "", nil, err
	}
	addr, err := freeAddr()
	if err != nil {
		return "", nil, err
	}
	buildDir := filepath.Join(root, ".gogi", "devbackend")
	if err := os.MkdirAll(buildDir, 0o755); err != nil {
		return "", nil, err
	}
	source := source(modulePath, manifest.Backend.Entry)
	if err := os.WriteFile(filepath.Join(buildDir, "main.go"), []byte(source), 0o644); err != nil {
		return "", nil, err
	}

	cmd := exec.Command("go", "run", "./.gogi/devbackend")
	cmd.Dir = root
	cmd.Env = append(os.Environ(), "GOGI_DEV_BACKEND_ADDR="+addr)
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	if err := cmd.Start(); err != nil {
		return "", nil, err
	}
	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()
	cleanup := func() {
		if cmd.Process != nil {
			_ = cmd.Process.Kill()
		}
		select {
		case <-done:
		case <-time.After(2 * time.Second):
		}
	}
	url := "http://" + addr
	if err := waitReady(url, done); err != nil {
		cleanup()
		return "", nil, err
	}
	return url, cleanup, nil
}

func readModulePath(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read %s: %w", path, err)
	}
	for _, line := range strings.Split(string(data), "\n") {
		fields := strings.Fields(line)
		if len(fields) == 2 && fields[0] == "module" {
			return fields[1], nil
		}
	}
	return "", fmt.Errorf("%s missing module declaration", path)
}

func freeAddr() (string, error) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return "", err
	}
	addr := listener.Addr().String()
	if err := listener.Close(); err != nil {
		return "", err
	}
	return addr, nil
}

func waitReady(url string, done <-chan error) error {
	deadline := time.Now().Add(15 * time.Second)
	client := &http.Client{Timeout: 250 * time.Millisecond}
	var lastErr error
	for time.Now().Before(deadline) {
		select {
		case err := <-done:
			if err == nil {
				return fmt.Errorf("dev backend exited before becoming ready")
			}
			return fmt.Errorf("dev backend exited before becoming ready: %w", err)
		default:
		}
		resp, err := client.Get(url + "/api/state")
		if err == nil {
			_ = resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return nil
			}
			lastErr = fmt.Errorf("status %d", resp.StatusCode)
		} else {
			lastErr = err
		}
		time.Sleep(100 * time.Millisecond)
	}
	return fmt.Errorf("dev backend did not become ready: %w", lastErr)
}

func source(modulePath string, backendEntry string) string {
	backendImport := modulePath + "/" + strings.Trim(backendEntry, "/")
	return fmt.Sprintf(`package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	userbackend "%s"
	"github.com/j0j1j2/gogi/payload/control"
	"github.com/j0j1j2/gogi/payload/mem"
	"github.com/j0j1j2/gogi/sdk"
)

func main() {
	logger := log.New(os.Stderr, "gogi backend: ", log.LstdFlags)
	registry := control.NewRegistry()
	ctx := sdk.NewContext()
	ctx.Logf = logger.Printf
	userbackend.Init(ctx)
	for _, patch := range sdk.RegisteredPatches() {
		registry.Register(mem.PatchSpec{
			ID:      patch.ID,
			Library: patch.Library,
			RVA:     patch.RVA,
			Expect:  patch.Expect,
			Replace: patch.Replace,
			Startup: patch.Startup,
		})
	}
	for _, action := range sdk.RegisteredActions() {
		action := action
		registry.RegisterAction(control.ActionSpec{
			ID: action.ID,
			Handler: func(req control.ActionRequest) (any, error) {
				return action.Handler(sdk.ActionRequest{ID: req.ID, Payload: req.Payload})
			},
		})
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/api/state", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-type", "application/json")
		_ = json.NewEncoder(w).Encode(registry.Snapshot())
	})
	mux.HandleFunc("/api/toggle/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		id := strings.TrimPrefix(r.URL.Path, "/api/toggle/")
		enabled := r.URL.Query().Get("enabled") == "true"
		if err := registry.Toggle(id, enabled); err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		w.Header().Set("content-type", "application/json")
		_ = json.NewEncoder(w).Encode(registry.Snapshot())
	})
	mux.HandleFunc("/api/action/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		id := strings.TrimPrefix(r.URL.Path, "/api/action/")
		payload, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		result, err := registry.DispatchAction(id, json.RawMessage(payload))
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		w.Header().Set("content-type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"ok": true, "result": result})
	})

	addr := os.Getenv("GOGI_DEV_BACKEND_ADDR")
	if addr == "" {
		addr = "127.0.0.1:0"
	}
	logger.Printf("listening on http://%%s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		logger.Fatal(err)
	}
}
`, backendImport)
}
