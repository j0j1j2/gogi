package main

import (
	"os"

	"github.com/j0j1j2/gogi/internal/cli"
)

func main() {
	os.Exit(cli.Run(os.Args[1:], os.Stdout, os.Stderr))
}
