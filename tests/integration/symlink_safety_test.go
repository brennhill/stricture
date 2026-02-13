// symlink_safety_test.go — Integration checks for symlink traversal safety.
//go:build integration

package integration

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestSymlinkOutsideProjectIsIgnored(t *testing.T) {
	projectDir := t.TempDir()
	outsideDir := t.TempDir()

	local := filepath.Join(projectDir, "local.ts")
	if err := os.WriteFile(local, []byte("export const local = 1;\n"), 0o644); err != nil {
		t.Fatalf("write local file: %v", err)
	}

	outside := filepath.Join(outsideDir, "outside.ts")
	if err := os.WriteFile(outside, []byte("export const outside = 1;\n"), 0o644); err != nil {
		t.Fatalf("write outside file: %v", err)
	}

	link := filepath.Join(projectDir, "linked.ts")
	if err := os.Symlink(outside, link); err != nil {
		t.Skipf("symlinks unsupported in this environment: %v", err)
	}

	stdout, stderr, code := runInDir(t, projectDir, "--format", "json", "--rule", "CONV-file-header", ".")
	if code != 1 {
		t.Fatalf("lint exit code = %d, want 1\nstderr=%q\nstdout=%q", code, stderr, stdout)
	}

	var result struct {
		Violations []struct {
			FilePath string `json:"filePath"`
		} `json:"violations"`
		Summary struct {
			TotalViolations float64 `json:"totalViolations"`
		} `json:"summary"`
	}
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Fatalf("unmarshal output JSON: %v\noutput=%q", err, stdout)
	}
	if len(result.Violations) != 1 || int(result.Summary.TotalViolations) != 1 {
		t.Fatalf("expected one violation from local file only, got len=%d total=%v", len(result.Violations), result.Summary.TotalViolations)
	}
	if filepath.Base(result.Violations[0].FilePath) != "local.ts" {
		t.Fatalf("expected only local.ts violation, got %q", result.Violations[0].FilePath)
	}
}
