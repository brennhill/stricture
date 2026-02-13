// baseline_test.go — Integration checks for --baseline behavior.
//go:build integration

package integration

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestBaselineBootstrapsAndFiltersExistingViolations(t *testing.T) {
	tmp := t.TempDir()
	for _, name := range []string{"a.ts", "b.ts"} {
		pathValue := filepath.Join(tmp, name)
		if err := os.WriteFile(pathValue, []byte("export const value = 1;\n"), 0o644); err != nil {
			t.Fatalf("write %s: %v", pathValue, err)
		}
	}

	baselinePath := filepath.Join(tmp, ".stricture-baseline.json")
	stdout, stderr, code := runInDir(t, tmp, "--format", "json", "--rule", "CONV-file-header", "--baseline", baselinePath, ".")
	if code != 0 {
		t.Fatalf("baseline bootstrap should exit 0, got %d\nstderr=%q\nstdout=%q", code, stderr, stdout)
	}

	if _, err := os.Stat(baselinePath); err != nil {
		t.Fatalf("baseline file should be created: %v", err)
	}

	var result struct {
		Violations []map[string]interface{} `json:"violations"`
		Summary    struct {
			TotalViolations float64 `json:"totalViolations"`
		} `json:"summary"`
	}
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Fatalf("unmarshal bootstrap JSON: %v\noutput=%q", err, stdout)
	}
	if len(result.Violations) != 0 || int(result.Summary.TotalViolations) != 0 {
		t.Fatalf("bootstrap run should report zero violations, got len=%d total=%v", len(result.Violations), result.Summary.TotalViolations)
	}

	stdout, stderr, code = runInDir(t, tmp, "--format", "json", "--rule", "CONV-file-header", "--baseline", baselinePath, ".")
	if code != 0 {
		t.Fatalf("baseline replay should exit 0, got %d\nstderr=%q\nstdout=%q", code, stderr, stdout)
	}

	if err := os.WriteFile(filepath.Join(tmp, "c.ts"), []byte("export const value = 1;\n"), 0o644); err != nil {
		t.Fatalf("write c.ts: %v", err)
	}
	stdout, stderr, code = runInDir(t, tmp, "--format", "json", "--rule", "CONV-file-header", "--baseline", baselinePath, ".")
	if code != 1 {
		t.Fatalf("new violation beyond baseline should exit 1, got %d\nstderr=%q\nstdout=%q", code, stderr, stdout)
	}
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Fatalf("unmarshal new violation JSON: %v\noutput=%q", err, stdout)
	}
	if len(result.Violations) != 1 || int(result.Summary.TotalViolations) != 1 {
		t.Fatalf("expected one non-baselined violation, got len=%d total=%v", len(result.Violations), result.Summary.TotalViolations)
	}
}

func TestBaselineRejectsMalformedFile(t *testing.T) {
	tmp := t.TempDir()
	baselinePath := filepath.Join(tmp, ".stricture-baseline.json")
	if err := os.WriteFile(baselinePath, []byte("{not-json"), 0o644); err != nil {
		t.Fatalf("write malformed baseline: %v", err)
	}
	_, stderr, code := runInDir(t, tmp, "--baseline", baselinePath, ".")
	if code != 2 {
		t.Fatalf("malformed baseline exit code = %d, want 2", code)
	}
	if stderr == "" {
		t.Fatalf("stderr should explain malformed baseline file")
	}
}

func TestBaselineWithMaxViolationsStillReportsNewFindings(t *testing.T) {
	tmp := t.TempDir()
	aPath := filepath.Join(tmp, "a.ts")
	bPath := filepath.Join(tmp, "b.ts")
	if err := os.WriteFile(aPath, []byte("export const a = 1;\n"), 0o644); err != nil {
		t.Fatalf("write a.ts: %v", err)
	}
	if err := os.WriteFile(bPath, []byte("export const b = 1;\n"), 0o644); err != nil {
		t.Fatalf("write b.ts: %v", err)
	}

	baselinePath := filepath.Join(tmp, ".stricture-baseline.json")
	_, stderr, code := runInDir(t, tmp, "--format", "json", "--rule", "CONV-file-header", "--baseline", baselinePath, ".")
	if code != 0 {
		t.Fatalf("baseline bootstrap failed: code=%d stderr=%q", code, stderr)
	}

	cPath := filepath.Join(tmp, "c.ts")
	if err := os.WriteFile(cPath, []byte("export const c = 1;\n"), 0o644); err != nil {
		t.Fatalf("write c.ts: %v", err)
	}

	stdout, stderr, code := runInDir(t, tmp, "--format", "json", "--rule", "CONV-file-header", "--baseline", baselinePath, "--max-violations", "1", ".")
	if code != 1 {
		t.Fatalf("expected new violation to be reported with exit 1, got code=%d stderr=%q stdout=%q", code, stderr, stdout)
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
		t.Fatalf("unmarshal JSON output: %v\noutput=%q", err, stdout)
	}
	if len(result.Violations) != 1 || int(result.Summary.TotalViolations) != 1 {
		t.Fatalf("expected one new violation with max-violations=1, got len=%d total=%v", len(result.Violations), result.Summary.TotalViolations)
	}
	if filepath.Base(result.Violations[0].FilePath) != "c.ts" {
		t.Fatalf("expected c.ts as new violation, got %q", result.Violations[0].FilePath)
	}
}
