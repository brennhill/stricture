// fix_command.go — The fix subcommand: applies auto-fixes or previews them with --dry-run.
package cli

import "fmt"

func runFix(args []string) {
	if hasHelpFlag(args) {
		printFixUsage()
		return
	}
	runLint(fixArgs(args))
}

// fixArgs maps the fix subcommand onto lint flags: it applies fixes unless
// --dry-run (or its --fix-dry-run alias) asks for a preview, since lint rejects both together.
func fixArgs(args []string) []string {
	out := make([]string, 0, len(args)+1)
	dryRun := false
	for _, arg := range args {
		switch arg {
		case "--dry-run", "-dry-run", "--fix-dry-run", "-fix-dry-run":
			dryRun = true
		default:
			out = append(out, arg)
		}
	}
	if dryRun {
		return append([]string{"--fix-dry-run"}, out...)
	}
	return append([]string{"--fix"}, out...)
}

func printFixUsage() {
	fmt.Println("Usage: strict fix [options] [paths...]")
	fmt.Println()
	fmt.Println("Apply auto-fixes for fixable violations.")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  --dry-run         Show what would change without modifying files")
	fmt.Println("  --fix-backup      Create .bak files before modifying sources")
	fmt.Println()
	fmt.Println("All 'strict lint' options also apply; run 'strict lint --help' for the full list.")
}
