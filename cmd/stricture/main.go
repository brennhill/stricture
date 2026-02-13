// main.go — Stricture CLI entry point.
package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"text/tabwriter"
	"time"
	"unicode/utf8"

	"github.com/stricture/stricture/internal/adapter"
	"github.com/stricture/stricture/internal/adapter/java"
	"github.com/stricture/stricture/internal/adapter/python"
	"github.com/stricture/stricture/internal/adapter/typescript"
	"github.com/stricture/stricture/internal/config"
	"github.com/stricture/stricture/internal/fix"
	"github.com/stricture/stricture/internal/lineage"
	"github.com/stricture/stricture/internal/model"
	"github.com/stricture/stricture/internal/plugins"
	"github.com/stricture/stricture/internal/rules/arch"
	"github.com/stricture/stricture/internal/rules/conv"
	"github.com/stricture/stricture/internal/rules/ctr"
	"github.com/stricture/stricture/internal/rules/tq"
	"github.com/stricture/stricture/internal/suppression"
	"gopkg.in/yaml.v3"
)

var version = "0.1.0-dev"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(0)
	}

	switch os.Args[1] {
	case "init":
		runInit(os.Args[2:])
	case "fix":
		runFix(os.Args[2:])
	case "inspect":
		runInspect(os.Args[2:])
	case "inspect-lineage":
		runInspectLineage(os.Args[2:])
	case "lineage-export":
		runLineageExport(os.Args[2:])
	case "lineage-diff":
		runLineageDiff(os.Args[2:])
	case "lineage-escalate":
		runLineageEscalate(os.Args[2:])
	case "list-rules":
		runListRules()
	case "explain":
		runExplain(os.Args[2:])
	case "validate-config":
		runValidateConfig(os.Args[2:])
	case "lint":
		runLint(os.Args[2:])
	case "--version", "-version", "version":
		fmt.Printf("stricture version %s\n", version)
	case "--help", "-help", "help":
		printUsage()
	default:
		first := os.Args[1]
		// No subcommand: treat all args as lint when first arg is a flag or likely path.
		if strings.HasPrefix(first, "-") || looksLikePathArg(first) {
			runLint(os.Args[1:])
			return
		}
		printUnknownCommand(first)
		os.Exit(2)
	}
}

// printUsage prints the top-level help message.
func printUsage() {
	fmt.Println("Stricture — A fast, language-agnostic linter")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  stricture [command] [options] [paths...]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  lint              Lint files (default when no command given)")
	fmt.Println("  fix               Apply auto-fixes for fixable violations")
	fmt.Println("  init              Create a default .stricture.yml")
	fmt.Println("  inspect <file>    Parse a file and print its UnifiedFileModel as JSON")
	fmt.Println("  inspect-lineage   Parse stricture-source annotations from a file")
	fmt.Println("  lineage-export    Build normalized lineage artifact from source files")
	fmt.Println("  lineage-diff      Diff two lineage artifacts and classify drift severity")
	fmt.Println("  lineage-escalate  Resolve emergency contacts upstream from a service")
	fmt.Println("  list-rules        List all registered rules")
	fmt.Println("  explain           Show details for a specific rule")
	fmt.Println("  validate-config   Check that a .stricture.yml file is valid")
	fmt.Println("  version           Print version and exit")
	fmt.Println("  help              Print this help message")
	fmt.Println()
	fmt.Println("Run 'stricture <command> --help' for details on a specific command.")
}

