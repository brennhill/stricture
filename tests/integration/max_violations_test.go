// max_violations_test.go — Integration checks for --max-violations behavior.
//go:build integration

package integration

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestMaxViolationsLimitsOutput(t *testing.T) {
	tmp := t.TempDir()
	for _, name := range []string{"a.ts", "b.ts"} {
		pathValue := filepath.Join(tmp, name)
		if err := os.WriteFile(pathValue, []byte("export const value = 1;\n"), 0o644); err != nil {
			t.Fatalf("write %s: %v", pathValue, err)
		}
	}

	stdout, stderr, code := runInDir(t, tmp, "--format", "json", "--rule", "CONV-file-header", ".")
	if code != 1 {
		t.Fatalf("baseline lint exit code = %d, want 1\nstderr=%q", code, stderr)
	}

	var baseline struct {
		Violations []map[string]interface{} `json:"violations"`
		Summary    struct {
			TotalViolations float64 `json:"totalViolations"`
		} `json:"summary"`
	}
	if err := json.Unmarshal([]byte(stdout), &baseline); err != nil {
		t.Fatalf("unmarshal baseline JSON: %v\noutput=%q", err, stdout)
	}
	if len(baseline.Violations) != 2 || int(baseline.Summary.TotalViolations) != 2 {
		t.Fatalf("baseline should report 2 violations, got len=%d total=%v", len(baseline.Violations), baseline.Summary.TotalViolations)
	}

	stdout, stderr, code = runInDir(t, tmp, "--format", "json", "--rule", "CONV-file-header", "--max-violations", "1", ".")
	if code != 1 {
		t.Fatalf("max-violations lint exit code = %d, want 1\nstderr=%q", code, stderr)
	}

	var limited struct {
		Violations []map[string]interface{} `json:"violations"`
		Summary    struct {
			TotalViolations float64 `json:"totalViolations"`
		} `json:"summary"`
	}
	if err := json.Unmarshal([]byte(stdout), &limited); err != nil {
		t.Fatalf("unmarshal limited JSON: %v\noutput=%q", err, stdout)
	}
	if len(limited.Violations) != 1 || int(limited.Summary.TotalViolations) != 1 {
		t.Fatalf("limited run should report 1 violation, got len=%d total=%v", len(limited.Violations), limited.Summary.TotalViolations)
	}
}

func TestMaxViolationsRejectsNegativeValue(t *testing.T) {
	tmp := t.TempDir()
	_, stderr, code := runInDir(t, tmp, "--max-violations", "-1", ".")
	if code != 2 {
		t.Fatalf("negative max-violations exit code = %d, want 2", code)
	}
	if stderr == "" {
		t.Fatalf("stderr should explain invalid max-violations")
	}
}
