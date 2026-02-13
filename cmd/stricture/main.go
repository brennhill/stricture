// main.go — Stricture CLI entry point.
package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/stricture/stricture/internal/config"
	"github.com/stricture/stricture/internal/lineage"
	"github.com/stricture/stricture/internal/model"
	"github.com/stricture/stricture/internal/rules/conv"
	"gopkg.in/yaml.v3"
)

var version = "0.1.0-dev"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(0)
	}

	switch os.Args[1] {
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
	case "validate-config":
		runValidateConfig(os.Args[2:])
	case "lint":
		runLint(os.Args[2:])
	case "--version", "-version", "version":
		fmt.Printf("stricture version %s\n", version)
	case "--help", "-help", "help":
		printUsage()
	default:
		// No subcommand: treat all args as lint targets.
		runLint(os.Args[1:])
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
	fmt.Println("  inspect <file>    Parse a file and print its UnifiedFileModel as JSON")
	fmt.Println("  inspect-lineage   Parse stricture-source annotations from a file")
	fmt.Println("  lineage-export    Build normalized lineage artifact from source files")
	fmt.Println("  lineage-diff      Diff two lineage artifacts and classify drift severity")
	fmt.Println("  lineage-escalate  Resolve emergency contacts upstream from a service")
	fmt.Println("  list-rules        List all registered rules")
	fmt.Println("  validate-config   Check that a .stricture.yml file is valid")
	fmt.Println("  version           Print version and exit")
	fmt.Println("  help              Print this help message")
	fmt.Println()
	fmt.Println("Run 'stricture <command> --help' for details on a specific command.")
}

// runLint is the default lint subcommand.
func runLint(args []string) {
	fs := flag.NewFlagSet("lint", flag.ExitOnError)
	format := fs.String("format", "text", "Output format (text, json, sarif, junit)")
	configPath := fs.String("config", ".stricture.yml", "Path to configuration file")
	rule := fs.String("rule", "", "Run a single rule by ID")
	category := fs.String("category", "", "Run all rules in a category")
	_ = fs.Bool("changed", false, "Lint only changed files")
	_ = fs.Bool("staged", false, "Lint only staged files")
	_ = fs.Bool("no-cache", false, "Disable caching")
	_ = fs.Parse(args)

	validFormats := map[string]bool{"text": true, "json": true, "sarif": true, "junit": true}
	if !validFormats[*format] {
		fmt.Fprintf(os.Stderr, "Error: invalid format %q (valid: text, json, sarif, junit)\n", *format)
		os.Exit(2)
	}

	registry := buildRegistry()

	cfg := config.Default()
	resolvedConfigPath := resolveConfigPath(*configPath)
	if loaded, err := config.Load(resolvedConfigPath); err == nil {
		cfg = loaded
	} else if !errors.Is(err, model.ErrConfigNotFound) {
		fmt.Fprintf(os.Stderr, "Error: invalid config %s: %v\n", resolvedConfigPath, err)
		os.Exit(1)
	}

	if unknown := config.UnknownRuleIDs(cfg, registry); len(unknown) > 0 {
		fmt.Fprintf(os.Stderr, "Warning: ignoring %d unknown rule(s): %s\n", len(unknown), strings.Join(unknown, ", "))
	}

	selectedRules, err := resolveLintRules(registry, cfg, *rule, *category)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(2)
	}

	paths := fs.Args()
	if len(paths) == 0 {
		paths = []string{"."}
	}

	filePaths, err := collectLintFilePaths(paths)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: collect files: %v\n", err)
		os.Exit(1)
	}

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
	violations := runLintRules(files, selectedRules, ctx)
	elapsed := time.Since(start).Milliseconds()

	sort.Slice(violations, func(i, j int) bool {
		if violations[i].FilePath != violations[j].FilePath {
			return violations[i].FilePath < violations[j].FilePath
		}
		if violations[i].StartLine != violations[j].StartLine {
			return violations[i].StartLine < violations[j].StartLine
		}
		return violations[i].RuleID < violations[j].RuleID
	})

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

	switch *format {
	case "json", "sarif", "junit":
		payload := map[string]interface{}{
			"version":    "1",
			"violations": violations,
			"summary":    summary,
		}
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		if err := enc.Encode(payload); err != nil {
			fmt.Fprintf(os.Stderr, "Error: write %s output: %v\n", *format, err)
			os.Exit(1)
		}
	default:
		if len(violations) == 0 {
			fmt.Println("No violations found.")
		} else {
			for _, v := range violations {
				fmt.Printf("%s:%d: %s %s: %s\n", v.FilePath, v.StartLine, strings.ToUpper(v.Severity), v.RuleID, v.Message)
			}
		}
		fmt.Printf("Summary: files=%d issues=%d violations=%d errors=%d warnings=%d elapsedMs=%d\n",
			summary["filesChecked"], summary["filesWithIssues"], summary["totalViolations"], summary["errors"], summary["warnings"], summary["elapsedMs"])
	}

	if errorCount > 0 {
		os.Exit(1)
	}
}

