// rule_test_helpers_test.go — Shared test contracts for CTR rules.
package ctr

import (
	"testing"

	"github.com/stricture/stricture/internal/model"
)

func assertRuleContract(t *testing.T, rule model.Rule) {
	t.Helper()

	if rule.Category() != "ctr" {
		t.Fatalf("category = %q, want ctr", rule.Category())
	}
	if rule.DefaultSeverity() != "error" {
		t.Fatalf("default severity = %q, want error", rule.DefaultSeverity())
	}
	if rule.ID() == "" {
		t.Fatal("rule ID must not be empty")
	}
	if rule.Description() == "" {
		t.Fatal("description must not be empty")
	}
	if rule.Why() == "" {
		t.Fatal("why must not be empty")
	}

	file := &model.UnifiedFileModel{
		Path:     "contracts/user.ts",
		Language: "typescript",
		Source:   []byte("// stricture-trigger " + rule.ID() + "\nexport type User = {}\n"),
	}

	violations := rule.Check(file, nil, model.RuleConfig{})
	if len(violations) != 1 {
		t.Fatalf("violations = %d, want 1", len(violations))
	}
	v := violations[0]
	if v.RuleID != rule.ID() {
		t.Fatalf("violation rule ID = %q, want %q", v.RuleID, rule.ID())
	}
	if v.Severity != "error" {
		t.Fatalf("severity = %q, want error", v.Severity)
	}
	if v.Message == "" {
		t.Fatal("message must not be empty")
	}
	if v.FilePath != file.Path {
		t.Fatalf("file path = %q, want %q", v.FilePath, file.Path)
	}
	if v.StartLine != 1 {
		t.Fatalf("start line = %d, want 1", v.StartLine)
	}

	warned := rule.Check(file, nil, model.RuleConfig{Severity: "warn"})
	if len(warned) != 1 || warned[0].Severity != "warn" {
		t.Fatalf("severity override failed: got %+v", warned)
	}

	clean := &model.UnifiedFileModel{
		Path:     "contracts/user.ts",
		Language: "typescript",
		Source:   []byte("export type User = {}\n"),
	}
	if got := rule.Check(clean, nil, model.RuleConfig{}); len(got) != 0 {
		t.Fatalf("clean source produced %d violations, want 0", len(got))
	}
}
