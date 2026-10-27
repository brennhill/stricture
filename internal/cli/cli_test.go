// cli_test.go — Unit tests for CLI argument handling, discovery, baseline, fix, and rule-running helpers.
package cli

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/stricture/stricture/internal/fix"
)

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

func TestLoadPolicyBindingFromConfig(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	pathValue := filepath.Join(dir, ".stricture.yml")
	content := "'strict:policy_url': https://policies.example.com/stricture/prod.yml\n'strict:policy_sha256': 9f0e6e7ccfa6c54f0b8a8ec61c8ab1a3c0c6c3f8875f0c6ad6d8b6c2c3d1f8a4\n"
	if err := os.WriteFile(pathValue, []byte(content), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	binding, err := loadPolicyBindingFromConfig(pathValue)
	if err != nil {
		t.Fatalf("loadPolicyBindingFromConfig error: %v", err)
	}
	if binding.PolicyURL != "https://policies.example.com/stricture/prod.yml" {
		t.Fatalf("PolicyURL = %q", binding.PolicyURL)
	}
	if binding.PolicySHA256 == "" {
		t.Fatalf("PolicySHA256 should not be empty")
	}
}

func TestLoadPolicyBindingFromConfigMissingURL(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	pathValue := filepath.Join(dir, ".stricture.yml")
	if err := os.WriteFile(pathValue, []byte("rules: {}\n"), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	if _, err := loadPolicyBindingFromConfig(pathValue); err == nil {
		t.Fatalf("expected missing strict:policy_url error")
	}
}

func TestResolveExpectedPolicyURL(t *testing.T) {
	t.Setenv("STRICTURE_POLICY_URL", "https://policies.example.com/stricture/prod.yml")
	got, source := resolveExpectedPolicyURL("", "STRICTURE_POLICY_URL")
	if got != "https://policies.example.com/stricture/prod.yml" {
		t.Fatalf("resolveExpectedPolicyURL env = %q", got)
	}
	if source != "env:STRICTURE_POLICY_URL" {
		t.Fatalf("resolveExpectedPolicyURL source = %q", source)
	}

	got, source = resolveExpectedPolicyURL("https://override.example/policy.yml", "STRICTURE_POLICY_URL")
	if got != "https://override.example/policy.yml" {
		t.Fatalf("resolveExpectedPolicyURL explicit should win, got %q", got)
	}
	if source != "flag:expected-url" {
		t.Fatalf("resolveExpectedPolicyURL explicit source = %q", source)
	}
}

func TestIsValidSHA256Hex(t *testing.T) {
	t.Parallel()

	if !isValidSHA256Hex("9f0e6e7ccfa6c54f0b8a8ec61c8ab1a3c0c6c3f8875f0c6ad6d8b6c2c3d1f8a4") {
		t.Fatalf("valid sha256 should pass")
	}
	if isValidSHA256Hex("short") {
		t.Fatalf("short value should fail")
	}
	if isValidSHA256Hex("zzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzz") {
		t.Fatalf("non-hex value should fail")
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

func TestFixArgs(t *testing.T) {
	cases := []struct {
		name string
		in   []string
		want []string
	}{
		{"applies by default", []string{"src"}, []string{"--fix", "src"}},
		{"dry-run previews instead of applying", []string{"--dry-run", "src"}, []string{"--fix-dry-run", "src"}},
		{"single-dash dry-run", []string{"-dry-run", "src"}, []string{"--fix-dry-run", "src"}},
		{"fix-dry-run alias is not combined with fix", []string{"--fix-dry-run", "src"}, []string{"--fix-dry-run", "src"}},
		{"other flags pass through", []string{"--fix-backup", "--rule", "CONV-file-header", "src"}, []string{"--fix", "--fix-backup", "--rule", "CONV-file-header", "src"}},
		{"dry-run=true previews", []string{"--dry-run=true", "src"}, []string{"--fix-dry-run", "src"}},
		{"dry-run=false applies", []string{"--dry-run=false", "src"}, []string{"--fix", "src"}},
		{"dry-run mixed with flags", []string{"--format", "json", "--dry-run"}, []string{"--fix-dry-run", "--format", "json"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := fixArgs(tc.in); !reflect.DeepEqual(got, tc.want) {
				t.Errorf("fixArgs(%v) = %v, want %v", tc.in, got, tc.want)
			}
		})
	}
}
