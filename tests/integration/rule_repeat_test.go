// rule_repeat_test.go — Integration checks for repeatable --rule filtering.
//go:build integration

package integration

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestRuleFlagCanBeRepeated(t *testing.T) {
	tmp := t.TempDir()
	file := filepath.Join(tmp, "BadName.go")
	if err := os.WriteFile(file, []byte("package main\n\nfunc main() {}\n"), 0o644); err != nil {
		t.Fatalf("write source: %v", err)
	}

	stdout, stderr, code := runInDir(t, tmp,
		"--format", "json",
		"--rule", "CONV-file-header",
		"--rule", "CONV-file-naming",
		".",
	)
	if code != 1 {
		t.Fatalf("repeat --rule run should fail with violations: code=%d stderr=%q stdout=%q", code, stderr, stdout)
	}

	var payload struct {
		Violations []struct {
			RuleID string `json:"ruleId"`
		} `json:"violations"`
		Summary struct {
			TotalViolations int `json:"totalViolations"`
		} `json:"summary"`
	}
	if err := json.Unmarshal([]byte(stdout), &payload); err != nil {
		t.Fatalf("unmarshal output: %v\noutput=%q", err, stdout)
	}
	if payload.Summary.TotalViolations != 2 || len(payload.Violations) != 2 {
		t.Fatalf("expected two violations from repeated --rule, got total=%d len=%d", payload.Summary.TotalViolations, len(payload.Violations))
	}

	ruleSeen := map[string]bool{}
	for _, v := range payload.Violations {
		ruleSeen[v.RuleID] = true
	}
	if !ruleSeen["CONV-file-header"] || !ruleSeen["CONV-file-naming"] {
		t.Fatalf("expected both rule IDs, got %+v", ruleSeen)
	}
}
