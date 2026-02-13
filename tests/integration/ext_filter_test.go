// ext_filter_test.go — Integration checks for --ext file targeting.
//go:build integration

package integration

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestExtFiltersLintTargetFiles(t *testing.T) {
	tmp := t.TempDir()
	goFile := filepath.Join(tmp, "bad.go")
	tsFile := filepath.Join(tmp, "bad.ts")
	if err := os.WriteFile(goFile, []byte("package main\n\nfunc main() {}\n"), 0o644); err != nil {
		t.Fatalf("write go file: %v", err)
	}
	if err := os.WriteFile(tsFile, []byte("export const x = 1;\n"), 0o644); err != nil {
		t.Fatalf("write ts file: %v", err)
	}

	stdout, stderr, code := runInDir(t, tmp,
		"--format", "json",
		"--rule", "CONV-file-header",
		".",
	)
	if code != 1 {
		t.Fatalf("baseline run should find violations: code=%d stderr=%q stdout=%q", code, stderr, stdout)
	}
	var baseline struct {
		Violations []struct {
			FilePath string `json:"filePath"`
		} `json:"violations"`
		Summary struct {
			TotalViolations int `json:"totalViolations"`
		} `json:"summary"`
	}
	if err := json.Unmarshal([]byte(stdout), &baseline); err != nil {
		t.Fatalf("unmarshal baseline output: %v\noutput=%q", err, stdout)
	}
	if baseline.Summary.TotalViolations != 2 || len(baseline.Violations) != 2 {
		t.Fatalf("expected 2 violations without ext filter, got total=%d len=%d", baseline.Summary.TotalViolations, len(baseline.Violations))
	}

	stdout, stderr, code = runInDir(t, tmp,
		"--format", "json",
		"--rule", "CONV-file-header",
		"--ext", "go",
		".",
	)
	if code != 1 {
		t.Fatalf("go extension filter run should find one violation: code=%d stderr=%q stdout=%q", code, stderr, stdout)
	}
	var goOnly struct {
		Violations []struct {
			FilePath string `json:"filePath"`
		} `json:"violations"`
		Summary struct {
			TotalViolations int `json:"totalViolations"`
		} `json:"summary"`
	}
	if err := json.Unmarshal([]byte(stdout), &goOnly); err != nil {
		t.Fatalf("unmarshal go-only output: %v\noutput=%q", err, stdout)
	}
	if goOnly.Summary.TotalViolations != 1 || len(goOnly.Violations) != 1 {
		t.Fatalf("expected 1 violation with --ext go, got total=%d len=%d", goOnly.Summary.TotalViolations, len(goOnly.Violations))
	}
	if !strings.HasSuffix(goOnly.Violations[0].FilePath, "bad.go") {
		t.Fatalf("--ext go should only report bad.go, got %q", goOnly.Violations[0].FilePath)
	}

	stdout, stderr, code = runInDir(t, tmp,
		"--format", "json",
		"--rule", "CONV-file-header",
		"--ext", ".ts",
		".",
	)
	if code != 1 {
		t.Fatalf("ts extension filter run should find one violation: code=%d stderr=%q stdout=%q", code, stderr, stdout)
	}
	var tsOnly struct {
		Violations []struct {
			FilePath string `json:"filePath"`
		} `json:"violations"`
		Summary struct {
			TotalViolations int `json:"totalViolations"`
		} `json:"summary"`
	}
	if err := json.Unmarshal([]byte(stdout), &tsOnly); err != nil {
		t.Fatalf("unmarshal ts-only output: %v\noutput=%q", err, stdout)
	}
	if tsOnly.Summary.TotalViolations != 1 || len(tsOnly.Violations) != 1 {
		t.Fatalf("expected 1 violation with --ext .ts, got total=%d len=%d", tsOnly.Summary.TotalViolations, len(tsOnly.Violations))
	}
	if !strings.HasSuffix(tsOnly.Violations[0].FilePath, "bad.ts") {
		t.Fatalf("--ext .ts should only report bad.ts, got %q", tsOnly.Violations[0].FilePath)
	}
}
