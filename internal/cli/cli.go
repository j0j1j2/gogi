package cli

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/j0j1j2/gogi/internal/apkbuild"
	"github.com/j0j1j2/gogi/internal/buildenv"
	"github.com/j0j1j2/gogi/internal/devbackend"
	"github.com/j0j1j2/gogi/internal/devserver"
	"github.com/j0j1j2/gogi/internal/project"
	gogitemplate "github.com/j0j1j2/gogi/internal/template"
	"github.com/j0j1j2/gogi/internal/version"
)

type runCommandFunc func(name string, args []string, env map[string]string, stdout io.Writer, stderr io.Writer) error
type devBackendStartFunc func(manifestPath string, stdout io.Writer, stderr io.Writer) (string, func(), error)

var commandRunner runCommandFunc = runCommand
var apkBuilder = apkbuild.BuildAPK
var xapkBuilder = apkbuild.BuildXAPK
var dependencyResolver = resolveProjectDependencies
var devServer = devserver.Serve
var devBackendStarter devBackendStartFunc = func(manifestPath string, stdout io.Writer, stderr io.Writer) (string, func(), error) {
	url, cleanup, err := devbackend.Start(manifestPath, stdout, stderr)
	if cleanup == nil {
		return url, nil, err
	}
	return url, func() { cleanup() }, err
}