func printUnknownCommand(command string) {
	fmt.Fprintf(os.Stderr, "Error: unknown command %q\n", command)
	fmt.Fprintln(os.Stderr, "Valid commands: lint, fix, init, inspect, inspect-lineage, lineage-export, lineage-diff, lineage-escalate, list-rules, explain, validate-config, version, help")
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

type repeatableFlag []string

func (f *repeatableFlag) String() string {
	return strings.Join(*f, ",")
}

func (f *repeatableFlag) Set(value string) error {
	for _, token := range strings.Split(value, ",") {
		trimmed := strings.TrimSpace(token)
		if trimmed == "" {
			continue
		}
		*f = append(*f, trimmed)
	}
	return nil
}

func (f *repeatableFlag) Values() []string {
	out := make([]string, len(*f))
	copy(out, *f)
	return out
}

// runLint is the default lint subcommand.
func runLint(args []string) {
	fs := flag.NewFlagSet("lint", flag.ExitOnError)
	format := fs.String("format", "text", "Output format (text, json, sarif, junit)")
	configPath := fs.String("config", ".stricture.yml", "Path to configuration file")
	noConfig := fs.Bool("no-config", false, "Ignore config file and use built-in defaults")
	var ruleFilters repeatableFlag
	fs.Var(&ruleFilters, "rule", "Run a single rule by ID (can be repeated)")
	category := fs.String("category", "", "Run all rules in a category")
	extFilter := fs.String("ext", "", "Only lint files with this extension (example: .go or .ts)")
	severityLevel := fs.String("severity", "", "Only report violations at this level or above (error, warn)")
	quiet := fs.Bool("quiet", false, "Only show errors, not warnings")
	forceColor := fs.Bool("color", false, "Force color output in text format")
	forceNoColor := fs.Bool("no-color", false, "Disable color output in text format")
	verbose := fs.Bool("verbose", false, "Show rule timing and debug info")
	concurrency := fs.Int("concurrency", runtime.NumCPU(), "Max parallel file processing")
	outputPath := fs.String("output", "", "Write report to file instead of stdout")
	maxViolations := fs.Int("max-violations", 0, "Stop after N violations (0 = unlimited)")
	baselinePath := fs.String("baseline", "", "Path to baseline file (existing violations are suppressed; missing file bootstraps baseline)")
	diffMode := fs.Bool("diff", false, "When used with --baseline, include added/resolved diff details against baseline")
	changedOnly := fs.Bool("changed", false, "Lint only changed files in git working tree/index")
	stagedOnly := fs.Bool("staged", false, "Lint only staged files in git index")
	fixApply := fs.Bool("fix", false, "Apply auto-fixes for fixable violations")
	fixDryRun := fs.Bool("fix-dry-run", false, "Show what --fix would change without modifying files")
	fixBackup := fs.Bool("fix-backup", false, "When used with --fix, create .bak files before modifying sources")
	cacheEnabled := fs.Bool("cache", false, "Enable caching (default behavior)")
	noCache := fs.Bool("no-cache", false, "Disable caching")
	_ = fs.Parse(args)

	if *fixApply && *fixDryRun {
		fmt.Fprintln(os.Stderr, "Error: --fix and --fix-dry-run are mutually exclusive")
		os.Exit(2)
	}
	if *fixBackup && !*fixApply {
		fmt.Fprintln(os.Stderr, "Error: --fix-backup requires --fix")
		os.Exit(2)
	}
	if *changedOnly && *stagedOnly {
		fmt.Fprintln(os.Stderr, "Error: --changed and --staged are mutually exclusive")
		os.Exit(2)
	}
	if *diffMode && strings.TrimSpace(*baselinePath) == "" {
		fmt.Fprintln(os.Stderr, "Error: --diff requires --baseline")
		os.Exit(2)
	}

	validFormats := map[string]bool{"text": true, "json": true, "sarif": true, "junit": true}
	if !validFormats[*format] {
		fmt.Fprintf(os.Stderr, "Error: invalid format %q (valid: text, json, sarif, junit)\n", *format)
		os.Exit(2)
	}
	if *maxViolations < 0 {
		fmt.Fprintln(os.Stderr, "Error: --max-violations must be >= 0")
		os.Exit(2)
	}
	if *forceColor && *forceNoColor {
		fmt.Fprintln(os.Stderr, "Error: --color and --no-color are mutually exclusive")
		os.Exit(2)
	}
	if *cacheEnabled && *noCache {
		fmt.Fprintln(os.Stderr, "Error: --cache and --no-cache are mutually exclusive")
		os.Exit(2)
	}
	if *concurrency < 1 {
		fmt.Fprintln(os.Stderr, "Error: --concurrency must be >= 1")
		os.Exit(2)
	}
	cacheActive := !*noCache
	if *cacheEnabled {
		cacheActive = true
	}
	minSeverity := strings.ToLower(strings.TrimSpace(*severityLevel))
	switch minSeverity {
	case "", "warn", "error":
		// Valid values.
	default:
		fmt.Fprintf(os.Stderr, "Error: invalid severity %q (valid: error, warn)\n", *severityLevel)
		os.Exit(2)
	}
	if *quiet {
		if minSeverity != "" && minSeverity != "error" {
			fmt.Fprintln(os.Stderr, "Error: --quiet cannot be combined with --severity=warn")
			os.Exit(2)
		}
		minSeverity = "error"
	}
	extensionAllowlist, err := parseExtensionFilter(*extFilter)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(2)
	}

	registry := buildRegistry()

	cfg := config.Default()
	resolvedConfigPath := resolveConfigPath(*configPath)
	if !*noConfig {
		if loaded, err := config.Load(resolvedConfigPath); err == nil {
			cfg = loaded
		} else if !errors.Is(err, model.ErrConfigNotFound) {
			fmt.Fprintf(os.Stderr, "Error: invalid config %s: %v\n", resolvedConfigPath, err)
			os.Exit(1)
		}

		if len(cfg.Plugins) > 0 {
			pluginPaths := resolvePluginPaths(resolvedConfigPath, cfg.Plugins)
			pluginRules, err := plugins.Load(pluginPaths)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: load plugins: %v\n", err)
				os.Exit(2)
			}
			for _, r := range pluginRules {
				registry.Register(r)
			}
		}

		if unknown := config.UnknownRuleIDs(cfg, registry); len(unknown) > 0 {
			fmt.Fprintf(os.Stderr, "Warning: ignoring %d unknown rule(s): %s\n", len(unknown), strings.Join(unknown, ", "))
		}
	}

	selectedRules, err := resolveLintRules(registry, cfg, ruleFilters.Values(), *category)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(2)
	}

	paths := fs.Args()
	if len(paths) == 0 {
		paths = []string{"."}
	}
	baselineConfigured := strings.TrimSpace(*baselinePath) != ""
	effectiveMaxViolations := *maxViolations
	if baselineConfigured {
		// Baseline filtering happens after rule evaluation; disabling early stop avoids
		// missing non-baselined findings when initial matches are all suppressed.
		effectiveMaxViolations = 0
	}

	filePaths, err := collectLintFilePaths(paths)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: collect files: %v\n", err)
		os.Exit(1)
	}
	filePaths = filterFilePathsByExtensions(filePaths, extensionAllowlist)
	verbosef(*verbose, "Verbose: collected %d candidate file(s)\n", len(filePaths))
	if *changedOnly || *stagedOnly {
		scoped, err := resolveGitScopedFileSet(*changedOnly, *stagedOnly)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(2)
		}

		cwd, err := os.Getwd()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: get working directory: %v\n", err)
			os.Exit(1)
		}
		filtered := make([]string, 0, len(filePaths))
		for _, p := range filePaths {
			if scoped[pathKeyFromBase(cwd, p)] {
				filtered = append(filtered, p)
			}
		}
		filePaths = filtered
	}
	cacheState := "off"
	if cacheActive {
		cacheState = "on"
	}
	verbosef(*verbose, "Verbose: using %d file(s) after scope filters; rules=%d cache=%s\n", len(filePaths), len(selectedRules), cacheState)

	files, err := buildUnifiedFiles(filePaths)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: parse files: %v\n", err)
		os.Exit(1)
	}

	ctx := &model.ProjectContext{Files: map[string]*model.UnifiedFileModel{}}
	for _, file := range files {
		ctx.Files[file.Path] = file
	}

	start := time.Now()
	violations := runLintRules(files, selectedRules, ctx, effectiveMaxViolations, *concurrency)
	baselineOpts := baselineOptions{BootstrapIfMissing: !*diffMode}
	baselineInfo, err := applyBaseline(strings.TrimSpace(*baselinePath), &violations, baselineOpts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(2)
	}
	violations = filterViolationsBySeverity(violations, minSeverity)
	elapsed := time.Since(start).Milliseconds()

	fixOps := make([]fix.Operation, 0)
	if *fixApply || *fixDryRun {
		planned, err := fix.Plan(violations)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: build fix plan: %v\n", err)
			os.Exit(1)
		}
		fixOps = planned

		if *fixApply && len(fixOps) > 0 {
			if *fixBackup {
				if err := writeFixBackups(fixOps); err != nil {
					fmt.Fprintf(os.Stderr, "Error: create fix backups: %v\n", err)
					os.Exit(1)
				}
			}
			if err := fix.Apply(fixOps); err != nil {
				fmt.Fprintf(os.Stderr, "Error: apply fixes: %v\n", err)
				os.Exit(1)
			}

			rewrittenPaths := rewritePathsAfterFix(paths, fixOps)
			filePaths, err = collectLintFilePaths(rewrittenPaths)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: collect files after fix: %v\n", err)
				os.Exit(1)
			}
			files, err = buildUnifiedFiles(filePaths)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: parse files after fix: %v\n", err)
				os.Exit(1)
			}
			ctx = &model.ProjectContext{Files: map[string]*model.UnifiedFileModel{}}
			for _, file := range files {
				ctx.Files[file.Path] = file
			}
			violations = runLintRules(files, selectedRules, ctx, effectiveMaxViolations, *concurrency)
			baselineInfo, err = applyBaseline(strings.TrimSpace(*baselinePath), &violations, baselineOpts)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(2)
			}
			violations = filterViolationsBySeverity(violations, minSeverity)
		}
	}

	sort.Slice(violations, func(i, j int) bool {
		if violations[i].FilePath != violations[j].FilePath {
			return violations[i].FilePath < violations[j].FilePath
		}
		if violations[i].StartLine != violations[j].StartLine {
			return violations[i].StartLine < violations[j].StartLine
		}
		return violations[i].RuleID < violations[j].RuleID
	})
	if *maxViolations > 0 && len(violations) > *maxViolations {
		violations = violations[:*maxViolations]
	}

	filesWithIssues := map[string]bool{}
	errorCount := 0
	warnCount := 0
	for _, v := range violations {
		filesWithIssues[v.FilePath] = true
		switch strings.ToLower(v.Severity) {
		case "error":
			errorCount++
		case "warn", "warning":
			warnCount++
		}
	}

	summary := map[string]interface{}{
		"filesChecked":    len(files),
		"filesWithIssues": len(filesWithIssues),
		"totalViolations": len(violations),
		"errors":          errorCount,
		"warnings":        warnCount,
		"elapsedMs":       elapsed,
	}
	if baselineInfo.Enabled {
		summary["baselinePath"] = filepath.ToSlash(baselineInfo.Path)
		summary["baselineSuppressed"] = baselineInfo.Suppressed
		summary["baselineBootstrapped"] = baselineInfo.Bootstrapped
	}
	if *diffMode {
		summary["diffEnabled"] = true
		summary["diffAdded"] = len(baselineInfo.Added)
		summary["diffResolved"] = len(baselineInfo.Resolved)
	}
	verbosef(*verbose, "Verbose: lint complete in %dms (violations=%d errors=%d warnings=%d)\n", elapsed, len(violations), errorCount, warnCount)

	var report []byte
	colorEnabled := shouldUseColor(*forceColor, *forceNoColor, strings.TrimSpace(*outputPath))
	switch *format {
	case "json", "sarif", "junit":
		payload := map[string]interface{}{
			"version":    "1",
			"violations": violations,
			"summary":    summary,
		}
		if baselineInfo.Enabled {
			payload["baseline"] = map[string]interface{}{
				"path":         filepath.ToSlash(baselineInfo.Path),
				"suppressed":   baselineInfo.Suppressed,
				"bootstrapped": baselineInfo.Bootstrapped,
				"entryCount":   baselineInfo.EntryCount,
			}
		}
		if *diffMode {
			payload["diff"] = map[string]interface{}{
				"enabled":  true,
				"added":    baselineInfo.Added,
				"resolved": baselineInfo.Resolved,
				"summary": map[string]int{
					"added":    len(baselineInfo.Added),
					"resolved": len(baselineInfo.Resolved),
				},
			}
		}
		if *fixApply || *fixDryRun {
			payload["fixes"] = renderFixOperations(fixOps)
			payload["fixMode"] = map[string]bool{
				"apply":   *fixApply,
				"dryRun":  *fixDryRun,
				"applied": *fixApply,
			}
		}
		encoded, err := json.MarshalIndent(payload, "", "  ")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: write %s output: %v\n", *format, err)
			os.Exit(1)
		}
		report = append(encoded, '\n')
	default:
		var out strings.Builder
		if baselineInfo.Enabled {
			if baselineInfo.Bootstrapped {
				fmt.Fprintf(&out, "Baseline created at %s with %d entry(s); existing violations suppressed.\n", baselineInfo.Path, baselineInfo.EntryCount)
			} else if baselineInfo.Suppressed > 0 {
				fmt.Fprintf(&out, "Baseline suppressed %d violation(s) from %s.\n", baselineInfo.Suppressed, baselineInfo.Path)
			}
		}
		if *diffMode {
			fmt.Fprintf(&out, "Diff: added=%d resolved=%d (baseline=%s)\n", len(baselineInfo.Added), len(baselineInfo.Resolved), baselineInfo.Path)
		}
		if *fixApply || *fixDryRun {
			out.WriteString(formatFixSummary(fixOps, *fixDryRun))
		}

		if len(violations) == 0 {
			fmt.Fprintln(&out, "No violations found.")
		} else {
			for _, v := range violations {
				severityLabel := strings.ToUpper(v.Severity)
				severityLabel = colorizeSeverityLabel(v.Severity, severityLabel, colorEnabled)
				fmt.Fprintf(&out, "%s:%d: %s %s: %s\n", v.FilePath, v.StartLine, severityLabel, v.RuleID, v.Message)
			}
		}
		fmt.Fprintf(&out, "Summary: files=%d issues=%d violations=%d errors=%d warnings=%d elapsedMs=%d\n",
			summary["filesChecked"], summary["filesWithIssues"], summary["totalViolations"], summary["errors"], summary["warnings"], summary["elapsedMs"])
		report = []byte(out.String())
	}

	targetOutput := strings.TrimSpace(*outputPath)
	if targetOutput == "" {
		if _, err := os.Stdout.Write(report); err != nil {
			fmt.Fprintf(os.Stderr, "Error: write output: %v\n", err)
			os.Exit(1)
		}
	} else {
		if err := os.MkdirAll(filepath.Dir(targetOutput), 0o755); err != nil {
			fmt.Fprintf(os.Stderr, "Error: create output directory for %s: %v\n", targetOutput, err)
			os.Exit(1)
		}
		if err := os.WriteFile(targetOutput, report, 0o644); err != nil {
			fmt.Fprintf(os.Stderr, "Error: write output file %s: %v\n", targetOutput, err)
			os.Exit(1)
		}
	}

	if errorCount > 0 {
		os.Exit(1)
	}
}

