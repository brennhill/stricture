package main

import (
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/stricture/stricture/internal/fix"
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

func TestLooksLikePathArg(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		value string
		want  bool
	}{
		{name: "empty", value: "", want: false},
		{name: "dot path", value: ".", want: true},
		{name: "absolute path", value: "/tmp/file.go", want: true},
		{name: "path with slash", value: "a/b", want: true},
		{name: "glob", value: "*.go", want: true},
		{name: "has dot", value: "main.go", want: true},
		{name: "plain token", value: "lint", want: false},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := looksLikePathArg(tc.value)
			if got != tc.want {
				t.Fatalf("looksLikePathArg(%q) = %v, want %v", tc.value, got, tc.want)
			}
		})
	}
}

func TestRepeatableFlagSetAndValues(t *testing.T) {
	t.Parallel()

	var f repeatableFlag
	if err := f.Set("A,B"); err != nil {
		t.Fatalf("Set(A,B): %v", err)
	}
	if err := f.Set(" C "); err != nil {
		t.Fatalf("Set(C): %v", err)
	}

	got := f.Values()
	want := []string{"A", "B", "C"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Values() = %#v, want %#v", got, want)
	}
	if f.String() != "A,B,C" {
		t.Fatalf("String() = %q, want %q", f.String(), "A,B,C")
	}
}

func TestSplitLintArgs(t *testing.T) {
	t.Parallel()

	flagArgs, pathArgs, err := splitLintArgs([]string{
		"--format", "json",
		".",
		"--rule", "CONV-file-header",
		"pkg",
		"--",
		"literal-arg",
	})
	if err != nil {
		t.Fatalf("splitLintArgs returned error: %v", err)
	}

	wantFlags := []string{"--format", "json", "--rule", "CONV-file-header"}
	wantPaths := []string{".", "pkg", "literal-arg"}
	if !reflect.DeepEqual(flagArgs, wantFlags) {
		t.Fatalf("flagArgs = %#v, want %#v", flagArgs, wantFlags)
	}
	if !reflect.DeepEqual(pathArgs, wantPaths) {
		t.Fatalf("pathArgs = %#v, want %#v", pathArgs, wantPaths)
	}
}

func TestSplitLintArgsMissingValue(t *testing.T) {
	t.Parallel()

	_, _, err := splitLintArgs([]string{"--format"})
	if err == nil || !strings.Contains(err.Error(), "requires a value") {
		t.Fatalf("expected missing value error, got %v", err)
	}
}

func TestSplitTraceArgs(t *testing.T) {
	t.Parallel()

	pathValue, flagArgs, err := splitTraceArgs([]string{"trace.json", "--manifest", "m.yml", "--service", "svc"})
	if err != nil {
		t.Fatalf("splitTraceArgs returned error: %v", err)
	}
	if pathValue != "trace.json" {
		t.Fatalf("pathValue = %q, want trace.json", pathValue)
	}
	wantFlags := []string{"--manifest", "m.yml", "--service", "svc"}
	if !reflect.DeepEqual(flagArgs, wantFlags) {
		t.Fatalf("flagArgs = %#v, want %#v", flagArgs, wantFlags)
	}
}

func TestSplitTraceArgsErrors(t *testing.T) {
	t.Parallel()

	t.Run("missing value", func(t *testing.T) {
		t.Parallel()
		_, _, err := splitTraceArgs([]string{"--manifest"})
		if err == nil {
			t.Fatalf("expected error for missing flag value")
		}
	})

	t.Run("multiple paths", func(t *testing.T) {
		t.Parallel()
		_, _, err := splitTraceArgs([]string{"a.json", "b.json"})
		if err == nil || !strings.Contains(err.Error(), "exactly one file argument") {
			t.Fatalf("expected exactly-one-path error, got %v", err)
		}
	})
}

func TestSplitAuditArgs(t *testing.T) {
	t.Parallel()

	flagArgs, pathArgs, err := splitAuditArgs([]string{"--format", "json", "svc-a", "--strictness", "strict", "svc-b"})
	if err != nil {
		t.Fatalf("splitAuditArgs returned error: %v", err)
	}
	wantFlags := []string{"--format", "json", "--strictness", "strict"}
	wantPaths := []string{"svc-a", "svc-b"}
	if !reflect.DeepEqual(flagArgs, wantFlags) {
		t.Fatalf("flagArgs = %#v, want %#v", flagArgs, wantFlags)
	}
	if !reflect.DeepEqual(pathArgs, wantPaths) {
		t.Fatalf("pathArgs = %#v, want %#v", pathArgs, wantPaths)
	}
}

