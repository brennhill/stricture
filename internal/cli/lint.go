// lint.go — The lint command: configuration, discovery, analysis, and auto-fix pipeline.
package cli

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/stricture/stricture/internal/baseline"
	"github.com/stricture/stricture/internal/config"
	"github.com/stricture/stricture/internal/discovery"
	"github.com/stricture/stricture/internal/engine"
	"github.com/stricture/stricture/internal/fix"
	"github.com/stricture/stricture/internal/model"
	"github.com/stricture/stricture/internal/plugins"
	"github.com/stricture/stricture/internal/rules/all"
)

// runLint is the default lint subcommand.
func runLint(args []string) {
	opts := parseLintOptions(args)

	registry := all.Registry()
	cfg := loadLintConfig(opts, registry)
	selected, err := engine.SelectRules(registry, cfg, opts.ruleFilters, opts.category)
	if err != nil {
		exitf(2, "%v", err)
	}

	run := &lintRun{opts: opts, rules: selected, fixOps: make([]fix.Operation, 0)}
	run.discover()
	run.analyze()
	run.applyFixes()

	report := run.report()
	writeReport(report.data, opts.outputPath)
	if report.errorCount > 0 {
		os.Exit(1)
	}
}

// loadLintConfig loads the config file (unless disabled), registers plugin rules, and warns about unknown rule ids.
func loadLintConfig(opts lintOptions, registry *model.RuleRegistry) *config.Config {
	cfg := config.Default()
	resolvedPath := resolveConfigPath(opts.configPath)
	if opts.noConfig {
		return cfg
	}

	if loaded, err := config.Load(resolvedPath); err == nil {
		cfg = loaded
	} else if !errors.Is(err, model.ErrConfigNotFound) {
		exitf(1, "invalid config %s: %v", resolvedPath, err)
	}

	registerPluginRules(cfg, resolvedPath, registry)

	if unknown := config.UnknownRuleIDs(cfg, registry); len(unknown) > 0 {
		fmt.Fprintf(os.Stderr, "Warning: ignoring %d unknown rule(s): %s\n", len(unknown), strings.Join(unknown, ", "))
	}
	return cfg
}

func registerPluginRules(cfg *config.Config, configPath string, registry *model.RuleRegistry) {
	if len(cfg.Plugins) == 0 {
		return
	}
	pluginRules, err := plugins.Load(resolvePluginPaths(configPath, cfg.Plugins))
	if err != nil {
		exitf(2, "load plugins: %v", err)
	}
	for _, r := range pluginRules {
		registry.Register(r)
	}
}

// lintRun carries the state of one lint invocation through its pipeline stages.
type lintRun struct {
	opts         lintOptions
	rules        []model.Rule
	filePaths    []string
	files        []*model.UnifiedFileModel
	violations   []model.Violation
	baselineInfo baseline.State
	fixOps       []fix.Operation
	elapsedMs    int64
}

// ruleLimit is the violation cap applied while rules run. Baselining filters afterwards, so
// stopping early could miss findings that are not baselined.
func (r *lintRun) ruleLimit() int {
	if strings.TrimSpace(r.opts.baselinePath) != "" {
		return 0
	}
	return r.opts.maxViolations
}

// discover finds the files to lint, applying extension and git scope filters.
func (r *lintRun) discover() {
	filePaths, err := discovery.CollectFiles(r.opts.paths)
	if err != nil {
		exitf(1, "collect files: %v", err)
	}
	filePaths = discovery.FilterByExtensions(filePaths, r.opts.extensions)
	verbosef(r.opts.verbose, "Verbose: collected %d candidate file(s)\n", len(filePaths))

	if r.opts.changedOnly || r.opts.stagedOnly {
		filePaths = scopeToGitChanges(filePaths, r.opts.changedOnly, r.opts.stagedOnly)
	}

	cacheState := "off"
	if r.opts.cacheActive {
		cacheState = "on"
	}
	verbosef(r.opts.verbose, "Verbose: using %d file(s) after scope filters; rules=%d cache=%s\n", len(filePaths), len(r.rules), cacheState)
	r.filePaths = filePaths
}

func scopeToGitChanges(filePaths []string, changedOnly, stagedOnly bool) []string {
	scoped, err := discovery.GitScopedFileSet(changedOnly, stagedOnly)
	if err != nil {
		exitf(2, "%v", err)
	}
	cwd, err := os.Getwd()
	if err != nil {
		exitf(1, "get working directory: %v", err)
	}
	filtered := make([]string, 0, len(filePaths))
	for _, p := range filePaths {
		if scoped[discovery.PathKey(cwd, p)] {
			filtered = append(filtered, p)
		}
	}
	return filtered
}

// analyze parses the discovered files and evaluates the selected rules against them.
func (r *lintRun) analyze() {
	files, err := engine.BuildFiles(r.filePaths)
	if err != nil {
		exitf(1, "parse files: %v", err)
	}
	r.files = files

	start := time.Now()
	r.evaluate()
	r.elapsedMs = time.Since(start).Milliseconds()
}

// evaluate runs the rules over r.files, then applies the baseline and the severity threshold.
func (r *lintRun) evaluate() {
	ctx := &model.ProjectContext{Files: map[string]*model.UnifiedFileModel{}}
	for _, file := range r.files {
		ctx.Files[file.Path] = file
	}

	r.violations = engine.RunRules(r.files, r.rules, ctx, r.ruleLimit(), r.opts.concurrency)
	info, err := baseline.Apply(strings.TrimSpace(r.opts.baselinePath), &r.violations, baseline.Options{BootstrapIfMissing: !r.opts.diffMode})
	if err != nil {
		exitf(2, "%v", err)
	}
	r.baselineInfo = info
	r.violations = engine.FilterBySeverity(r.violations, r.opts.minSeverity)
}

// applyFixes plans fixes for the current violations and, with --fix, writes them and re-lints.
func (r *lintRun) applyFixes() {
	if !r.opts.fixApply && !r.opts.fixDryRun {
		return
	}
	planned, err := fix.Plan(r.violations)
	if err != nil {
		exitf(1, "build fix plan: %v", err)
	}
	r.fixOps = planned
	if r.opts.fixApply && len(planned) > 0 {
		r.writeFixesAndRelint()
	}
}

func (r *lintRun) writeFixesAndRelint() {
	if r.opts.fixBackup {
		if err := writeFixBackups(r.fixOps); err != nil {
			exitf(1, "create fix backups: %v", err)
		}
	}
	if err := fix.Apply(r.fixOps); err != nil {
		exitf(1, "apply fixes: %v", err)
	}

	filePaths, err := discovery.CollectFiles(rewritePathsAfterFix(r.opts.paths, r.fixOps))
	if err != nil {
		exitf(1, "collect files after fix: %v", err)
	}
	files, err := engine.BuildFiles(filePaths)
	if err != nil {
		exitf(1, "parse files after fix: %v", err)
	}
	r.filePaths, r.files = filePaths, files
	r.evaluate()
}