func runFix(args []string) {
	runLint(append([]string{"--fix"}, args...))
}

func runInit(args []string) {
	fs := flag.NewFlagSet("init", flag.ExitOnError)
	force := fs.Bool("force", false, "Overwrite existing .stricture.yml if it exists")
	pathValue := fs.String("path", ".stricture.yml", "Destination config path")
	fs.Usage = func() {
		fmt.Println("Usage: stricture init [options]")
		fmt.Println()
		fmt.Println("Create a default .stricture.yml with recommended settings.")
		fs.PrintDefaults()
	}
	_ = fs.Parse(args)

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

type baselineState struct {
	Enabled      bool
	Path         string
	EntryCount   int
	Suppressed   int
	Bootstrapped bool
	Entries      []baselineEntry
	Added        []model.Violation
	Resolved     []baselineEntry
}

type baselineOptions struct {
	BootstrapIfMissing bool
}

type baselineFile struct {
	Version     string          `json:"version"`
	GeneratedAt string          `json:"generatedAt"`
	Entries     []baselineEntry `json:"entries"`
}

type baselineEntry struct {
	RuleID    string `json:"ruleId"`
	FilePath  string `json:"filePath"`
	StartLine int    `json:"startLine"`
	Message   string `json:"message"`
}

func applyBaseline(pathValue string, violations *[]model.Violation, options baselineOptions) (baselineState, error) {
	state := baselineState{}
	if strings.TrimSpace(pathValue) == "" {
		return state, nil
	}
	if violations == nil {
		return state, fmt.Errorf("internal baseline error: violations pointer is nil")
	}

	state.Enabled = true
	state.Path = pathValue

	data, err := os.ReadFile(pathValue)
	if err != nil {
		if !os.IsNotExist(err) {
			return state, fmt.Errorf("read baseline %s: %w", pathValue, err)
		}
		if !options.BootstrapIfMissing {
			return state, fmt.Errorf("baseline %s does not exist (run once without --diff to bootstrap)", pathValue)
		}

		entries := make([]baselineEntry, 0, len(*violations))
		for _, v := range *violations {
			entries = append(entries, baselineEntry{
				RuleID:    strings.TrimSpace(v.RuleID),
				FilePath:  filepath.ToSlash(v.FilePath),
				StartLine: v.StartLine,
				Message:   strings.TrimSpace(v.Message),
			})
		}
		sort.Slice(entries, func(i, j int) bool {
			if entries[i].FilePath != entries[j].FilePath {
				return entries[i].FilePath < entries[j].FilePath
			}
			if entries[i].StartLine != entries[j].StartLine {
				return entries[i].StartLine < entries[j].StartLine
			}
			if entries[i].RuleID != entries[j].RuleID {
				return entries[i].RuleID < entries[j].RuleID
			}
			return entries[i].Message < entries[j].Message
		})

		doc := baselineFile{
			Version:     "1",
			GeneratedAt: time.Now().UTC().Format(time.RFC3339),
			Entries:     entries,
		}
		encoded, err := json.MarshalIndent(doc, "", "  ")
		if err != nil {
			return state, fmt.Errorf("marshal baseline %s: %w", pathValue, err)
		}
		encoded = append(encoded, '\n')
		if err := os.MkdirAll(filepath.Dir(pathValue), 0o755); err != nil {
			return state, fmt.Errorf("create baseline directory for %s: %w", pathValue, err)
		}
		if err := os.WriteFile(pathValue, encoded, 0o644); err != nil {
			return state, fmt.Errorf("write baseline %s: %w", pathValue, err)
		}

		state.EntryCount = len(entries)
		state.Suppressed = len(*violations)
		state.Bootstrapped = true
		state.Entries = entries
		state.Resolved = []baselineEntry{}
		state.Added = []model.Violation{}
		*violations = []model.Violation{}
		return state, nil
	}

	var doc baselineFile
	if err := json.Unmarshal(data, &doc); err != nil {
		return state, fmt.Errorf("parse baseline %s: %w", pathValue, err)
	}

	lookup := map[string]bool{}
	for _, entry := range doc.Entries {
		lookup[baselineKeyFromEntry(entry)] = true
	}

	rawCurrent := append([]model.Violation(nil), (*violations)...)
	filtered := make([]model.Violation, 0, len(*violations))
	for _, v := range *violations {
		if lookup[baselineKeyFromViolation(v)] {
			state.Suppressed++
			continue
		}
		filtered = append(filtered, v)
	}

	state.EntryCount = len(doc.Entries)
	state.Entries = append([]baselineEntry(nil), doc.Entries...)
	state.Added = append([]model.Violation(nil), filtered...)
	state.Resolved = baselineResolvedEntries(rawCurrent, doc.Entries)
	*violations = filtered
	return state, nil
}

func baselineResolvedEntries(current []model.Violation, entries []baselineEntry) []baselineEntry {
	currentLookup := map[string]bool{}
	for _, v := range current {
		currentLookup[baselineKeyFromViolation(v)] = true
	}

	out := make([]baselineEntry, 0)
	for _, entry := range entries {
		if currentLookup[baselineKeyFromEntry(entry)] {
			continue
		}
		out = append(out, entry)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].FilePath != out[j].FilePath {
			return out[i].FilePath < out[j].FilePath
		}
		if out[i].StartLine != out[j].StartLine {
			return out[i].StartLine < out[j].StartLine
		}
		if out[i].RuleID != out[j].RuleID {
			return out[i].RuleID < out[j].RuleID
		}
		return out[i].Message < out[j].Message
	})
	return out
}