func TestParseExtensionFilter(t *testing.T) {
	t.Parallel()

	got, err := parseExtensionFilter("go, .TS,py")
	if err != nil {
		t.Fatalf("parseExtensionFilter returned error: %v", err)
	}
	want := map[string]bool{
		".go": true,
		".ts": true,
		".py": true,
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("parseExtensionFilter result = %#v, want %#v", got, want)
	}
}

func TestParseExtensionFilterInvalid(t *testing.T) {
	t.Parallel()

	if _, err := parseExtensionFilter(" , , . "); err == nil {
		t.Fatalf("expected invalid extension filter error")
	}
}

func TestFilterFilePathsByExtensions(t *testing.T) {
	t.Parallel()

	paths := []string{"a.go", "b.ts", "c.py", "README.md"}
	allowlist := map[string]bool{".go": true, ".py": true}
	got := filterFilePathsByExtensions(paths, allowlist)
	want := []string{"a.go", "c.py"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("filterFilePathsByExtensions = %#v, want %#v", got, want)
	}
}

func TestInferTraceFormat(t *testing.T) {
	t.Parallel()

	if got := inferTraceFormat("trace.har", []byte("{}")); got != "har" {
		t.Fatalf("inferTraceFormat(.har) = %q, want har", got)
	}
	if got := inferTraceFormat("trace.otel", []byte("{}")); got != "otel" {
		t.Fatalf("inferTraceFormat(.otel) = %q, want otel", got)
	}
	if got := inferTraceFormat("trace.json", []byte(`{"resourceSpans":[]}`)); got != "otel" {
		t.Fatalf("inferTraceFormat(resourceSpans) = %q, want otel", got)
	}
	if got := inferTraceFormat("trace.json", []byte(`{"log":{}}`)); got != "har" {
		t.Fatalf("inferTraceFormat(log) = %q, want har", got)
	}
	if got := inferTraceFormat("trace.json", []byte(`{"x":1}`)); got != "custom" {
		t.Fatalf("inferTraceFormat(custom) = %q, want custom", got)
	}
}

func TestValidateTracePayload(t *testing.T) {
	t.Parallel()

	if err := validateTracePayload("har", []byte(`{"log":{"entries":[]}}`), true); err != nil {
		t.Fatalf("har strict should pass: %v", err)
	}
	if err := validateTracePayload("har", []byte(`{"log":{}}`), true); err == nil {
		t.Fatalf("har strict empty log should fail")
	}
	if err := validateTracePayload("otel", []byte(`{"resourceSpans":[{}]}`), true); err != nil {
		t.Fatalf("otel strict should pass: %v", err)
	}
	if err := validateTracePayload("otel", []byte(`{"resourceSpans":[]}`), true); err == nil {
		t.Fatalf("otel strict empty resourceSpans should fail")
	}
	if err := validateTracePayload("custom", []byte(`{}`), true); err == nil {
		t.Fatalf("custom strict empty object should fail")
	}
	if err := validateTracePayload("custom", []byte(`{"ok":true}`), true); err != nil {
		t.Fatalf("custom strict non-empty should pass: %v", err)
	}
	if err := validateTracePayload("wat", []byte(`{"ok":true}`), false); err == nil {
		t.Fatalf("unsupported format should fail")
	}
}

func TestParseTraceEnvelope(t *testing.T) {
	t.Parallel()

	if _, err := parseTraceEnvelope([]byte("not-json")); err == nil {
		t.Fatalf("invalid JSON should fail")
	}
	out, err := parseTraceEnvelope([]byte(`{"a":1}`))
	if err != nil {
		t.Fatalf("valid JSON object should pass: %v", err)
	}
	if len(out) != 1 {
		t.Fatalf("expected one key, got %d", len(out))
	}
}

func TestValidStrictness(t *testing.T) {
	t.Parallel()

	valid := []string{"minimal", "basic", "standard", "strict", "exhaustive"}
	for _, value := range valid {
		if !validStrictness(value) {
			t.Fatalf("validStrictness(%q) = false, want true", value)
		}
	}
	if validStrictness("wild") {
		t.Fatalf("validStrictness(wild) = true, want false")
	}
}

func TestRewritePathsAfterFix(t *testing.T) {
	t.Parallel()

	paths := []string{"a.go", "b.go"}
	ops := []fix.Operation{
		{Kind: "rename", Path: "a.go", NewPath: "c.go"},
	}
	got := rewritePathsAfterFix(paths, ops)
	want := []string{"c.go", "b.go"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("rewritePathsAfterFix = %#v, want %#v", got, want)
	}
}

