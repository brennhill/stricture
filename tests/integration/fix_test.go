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

func TestFixBackupCreatesBackupFile(t *testing.T) {
	tmp := t.TempDir()
	target := filepath.Join(tmp, "user_service.ts")
	original := "export const value = 1;\n"
	if err := os.WriteFile(target, []byte(original), 0o644); err != nil {
		t.Fatalf("write target: %v", err)
	}

	_, stderr, code := run(t, "--fix", "--fix-backup", target)
	if code != 0 {
		t.Fatalf("--fix --fix-backup exit code = %d, want 0\nstderr=%q", code, stderr)
	}

	backup, err := os.ReadFile(target + ".bak")
	if err != nil {
		t.Fatalf("read backup: %v", err)
	}
	if string(backup) != original {
		t.Fatalf("backup content mismatch: got %q want %q", string(backup), original)
	}
}

func TestFixBackupRequiresFix(t *testing.T) {
	tmp := t.TempDir()
	target := filepath.Join(tmp, "user_service.ts")
	if err := os.WriteFile(target, []byte("export const value = 1;\n"), 0o644); err != nil {
		t.Fatalf("write target: %v", err)
	}

	_, stderr, code := run(t, "--fix-backup", target)
	if code != 2 {
		t.Fatalf("--fix-backup without --fix exit code = %d, want 2\nstderr=%q", code, stderr)
	}
}

func TestFixBackupFailsWhenBackupAlreadyExists(t *testing.T) {
	tmp := t.TempDir()
	target := filepath.Join(tmp, "user_service.ts")
	original := "export const value = 1;\n"
	if err := os.WriteFile(target, []byte(original), 0o644); err != nil {
		t.Fatalf("write target: %v", err)
	}
	if err := os.WriteFile(target+".bak", []byte("existing backup\n"), 0o644); err != nil {
		t.Fatalf("write existing backup: %v", err)
	}

	_, stderr, code := run(t, "--fix", "--fix-backup", target)
	if code != 1 {
		t.Fatalf("expected backup collision to fail with exit 1, got %d\nstderr=%q", code, stderr)
	}
	if !strings.Contains(strings.ToLower(stderr), "backup already exists") {
		t.Fatalf("stderr should mention backup collision, got %q", stderr)
	}

	after, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("read target: %v", err)
	}
	if string(after) != original {
		t.Fatalf("source file should remain unchanged when backup creation fails")
	}
}
