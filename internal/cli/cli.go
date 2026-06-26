package cli

import (
	"fmt"
	"io"
	"os"
	"runtime"
	"strconv"

	"gogi/internal/buildenv"
	"gogi/internal/project"
	gogitemplate "gogi/internal/template"
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
	fmt.Fprintln(w, "  gogi build [--abi arm64-v8a] [--api 24] [--menu webview]")
}
