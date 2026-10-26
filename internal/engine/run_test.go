// run_test.go — Tests rule running, severity filtering, and file helpers.
package engine

import (
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"

	"github.com/stricture/stricture/internal/model"
)

type fakeRule struct {
	id          string
	violations  []model.Violation
	shouldPanic bool
}

func (r fakeRule) ID() string {
	return r.id
}

func (r fakeRule) Category() string {
	return "TEST"
}

func (r fakeRule) Description() string {
	return "test rule"
}

func (r fakeRule) DefaultSeverity() string {
	return "error"
}

func (r fakeRule) Check(_ *model.UnifiedFileModel, _ *model.ProjectContext, _ model.RuleConfig) []model.Violation {
	if r.shouldPanic {
		panic("boom")
	}
	return append([]model.Violation(nil), r.violations...)
}

func (r fakeRule) NeedsProjectContext() bool {
	return false
}

func (r fakeRule) Why() string {
	return "tests"
}

func TestFilterBySeverity(t *testing.T) {
	t.Parallel()

	violations := []model.Violation{
		{RuleID: "A", Severity: "warn"},
		{RuleID: "B", Severity: "error"},
	}
	got := FilterBySeverity(violations, "error")
	if len(got) != 1 || got[0].RuleID != "B" {
		t.Fatalf("FilterBySeverity(error) = %#v, want only B", got)
	}
}

func TestCountLinesAndLooksLikeTestFile(t *testing.T) {
	t.Parallel()

	if got := CountLines([]byte("a\nb\n")); got != 3 {
		t.Fatalf("countLines = %d, want 3", got)
	}
	if got := CountLines(nil); got != 0 {
		t.Fatalf("CountLines(nil) = %d, want 0", got)
	}

	testNames := []string{"foo_test.go", "foo.test.ts", "bar.spec.js", "test_sample.py", "serviceTest.java"}
	for _, name := range testNames {
		if !LooksLikeTestFile(name) {
			t.Fatalf("LooksLikeTestFile(%q) = false, want true", name)
		}
	}
	if LooksLikeTestFile("main.go") {
		t.Fatalf("main.go should not look like a test file")
	}
}

func TestRunForFileRecoversFromPanic(t *testing.T) {
	t.Parallel()

	file := &model.UnifiedFileModel{Path: "a.go", Source: []byte("package main\n")}
	rules := []model.Rule{
		fakeRule{id: "PANIC-rule", shouldPanic: true},
	}
	out := runLintRulesForFile(file, rules, &model.ProjectContext{}, 0)
	if len(out) != 1 {
		t.Fatalf("expected 1 panic violation, got %d", len(out))
	}
	if out[0].RuleID != "PANIC-rule" {
		t.Fatalf("panic violation RuleID = %q, want PANIC-rule", out[0].RuleID)
	}
}

func TestRunForFileAppliesSuppressionAndRuleFallback(t *testing.T) {
	t.Parallel()

	file := &model.UnifiedFileModel{
		Path:   "a.go",
		Source: []byte("// stricture-disable-next-line RULE-a\nx\n"),
	}
	rules := []model.Rule{
		fakeRule{
			id: "RULE-a",
			violations: []model.Violation{
				{RuleID: "", Severity: "error", FilePath: "a.go", StartLine: 2, Message: "suppressed"},
			},
		},
		fakeRule{
			id: "RULE-b",
			violations: []model.Violation{
				{RuleID: "", Severity: "warn", FilePath: "a.go", StartLine: 3, Message: "kept"},
			},
		},
	}
	out := runLintRulesForFile(file, rules, &model.ProjectContext{}, 0)
	if len(out) != 1 {
		t.Fatalf("expected 1 remaining violation, got %d (%+v)", len(out), out)
	}
	if out[0].RuleID != "RULE-b" {
		t.Fatalf("fallback RuleID should be RULE-b, got %q", out[0].RuleID)
	}
}

func TestRunParallelMatchesSequentialSet(t *testing.T) {
	t.Parallel()

	files := []*model.UnifiedFileModel{
		{Path: "a.go", Source: []byte("package a\n")},
		{Path: "b.go", Source: []byte("package b\n")},
	}
	rules := []model.Rule{
		fakeRule{
			id: "RULE-a",
			violations: []model.Violation{
				{RuleID: "RULE-a", Severity: "error", FilePath: "a.go", StartLine: 1, Message: "x"},
			},
		},
	}
	ctx := &model.ProjectContext{}

	seq := runLintRulesSequential(files, rules, ctx, 0)
	par := runLintRulesParallel(files, rules, ctx, 4)

	normalize := func(in []model.Violation) []string {
		out := make([]string, 0, len(in))
		for _, v := range in {
			out = append(out, v.RuleID+"|"+v.FilePath+"|"+v.Message)
		}
		sort.Strings(out)
		return out
	}

	if !reflect.DeepEqual(normalize(seq), normalize(par)) {
		t.Fatalf("parallel result differs from sequential\nseq=%v\npar=%v", normalize(seq), normalize(par))
	}
}

func TestBuildFilesReadsSourceLanguageAndTestFlag(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "svc_test.go")
	if err := os.WriteFile(src, []byte("package x\n\nfunc F() {}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	files, err := BuildFiles([]string{src})
	if err != nil {
		t.Fatalf("BuildFiles error: %v", err)
	}
	if len(files) != 1 {
		t.Fatalf("got %d files, want 1", len(files))
	}
	f := files[0]
	if f.Language != "go" || !f.IsTestFile || f.LineCount != 4 || string(f.Source) != "package x\n\nfunc F() {}\n" {
		t.Errorf("unexpected model: language=%q test=%v lines=%d", f.Language, f.IsTestFile, f.LineCount)
	}
	if f.Path != filepath.ToSlash(src) {
		t.Errorf("Path = %q, want slash-separated %q", f.Path, filepath.ToSlash(src))
	}

	if _, err := BuildFiles([]string{filepath.Join(dir, "missing.go")}); err == nil {
		t.Error("BuildFiles should fail for an unreadable file")
	}
}
