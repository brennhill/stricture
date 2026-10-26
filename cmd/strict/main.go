// main.go — Stricture CLI entry point.
package main

import (
	"os"

	"github.com/stricture/stricture/internal/cli"
)

// version is overridden at build time with -ldflags "-X main.version=...".
var version = "0.1.0-dev"

func main() {
	cli.Run(os.Args, version)
}
