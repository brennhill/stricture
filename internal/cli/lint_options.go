// lint_options.go — Parses and validates lint command flags into lintOptions.
package cli

import (
	"flag"
	"fmt"
	"runtime"
	"strings"

	"github.com/stricture/stricture/internal/discovery"
)

// lintOptions holds validated lint flags.
type lintOptions struct {
	format        string
	configPath    string
	noConfig      bool
	ruleFilters   []string
	category      string
	minSeverity   string
	forceColor    bool
	forceNoColor  bool
	verbose       bool
	concurrency   int
	outputPath    string
	maxViolations int
	baselinePath  string
	diffMode      bool
	changedOnly   bool
	stagedOnly    bool
	fixApply      bool
	fixDryRun     bool
	fixBackup     bool
	cacheActive   bool
	extensions    map[string]bool
	paths         []string
}

// rawLintFlags are the lint flags exactly as typed, before validation.
type rawLintFlags struct {
	format, configPath, category, extFilter, severity, outputPath, baselinePath string
	noConfig, quiet, forceColor, forceNoColor, verbose, diffMode                bool
	changedOnly, stagedOnly, fixApply, fixDryRun, fixBackup                     bool
	cacheEnabled, noCache                                                       bool
	concurrency, maxViolations                                                  int
	ruleFilters                                                                 repeatableFlag
}

func registerLintFlags(fs *flag.FlagSet, f *rawLintFlags) {
	fs.StringVar(&f.format, "format", "text", "Output format (text, json, sarif, junit)")
	fs.StringVar(&f.configPath, "config", ".stricture.yml", "Path to configuration file")
	fs.BoolVar(&f.noConfig, "no-config", false, "Ignore config file and use built-in defaults")
	fs.Var(&f.ruleFilters, "rule", "Run a single rule by ID (can be repeated)")
	fs.StringVar(&f.category, "category", "", "Run all rules in a category")
	fs.StringVar(&f.extFilter, "ext", "", "Only lint files with this extension (example: .go or .ts)")
	fs.StringVar(&f.severity, "severity", "", "Only report violations at this level or above (error, warn)")
	fs.BoolVar(&f.quiet, "quiet", false, "Only show errors, not warnings")
	fs.BoolVar(&f.forceColor, "color", false, "Force color output in text format")
	fs.BoolVar(&f.forceNoColor, "no-color", false, "Disable color output in text format")
	fs.BoolVar(&f.verbose, "verbose", false, "Show rule timing and debug info")
	fs.IntVar(&f.concurrency, "concurrency", runtime.NumCPU(), "Max parallel file processing")
	fs.StringVar(&f.outputPath, "output", "", "Write report to file instead of stdout")
	fs.IntVar(&f.maxViolations, "max-violations", 0, "Stop after N violations (0 = unlimited)")
	fs.StringVar(&f.baselinePath, "baseline", "", "Path to baseline file (existing violations are suppressed; missing file bootstraps baseline)")
	fs.BoolVar(&f.diffMode, "diff", false, "When used with --baseline, include added/resolved diff details against baseline")
	fs.BoolVar(&f.changedOnly, "changed", false, "Lint only changed files in git working tree/index")
	fs.BoolVar(&f.stagedOnly, "staged", false, "Lint only staged files in git index")
	fs.BoolVar(&f.fixApply, "fix", false, "Apply auto-fixes for fixable violations")
	fs.BoolVar(&f.fixDryRun, "fix-dry-run", false, "Show what --fix would change without modifying files")
	fs.BoolVar(&f.fixBackup, "fix-backup", false, "When used with --fix, create .bak files before modifying sources")
	fs.BoolVar(&f.cacheEnabled, "cache", false, "Enable caching (default behavior)")
	fs.BoolVar(&f.noCache, "no-cache", false, "Disable caching")
}

// parseLintOptions parses and validates lint arguments, exiting with code 2 on any usage error.
func parseLintOptions(args []string) lintOptions {
	flagArgs, pathArgs, argErr := splitLintArgs(args)
	if argErr != nil {
		exitf(2, "%v", argErr)
	}

	var raw rawLintFlags
	fs := flag.NewFlagSet("lint", flag.ExitOnError)
	registerLintFlags(fs, &raw)
	parseFlagSetOrExit(fs, flagArgs)

	return validateLintFlags(raw, pathArgs)
}

// usageCheck is one rejected flag combination; checks run in order and the first failure wins.
type usageCheck struct {
	failed  bool
	message string
}

func validateLintFlags(raw rawLintFlags, pathArgs []string) lintOptions {
	for _, c := range usageChecks(raw) {
		if c.failed {
			exitf(2, "%s", c.message)
		}
	}

	minSeverity, err := resolveMinSeverity(raw.severity, raw.quiet)
	if err != nil {
		exitf(2, "%v", err)
	}
	extensions, err := discovery.ParseExtensionFilter(raw.extFilter)
	if err != nil {
		exitf(2, "%v", err)
	}

	paths := pathArgs
	if len(paths) == 0 {
		paths = []string{"."}
	}
	return lintOptions{
		format:        raw.format,
		configPath:    raw.configPath,
		noConfig:      raw.noConfig,
		ruleFilters:   raw.ruleFilters.Values(),
		category:      raw.category,
		minSeverity:   minSeverity,
		forceColor:    raw.forceColor,
		forceNoColor:  raw.forceNoColor,
		verbose:       raw.verbose,
		concurrency:   raw.concurrency,
		outputPath:    raw.outputPath,
		maxViolations: raw.maxViolations,
		baselinePath:  raw.baselinePath,
		diffMode:      raw.diffMode,
		changedOnly:   raw.changedOnly,
		stagedOnly:    raw.stagedOnly,
		fixApply:      raw.fixApply,
		fixDryRun:     raw.fixDryRun,
		fixBackup:     raw.fixBackup,
		cacheActive:   !raw.noCache || raw.cacheEnabled,
		extensions:    extensions,
		paths:         paths,
	}
}

// usageChecks lists rejected flag combinations in the order they are reported.
func usageChecks(raw rawLintFlags) []usageCheck {
	validFormats := map[string]bool{"text": true, "json": true, "sarif": true, "junit": true}
	return []usageCheck{
		{raw.fixApply && raw.fixDryRun, "--fix and --fix-dry-run are mutually exclusive"},
		{raw.fixBackup && !raw.fixApply, "--fix-backup requires --fix"},
		{raw.changedOnly && raw.stagedOnly, "--changed and --staged are mutually exclusive"},
		{raw.diffMode && strings.TrimSpace(raw.baselinePath) == "", "--diff requires --baseline"},
		{!validFormats[raw.format], fmt.Sprintf("invalid format %q (valid: text, json, sarif, junit)", raw.format)},
		{raw.maxViolations < 0, "--max-violations must be >= 0"},
		{raw.forceColor && raw.forceNoColor, "--color and --no-color are mutually exclusive"},
		{raw.cacheEnabled && raw.noCache, "--cache and --no-cache are mutually exclusive"},
		{raw.concurrency < 1, "--concurrency must be >= 1"},
	}
}

// resolveMinSeverity normalizes --severity and applies --quiet, which means errors only.
func resolveMinSeverity(severity string, quiet bool) (string, error) {
	level := strings.ToLower(strings.TrimSpace(severity))
	switch level {
	case "", "warn", "error":
		// Valid values.
	default:
		return "", fmt.Errorf("invalid severity %q (valid: error, warn)", severity)
	}
	if !quiet {
		return level, nil
	}
	if level != "" && level != "error" {
		return "", fmt.Errorf("--quiet cannot be combined with --severity=warn")
	}
	return "error", nil
}