func baselineKeyFromEntry(entry baselineEntry) string {
	return fmt.Sprintf("%s|%s|%d|%s",
		strings.TrimSpace(entry.RuleID),
		filepath.ToSlash(strings.TrimSpace(entry.FilePath)),
		entry.StartLine,
		strings.TrimSpace(entry.Message))
}

func baselineKeyFromViolation(v model.Violation) string {
	return fmt.Sprintf("%s|%s|%d|%s",
		strings.TrimSpace(v.RuleID),
		filepath.ToSlash(strings.TrimSpace(v.FilePath)),
		v.StartLine,
		strings.TrimSpace(v.Message))
}

func resolveLintRules(registry *model.RuleRegistry, cfg *config.Config, requestedRules []string, category string) ([]model.Rule, error) {
	selected := make([]model.Rule, 0)
	targetCategory := strings.ToLower(strings.TrimSpace(category))

	ruleFilter := map[string]bool{}
	for _, raw := range requestedRules {
		id := strings.TrimSpace(raw)
		if id == "" {
			continue
		}
		if _, ok := registry.ByID(id); !ok {
			return nil, fmt.Errorf("unknown rule %q", id)
		}
		ruleFilter[id] = true
	}
	hasRuleFilter := len(ruleFilter) > 0

	candidates := make([]model.Rule, 0)
	switch {
	case hasRuleFilter:
		ids := make([]string, 0, len(ruleFilter))
		for id := range ruleFilter {
			ids = append(ids, id)
		}
		sort.Strings(ids)
		for _, id := range ids {
			r, _ := registry.ByID(id)
			candidates = append(candidates, r)
		}
	case cfg != nil && len(cfg.Rules) > 0:
		ids := make([]string, 0, len(cfg.Rules))
		for id := range cfg.Rules {
			ids = append(ids, id)
		}
		sort.Strings(ids)
		for _, id := range ids {
			if r, ok := registry.ByID(id); ok {
				candidates = append(candidates, r)
			}
		}
	default:
		candidates = append(candidates, registry.All()...)
	}

	for _, r := range candidates {
		if hasRuleFilter && !ruleFilter[r.ID()] {
			continue
		}
		if targetCategory != "" && strings.ToLower(r.Category()) != targetCategory {
			continue
		}

		ruleCfg := model.RuleConfig{
			Severity: r.DefaultSeverity(),
			Options:  map[string]interface{}{},
		}
		if cfg != nil {
			if override, ok := cfg.Rules[r.ID()]; ok {
				if strings.TrimSpace(override.Severity) != "" {
					ruleCfg.Severity = override.Severity
				}
				if override.Options != nil {
					ruleCfg.Options = override.Options
				}
			}
		}
		if strings.EqualFold(ruleCfg.Severity, "off") {
			continue
		}

		selected = append(selected, lintRuleWithConfig{Rule: r, Config: ruleCfg})
	}

	return selected, nil
}

type lintRuleWithConfig struct {
	model.Rule
	Config model.RuleConfig
}

func rewritePathsAfterFix(paths []string, ops []fix.Operation) []string {
	renames := map[string]string{}
	for _, op := range ops {
		if op.Kind != "rename" {
			continue
		}
		renames[filepath.Clean(op.Path)] = filepath.Clean(op.NewPath)
	}
	if len(renames) == 0 {
		return append([]string(nil), paths...)
	}

	rewritten := make([]string, 0, len(paths))
	for _, pathValue := range paths {
		clean := filepath.Clean(pathValue)
		if newPath, ok := renames[clean]; ok {
			rewritten = append(rewritten, newPath)
			continue
		}
		rewritten = append(rewritten, pathValue)
	}
	return rewritten
}

func renderFixOperations(ops []fix.Operation) []map[string]string {
	out := make([]map[string]string, 0, len(ops))
	for _, op := range ops {
		entry := map[string]string{
			"ruleId":      op.RuleID,
			"kind":        op.Kind,
			"path":        filepath.ToSlash(op.Path),
			"description": op.Description,
		}
		if op.NewPath != "" {
			entry["newPath"] = filepath.ToSlash(op.NewPath)
		}
		out = append(out, entry)
	}
	return out
}

func formatFixSummary(ops []fix.Operation, dryRun bool) string {
	var out strings.Builder
	mode := "apply"
	if dryRun {
		mode = "dry-run"
	}
	fmt.Fprintf(&out, "Fixes: %d operation(s) (%s)\n", len(ops), mode)
	for _, op := range ops {
		fmt.Fprintf(&out, "  - [%s] %s\n", op.RuleID, op.Description)
	}
	return out.String()
}

func printFixSummary(ops []fix.Operation, dryRun bool) {
	fmt.Print(formatFixSummary(ops, dryRun))
}

func writeFixBackups(ops []fix.Operation) error {
	paths := map[string]bool{}
	for _, op := range ops {
		switch op.Kind {
		case "edit", "rename":
			if strings.TrimSpace(op.Path) != "" {
				paths[filepath.Clean(op.Path)] = true
			}
		}
	}

	ordered := make([]string, 0, len(paths))
	for pathValue := range paths {
		ordered = append(ordered, pathValue)
	}
	sort.Strings(ordered)

	for _, pathValue := range ordered {
		data, err := os.ReadFile(pathValue)
		if err != nil {
			return fmt.Errorf("read %s for backup: %w", pathValue, err)
		}

		backupPath := pathValue + ".bak"
		if _, err := os.Stat(backupPath); err == nil {
			return fmt.Errorf("backup already exists: %s", backupPath)
		} else if !os.IsNotExist(err) {
			return fmt.Errorf("check backup %s: %w", backupPath, err)
		}

		if err := os.WriteFile(backupPath, data, 0o644); err != nil {
			return fmt.Errorf("write backup %s: %w", backupPath, err)
		}
	}
	return nil
}

func verbosef(enabled bool, format string, args ...interface{}) {
	if !enabled {
		return
	}
	fmt.Fprintf(os.Stderr, format, args...)
}

func shouldUseColor(forceColor bool, forceNoColor bool, outputPath string) bool {
	if forceNoColor {
		return false
	}
	if forceColor {
		return true
	}
	if strings.TrimSpace(outputPath) != "" {
		return false
	}
	term := strings.ToLower(strings.TrimSpace(os.Getenv("TERM")))
	if term == "" || term == "dumb" {
		return false
	}
	info, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return info.Mode()&os.ModeCharDevice != 0
}

