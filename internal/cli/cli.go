package cli

import (
	"fmt"
	"io"

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
		fmt.Fprintln(stderr, "build command is unavailable until android build support is added")
		return 1
	default:
		fmt.Fprintf(stderr, "unknown command %q\n", args[0])
		printHelp(stderr)
		return 2
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
