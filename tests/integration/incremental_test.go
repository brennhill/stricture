// incremental_test.go — Integration checks for --changed / --staged lint scopes.
//go:build integration

package integration

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestChangedOnlyLintsChangedFiles(t *testing.T) {
	tmp := t.TempDir()
	initGitRepo(t, tmp)

	aPath := filepath.Join(tmp, "a.ts")
	bPath := filepath.Join(tmp, "b.ts")
	if err := os.WriteFile(aPath, []byte("// a.ts — stable\nexport const a = 1;\n"), 0o644); err != nil {
		t.Fatalf("write a.ts: %v", err)
	}
	if err := os.WriteFile(bPath, []byte("// b.ts — stable\nexport const b = 1;\n"), 0o644); err != nil {
		t.Fatalf("write b.ts: %v", err)
	}
	runGit(t, tmp, "add", ".")
	runGit(t, tmp, "commit", "-m", "init")

	if err := os.WriteFile(aPath, []byte("export const a = 2;\n"), 0o644); err != nil {
		t.Fatalf("mutate a.ts: %v", err)
	}

	stdout, stderr, code := runInDir(t, tmp, "--format", "json", "--rule", "CONV-file-header", "--changed", ".")
	if code != 1 {
		t.Fatalf("--changed expected exit 1, got %d\nstderr=%q\nstdout=%q", code, stderr, stdout)
	}

	var payload struct {
		Violations []struct {
			FilePath string `json:"filePath"`
		} `json:"violations"`
	}
	if err := json.Unmarshal([]byte(stdout), &payload); err != nil {
		t.Fatalf("unmarshal changed output: %v\noutput=%q", err, stdout)
	}
	if len(payload.Violations) != 1 || filepath.Base(payload.Violations[0].FilePath) != "a.ts" {
		t.Fatalf("changed scope should only report a.ts, got %+v", payload.Violations)
	}
}

func TestStagedOnlyLintsStagedFiles(t *testing.T) {
	tmp := t.TempDir()
	initGitRepo(t, tmp)

	aPath := filepath.Join(tmp, "a.ts")
	bPath := filepath.Join(tmp, "b.ts")
	if err := os.WriteFile(aPath, []byte("// a.ts — stable\nexport const a = 1;\n"), 0o644); err != nil {
		t.Fatalf("write a.ts: %v", err)
	}
	if err := os.WriteFile(bPath, []byte("// b.ts — stable\nexport const b = 1;\n"), 0o644); err != nil {
		t.Fatalf("write b.ts: %v", err)
	}
	runGit(t, tmp, "add", ".")
	runGit(t, tmp, "commit", "-m", "init")

	if err := os.WriteFile(aPath, []byte("export const a = 2;\n"), 0o644); err != nil {
		t.Fatalf("mutate a.ts: %v", err)
	}
	if err := os.WriteFile(bPath, []byte("export const b = 2;\n"), 0o644); err != nil {
		t.Fatalf("mutate b.ts: %v", err)
	}
	runGit(t, tmp, "add", "b.ts")

	stdout, stderr, code := runInDir(t, tmp, "--format", "json", "--rule", "CONV-file-header", "--staged", ".")
	if code != 1 {
		t.Fatalf("--staged expected exit 1, got %d\nstderr=%q\nstdout=%q", code, stderr, stdout)
	}

	var payload struct {
		Violations []struct {
			FilePath string `json:"filePath"`
		} `json:"violations"`
	}
	if err := json.Unmarshal([]byte(stdout), &payload); err != nil {
		t.Fatalf("unmarshal staged output: %v\noutput=%q", err, stdout)
	}
	if len(payload.Violations) != 1 || filepath.Base(payload.Violations[0].FilePath) != "b.ts" {
		t.Fatalf("staged scope should only report b.ts, got %+v", payload.Violations)
	}
}

func TestChangedAndStagedMutuallyExclusive(t *testing.T) {
	tmp := t.TempDir()
	initGitRepo(t, tmp)
	if err := os.WriteFile(filepath.Join(tmp, "a.ts"), []byte("export const a = 1;\n"), 0o644); err != nil {
		t.Fatalf("write a.ts: %v", err)
	}

	_, stderr, code := runInDir(t, tmp, "--changed", "--staged", ".")
	if code != 2 {
		t.Fatalf("expected exit 2 for mutually exclusive flags, got %d", code)
	}
	if !strings.Contains(strings.ToLower(stderr), "mutually exclusive") {
		t.Fatalf("stderr should mention mutually exclusive flags, got %q", stderr)
	}
}

func initGitRepo(t *testing.T, dir string) {
	t.Helper()
	runGit(t, dir, "init")
	runGit(t, dir, "config", "user.email", "stricture-test@example.com")
	runGit(t, dir, "config", "user.name", "Stricture Test")
}

func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %s failed: %v\n%s", strings.Join(args, " "), err, string(out))
	}
}