func Run(args []string, stdout io.Writer, stderr io.Writer) int {
	if len(args) == 0 {
		printHelp(stdout)
		return 0
	}

	switch args[0] {
	case "help", "-h", "--help":
		printHelp(stdout)
		return 0
	case "version", "-v", "--version":
		fmt.Fprint(stdout, version.Current().String())
		return 0
	case "init":
		if len(args) != 2 {
			fmt.Fprintln(stderr, "usage: gogi init <name>")
			return 2
		}
		if err := gogitemplate.InitProject(args[1], args[1]); err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		if err := dependencyResolver(args[1], stdout, stderr); err != nil {
			fmt.Fprintf(stderr, "warning: initialize sdk dependency: %v\n", err)
		}
		fmt.Fprintf(stdout, "created %s\n", args[1])
		return 0
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
	case "dev":
		addr := "127.0.0.1:17374"
		proxy := ""
		frontendDir := "frontend"
		manifest, ok, err := loadManifestIfPresent("gogi.toml")
		if err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		if !ok {
			fmt.Fprintln(stderr, "gogi.toml not found; run gogi dev inside a gogi project")
			return 1
		}
		if ok && manifest.Frontend.Entry != "" {
			frontendDir = filepath.Dir(manifest.Frontend.Entry)
			if frontendDir == "." {
				frontendDir = "frontend"
			}
		}
		for i := 1; i < len(args); i++ {
			switch args[i] {
			case "--addr":
				i++
				if i >= len(args) {
					fmt.Fprintln(stderr, "--addr requires a value")
					return 2
				}
				addr = args[i]
			case "--proxy":
				i++
				if i >= len(args) {
					fmt.Fprintln(stderr, "--proxy requires a value")
					return 2
				}
				proxy = args[i]
			case "--frontend":
				i++
				if i >= len(args) {
					fmt.Fprintln(stderr, "--frontend requires a value")
					return 2
				}
				frontendDir = args[i]
			default:
				fmt.Fprintf(stderr, "unknown dev flag %q\n", args[i])
				return 2
			}
		}
		var cleanup func()
		if proxy == "" {
			backendURL, backendCleanup, err := devBackendStarter("gogi.toml", stdout, stderr)
			if err != nil {
				fmt.Fprintf(stderr, "warning: start go backend dev runner: %v\n", err)
			} else {
				proxy = backendURL
				cleanup = backendCleanup
				fmt.Fprintf(stdout, "backend connected on %s\n", backendURL)
			}
		}
		if cleanup != nil {
			defer cleanup()
		}
		if err := devServer(devserver.Options{
			FrontendDir: frontendDir,
			Addr:        addr,
			Proxy:       proxy,
			Stdout:      stdout,
			Overlay: devserver.OverlayOptions{
				Width:         manifest.Overlay.Width,
				Height:        manifest.Overlay.Height,
				CollapsedSize: manifest.Overlay.CollapsedSize,
				Draggable:     manifest.Overlay.Draggable,
			},
		}); err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		return 0
	case "compile":
		abi := "arm64-v8a"
		api := 24
		manifest, ok, err := loadManifestIfPresent("gogi.toml")
		if err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		if ok {
			if len(manifest.Build.ABIs) > 0 {
				abi = manifest.Build.ABIs[0]
			}
			if manifest.Build.MinSDK > 0 {
				api = manifest.Build.MinSDK
			}
		}
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
			default:
				fmt.Fprintf(stderr, "unknown compile flag %q\n", args[i])
				return 2
			}
		}
		env := map[string]string{
			"ANDROID_NDK_HOME": os.Getenv("ANDROID_NDK_HOME"),
			"ANDROID_NDK_ROOT": os.Getenv("ANDROID_NDK_ROOT"),
			"ANDROID_HOME":     os.Getenv("ANDROID_HOME"),
			"ANDROID_SDK_ROOT": os.Getenv("ANDROID_SDK_ROOT"),
		}
		cfg, err := buildenv.ResolveAndroid(env, abi, api, defaultHostTag())
		if err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		outPath := filepath.Join("dist", cfg.ABI, "libgogi.so")
		if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		buildEnv := map[string]string{
			"GOOS":        cfg.GoOS,
			"GOARCH":      cfg.GoArch,
			"CGO_ENABLED": "1",
			"CC":          cfg.CC,
			"GOPROXY":     "direct",
		}
		payloadPkg, err := payloadPackage(manifest)
		if err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		if payloadPkg == generatedPayloadPackage {
			if err := commandRunner("go", []string{"get", "github.com/j0j1j2/gogi@main"}, buildEnv, stdout, stderr); err != nil {
				fmt.Fprintln(stderr, err)
				return 1
			}
		}
		buildArgs := []string{"build", "-buildmode=c-shared", "-o", outPath, payloadPkg}
		if ldflags := overlayLDFlags(manifest); len(ldflags) > 0 {
			buildArgs = []string{"build", "-ldflags", joinLDFlags(ldflags), "-buildmode=c-shared", "-o", outPath, payloadPkg}
		}
		if err := commandRunner("go", buildArgs, buildEnv, stdout, stderr); err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		fmt.Fprintf(stdout, "built %s\n", outPath)
		return 0
	case "build":
		target := ""
		targetKind := ""
		out := ""
		abi := "arm64-v8a"
		api := 24
		for i := 1; i < len(args); i++ {
			switch args[i] {
			case "--apk":
				i++
				if i >= len(args) {
					fmt.Fprintf(stderr, "%s requires a value\n", args[i-1])
					return 2
				}
				target = args[i]
				targetKind = "apk"
			case "--xapk":
				i++
				if i >= len(args) {
					fmt.Fprintf(stderr, "%s requires a value\n", args[i-1])
					return 2
				}
				target = args[i]
				targetKind = "xapk"
			case "--out":
				i++
				if i >= len(args) {
					fmt.Fprintln(stderr, "--out requires a value")
					return 2
				}
				out = args[i]
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
			default:
				fmt.Fprintf(stderr, "unknown build flag %q\n", args[i])
				return 2
			}
		}
		if target == "" {
			fmt.Fprintln(stderr, "usage: gogi build --apk <path>|--xapk <path> [--out <path>]")
			return 2
		}
		if out == "" {
			fmt.Fprintln(stderr, "--out is required")
			return 2
		}
		compileArgs := []string{"compile", "--abi", abi, "--api", strconv.Itoa(api)}
		if code := Run(compileArgs, stdout, stderr); code != 0 {
			return code
		}
		libPath := filepath.Join("dist", abi, "libgogi.so")
		switch targetKind {
		case "apk":
			if err := apkBuilder(apkbuild.APKOptions{
				APKPath:     target,
				OutPath:     out,
				ABI:         abi,
				LibraryPath: libPath,
				Runner:      apkbuild.Runner(commandRunner),
				Stdout:      stdout,
				Stderr:      stderr,
			}); err != nil {
				fmt.Fprintln(stderr, err)
				return 1
			}
		case "xapk":
			if err := xapkBuilder(apkbuild.XAPKOptions{
				XAPKPath:    target,
				OutPath:     out,
				ABI:         abi,
				LibraryPath: libPath,
				Runner:      apkbuild.Runner(commandRunner),
				Stdout:      stdout,
				Stderr:      stderr,
			}); err != nil {
				fmt.Fprintln(stderr, err)
				return 1
			}
		}
		fmt.Fprintf(stdout, "built %s\n", out)
		return 0
	default:
		fmt.Fprintf(stderr, "unknown command %q\n", args[0])
		printHelp(stderr)
		return 2
	}
}

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

func printHelp(w io.Writer) {
	fmt.Fprintln(w, "gogi - Go-based Android injectable .so builder")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Usage:")
	fmt.Fprintln(w, "  gogi init <name>")
	fmt.Fprintln(w, "  gogi validate [manifest]")
	fmt.Fprintln(w, "  gogi compile [--abi arm64-v8a] [--api 24]")
	fmt.Fprintln(w, "  gogi build --apk <path>|--xapk <path> --out <path>")
	fmt.Fprintln(w, "  gogi dev [--addr 127.0.0.1:17374] [--proxy http://host:port] [--frontend dir]")
	fmt.Fprintln(w, "  gogi version")
}

func runCommand(name string, args []string, env map[string]string, stdout io.Writer, stderr io.Writer) error {
	name = resolveExecutable(name)
	cmd := exec.Command(name, args...)
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	cmd.Env = append(os.Environ(), envSlice(env)...)
	return cmd.Run()
}