func colorizeSeverityLabel(rawSeverity string, label string, enabled bool) string {
	if !enabled {
		return label
	}
	switch strings.ToLower(strings.TrimSpace(rawSeverity)) {
	case "error":
		return "\x1b[31m" + label + "\x1b[0m"
	case "warn", "warning":
		return "\x1b[33m" + label + "\x1b[0m"
	default:
		return label
	}
}

func parseExtensionFilter(raw string) (map[string]bool, error) {
	filter := map[string]bool{}
	value := strings.TrimSpace(raw)
	if value == "" {
		return filter, nil
	}

	for _, token := range strings.Split(value, ",") {
		normalized := strings.ToLower(strings.TrimSpace(token))
		if normalized == "" {
			continue
		}
		if !strings.HasPrefix(normalized, ".") {
			normalized = "." + normalized
		}
		if normalized == "." {
			continue
		}
		filter[normalized] = true
	}

	if len(filter) == 0 {
		return nil, fmt.Errorf("invalid --ext %q (expected value like .go or ts)", raw)
	}
	return filter, nil
}

func filterFilePathsByExtensions(paths []string, allowlist map[string]bool) []string {
	if len(allowlist) == 0 {
		return paths
	}
	filtered := make([]string, 0, len(paths))
	for _, pathValue := range paths {
		ext := strings.ToLower(filepath.Ext(pathValue))
		if allowlist[ext] {
			filtered = append(filtered, pathValue)
		}
	}
	return filtered
}

func collectLintFilePaths(paths []string) ([]string, error) {
	files := make([]string, 0)
	seen := map[string]bool{}
	projectRoot := currentProjectRoot()

	for _, raw := range paths {
		pathValue := strings.TrimSpace(raw)
		if pathValue == "" {
			continue
		}

		info, err := os.Stat(pathValue)
		if err != nil {
			return nil, err
		}

		if !info.IsDir() {
			if isLintSourceFile(pathValue) {
				outside, err := symlinkResolvesOutsideProject(pathValue, projectRoot)
				if err != nil {
					return nil, err
				}
				if outside {
					continue
				}
				canonical := filepath.ToSlash(pathValue)
				if !seen[canonical] {
					seen[canonical] = true
					files = append(files, canonical)
				}
			}
			continue
		}

		err = filepath.WalkDir(pathValue, func(current string, entry fs.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if entry.IsDir() {
				if shouldSkipLintDir(current) {
					return filepath.SkipDir
				}
				return nil
			}
			if !isLintSourceFile(current) {
				return nil
			}
			outside, err := symlinkResolvesOutsideProject(current, projectRoot)
			if err != nil {
				return err
			}
			if outside {
				return nil
			}

			canonical := filepath.ToSlash(current)
			if !seen[canonical] {
				seen[canonical] = true
				files = append(files, canonical)
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
	}

	sort.Strings(files)
	return files, nil
}

func resolveGitScopedFileSet(changedOnly bool, stagedOnly bool) (map[string]bool, error) {
	rootRaw, err := gitOutput("rev-parse", "--show-toplevel")
	if err != nil {
		return nil, fmt.Errorf("git-scoped lint requires a git repository: %w", err)
	}
	root := strings.TrimSpace(rootRaw)
	if root == "" {
		return nil, fmt.Errorf("unable to resolve git repository root")
	}

	combined := make([]string, 0)
	switch {
	case stagedOnly:
		staged, err := gitOutputLines("diff", "--name-only", "--cached", "--diff-filter=ACMRT")
		if err != nil {
			return nil, err
		}
		combined = append(combined, staged...)
	case changedOnly:
		working, err := gitOutputLines("diff", "--name-only", "--diff-filter=ACMRT")
		if err != nil {
			return nil, err
		}
		staged, err := gitOutputLines("diff", "--name-only", "--cached", "--diff-filter=ACMRT")
		if err != nil {
			return nil, err
		}
		untracked, err := gitOutputLines("ls-files", "--others", "--exclude-standard")
		if err != nil {
			return nil, err
		}
		combined = append(combined, working...)
		combined = append(combined, staged...)
		combined = append(combined, untracked...)
	default:
		return map[string]bool{}, nil
	}

	out := map[string]bool{}
	for _, rel := range combined {
		r := strings.TrimSpace(rel)
		if r == "" {
			continue
		}
		out[pathKeyFromBase(root, r)] = true
	}
	return out, nil
}

func gitOutputLines(args ...string) ([]string, error) {
	out, err := gitOutput(args...)
	if err != nil {
		return nil, err
	}
	lines := strings.Split(out, "\n")
	trimmed := make([]string, 0, len(lines))
	for _, line := range lines {
		v := strings.TrimSpace(line)
		if v == "" {
			continue
		}
		trimmed = append(trimmed, v)
	}
	return trimmed, nil
}

func gitOutput(args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		msg := strings.TrimSpace(string(out))
		if msg == "" {
			msg = err.Error()
		}
		return "", fmt.Errorf("git %s failed: %s", strings.Join(args, " "), msg)
	}
	return string(out), nil
}

func pathKeyFromBase(baseDir string, pathValue string) string {
	candidate := pathValue
	if !filepath.IsAbs(candidate) {
		candidate = filepath.Join(baseDir, candidate)
	}
	abs, err := filepath.Abs(candidate)
	if err == nil {
		candidate = abs
	}
	if resolved, err := filepath.EvalSymlinks(candidate); err == nil {
		candidate = resolved
	}
	return filepath.ToSlash(filepath.Clean(candidate))
}

func currentProjectRoot() string {
	wd, err := os.Getwd()
	if err != nil {
		return ""
	}
	abs, err := filepath.Abs(wd)
	if err != nil {
		return filepath.Clean(wd)
	}
	resolved, err := filepath.EvalSymlinks(abs)
	if err == nil {
		return filepath.Clean(resolved)
	}
	return filepath.Clean(abs)
}

func symlinkResolvesOutsideProject(pathValue string, projectRoot string) (bool, error) {
	if strings.TrimSpace(projectRoot) == "" {
		return false, nil
	}

	info, err := os.Lstat(pathValue)
	if err != nil {
		return false, err
	}
	if info.Mode()&os.ModeSymlink == 0 {
		return false, nil
	}

	resolved, err := filepath.EvalSymlinks(pathValue)
	if err != nil {
		// Broken or inaccessible symlink should not be linted.
		return true, nil
	}
	absResolved, err := filepath.Abs(resolved)
	if err != nil {
		return false, err
	}
	return !isPathWithinRoot(absResolved, projectRoot), nil
}

func isPathWithinRoot(pathValue string, root string) bool {
	rel, err := filepath.Rel(root, pathValue)
	if err != nil {
		return false
	}
	if rel == "." {
		return true
	}
	if rel == ".." {
		return false
	}
	return !strings.HasPrefix(rel, ".."+string(filepath.Separator))
}

func shouldSkipLintDir(dir string) bool {
	base := filepath.Base(dir)
	switch base {
	case ".git", "node_modules", "bin", ".stricture-cache", "docs", "tests":
		return true
	}

	normalized := filepath.ToSlash(dir)
	return strings.Contains(normalized, "tests/fixtures") || strings.Contains(normalized, "tests/benchmark")
}

func isLintSourceFile(path string) bool {
	if isGeneratedSourceFile(path) {
		return false
	}
	switch strings.ToLower(filepath.Ext(path)) {
	case ".go", ".ts", ".tsx", ".js", ".jsx", ".py", ".java", ".kt", ".rs":
		return true
	default:
		return false
	}
}

func isGeneratedSourceFile(path string) bool {
	name := strings.ToLower(filepath.Base(path))
	return strings.Contains(name, ".generated.") ||
		strings.HasSuffix(name, ".pb.go") ||
		strings.HasSuffix(name, ".pb.ts")
}

func buildUnifiedFiles(paths []string) ([]*model.UnifiedFileModel, error) {
	files := make([]*model.UnifiedFileModel, 0, len(paths))
	for _, pathValue := range paths {
		data, err := os.ReadFile(pathValue)
		if err != nil {
			return nil, err
		}

		file := &model.UnifiedFileModel{
			Path:       filepath.ToSlash(pathValue),
			Language:   detectLanguage(pathValue),
			Source:     data,
			LineCount:  countLines(data),
			IsTestFile: looksLikeTestFile(pathValue),
		}
		files = append(files, file)
	}
	return files, nil
}

func countLines(data []byte) int {
	if len(data) == 0 {
		return 0
	}
	count := 1
	for _, b := range data {
		if b == '\n' {
			count++
		}
	}
	return count
}

func looksLikeTestFile(pathValue string) bool {
	name := strings.ToLower(filepath.Base(pathValue))
	return strings.HasSuffix(name, "_test.go") ||
		strings.Contains(name, ".test.") ||
		strings.Contains(name, ".spec.") ||
		strings.HasPrefix(name, "test_") ||
		strings.HasSuffix(name, "test.java")
}

func runLintRules(files []*model.UnifiedFileModel, rules []model.Rule, ctx *model.ProjectContext, maxViolations int, _ int) []model.Violation {
	violations := make([]model.Violation, 0)
	stop := false
	for _, file := range files {
		if stop {
			break
		}
		policy := suppression.Compile(file.Source)
		for _, rawRule := range rules {
			if stop {
				break
			}
			ruleCfg := model.RuleConfig{Severity: rawRule.DefaultSeverity(), Options: map[string]interface{}{}}
			if withCfg, ok := rawRule.(lintRuleWithConfig); ok {
				rawRule = withCfg.Rule
				ruleCfg = withCfg.Config
			}

			func() {
				defer func() {
					if recovered := recover(); recovered != nil {
						violations = append(violations, model.Violation{
							RuleID:    rawRule.ID(),
							Severity:  "error",
							Message:   fmt.Sprintf("Rule panicked: %v", recovered),
							FilePath:  file.Path,
							StartLine: 1,
						})
						if maxViolations > 0 && len(violations) >= maxViolations {
							stop = true
						}
					}
				}()
				rawViolations := rawRule.Check(file, ctx, ruleCfg)
				for _, v := range rawViolations {
					ruleID := strings.TrimSpace(v.RuleID)
					if ruleID == "" {
						ruleID = rawRule.ID()
						v.RuleID = ruleID
					}
					line := v.StartLine
					if line <= 0 {
						line = 1
					}
					if policy.Suppressed(ruleID, line) {
						continue
					}
					violations = append(violations, v)
					if maxViolations > 0 && len(violations) >= maxViolations {
						stop = true
						break
					}
				}
			}()
		}
	}
	return violations
}

func filterViolationsBySeverity(violations []model.Violation, minSeverity string) []model.Violation {
	threshold := strings.ToLower(strings.TrimSpace(minSeverity))
	if threshold == "" {
		return violations
	}
	minRank := severityRank(threshold)
	filtered := make([]model.Violation, 0, len(violations))
	for _, v := range violations {
		if severityRank(v.Severity) >= minRank {
			filtered = append(filtered, v)
		}
	}
	return filtered
}

func severityRank(severity string) int {
	switch strings.ToLower(strings.TrimSpace(severity)) {
	case "warn", "warning":
		return 1
	case "error":
		return 2
	default:
		// Unknown severities are treated as errors to avoid accidental suppression.
		return 2
	}
}

func resolveConfigPath(configPath string) string {
	if strings.TrimSpace(configPath) == "" || filepath.IsAbs(configPath) {
		return configPath
	}

	if _, err := os.Stat(configPath); err == nil {
		return configPath
	}

	wd, err := os.Getwd()
	if err != nil {
		return configPath
	}

	current := wd
	for {
		candidate := filepath.Join(current, configPath)
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
		parent := filepath.Dir(current)
		if parent == current {
			break
		}
		current = parent
	}

	return configPath
}

func resolvePluginPaths(configPath string, pluginPaths []string) []string {
	resolved := make([]string, 0, len(pluginPaths))
	configDir := filepath.Dir(configPath)
	for _, pathValue := range pluginPaths {
		p := strings.TrimSpace(pathValue)
		if p == "" {
			continue
		}
		if filepath.IsAbs(p) {
			resolved = append(resolved, p)
			continue
		}
		if strings.Contains(p, "://") {
			resolved = append(resolved, p)
			continue
		}
		resolved = append(resolved, filepath.Join(configDir, p))
	}
	return resolved
}

// runInspect parses a file and prints its UnifiedFileModel as JSON.
func runInspect(args []string) {
	fs := flag.NewFlagSet("inspect", flag.ExitOnError)
	fs.Usage = func() {
		fmt.Println("Usage: stricture inspect [options] <file>")
		fmt.Println()
		fmt.Println("Parse a file and print its UnifiedFileModel as formatted JSON.")
		fmt.Println("This is a debugging tool for adapter development.")
		fs.PrintDefaults()
	}
	_ = fs.Parse(args)

	if fs.NArg() == 0 {
		fmt.Fprintln(os.Stderr, "Error: inspect requires a file path argument.")
		fs.Usage()
		os.Exit(2)
	}

	filePath := fs.Arg(0)
	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "Error: file not found: %s\n", filePath)
			os.Exit(2)
		}
		fmt.Fprintf(os.Stderr, "Error: cannot read %s: %v\n", filePath, err)
		os.Exit(2)
	}
	if isLikelyBinary(data) {
		fmt.Fprintf(os.Stderr, "Error: cannot inspect binary file: %s\n", filePath)
		os.Exit(2)
	}

	lang := detectLanguage(filePath)
	if lang == "unknown" {
		fmt.Fprintf(os.Stderr, "Error: no language adapter for %q files. Supported: %s\n", filepath.Ext(filePath), strings.Join(supportedInspectLanguages(), ", "))
		os.Exit(2)
	}

	parsed, err := inspectParseFile(filePath, data)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: inspect parse failed for %s: %v\n", filePath, err)
		os.Exit(2)
	}

	out, err := json.MarshalIndent(parsed, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to marshal model: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(string(out))
}

