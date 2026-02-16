// explain_test.go — Integration checks for strict explain command.
//go:build integration

package integration

import (
	"strings"
	"testing"
)

func TestExplainKnownRule(t *testing.T) {
	stdout, stderr, code := run(t, "explain", "CONV-file-header")
	if code != 0 {
		t.Fatalf("explain exit code = %d, want 0\nstderr=%q", code, stderr)
	}

	mustContain := []string{
		"ID: CONV-file-header",
		"Category: CONV",
		"Default Severity: error",
		"Fixable: Yes",
		"Description:",
		"Why:",
	}
	for _, token := range mustContain {
		if !strings.Contains(stdout, token) {
			t.Fatalf("explain output missing %q\noutput=%q", token, stdout)
		}
	}
}

func TestExplainUnknownRuleExitsTwo(t *testing.T) {
	_, stderr, code := run(t, "explain", "NOT-A-RULE")
	if code != 2 {
		t.Fatalf("explain unknown rule exit code = %d, want 2", code)
	}
	if !strings.Contains(strings.ToLower(stderr), "unknown rule") {
		t.Fatalf("stderr should mention unknown rule, got %q", stderr)
	}
}

func TestExplainRequiresRuleID(t *testing.T) {
	_, stderr, code := run(t, "explain")
	if code != 2 {
		t.Fatalf("explain missing rule id exit code = %d, want 2", code)
	}
	if !strings.Contains(strings.ToLower(stderr), "requires a rule id") {
		t.Fatalf("stderr should mention required rule id, got %q", stderr)
	}
}