func resolveLintRules(registry *model.RuleRegistry, cfg *config.Config, singleRule string, category string) ([]model.Rule, error) {
	selected := make([]model.Rule, 0)
	targetCategory := strings.ToLower(strings.TrimSpace(category))

	var single model.Rule
	if strings.TrimSpace(singleRule) != "" {
		found, ok := registry.ByID(strings.TrimSpace(singleRule))
		if !ok {
			return nil, fmt.Errorf("unknown rule %q", singleRule)
		}
		single = found
	}

	candidates := make([]model.Rule, 0)
	switch {
	case single != nil:
		candidates = append(candidates, single)
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
		if single != nil && r.ID() != single.ID() {
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

func collectLintFilePaths(paths []string) ([]string, error) {
	files := make([]string, 0)
	seen := map[string]bool{}

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
	switch strings.ToLower(filepath.Ext(path)) {
	case ".go", ".ts", ".tsx", ".js", ".jsx", ".py", ".java", ".kt", ".rs":
		return true
	default:
		return false
	}
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

func runLintRules(files []*model.UnifiedFileModel, rules []model.Rule, ctx *model.ProjectContext) []model.Violation {
	violations := make([]model.Violation, 0)
	for _, file := range files {
		for _, rawRule := range rules {
			ruleCfg := model.RuleConfig{Severity: rawRule.DefaultSeverity(), Options: map[string]interface{}{}}
			if withCfg, ok := rawRule.(lintRuleWithConfig); ok {
				rawRule = withCfg.Rule
				ruleCfg = withCfg.Config
			}
			if isRuleSuppressed(file.Source, rawRule.ID()) {
				continue
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
					}
				}()
				violations = append(violations, rawRule.Check(file, ctx, ruleCfg)...)
			}()
		}
	}
	return violations
}

func isRuleSuppressed(source []byte, ruleID string) bool {
	text := string(source)
	return strings.Contains(text, "stricture-disable "+ruleID) ||
		strings.Contains(text, "stricture-disable-next-line "+ruleID)
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
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Error: file not found: %s\n", filePath)
		os.Exit(1)
	}

	// Stub: return an empty UnifiedFileModel until adapters are built.
	stub := model.UnifiedFileModel{
		Path:     filePath,
		Language: detectLanguage(filePath),
	}

	out, err := json.MarshalIndent(stub, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to marshal model: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(string(out))
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
	fmt.Fprintln(w, "ID\tCATEGORY\tDEFAULT\tDESCRIPTION")
	fmt.Fprintln(w, "--\t--------\t-------\t-----------")
	for _, r := range registry.All() {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
			r.ID(), strings.ToUpper(r.Category()), r.DefaultSeverity(), r.Description())
	}
	w.Flush()
	fmt.Printf("\n%d rules registered.\n", len(registry.All()))
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

// buildRegistry creates a RuleRegistry with all known rules.
func buildRegistry() *model.RuleRegistry {
	r := model.NewRuleRegistry()
	r.Register(&conv.FileNaming{})
	r.Register(&conv.FileHeader{})
	r.Register(&conv.ErrorFormat{})
	r.Register(&conv.ExportNaming{})
	r.Register(&conv.TestFileLocation{})
	r.Register(&conv.RequiredExports{})
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