func inspectParseFile(path string, source []byte) (*model.UnifiedFileModel, error) {
	lang := detectLanguage(path)
	cfg := adapter.AdapterConfig{}

	switch lang {
	case "typescript", "javascript":
		return (&typescript.Adapter{}).Parse(path, source, cfg)
	case "python":
		return (&python.Adapter{}).Parse(path, source, cfg)
	case "java":
		return (&java.Adapter{}).Parse(path, source, cfg)
	case "go":
		return parseGoInspect(path, source)
	default:
		return nil, fmt.Errorf("unsupported language %q", lang)
	}
}

func parseGoInspect(path string, source []byte) (*model.UnifiedFileModel, error) {
	fset := token.NewFileSet()
	parsed, err := parser.ParseFile(fset, path, source, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	ufm := &model.UnifiedFileModel{
		Path:       filepath.ToSlash(path),
		Language:   "go",
		Source:     append([]byte(nil), source...),
		LineCount:  countLines(source),
		IsTestFile: strings.HasSuffix(strings.ToLower(filepath.Base(path)), "_test.go"),
		Imports:    []model.ImportDecl{},
		Functions:  []model.FuncModel{},
		Types:      []model.TypeModel{},
	}

	for _, imp := range parsed.Imports {
		importPath := strings.Trim(imp.Path.Value, `"`)
		name := ""
		if imp.Name != nil {
			name = imp.Name.Name
		}
		ufm.Imports = append(ufm.Imports, model.ImportDecl{
			Path:      importPath,
			Alias:     name,
			StartLine: fset.Position(imp.Pos()).Line,
			EndLine:   fset.Position(imp.End()).Line,
		})
	}

	for _, decl := range parsed.Decls {
		switch d := decl.(type) {
		case *ast.FuncDecl:
			fn := model.FuncModel{
				Name:       d.Name.Name,
				IsExported: ast.IsExported(d.Name.Name),
				StartLine:  fset.Position(d.Pos()).Line,
				EndLine:    fset.Position(d.End()).Line,
			}
			ufm.Functions = append(ufm.Functions, fn)
		case *ast.GenDecl:
			if d.Tok != token.TYPE {
				continue
			}
			for _, spec := range d.Specs {
				ts, ok := spec.(*ast.TypeSpec)
				if !ok {
					continue
				}
				ufm.Types = append(ufm.Types, model.TypeModel{
					Name:      ts.Name.Name,
					Exported:  ast.IsExported(ts.Name.Name),
					StartLine: fset.Position(ts.Pos()).Line,
					EndLine:   fset.Position(ts.End()).Line,
				})
			}
		}
	}

	return ufm, nil
}

func isLikelyBinary(data []byte) bool {
	if len(data) == 0 {
		return false
	}
	if !utf8.Valid(data) {
		return true
	}
	sample := data
	if len(sample) > 1024 {
		sample = sample[:1024]
	}
	for _, b := range sample {
		if b == 0x00 {
			return true
		}
	}
	return false
}

func supportedInspectLanguages() []string {
	return []string{"go", "typescript", "javascript", "python", "java"}
}

// runInspectLineage parses stricture-source annotations and prints JSON output.
func runInspectLineage(args []string) {
	fs := flag.NewFlagSet("inspect-lineage", flag.ExitOnError)
	fs.Usage = func() {
		fmt.Println("Usage: stricture inspect-lineage <file>")
		fmt.Println()
		fmt.Println("Parse stricture-source annotations from comments and print them as JSON.")
	}
	_ = fs.Parse(args)

	if fs.NArg() == 0 {
		fmt.Fprintln(os.Stderr, "Error: inspect-lineage requires a file path argument.")
		fs.Usage()
		os.Exit(2)
	}

	filePath := fs.Arg(0)
	data, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: cannot read %s: %v\n", filePath, err)
		os.Exit(1)
	}

	annotations, parseErrs := lineage.Parse(data)
	payload := map[string]interface{}{
		"file":        filePath,
		"annotations": annotations,
		"errors":      parseErrs,
	}

	out, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to serialize lineage result: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(string(out))

	if len(parseErrs) > 0 {
		os.Exit(1)
	}
}

