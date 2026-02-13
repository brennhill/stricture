// output_test.go — Integration checks for --output report file behavior.
//go:build integration

package integration

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestOutputWritesJSONReportFile(t *testing.T) {
	tmp := t.TempDir()
	target := filepath.Join(tmp, "a.ts")
	if err := os.WriteFile(target, []byte("export const a = 1;\n"), 0o644); err != nil {
		t.Fatalf("write target: %v", err)
	}

	report := filepath.Join(tmp, "reports", "stricture.json")
	stdout, stderr, code := runInDir(t, tmp, "--format", "json", "--rule", "CONV-file-header", "--output", report, ".")
	if code != 1 {
		t.Fatalf("expected violations to return exit 1, got %d\nstderr=%q\nstdout=%q", code, stderr, stdout)
	}
	if strings.TrimSpace(stdout) != "" {
		t.Fatalf("stdout should be empty when --output is used, got %q", stdout)
	}

	data, err := os.ReadFile(report)
	if err != nil {
		t.Fatalf("read report file: %v", err)
	}

	var payload struct {
		Violations []struct {
			RuleID string `json:"ruleId"`
		} `json:"violations"`
		Summary map[string]interface{} `json:"summary"`
	}
	if err := json.Unmarshal(data, &payload); err != nil {
		t.Fatalf("report file must contain valid JSON: %v", err)
	}
	if len(payload.Violations) != 1 || payload.Violations[0].RuleID != "CONV-file-header" {
		t.Fatalf("unexpected violations payload: %+v", payload.Violations)
	}
	if payload.Summary == nil {
		t.Fatalf("missing summary payload")
	}
}

func TestOutputWritesTextReportFile(t *testing.T) {
	tmp := t.TempDir()
	target := filepath.Join(tmp, "a.ts")
	if err := os.WriteFile(target, []byte("export const a = 1;\n"), 0o644); err != nil {
		t.Fatalf("write target: %v", err)
	}

	report := filepath.Join(tmp, "nested", "stricture.txt")
	stdout, stderr, code := runInDir(t, tmp, "--rule", "CONV-file-header", "--output", report, ".")
	if code != 1 {
		t.Fatalf("expected violations to return exit 1, got %d\nstderr=%q\nstdout=%q", code, stderr, stdout)
	}
	if strings.TrimSpace(stdout) != "" {
		t.Fatalf("stdout should be empty when --output is used, got %q", stdout)
	}

	data, err := os.ReadFile(report)
	if err != nil {
		t.Fatalf("read report file: %v", err)
	}
	text := string(data)
	if !strings.Contains(text, "CONV-file-header") || !strings.Contains(text, "Summary:") {
		t.Fatalf("text report missing expected content: %q", text)
	}
}
