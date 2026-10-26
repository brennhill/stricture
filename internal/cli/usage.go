// usage.go — Top-level usage text and argument classification helpers.
package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// printUsage prints the top-level help message.
func printUsage() {
	fmt.Println("Stricture — A fast, language-agnostic linter")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  strict [command] [options] [paths...]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  lint              Lint files (default when no command given)")
	fmt.Println("  fix               Apply auto-fixes for fixable violations (--dry-run to preview)")
	fmt.Println("  init              Create a default .stricture.yml")
	fmt.Println("  inspect <file>    Parse a file and print its UnifiedFileModel as JSON")
	fmt.Println("  audit             Run cross-service strictness audit checks")
	fmt.Println("  trace <file>      Validate a trace artifact against basic constraints")
	fmt.Println("  policy            Policy URL binding and compliance checks")
	fmt.Println("  inspect-lineage   Parse strict-source annotations from a file")
	fmt.Println("  lineage-export    Build normalized lineage artifact from source files")
	fmt.Println("  lineage-diff      Diff two lineage artifacts and classify drift severity")
	fmt.Println("  lineage-escalate  Resolve emergency contacts upstream from a service")
	fmt.Println("  list-rules        List all registered rules")
	fmt.Println("  explain           Show details for a specific rule")
	fmt.Println("  validate-config   Check that a .stricture.yml file is valid")
	fmt.Println("  version           Print version and exit")
	fmt.Println("  help              Print this help message")
	fmt.Println()
	fmt.Println("Run 'strict <command> --help' for details on a specific command.")
}

func printUnknownCommand(command string) {
	fmt.Fprintf(os.Stderr, "Error: unknown command %q\n", command)
	fmt.Fprintln(os.Stderr, "Valid commands: lint, fix, init, inspect, audit, trace, policy, inspect-lineage, lineage-export, lineage-diff, lineage-escalate, list-rules, explain, validate-config, version, help")
}

func looksLikePathArg(value string) bool {
	v := strings.TrimSpace(value)
	if v == "" {
		return false
	}
	if strings.HasPrefix(v, ".") || strings.HasPrefix(v, "/") || strings.HasPrefix(v, "~") {
		return true
	}
	if strings.Contains(v, string(filepath.Separator)) || strings.Contains(v, "/") {
		return true
	}
	if strings.ContainsAny(v, "*?[]{}") {
		return true
	}
	if strings.Contains(v, ".") {
		return true
	}
	if _, err := os.Stat(v); err == nil {
		return true
	}
	return false
}

func hasHelpFlag(args []string) bool {
	for _, arg := range args {
		switch strings.TrimSpace(arg) {
		case "-h", "--help", "-help", "help":
			return true
		}
	}
	return false
}
