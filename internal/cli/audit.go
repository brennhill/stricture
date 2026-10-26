// audit.go — The audit command: contract-focused lint workflow.
package cli

import (
	"flag"
	"fmt"
	"os"
	"strings"

	manifestpkg "github.com/stricture/stricture/internal/manifest"
)

func runAudit(args []string) {
	if hasHelpFlag(args) {
		printAuditUsage()
		return
	}

	flagArgs, pathArgs, argErr := splitAuditArgs(args)
	if argErr != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", argErr)
		os.Exit(2)
	}

	fs := flag.NewFlagSet("audit", flag.ExitOnError)
	manifestPath := fs.String("manifest", "", "Path to stricture manifest file")
	service := fs.String("service", "", "Service name to scope audit output")
	remote := fs.Bool("remote", false, "Fetch remote repositories for cross-validation")
	strictness := fs.String("strictness", "", "Strictness override (minimal|basic|standard|strict|exhaustive)")
	format := fs.String("format", "text", "Output format (text, json, sarif, junit)")
	outputPath := fs.String("output", "", "Write report to file instead of stdout")
	configPath := fs.String("config", ".stricture.yml", "Path to configuration file")
	noConfig := fs.Bool("no-config", false, "Ignore config file and use built-in defaults")
	parseFlagSetOrExit(fs, flagArgs)

	strictnessValue := strings.ToLower(strings.TrimSpace(*strictness))
	if strictnessValue != "" && !validStrictness(strictnessValue) {
		fmt.Fprintf(os.Stderr, "Error: invalid strictness %q (valid: minimal, basic, standard, strict, exhaustive)\n", *strictness)
		os.Exit(2)
	}

	manifest := strings.TrimSpace(*manifestPath)
	if manifest == "" {
		manifest = autoDetectManifestPath()
	}
	if manifest != "" {
		if _, err := manifestpkg.Load(manifest); err != nil {
			fmt.Fprintf(os.Stderr, "Error: manifest %s is invalid or unreadable: %v\n", manifest, err)
			os.Exit(2)
		}
	}

	if *remote {
		fmt.Fprintln(os.Stderr, "Warning: --remote is not implemented yet; running local audit only.")
	}
	if strings.TrimSpace(*service) != "" {
		fmt.Fprintf(os.Stderr, "Info: audit scoped to service %q\n", strings.TrimSpace(*service))
	}
	if strictnessValue != "" {
		fmt.Fprintf(os.Stderr, "Info: strictness override %q is accepted.\n", strictnessValue)
	}

	lintArgs := []string{"--category", "ctr", "--format", *format}
	if strings.TrimSpace(*configPath) != "" {
		lintArgs = append(lintArgs, "--config", *configPath)
	}
	if *noConfig {
		lintArgs = append(lintArgs, "--no-config")
	}
	if strings.TrimSpace(*outputPath) != "" {
		lintArgs = append(lintArgs, "--output", *outputPath)
	}
	lintArgs = append(lintArgs, pathArgs...)
	runLint(lintArgs)
}

func printAuditUsage() {
	fmt.Println("Usage: strict audit [options] [paths...]")
	fmt.Println()
	fmt.Println("Run contract-focused lint checks (CTR category) as an audit workflow.")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  --manifest <path>    Path to stricture-manifest.yml")
	fmt.Println("  --service <name>     Scope audit messaging to one service")
	fmt.Println("  --remote             Attempt remote cross-validation (compatibility flag)")
	fmt.Println("  --strictness <lvl>   Strictness override (compatibility flag)")
	fmt.Println("  --format <fmt>       Output format: text, json, sarif, junit")
	fmt.Println("  --output <file>      Write report to file")
	fmt.Println("  --config <path>      Use a specific config file")
	fmt.Println("  --no-config          Ignore .stricture.yml, use defaults only")
}

func autoDetectManifestPath() string {
	candidates := []string{
		"stricture-manifest.yml",
		".stricture-manifest.yml",
		"stricture-manifest.yaml",
		".stricture-manifest.yaml",
	}
	for _, candidate := range candidates {
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}
	return ""
}

func validStrictness(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "minimal", "basic", "standard", "strict", "exhaustive":
		return true
	default:
		return false
	}
}
