// output_controls_test.go — Integration checks for color and verbose output flags.
//go:build integration

package integration

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestColorFlagsControlTextOutput(t *testing.T) {
	tmp := t.TempDir()
	source := filepath.Join(tmp, "bad.go")
	if err := os.WriteFile(source, []byte("package main\n\nfunc main() {}\n"), 0o644); err != nil {
		t.Fatalf("write source: %v", err)
	}

	stdout, stderr, code := runInDir(t, tmp,
		"--format", "text",
		"--rule", "CONV-file-header",
		"--color",
		".",
	)
	if code != 1 {
		t.Fatalf("--color run should fail with one violation: code=%d stderr=%q stdout=%q", code, stderr, stdout)
	}
	if !strings.Contains(stdout, "\x1b[") {
		t.Fatalf("expected ANSI color escape sequence with --color, got %q", stdout)
	}

	stdout, stderr, code = runInDir(t, tmp,
		"--format", "text",
		"--rule", "CONV-file-header",
		"--no-color",
		".",
	)
	if code != 1 {
		t.Fatalf("--no-color run should fail with one violation: code=%d stderr=%q stdout=%q", code, stderr, stdout)
	}
	if strings.Contains(stdout, "\x1b[") {
		t.Fatalf("did not expect ANSI color escape sequence with --no-color, got %q", stdout)
	}

	_, stderr, code = runInDir(t, tmp,
		"--format", "text",
		"--rule", "CONV-file-header",
		"--color",
		"--no-color",
		".",
	)
	if code != 2 {
		t.Fatalf("--color and --no-color together should exit 2, got %d", code)
	}
	if !strings.Contains(strings.ToLower(stderr), "mutually exclusive") {
		t.Fatalf("stderr should explain color flag conflict, got %q", stderr)
	}
}

func TestVerboseWritesDiagnosticsToStderr(t *testing.T) {
	tmp := t.TempDir()
	source := filepath.Join(tmp, "bad.go")
	if err := os.WriteFile(source, []byte("package main\n\nfunc main() {}\n"), 0o644); err != nil {
		t.Fatalf("write source: %v", err)
	}

	stdout, stderr, code := runInDir(t, tmp,
		"--format", "json",
		"--rule", "CONV-file-header",
		"--verbose",
		".",
	)
	if code != 1 {
		t.Fatalf("--verbose run should fail with one violation: code=%d stderr=%q stdout=%q", code, stderr, stdout)
	}
	if !strings.Contains(stderr, "Verbose:") {
		t.Fatalf("expected verbose diagnostics on stderr, got %q", stderr)
	}

	var payload map[string]interface{}
	if err := json.Unmarshal([]byte(stdout), &payload); err != nil {
		t.Fatalf("stdout should remain valid JSON with --verbose: %v\noutput=%q", err, stdout)
	}
}
