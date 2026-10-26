// lint_report_test.go — Tests violation counting, ordering, summaries, and report rendering.
package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stricture/stricture/internal/baseline"
	"github.com/stricture/stricture/internal/model"
)

func sampleRun() *lintRun {
	return &lintRun{
		opts: lintOptions{format: "text", maxViolations: 0},
		files: []*model.UnifiedFileModel{
			{Path: "a.go"}, {Path: "b.go"}, {Path: "c.go"},
		},
		violations: []model.Violation{
			{RuleID: "R2", Severity: "warn", FilePath: "b.go", StartLine: 3, Message: "later"},
			{RuleID: "R1", Severity: "error", FilePath: "a.go", StartLine: 9, Message: "second"},
			{RuleID: "R0", Severity: "error", FilePath: "a.go", StartLine: 2, Message: "first"},
			{RuleID: "R3", Severity: "WARNING", FilePath: "b.go", StartLine: 3, Message: "tie"},
		},
		elapsedMs: 7,
	}
}

func TestCountViolations(t *testing.T) {
	c := countViolations(sampleRun().violations)
	if c.errors != 2 || c.warnings != 2 || c.filesWithIssues != 2 {
		t.Errorf("counts = %+v, want 2 errors, 2 warnings, 2 files", c)
	}
}

func TestSortAndTruncate(t *testing.T) {
	r := sampleRun()
	r.sortAndTruncate()
	got := make([]string, 0, len(r.violations))
	for _, v := range r.violations {
		got = append(got, v.RuleID)
	}
	if strings.Join(got, ",") != "R0,R1,R2,R3" {
		t.Errorf("order = %v, want by file, line, then rule", got)
	}

	r.opts.maxViolations = 2
	r.sortAndTruncate()
	if len(r.violations) != 2 {
		t.Errorf("truncated to %d, want 2", len(r.violations))
	}
}

func TestRuleLimitIgnoresMaxViolationsWhenBaselining(t *testing.T) {
	r := &lintRun{opts: lintOptions{maxViolations: 5}}
	if r.ruleLimit() != 5 {
		t.Errorf("ruleLimit = %d, want 5", r.ruleLimit())
	}
	r.opts.baselinePath = "b.json"
	if r.ruleLimit() != 0 {
		t.Errorf("baseline filtering happens afterwards, so ruleLimit should be 0, got %d", r.ruleLimit())
	}
}

func TestSummaryIncludesBaselineAndDiffOnlyWhenEnabled(t *testing.T) {
	r := sampleRun()
	s := r.summary(countViolations(r.violations))
	if s["filesChecked"] != 3 || s["totalViolations"] != 4 || s["elapsedMs"] != int64(7) {
		t.Errorf("summary = %v", s)
	}
	if _, ok := s["baselinePath"]; ok {
		t.Error("baseline keys should be absent when no baseline is enabled")
	}

	r.baselineInfo = baseline.State{Enabled: true, Path: filepath.Join("dir", "b.json"), Suppressed: 3, Bootstrapped: true}
	r.opts.diffMode = true
	s = r.summary(countViolations(r.violations))
	if s["baselinePath"] != "dir/b.json" || s["baselineSuppressed"] != 3 || s["baselineBootstrapped"] != true || s["diffEnabled"] != true {
		t.Errorf("summary = %v", s)
	}
}

func TestRenderTextListsViolationsAndSummary(t *testing.T) {
	r := sampleRun()
	r.sortAndTruncate()
	out := string(r.renderText(r.summary(countViolations(r.violations)), false))
	for _, want := range []string{
		"a.go:2: ERROR R0: first\n",
		"b.go:3: WARN R2: later\n",
		"Summary: files=3 issues=2 violations=4 errors=2 warnings=2 elapsedMs=7\n",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("text report missing %q:\n%s", want, out)
		}
	}
}

func TestRenderTextReportsCleanRunAndBaselineLines(t *testing.T) {
	r := &lintRun{opts: lintOptions{diffMode: true}, files: nil}
	r.baselineInfo = baseline.State{Enabled: true, Path: "b.json", EntryCount: 4, Bootstrapped: true}
	out := string(r.renderText(r.summary(countViolations(nil)), false))
	for _, want := range []string{
		"Baseline created at b.json with 4 entry(s); existing violations suppressed.\n",
		"Diff: added=0 resolved=0 (baseline=b.json)\n",
		"No violations found.\n",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("clean report missing %q:\n%s", want, out)
		}
	}

	r.baselineInfo = baseline.State{Enabled: true, Path: "b.json", Suppressed: 2}
	if out := string(r.renderText(r.summary(countViolations(nil)), false)); !strings.Contains(out, "Baseline suppressed 2 violation(s) from b.json.") {
		t.Errorf("missing suppression line:\n%s", out)
	}
}

func TestRenderStructuredEnvelope(t *testing.T) {
	r := sampleRun()
	r.opts.format = "json"
	r.opts.fixDryRun = true
	r.opts.diffMode = true
	r.baselineInfo = baseline.State{Enabled: true, Path: "b.json", EntryCount: 1}
	r.fixOps = nil

	var payload map[string]json.RawMessage
	if err := json.Unmarshal(r.renderStructured(r.summary(countViolations(r.violations))), &payload); err != nil {
		t.Fatalf("report is not valid JSON: %v", err)
	}
	for _, key := range []string{"version", "violations", "summary", "baseline", "diff", "fixes", "fixMode"} {
		if _, ok := payload[key]; !ok {
			t.Errorf("structured report missing %q", key)
		}
	}

	plain := sampleRun()
	var minimal map[string]json.RawMessage
	if err := json.Unmarshal(plain.renderStructured(plain.summary(countViolations(plain.violations))), &minimal); err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"baseline", "diff", "fixes", "fixMode"} {
		if _, ok := minimal[key]; ok {
			t.Errorf("plain report should omit %q", key)
		}
	}
}

func TestReportPicksFormatAndCountsErrors(t *testing.T) {
	r := sampleRun()
	r.opts.format = "sarif"
	rep := r.report()
	if rep.errorCount != 2 || !strings.HasPrefix(string(rep.data), "{") {
		t.Errorf("sarif report: errors=%d data=%.20s", rep.errorCount, rep.data)
	}

	r = sampleRun()
	rep = r.report()
	if rep.errorCount != 2 || !strings.Contains(string(rep.data), "Summary:") {
		t.Errorf("text report: errors=%d data=%.40s", rep.errorCount, rep.data)
	}
}

func TestWriteReportCreatesParentDirectories(t *testing.T) {
	target := filepath.Join(t.TempDir(), "nested", "out", "report.txt")
	writeReport([]byte("hello\n"), "  "+target+"  ")
	got, err := os.ReadFile(target)
	if err != nil || string(got) != "hello\n" {
		t.Fatalf("report file = %q, %v", got, err)
	}
}