// runLineageExport builds a normalized lineage artifact from source files.
func runLineageExport(args []string) {
	fs := flag.NewFlagSet("lineage-export", flag.ExitOnError)
	outPath := fs.String("out", "", "Write artifact JSON to this path (stdout if empty)")
	strict := fs.Bool("strict", true, "Exit non-zero if parse errors are found")
	fs.Usage = func() {
		fmt.Println("Usage: stricture lineage-export [options] [paths...]")
		fmt.Println()
		fmt.Println("Build a normalized lineage artifact from source files.")
		fs.PrintDefaults()
	}
	_ = fs.Parse(args)

	paths := fs.Args()
	if len(paths) == 0 {
		paths = []string{"."}
	}

	artifact, parseErrs, err := lineage.Collect(paths)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: lineage-export failed: %v\n", err)
		os.Exit(1)
	}

	if *outPath != "" {
		if err := lineage.WriteArtifact(*outPath, artifact); err != nil {
			fmt.Fprintf(os.Stderr, "Error: write artifact: %v\n", err)
			os.Exit(1)
		}
	} else {
		out, err := json.MarshalIndent(artifact, "", "  ")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: marshal artifact: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(string(out))
	}

	if len(parseErrs) > 0 {
		errOut, _ := json.MarshalIndent(parseErrs, "", "  ")
		fmt.Fprintf(os.Stderr, "Lineage parse errors (%d):\n%s\n", len(parseErrs), string(errOut))
		if *strict {
			os.Exit(1)
		}
	}
}

// runLineageDiff diffs two lineage artifacts and classifies drift severity.
func runLineageDiff(args []string) {
	fs := flag.NewFlagSet("lineage-diff", flag.ExitOnError)
	basePath := fs.String("base", "", "Path to base lineage artifact JSON")
	headPath := fs.String("head", "", "Path to head lineage artifact JSON")
	outPath := fs.String("out", "", "Write diff JSON to this path (stdout if empty)")
	failOn := fs.String("fail-on", "high", "Fail when drift at/above severity (high|medium|low|info|none)")
	modeRaw := fs.String("mode", string(lineage.ModeBlock), "Enforcement mode: block (exit non-zero) or warn (always exit zero)")
	fs.Usage = func() {
		fmt.Println("Usage: stricture lineage-diff --base <file> --head <file> [options]")
		fmt.Println()
		fmt.Println("Diff two lineage artifacts and classify drift severity.")
		fs.PrintDefaults()
	}
	_ = fs.Parse(args)

	if strings.TrimSpace(*basePath) == "" || strings.TrimSpace(*headPath) == "" {
		fmt.Fprintln(os.Stderr, "Error: --base and --head are required")
		fs.Usage()
		os.Exit(2)
	}

	base, err := lineage.LoadArtifact(*basePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: load base artifact: %v\n", err)
		os.Exit(1)
	}
	head, err := lineage.LoadArtifact(*headPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: load head artifact: %v\n", err)
		os.Exit(1)
	}

	result := lineage.DiffArtifacts(base, head)
	threshold, err := lineage.ParseSeverity(*failOn)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(2)
	}
	mode, err := lineage.ParseEnforcementMode(*modeRaw)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(2)
	}

	out, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: marshal diff result: %v\n", err)
		os.Exit(1)
	}

	if *outPath != "" {
		if err := os.WriteFile(*outPath, append(out, '\n'), 0o644); err != nil {
			fmt.Fprintf(os.Stderr, "Error: write diff output: %v\n", err)
			os.Exit(1)
		}
	} else {
		fmt.Println(string(out))
	}

	thresholdExceeded := lineage.ShouldFailAtThreshold(result, threshold)
	if thresholdExceeded && mode == lineage.ModeWarn {
		fmt.Fprintf(os.Stderr, "WARN: drift at/above %s detected, but mode=warn so exit code remains 0\n", threshold)
	}
	if thresholdExceeded && lineage.ShouldFailAtThresholdWithMode(result, threshold, mode) {
		os.Exit(1)
	}
}

// runLineageEscalate resolves emergency contacts upstream from a service.
func runLineageEscalate(args []string) {
	fs := flag.NewFlagSet("lineage-escalate", flag.ExitOnError)
	serviceID := fs.String("service", "", "Service/system ID to investigate")
	artifactPath := fs.String("artifact", "", "Path to lineage artifact JSON")
	systemsPath := fs.String("systems", "", "Path to system registry YAML (optional)")
	maxDepth := fs.Int("max-depth", 8, "Maximum upstream depth to traverse")
	fs.Usage = func() {
		fmt.Println("Usage: stricture lineage-escalate --service <id> --artifact <file> [options]")
		fmt.Println()
		fmt.Println("Show emergency contacts for a service and its upstream dependencies.")
		fs.PrintDefaults()
	}
	_ = fs.Parse(args)

	if strings.TrimSpace(*serviceID) == "" || strings.TrimSpace(*artifactPath) == "" {
		fmt.Fprintln(os.Stderr, "Error: --service and --artifact are required")
		fs.Usage()
		os.Exit(2)
	}

	artifact, err := lineage.LoadArtifact(*artifactPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: load artifact: %v\n", err)
		os.Exit(1)
	}

	registry := lineage.SystemRegistry{}
	if strings.TrimSpace(*systemsPath) != "" {
		registry, err = lineage.LoadSystemRegistry(*systemsPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: load systems registry: %v\n", err)
			os.Exit(1)
		}
	}

	steps, err := lineage.BuildEscalationChain(*serviceID, artifact, registry, *maxDepth)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: build escalation chain: %v\n", err)
		os.Exit(1)
	}

	payload := map[string]interface{}{
		"service": *serviceID,
		"steps":   steps,
	}
	out, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: marshal escalation chain: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(string(out))
}

