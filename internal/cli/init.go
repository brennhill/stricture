// init.go — The init command: default configuration scaffolding.
package cli

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

func runInit(args []string) {
	fs := flag.NewFlagSet("init", flag.ExitOnError)
	force := fs.Bool("force", false, "Overwrite existing .stricture.yml if it exists")
	pathValue := fs.String("path", ".stricture.yml", "Destination config path")
	fs.Usage = func() {
		fmt.Println("Usage: strict init [options]")
		fmt.Println()
		fmt.Println("Create a default .stricture.yml with recommended settings.")
		fs.PrintDefaults()
	}
	parseFlagSetOrExit(fs, args)

	target := strings.TrimSpace(*pathValue)
	if target == "" {
		target = ".stricture.yml"
	}

	if _, err := os.Stat(target); err == nil && !*force {
		fmt.Fprintf(os.Stderr, "Error: %s already exists. Re-run with --force to overwrite.\n", target)
		os.Exit(2)
	}

	content := defaultInitConfig()
	if err := os.WriteFile(target, []byte(content), 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "Error: write %s: %v\n", target, err)
		os.Exit(1)
	}

	fmt.Printf("Created %s\n", target)
}

func defaultInitConfig() string {
	return `version: "1.0"

rules:
  CONV-file-naming: error
  CONV-file-header: error
  CONV-error-format: error
  CONV-export-naming: error
  CONV-test-file-location: error
  CONV-required-exports: error
  ARCH-dependency-direction: error
  ARCH-import-boundary: error
  ARCH-no-circular-deps: error
  ARCH-max-file-lines: error
  ARCH-layer-violation: error
  ARCH-module-boundary: error
  TQ-no-shallow-assertions: error
  TQ-return-type-verified: error
  TQ-schema-conformance: error
  TQ-error-path-coverage: error
  TQ-assertion-depth: error
  TQ-boundary-tested: error
  TQ-mock-scope: error
  TQ-test-isolation: error
  TQ-negative-cases: error
  TQ-test-naming: error
  CTR-request-shape: error
  CTR-response-shape: error
  CTR-status-code-handling: error
  CTR-shared-type-sync: error
  CTR-json-tag-match: error
  CTR-dual-test: error
  CTR-strictness-parity: error
  CTR-manifest-conformance: error
`
}
