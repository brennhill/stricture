// audit_trace_test.go — Integration checks for audit/trace command availability.
//go:build integration

package integration

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAuditHelp(t *testing.T) {
	stdout, stderr, code := run(t, "audit", "--help")
	if code != 0 {
		t.Fatalf("audit --help exit code = %d, want 0\nstderr=%q\nstdout=%q", code, stderr, stdout)
	}
	if !strings.Contains(stdout, "Usage: stricture audit") {
		t.Fatalf("audit --help missing usage line, got %q", stdout)
	}
}

func TestAuditMissingManifestExitsTwo(t *testing.T) {
	tmp := t.TempDir()
	_, stderr, code := runInDir(t, tmp, "audit", "--manifest", "missing-manifest.yml", ".")
	if code != 2 {
		t.Fatalf("audit with missing manifest exit code = %d, want 2", code)
	}
	if !strings.Contains(strings.ToLower(stderr), "manifest") {
		t.Fatalf("stderr should mention missing manifest, got %q", stderr)
	}
}

func TestAuditInvalidManifestExitsTwo(t *testing.T) {
	tmp := t.TempDir()
	manifestPath := filepath.Join(tmp, "stricture-manifest.yml")
	if err := os.WriteFile(manifestPath, []byte("manifest_version: \"\"\ncontracts: []\n"), 0o644); err != nil {
		t.Fatalf("write manifest: %v", err)
	}

	_, stderr, code := runInDir(t, tmp, "audit", "--manifest", manifestPath, ".")
	if code != 2 {
		t.Fatalf("audit with invalid manifest exit code = %d, want 2", code)
	}
	if !strings.Contains(strings.ToLower(stderr), "manifest") {
		t.Fatalf("stderr should mention invalid manifest, got %q", stderr)
	}
}

func TestAuditValidManifestRuns(t *testing.T) {
	tmp := t.TempDir()
	manifestPath := filepath.Join(tmp, "stricture-manifest.yml")
	manifest := "manifest_version: \"v1\"\nstrictness_level: \"standard\"\ncontracts:\n  - id: users.v1\n    endpoint: /users\n    method: GET\n"
	if err := os.WriteFile(manifestPath, []byte(manifest), 0o644); err != nil {
		t.Fatalf("write manifest: %v", err)
	}

	stdout, stderr, code := runInDir(t, tmp, "audit", "--manifest", manifestPath, ".")
	if code != 0 {
		t.Fatalf("audit with valid manifest exit code = %d, want 0\nstderr=%q\nstdout=%q", code, stderr, stdout)
	}
}

func TestTraceRequiresFileArgument(t *testing.T) {
	_, stderr, code := run(t, "trace")
	if code != 2 {
		t.Fatalf("trace without file exit code = %d, want 2", code)
	}
	if !strings.Contains(strings.ToLower(stderr), "requires a trace file") {
		t.Fatalf("stderr should mention required trace file, got %q", stderr)
	}
}

func TestTraceParsesJSONFile(t *testing.T) {
	tmp := t.TempDir()
	tracePath := filepath.Join(tmp, "trace.json")
	if err := os.WriteFile(tracePath, []byte(`{"events":[]}`), 0o644); err != nil {
		t.Fatalf("write trace file: %v", err)
	}

	stdout, stderr, code := runInDir(t, tmp, "trace", "trace.json", "--trace-format", "custom")
	if code != 0 {
		t.Fatalf("trace parse exit code = %d, want 0\nstderr=%q\nstdout=%q", code, stderr, stdout)
	}
	if !strings.Contains(stdout, "Trace parsed") {
		t.Fatalf("trace output should confirm parse, got %q", stdout)
	}
}

func TestTraceMissingManifestExitsTwo(t *testing.T) {
	tmp := t.TempDir()
	tracePath := filepath.Join(tmp, "trace.json")
	if err := os.WriteFile(tracePath, []byte(`{"events":[]}`), 0o644); err != nil {
		t.Fatalf("write trace file: %v", err)
	}

	_, stderr, code := runInDir(t, tmp, "trace", "trace.json", "--manifest", "missing.yml")
	if code != 2 {
		t.Fatalf("trace with missing manifest exit code = %d, want 2", code)
	}
	if !strings.Contains(strings.ToLower(stderr), "manifest") {
		t.Fatalf("stderr should mention missing manifest, got %q", stderr)
	}
}

func TestTraceInvalidJSONExitsTwo(t *testing.T) {
	tmp := t.TempDir()
	tracePath := filepath.Join(tmp, "trace.json")
	if err := os.WriteFile(tracePath, []byte("{not-json"), 0o644); err != nil {
		t.Fatalf("write trace file: %v", err)
	}

	_, stderr, code := runInDir(t, tmp, "trace", "trace.json", "--trace-format", "custom")
	if code != 2 {
		t.Fatalf("trace invalid JSON exit code = %d, want 2", code)
	}
	if !strings.Contains(strings.ToLower(stderr), "not valid json") {
		t.Fatalf("stderr should mention invalid JSON, got %q", stderr)
	}
}

func TestAuditRejectsInvalidStrictness(t *testing.T) {
	tmp := t.TempDir()
	_, stderr, code := runInDir(t, tmp, "audit", "--strictness", "wild", ".")
	if code != 2 {
		t.Fatalf("audit invalid strictness exit code = %d, want 2", code)
	}
	if !strings.Contains(strings.ToLower(stderr), "strictness") {
		t.Fatalf("stderr should mention strictness validation, got %q", stderr)
	}
}

func TestTraceHarValidationRequiresLogEnvelope(t *testing.T) {
	tmp := t.TempDir()
	tracePath := filepath.Join(tmp, "trace.har")
	if err := os.WriteFile(tracePath, []byte(`{"events":[]}`), 0o644); err != nil {
		t.Fatalf("write trace file: %v", err)
	}

	_, stderr, code := runInDir(t, tmp, "trace", "trace.har", "--trace-format", "har")
	if code != 2 {
		t.Fatalf("har trace without log envelope exit code = %d, want 2", code)
	}
	if !strings.Contains(strings.ToLower(stderr), "har") {
		t.Fatalf("stderr should mention har validation, got %q", stderr)
	}
}

func TestTraceOtelValidationRequiresResourceSpans(t *testing.T) {
	tmp := t.TempDir()
	tracePath := filepath.Join(tmp, "trace.json")
	if err := os.WriteFile(tracePath, []byte(`{"events":[]}`), 0o644); err != nil {
		t.Fatalf("write trace file: %v", err)
	}

	_, stderr, code := runInDir(t, tmp, "trace", "trace.json", "--trace-format", "otel")
	if code != 2 {
		t.Fatalf("otel trace without resourceSpans exit code = %d, want 2", code)
	}
	if !strings.Contains(strings.ToLower(stderr), "resource") {
		t.Fatalf("stderr should mention resourceSpans validation, got %q", stderr)
	}
}

func TestTraceAutoDetectsHarFromExtension(t *testing.T) {
	tmp := t.TempDir()
	tracePath := filepath.Join(tmp, "trace.har")
	if err := os.WriteFile(tracePath, []byte(`{"log":{"version":"1.2"}}`), 0o644); err != nil {
		t.Fatalf("write trace file: %v", err)
	}

	stdout, stderr, code := runInDir(t, tmp, "trace", "trace.har")
	if code != 0 {
		t.Fatalf("auto-detect HAR exit code = %d, want 0\nstderr=%q\nstdout=%q", code, stderr, stdout)
	}
	if !strings.Contains(strings.ToLower(stdout), "format=har") {
		t.Fatalf("trace output should report har format, got %q", stdout)
	}
}