func resolveProjectDependencies(root string, stdout io.Writer, stderr io.Writer) error {
	cmd := exec.Command("go", "get", "github.com/j0j1j2/gogi@main")
	cmd.Dir = root
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	cmd.Env = append(os.Environ(), "GOPROXY=direct")
	return cmd.Run()
}

func resolveExecutable(name string) string {
	if path, err := exec.LookPath(name); err == nil {
		return path
	}
	if tool := findAndroidBuildTool(name); tool != "" {
		return tool
	}
	return name
}

func findAndroidBuildTool(name string) string {
	sdk := os.Getenv("ANDROID_HOME")
	if sdk == "" {
		sdk = os.Getenv("ANDROID_SDK_ROOT")
	}
	if sdk == "" {
		return ""
	}
	root := filepath.Join(sdk, "build-tools")
	entries, err := os.ReadDir(root)
	if err != nil {
		return ""
	}
	for i := len(entries) - 1; i >= 0; i-- {
		if !entries[i].IsDir() {
			continue
		}
		candidate := filepath.Join(root, entries[i].Name(), name)
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}
	return ""
}

func envSlice(env map[string]string) []string {
	items := make([]string, 0, len(env))
	for key, value := range env {
		items = append(items, key+"="+value)
	}
	return items
}

const generatedPayloadPackage = "./.gogi/build"

func payloadPackage(manifest *project.Manifest) (string, error) {
	if info, err := os.Stat("payload"); err == nil && info.IsDir() {
		return "./payload", nil
	}
	if manifest == nil {
		return "github.com/j0j1j2/gogi/payload", nil
	}
	if err := writeGeneratedPayload(manifest); err != nil {
		return "", err
	}
	return generatedPayloadPackage, nil
}

func loadManifestIfPresent(path string) (*project.Manifest, bool, error) {
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return nil, false, nil
		}
		return nil, false, err
	}
	manifest, err := project.LoadManifest(path)
	if err != nil {
		return nil, false, err
	}
	if err := manifest.Validate(); err != nil {
		return nil, false, err
	}
	return manifest, true, nil
}

func overlayLDFlags(manifest *project.Manifest) []string {
	if manifest == nil || !manifest.Overlay.Enabled {
		return nil
	}
	prefix := "github.com/j0j1j2/gogi/payload/runtime."
	var flags []string
	if manifest.Overlay.Width > 0 {
		flags = append(flags, "-X", prefix+"overlayWidth="+strconv.Itoa(manifest.Overlay.Width))
	}
	if manifest.Overlay.Height > 0 {
		flags = append(flags, "-X", prefix+"overlayHeight="+strconv.Itoa(manifest.Overlay.Height))
	}
	if manifest.Overlay.CollapsedSize > 0 {
		flags = append(flags, "-X", prefix+"overlayCollapsedSize="+strconv.Itoa(manifest.Overlay.CollapsedSize))
	}
	flags = append(flags, "-X", prefix+"overlayDraggable="+strconv.FormatBool(manifest.Overlay.Draggable))
	return flags
}

func joinLDFlags(flags []string) string {
	result := ""
	for i, flag := range flags {
		if i > 0 {
			result += " "
		}
		result += flag
	}
	return result
}

func writeGeneratedPayload(manifest *project.Manifest) error {
	modulePath, err := readModulePath("go.mod")
	if err != nil {
		return err
	}
	buildDir := filepath.Join(".gogi", "build")
	if err := os.RemoveAll(buildDir); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Join(buildDir, "frontend"), 0o755); err != nil {
		return err
	}
	if err := copyDir("frontend", filepath.Join(buildDir, "frontend")); err != nil {
		return fmt.Errorf("copy frontend: %w", err)
	}
	source := generatedPayloadSource(modulePath, manifest.Backend.Entry)
	return os.WriteFile(filepath.Join(buildDir, "main_android.go"), []byte(source), 0o644)
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

func copyDir(src string, dst string) error {
	return filepath.WalkDir(src, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		if entry.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(target, data, 0o644)
	})
}

func generatedPayloadSource(modulePath string, backendEntry string) string {
	backendImport := modulePath + "/" + strings.Trim(backendEntry, "/")
	return fmt.Sprintf(`package main

/*
#include <jni.h>
*/
import "C"

import (
	"embed"
	"unsafe"

	userbackend "%s"
	"github.com/j0j1j2/gogi/sdk"
	gogiruntime "github.com/j0j1j2/gogi/payload/runtime"
)

//go:embed frontend/*
var frontendFiles embed.FS

func init() {
	gogiruntime.SetFrontendAssets(frontendFiles, "frontend")
	ctx := sdk.NewContext()
	ctx.Logf = gogiruntime.Logf
	userbackend.Init(ctx)
}

//export ModInit
func ModInit() {
	gogiruntime.Start(nil)
}

//export JNI_OnLoad
func JNI_OnLoad(vm *C.JavaVM, reserved unsafe.Pointer) C.jint {
	gogiruntime.Start(unsafe.Pointer(vm))
	return C.JNI_VERSION_1_6
}

func main() {}
`, backendImport)
}
