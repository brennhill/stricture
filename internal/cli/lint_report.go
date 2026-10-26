// lint_report.go — Builds the lint summary and renders text or machine-readable reports.
package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/stricture/stricture/internal/model"
)

// lintReport is a rendered report plus the number of error-severity violations in it.
type lintReport struct {
	data       []byte
	errorCount int
}

// report orders and truncates violations, then renders them in the requested format.
func (r *lintRun) report() lintReport {
	r.sortAndTruncate()
	counts := countViolations(r.violations)
	summary := r.summary(counts)
	verbosef(r.opts.verbose, "Verbose: lint complete in %dms (violations=%d errors=%d warnings=%d)\n", r.elapsedMs, len(r.violations), counts.errors, counts.warnings)

	switch r.opts.format {
	case "json", "sarif", "junit":
		return lintReport{data: r.renderStructured(summary), errorCount: counts.errors}
	default:
		colorEnabled := shouldUseColor(r.opts.forceColor, r.opts.forceNoColor, strings.TrimSpace(r.opts.outputPath))
		return lintReport{data: r.renderText(summary, colorEnabled), errorCount: counts.errors}
	}
}

func (r *lintRun) sortAndTruncate() {
	v := r.violations
	sort.Slice(v, func(i, j int) bool {
		if v[i].FilePath != v[j].FilePath {
			return v[i].FilePath < v[j].FilePath
		}
		if v[i].StartLine != v[j].StartLine {
			return v[i].StartLine < v[j].StartLine
		}
		return v[i].RuleID < v[j].RuleID
	})
	if r.opts.maxViolations > 0 && len(v) > r.opts.maxViolations {
		r.violations = v[:r.opts.maxViolations]
	}
}

type violationCounts struct {
	filesWithIssues int
	errors          int
	warnings        int
}

func countViolations(violations []model.Violation) violationCounts {
	files := map[string]bool{}
	var c violationCounts
	for _, v := range violations {
		files[v.FilePath] = true
		switch strings.ToLower(v.Severity) {
		case "error":
			c.errors++
		case "warn", "warning":
			c.warnings++
		}
	}
	c.filesWithIssues = len(files)
	return c
}

func (r *lintRun) summary(c violationCounts) map[string]interface{} {
	summary := map[string]interface{}{
		"filesChecked":    len(r.files),
		"filesWithIssues": c.filesWithIssues,
		"totalViolations": len(r.violations),
		"errors":          c.errors,
		"warnings":        c.warnings,
		"elapsedMs":       r.elapsedMs,
	}
	if r.baselineInfo.Enabled {
		summary["baselinePath"] = filepath.ToSlash(r.baselineInfo.Path)
		summary["baselineSuppressed"] = r.baselineInfo.Suppressed
		summary["baselineBootstrapped"] = r.baselineInfo.Bootstrapped
	}
	if r.opts.diffMode {
		summary["diffEnabled"] = true
		summary["diffAdded"] = len(r.baselineInfo.Added)
		summary["diffResolved"] = len(r.baselineInfo.Resolved)
	}
	return summary
}

// renderStructured renders the JSON envelope used by the json, sarif, and junit formats.
func (r *lintRun) renderStructured(summary map[string]interface{}) []byte {
	payload := map[string]interface{}{
		"version":    "1",
		"violations": r.violations,
		"summary":    summary,
	}
	if r.baselineInfo.Enabled {
		payload["baseline"] = map[string]interface{}{
			"path":         filepath.ToSlash(r.baselineInfo.Path),
			"suppressed":   r.baselineInfo.Suppressed,
			"bootstrapped": r.baselineInfo.Bootstrapped,
			"entryCount":   r.baselineInfo.EntryCount,
		}
	}
	if r.opts.diffMode {
		payload["diff"] = map[string]interface{}{
			"enabled":  true,
			"added":    r.baselineInfo.Added,
			"resolved": r.baselineInfo.Resolved,
			"summary": map[string]int{
				"added":    len(r.baselineInfo.Added),
				"resolved": len(r.baselineInfo.Resolved),
			},
		}
	}
	if r.opts.fixApply || r.opts.fixDryRun {
		payload["fixes"] = renderFixOperations(r.fixOps)
		payload["fixMode"] = map[string]bool{
			"apply":   r.opts.fixApply,
			"dryRun":  r.opts.fixDryRun,
			"applied": r.opts.fixApply,
		}
	}
	encoded, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		exitf(1, "write %s output: %v", r.opts.format, err)
	}
	return append(encoded, '\n')
}

func (r *lintRun) renderText(summary map[string]interface{}, colorEnabled bool) []byte {
	var out strings.Builder
	r.writeBaselineLines(&out)
	if r.opts.fixApply || r.opts.fixDryRun {
		out.WriteString(formatFixSummary(r.fixOps, r.opts.fixDryRun))
	}

	if len(r.violations) == 0 {
		fmt.Fprintln(&out, "No violations found.")
	}
	for _, v := range r.violations {
		label := colorizeSeverityLabel(v.Severity, strings.ToUpper(v.Severity), colorEnabled)
		fmt.Fprintf(&out, "%s:%d: %s %s: %s\n", v.FilePath, v.StartLine, label, v.RuleID, v.Message)
	}
	fmt.Fprintf(&out, "Summary: files=%d issues=%d violations=%d errors=%d warnings=%d elapsedMs=%d\n",
		summary["filesChecked"], summary["filesWithIssues"], summary["totalViolations"], summary["errors"], summary["warnings"], summary["elapsedMs"])
	return []byte(out.String())
}

func (r *lintRun) writeBaselineLines(out *strings.Builder) {
	info := r.baselineInfo
	if info.Enabled {
		if info.Bootstrapped {
			fmt.Fprintf(out, "Baseline created at %s with %d entry(s); existing violations suppressed.\n", info.Path, info.EntryCount)
		} else if info.Suppressed > 0 {
			fmt.Fprintf(out, "Baseline suppressed %d violation(s) from %s.\n", info.Suppressed, info.Path)
		}
	}
	if r.opts.diffMode {
		fmt.Fprintf(out, "Diff: added=%d resolved=%d (baseline=%s)\n", len(info.Added), len(info.Resolved), info.Path)
	}
}

// writeReport writes the report to stdout, or to outputPath (creating parent directories).
func writeReport(report []byte, outputPath string) {
	target := strings.TrimSpace(outputPath)
	if target == "" {
		if _, err := os.Stdout.Write(report); err != nil {
			exitf(1, "write output: %v", err)
		}
		return
	}
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		exitf(1, "create output directory for %s: %v", target, err)
	}
	if err := os.WriteFile(target, report, 0o644); err != nil {
		exitf(1, "write output file %s: %v", target, err)
	}
}