func TestRenderFixOperations(t *testing.T) {
	t.Parallel()

	ops := []fix.Operation{
		{RuleID: "CONV-file-naming", Kind: "rename", Path: "a.go", NewPath: "b.go", Description: "rename"},
	}
	out := renderFixOperations(ops)
	if len(out) != 1 {
		t.Fatalf("len(out) = %d, want 1", len(out))
	}
	if out[0]["newPath"] != "b.go" {
		t.Fatalf("newPath = %q, want b.go", out[0]["newPath"])
	}
}

func TestFormatFixSummary(t *testing.T) {
	t.Parallel()

	summary := formatFixSummary([]fix.Operation{{RuleID: "X", Description: "did thing"}}, true)
	if !strings.Contains(summary, "dry-run") {
		t.Fatalf("expected dry-run marker in summary, got %q", summary)
	}
	if !strings.Contains(summary, "[X] did thing") {
		t.Fatalf("expected operation details in summary, got %q", summary)
	}
}

func TestWriteFixBackups(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	pathValue := filepath.Join(dir, "a.go")
	if err := os.WriteFile(pathValue, []byte("package main\n"), 0o644); err != nil {
		t.Fatalf("write source file: %v", err)
	}

	ops := []fix.Operation{{Kind: "edit", Path: pathValue}}
	if err := writeFixBackups(ops); err != nil {
		t.Fatalf("writeFixBackups failed: %v", err)
	}
	if _, err := os.Stat(pathValue + ".bak"); err != nil {
		t.Fatalf("expected backup file, got stat error: %v", err)
	}
}

func TestFilterViolationsBySeverity(t *testing.T) {
	t.Parallel()

	violations := []model.Violation{
		{RuleID: "A", Severity: "warn"},
		{RuleID: "B", Severity: "error"},
	}
	got := filterViolationsBySeverity(violations, "error")
	if len(got) != 1 || got[0].RuleID != "B" {
		t.Fatalf("filterViolationsBySeverity(error) = %#v, want only B", got)
	}
}

func TestIsPathWithinRoot(t *testing.T) {
	t.Parallel()

	root := filepath.FromSlash("/workspace/project")
	inside := filepath.FromSlash("/workspace/project/src/main.go")
	outside := filepath.FromSlash("/workspace/other/file.go")
	if !isPathWithinRoot(inside, root) {
		t.Fatalf("inside path should be within root")
	}
	if isPathWithinRoot(outside, root) {
		t.Fatalf("outside path should not be within root")
	}
}

func TestShouldSkipLintDir(t *testing.T) {
	t.Parallel()

	if !shouldSkipLintDir("node_modules") {
		t.Fatalf("node_modules should be skipped")
	}
	if !shouldSkipLintDir("tests/fixtures/sample") {
		t.Fatalf("tests/fixtures path should be skipped")
	}
	if shouldSkipLintDir("src") {
		t.Fatalf("src should not be skipped")
	}
}

func TestSourceFileHeuristics(t *testing.T) {
	t.Parallel()

	if isLintSourceFile("x.generated.ts") {
		t.Fatalf("generated file should not be lint source")
	}
	if isLintSourceFile("x.pb.go") {
		t.Fatalf("protobuf-generated file should not be lint source")
	}
	if !isLintSourceFile("x.go") {
		t.Fatalf("go file should be lint source")
	}
	if !isLintSourceFile("x.ts") {
		t.Fatalf("ts file should be lint source")
	}
	if isLintSourceFile("x.md") {
		t.Fatalf("markdown should not be lint source")
	}
}

func TestCountLinesAndLooksLikeTestFile(t *testing.T) {
	t.Parallel()

	if got := countLines([]byte("a\nb\n")); got != 3 {
		t.Fatalf("countLines = %d, want 3", got)
	}
	if got := countLines(nil); got != 0 {
		t.Fatalf("countLines(nil) = %d, want 0", got)
	}

	testNames := []string{"foo_test.go", "foo.test.ts", "bar.spec.js", "test_sample.py", "serviceTest.java"}
	for _, name := range testNames {
		if !looksLikeTestFile(name) {
			t.Fatalf("looksLikeTestFile(%q) = false, want true", name)
		}
	}
	if looksLikeTestFile("main.go") {
		t.Fatalf("main.go should not look like a test file")
	}
}

func TestRunLintRulesForFileRecoversFromPanic(t *testing.T) {
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

func TestRunLintRulesForFileAppliesSuppressionAndRuleFallback(t *testing.T) {
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

func TestRunLintRulesParallelMatchesSequentialSet(t *testing.T) {
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
