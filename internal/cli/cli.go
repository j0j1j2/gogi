package cli

import (
	"fmt"
	"io"
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
		fmt.Fprintln(stderr, "init command is unavailable until project templates are added")
		return 1
	case "validate":
		fmt.Fprintln(stderr, "validate command is unavailable until manifest support is added")
		return 1
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
