// severity_filter_test.go — Integration checks for warning filtering flags.
//go:build integration

package integration

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type lintJSONPayload struct {
	Violations []struct {
		Severity string `json:"severity"`
	} `json:"violations"`
	Summary struct {
		TotalViolations int `json:"totalViolations"`
		Errors          int `json:"errors"`
		Warnings        int `json:"warnings"`
	} `json:"summary"`
}

func createWarnOnlyProject(t *testing.T) (projectDir string, configPath string) {
	t.Helper()

	projectDir = t.TempDir()
	src := filepath.Join(projectDir, "bad.go")
	if err := os.WriteFile(src, []byte("package main\n\nfunc main() {}\n"), 0o644); err != nil {
		t.Fatalf("write source: %v", err)
	}

	configPath = filepath.Join(projectDir, ".stricture.yml")
	cfg := `rules:
  CONV-file-header: warn
`
	if err := os.WriteFile(configPath, []byte(cfg), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	return projectDir, configPath
}

func parseLintJSON(t *testing.T, stdout string) lintJSONPayload {
	t.Helper()
	var payload lintJSONPayload
	if err := json.Unmarshal([]byte(stdout), &payload); err != nil {
		t.Fatalf("unmarshal lint JSON: %v\noutput=%q", err, stdout)
	}
	return payload
}

func TestSeverityWarnIncludesWarnings(t *testing.T) {
	projectDir, configPath := createWarnOnlyProject(t)

	stdout, stderr, code := runInDir(t, projectDir,
		"--format", "json",
		"--config", configPath,
		"--rule", "CONV-file-header",
		"--severity", "warn",
		".",
	)
	if code != 0 {
		t.Fatalf("warn-only run should exit 0, got %d\nstderr=%q\nstdout=%q", code, stderr, stdout)
	}

	payload := parseLintJSON(t, stdout)
	if len(payload.Violations) != 1 || payload.Summary.TotalViolations != 1 {
		t.Fatalf("expected one warning violation, got len=%d total=%d", len(payload.Violations), payload.Summary.TotalViolations)
	}
	if strings.ToLower(payload.Violations[0].Severity) != "warn" {
		t.Fatalf("expected warn severity, got %q", payload.Violations[0].Severity)
	}
	if payload.Summary.Errors != 0 || payload.Summary.Warnings != 1 {
		t.Fatalf("unexpected summary counts: errors=%d warnings=%d", payload.Summary.Errors, payload.Summary.Warnings)
	}
}

func TestSeverityErrorFiltersWarnings(t *testing.T) {
	projectDir, configPath := createWarnOnlyProject(t)

	stdout, stderr, code := runInDir(t, projectDir,
		"--format", "json",
		"--config", configPath,
		"--rule", "CONV-file-header",
		"--severity", "error",
		".",
	)
	if code != 0 {
		t.Fatalf("error-filtered warn-only run should exit 0, got %d\nstderr=%q\nstdout=%q", code, stderr, stdout)
	}

	payload := parseLintJSON(t, stdout)
	if len(payload.Violations) != 0 || payload.Summary.TotalViolations != 0 {
		t.Fatalf("warnings should be filtered at error severity, got len=%d total=%d", len(payload.Violations), payload.Summary.TotalViolations)
	}
	if payload.Summary.Errors != 0 || payload.Summary.Warnings != 0 {
		t.Fatalf("unexpected summary counts: errors=%d warnings=%d", payload.Summary.Errors, payload.Summary.Warnings)
	}
}

func TestQuietSuppressesWarnings(t *testing.T) {
	projectDir, configPath := createWarnOnlyProject(t)

	stdout, stderr, code := runInDir(t, projectDir,
		"--format", "json",
		"--config", configPath,
		"--rule", "CONV-file-header",
		"--quiet",
		".",
	)
	if code != 0 {
		t.Fatalf("quiet warn-only run should exit 0, got %d\nstderr=%q\nstdout=%q", code, stderr, stdout)
	}

	payload := parseLintJSON(t, stdout)
	if len(payload.Violations) != 0 || payload.Summary.TotalViolations != 0 {
		t.Fatalf("quiet should suppress warnings, got len=%d total=%d", len(payload.Violations), payload.Summary.TotalViolations)
	}
}

func TestSeverityAndQuietConflict(t *testing.T) {
	projectDir, configPath := createWarnOnlyProject(t)

	_, stderr, code := runInDir(t, projectDir,
		"--config", configPath,
		"--rule", "CONV-file-header",
		"--severity", "warn",
		"--quiet",
		".",
	)
	if code != 2 {
		t.Fatalf("--severity warn with --quiet exit code = %d, want 2", code)
	}
	if !strings.Contains(strings.ToLower(stderr), "cannot be combined") {
		t.Fatalf("stderr should explain severity/quiet conflict, got %q", stderr)
	}
}

func TestInvalidSeverityValueExitsTwo(t *testing.T) {
	projectDir, configPath := createWarnOnlyProject(t)

	_, stderr, code := runInDir(t, projectDir,
		"--config", configPath,
		"--rule", "CONV-file-header",
		"--severity", "fatal",
		".",
	)
	if code != 2 {
		t.Fatalf("invalid severity exit code = %d, want 2", code)
	}
	if !strings.Contains(strings.ToLower(stderr), "invalid severity") {
		t.Fatalf("stderr should explain invalid severity, got %q", stderr)
	}
}
