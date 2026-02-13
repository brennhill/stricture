// diff_test.go — Integration checks for baseline diff mode.
//go:build integration

package integration

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDiffRequiresBaseline(t *testing.T) {
	tmp := t.TempDir()
	if err := os.WriteFile(filepath.Join(tmp, "a.ts"), []byte("export const value = 1;\n"), 0o644); err != nil {
		t.Fatalf("write a.ts: %v", err)
	}

	_, stderr, code := runInDir(t, tmp, "--diff", ".")
	if code != 2 {
		t.Fatalf("--diff without --baseline exit code = %d, want 2", code)
	}
	if !strings.Contains(strings.ToLower(stderr), "requires --baseline") {
		t.Fatalf("stderr should explain baseline requirement, got %q", stderr)
	}
}

func TestDiffMissingBaselineFails(t *testing.T) {
	tmp := t.TempDir()
	if err := os.WriteFile(filepath.Join(tmp, "a.ts"), []byte("export const value = 1;\n"), 0o644); err != nil {
		t.Fatalf("write a.ts: %v", err)
	}

	missing := filepath.Join(tmp, "missing-baseline.json")
	_, stderr, code := runInDir(t, tmp, "--baseline", missing, "--diff", ".")
	if code != 2 {
		t.Fatalf("--diff with missing baseline exit code = %d, want 2", code)
	}
	if !strings.Contains(strings.ToLower(stderr), "does not exist") {
		t.Fatalf("stderr should mention missing baseline, got %q", stderr)
	}
}

func TestDiffReportsAddedAndResolved(t *testing.T) {
	tmp := t.TempDir()
	aPath := filepath.Join(tmp, "a.ts")
	if err := os.WriteFile(aPath, []byte("export const value = 1;\n"), 0o644); err != nil {
		t.Fatalf("write a.ts: %v", err)
	}
	baselinePath := filepath.Join(tmp, ".stricture-baseline.json")

	_, stderr, code := runInDir(t, tmp, "--format", "json", "--rule", "CONV-file-header", "--baseline", baselinePath, ".")
	if code != 0 {
		t.Fatalf("baseline bootstrap failed: code=%d stderr=%q", code, stderr)
	}

	// Resolve existing baseline issue and add a new one.
	if err := os.WriteFile(aPath, []byte("// a.ts — module\nexport const value = 1;\n"), 0o644); err != nil {
		t.Fatalf("rewrite a.ts: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmp, "b.ts"), []byte("export const value = 2;\n"), 0o644); err != nil {
		t.Fatalf("write b.ts: %v", err)
	}

	stdout, stderr, code := runInDir(t, tmp, "--format", "json", "--rule", "CONV-file-header", "--baseline", baselinePath, "--diff", ".")
	if code != 1 {
		t.Fatalf("diff run should fail on added violation: code=%d stderr=%q stdout=%q", code, stderr, stdout)
	}

	var result struct {
		Violations []struct {
			FilePath string `json:"filePath"`
		} `json:"violations"`
		Diff struct {
			Enabled bool `json:"enabled"`
			Added   []struct {
				FilePath string `json:"filePath"`
			} `json:"added"`
			Resolved []struct {
				FilePath string `json:"filePath"`
			} `json:"resolved"`
			Summary struct {
				Added    int `json:"added"`
				Resolved int `json:"resolved"`
			} `json:"summary"`
		} `json:"diff"`
	}
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Fatalf("unmarshal diff output: %v\noutput=%q", err, stdout)
	}
	if !result.Diff.Enabled {
		t.Fatalf("diff payload should be enabled")
	}
	if result.Diff.Summary.Added != 1 || result.Diff.Summary.Resolved != 1 {
		t.Fatalf("unexpected diff summary: %+v", result.Diff.Summary)
	}
	if len(result.Violations) != 1 || filepath.Base(result.Violations[0].FilePath) != "b.ts" {
		t.Fatalf("violations should contain only new b.ts issue: %+v", result.Violations)
	}
	if len(result.Diff.Added) != 1 || filepath.Base(result.Diff.Added[0].FilePath) != "b.ts" {
		t.Fatalf("diff added should contain only b.ts: %+v", result.Diff.Added)
	}
	if len(result.Diff.Resolved) != 1 || filepath.Base(result.Diff.Resolved[0].FilePath) != "a.ts" {
		t.Fatalf("diff resolved should contain only a.ts: %+v", result.Diff.Resolved)
	}
}
