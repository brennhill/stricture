// no_config_test.go — Integration checks for --no-config behavior.
//go:build integration

package integration

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestNoConfigIgnoresProjectRuleOverrides(t *testing.T) {
	tmp := t.TempDir()
	source := filepath.Join(tmp, "bad.go")
	if err := os.WriteFile(source, []byte("package main\n\nfunc main() {}\n"), 0o644); err != nil {
		t.Fatalf("write source: %v", err)
	}

	cfgPath := filepath.Join(tmp, ".stricture.yml")
	cfg := `rules:
  CONV-file-header: off
`
	if err := os.WriteFile(cfgPath, []byte(cfg), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	stdout, stderr, code := runInDir(t, tmp,
		"--format", "json",
		"--rule", "CONV-file-header",
		".",
	)
	if code != 0 {
		t.Fatalf("config override run should pass: code=%d stderr=%q stdout=%q", code, stderr, stdout)
	}
	var withConfig struct {
		Summary struct {
			TotalViolations int `json:"totalViolations"`
		} `json:"summary"`
	}
	if err := json.Unmarshal([]byte(stdout), &withConfig); err != nil {
		t.Fatalf("unmarshal with-config output: %v\noutput=%q", err, stdout)
	}
	if withConfig.Summary.TotalViolations != 0 {
		t.Fatalf("expected 0 violations with config override, got %d", withConfig.Summary.TotalViolations)
	}

	stdout, stderr, code = runInDir(t, tmp,
		"--format", "json",
		"--rule", "CONV-file-header",
		"--no-config",
		".",
	)
	if code != 1 {
		t.Fatalf("--no-config should re-enable default rule behavior: code=%d stderr=%q stdout=%q", code, stderr, stdout)
	}
	var noConfig struct {
		Summary struct {
			TotalViolations int `json:"totalViolations"`
			Errors          int `json:"errors"`
		} `json:"summary"`
	}
	if err := json.Unmarshal([]byte(stdout), &noConfig); err != nil {
		t.Fatalf("unmarshal --no-config output: %v\noutput=%q", err, stdout)
	}
	if noConfig.Summary.TotalViolations != 1 || noConfig.Summary.Errors != 1 {
		t.Fatalf("expected one error violation with --no-config, got total=%d errors=%d", noConfig.Summary.TotalViolations, noConfig.Summary.Errors)
	}
}
