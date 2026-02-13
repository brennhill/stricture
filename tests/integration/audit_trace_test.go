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
