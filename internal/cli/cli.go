package cli

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"

	"github.com/j0j1j2/gogi/internal/buildenv"
	"github.com/j0j1j2/gogi/internal/project"
	gogitemplate "github.com/j0j1j2/gogi/internal/template"
)

type runCommandFunc func(name string, args []string, env map[string]string, stdout io.Writer, stderr io.Writer) error

var commandRunner runCommandFunc = runCommand

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
		}
		buildArgs := []string{"build", "-buildmode=c-shared", "-o", outPath, payloadPackage()}
		if ldflags := overlayLDFlags(manifest); len(ldflags) > 0 {
			buildArgs = []string{"build", "-ldflags", joinLDFlags(ldflags), "-buildmode=c-shared", "-o", outPath, payloadPackage()}
		}
		if err := commandRunner("go", buildArgs, buildEnv, stdout, stderr); err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		fmt.Fprintf(stdout, "built %s\n", outPath)
		return 0
	case "build":
		target := ""
		out := ""
		for i := 1; i < len(args); i++ {
			switch args[i] {
			case "--apk", "--xapk":
				i++
				if i >= len(args) {
					fmt.Fprintf(stderr, "%s requires a value\n", args[i-1])
					return 2
				}
				target = args[i]
			case "--out":
				i++
				if i >= len(args) {
					fmt.Fprintln(stderr, "--out requires a value")
					return 2
				}
				out = args[i]
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
		fmt.Fprintln(stderr, "APK/XAPK integration is not implemented yet")
		return 1
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
}

func runCommand(name string, args []string, env map[string]string, stdout io.Writer, stderr io.Writer) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	cmd.Env = append(os.Environ(), envSlice(env)...)
	return cmd.Run()
}

func envSlice(env map[string]string) []string {
	items := make([]string, 0, len(env))
	for key, value := range env {
		items = append(items, key+"="+value)
	}
	return items
}

func payloadPackage() string {
	if info, err := os.Stat("payload"); err == nil && info.IsDir() {
		return "./payload"
	}
	return "github.com/j0j1j2/gogi/payload"
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
