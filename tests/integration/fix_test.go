// fix_test.go — Integration checks for fix command behavior.
//go:build integration

package integration

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFixDryRunDoesNotModifyFiles(t *testing.T) {
	tmp := t.TempDir()
	target := filepath.Join(tmp, "user-service.ts")
	original := "export const value = 1;\n"
	if err := os.WriteFile(target, []byte(original), 0o644); err != nil {
		t.Fatalf("write target: %v", err)
	}

	stdout, stderr, code := run(t, "--fix-dry-run", target)
	if code == 2 {
		t.Fatalf("--fix-dry-run must be implemented, got exit 2\nstderr=%q", stderr)
	}
	if !strings.Contains(stdout, "Fixes:") {
		t.Fatalf("dry-run output should include fixes summary, got %q", stdout)
	}

	after, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("read target: %v", err)
	}
	if string(after) != original {
		t.Fatalf("dry-run modified file unexpectedly")
	}
}

func TestFixAppliesHeaderForFixableRule(t *testing.T) {
	tmp := t.TempDir()
	target := filepath.Join(tmp, "user_service.ts")
	if err := os.WriteFile(target, []byte("export const value = 1;\n"), 0o644); err != nil {
		t.Fatalf("write target: %v", err)
	}

	_, stderr, code := run(t, "--fix", target)
	if code != 0 {
		t.Fatalf("--fix exit code = %d, want 0\nstderr=%q", code, stderr)
	}

	after, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("read target: %v", err)
	}
	if !strings.HasPrefix(string(after), "// user_service.ts — ") {
		t.Fatalf("fixed file missing header, got:\n%s", string(after))
	}
}