// runListRules prints a table of all registered rules.
func runListRules() {
	registry := buildRegistry()

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tCATEGORY\tDEFAULT\tFIXABLE\tDESCRIPTION")
	fmt.Fprintln(w, "--\t--------\t-------\t-------\t-----------")
	for _, r := range sortedRulesForDisplay(registry) {
		meta := ruleMetadata(r.ID())
		desc := r.Description()
		if meta.RequiresManifest {
			desc += " (requires manifest)"
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
			r.ID(), strings.ToUpper(r.Category()), r.DefaultSeverity(), meta.Fixability, desc)
	}
	w.Flush()
	fmt.Printf("\n%d rules registered.\n", len(registry.All()))
}

func runExplain(args []string) {
	fs := flag.NewFlagSet("explain", flag.ExitOnError)
	fs.Usage = func() {
		fmt.Println("Usage: stricture explain <rule-id>")
		fmt.Println()
		fmt.Println("Show details for a specific rule.")
	}
	_ = fs.Parse(args)

	if fs.NArg() == 0 {
		fmt.Fprintln(os.Stderr, "Error: explain requires a rule ID argument.")
		fs.Usage()
		os.Exit(2)
	}

	ruleID := strings.TrimSpace(fs.Arg(0))
	registry := buildRegistry()
	ruleDef, ok := registry.ByID(ruleID)
	if !ok {
		fmt.Fprintf(os.Stderr, "Error: unknown rule %q\n", ruleID)
		fmt.Fprintln(os.Stderr, "Run 'stricture list-rules' to see available rules.")
		os.Exit(2)
	}

	meta := ruleMetadata(ruleDef.ID())
	requiresManifest := "No"
	if meta.RequiresManifest {
		requiresManifest = "Yes"
	}

	fmt.Printf("ID: %s\n", ruleDef.ID())
	fmt.Printf("Category: %s\n", strings.ToUpper(ruleDef.Category()))
	fmt.Printf("Default Severity: %s\n", ruleDef.DefaultSeverity())
	fmt.Printf("Fixable: %s\n", meta.Fixability)
	fmt.Printf("Needs Project Context: %t\n", ruleDef.NeedsProjectContext())
	fmt.Printf("Requires Manifest: %s\n", requiresManifest)
	fmt.Printf("Description: %s\n", ruleDef.Description())
	fmt.Printf("Why: %s\n", ruleDef.Why())
}

// runValidateConfig checks that a config file is valid YAML with recognized rule IDs.
func runValidateConfig(args []string) {
	fs := flag.NewFlagSet("validate-config", flag.ExitOnError)
	fs.Usage = func() {
		fmt.Println("Usage: stricture validate-config [path]")
		fmt.Println()
		fmt.Println("Validate a .stricture.yml configuration file.")
		fmt.Println("Checks YAML syntax and verifies all rule IDs are recognized.")
	}
	_ = fs.Parse(args)

	configPath := ".stricture.yml"
	if fs.NArg() > 0 {
		configPath = fs.Arg(0)
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: cannot read %s: %v\n", configPath, err)
		os.Exit(1)
	}

	var cfg struct {
		Rules   map[string]string                 `yaml:"rules"`
		Options map[string]map[string]interface{} `yaml:"options"`
	}
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		fmt.Fprintf(os.Stderr, "Error: invalid YAML in %s: %v\n", configPath, err)
		os.Exit(1)
	}

	registry := buildRegistry()
	var unknown []string
	for id := range cfg.Rules {
		if _, ok := registry.ByID(id); !ok {
			unknown = append(unknown, id)
		}
	}

	if len(unknown) > 0 {
		fmt.Fprintf(os.Stderr, "Warning: %d unrecognized rule(s): %s\n",
			len(unknown), strings.Join(unknown, ", "))
		fmt.Fprintf(os.Stderr, "(These may be valid rules not yet registered in this build.)\n")
	}

	fmt.Printf("Config %s: valid YAML, %d rules configured.\n", configPath, len(cfg.Rules))
}

type ruleMeta struct {
	Fixability       string
	RequiresManifest bool
}

func ruleMetadata(ruleID string) ruleMeta {
	switch ruleID {
	case "CONV-file-header", "CONV-file-naming", "CONV-test-file-location":
		return ruleMeta{Fixability: "Yes"}
	case "TQ-mock-scope":
		return ruleMeta{Fixability: "Partial"}
	case "CTR-strictness-parity", "CTR-manifest-conformance":
		return ruleMeta{Fixability: "No", RequiresManifest: true}
	default:
		return ruleMeta{Fixability: "No"}
	}
}

func sortedRulesForDisplay(registry *model.RuleRegistry) []model.Rule {
	all := append([]model.Rule(nil), registry.All()...)
	sort.SliceStable(all, func(i, j int) bool {
		ci := categoryOrder(strings.ToLower(all[i].Category()))
		cj := categoryOrder(strings.ToLower(all[j].Category()))
		if ci != cj {
			return ci < cj
		}
		return all[i].ID() < all[j].ID()
	})
	return all
}

func categoryOrder(category string) int {
	switch strings.ToLower(category) {
	case "tq":
		return 0
	case "arch":
		return 1
	case "conv":
		return 2
	case "ctr":
		return 3
	default:
		return 4
	}
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

// buildRegistry creates a RuleRegistry with all known rules.
func buildRegistry() *model.RuleRegistry {
	r := model.NewRuleRegistry()

	// CONV
	r.Register(&conv.FileNaming{})
	r.Register(&conv.FileHeader{})
	r.Register(&conv.ErrorFormat{})
	r.Register(&conv.ExportNaming{})
	r.Register(&conv.TestFileLocation{})
	r.Register(&conv.RequiredExports{})

	// ARCH
	r.Register(&arch.DependencyDirection{})
	r.Register(&arch.ImportBoundary{})
	r.Register(&arch.NoCircularDeps{})
	r.Register(&arch.MaxFileLines{})
	r.Register(&arch.LayerViolation{})
	r.Register(&arch.ModuleBoundary{})

	// TQ
	r.Register(&tq.NoShallowAssertions{})
	r.Register(&tq.ReturnTypeVerified{})
	r.Register(&tq.SchemaConformance{})
	r.Register(&tq.ErrorPathCoverage{})
	r.Register(&tq.AssertionDepth{})
	r.Register(&tq.BoundaryTested{})
	r.Register(&tq.MockScope{})
	r.Register(&tq.TestIsolation{})
	r.Register(&tq.NegativeCases{})
	r.Register(&tq.TestNaming{})

	// CTR
	r.Register(&ctr.RequestShape{})
	r.Register(&ctr.ResponseShape{})
	r.Register(&ctr.StatusCodeHandling{})
	r.Register(&ctr.SharedTypeSync{})
	r.Register(&ctr.JSONTagMatch{})
	r.Register(&ctr.DualTest{})
	r.Register(&ctr.StrictnessParity{})
	r.Register(&ctr.ManifestConformance{})

	return r
}

// Language detection by file extension.
var extLanguages = map[string]string{
	".go": "go", ".ts": "typescript", ".tsx": "typescript",
	".js": "javascript", ".jsx": "javascript", ".py": "python",
	".java": "java", ".rs": "rust",
}

func detectLanguage(path string) string {
	for ext, lang := range extLanguages {
		if strings.HasSuffix(path, ext) {
			return lang
		}
	}
	return "unknown"
}
