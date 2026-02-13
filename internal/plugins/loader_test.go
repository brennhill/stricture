package plugins

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stricture/stricture/internal/model"
)

func TestLoadYAMLPluginAndRunRule(t *testing.T) {
	tmp := t.TempDir()
	pluginPath := filepath.Join(tmp, "custom.yml")
	content := `rules:
  - id: CUSTOM-no-fmt-print
    category: custom
    severity: warn
    description: "Disallow fmt.Print"
    match:
      languages: ["go"]
      paths: ["**/*.go"]
    check:
      must_not_contain:
        pattern: "fmt\\.Print"
        message: "Use structured logging instead of fmt.Print"
`
	if err := os.WriteFile(pluginPath, []byte(content), 0o644); err != nil {
		t.Fatalf("write plugin: %v", err)
	}

	rules, err := Load([]string{pluginPath})
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if len(rules) != 1 {
		t.Fatalf("rules len = %d, want 1", len(rules))
	}

	file := &model.UnifiedFileModel{
		Path:     "internal/app/main.go",
		Language: "go",
		Source:   []byte("package main\nimport \"fmt\"\nfunc main(){fmt.Print(\"x\")}\n"),
	}
	violations := rules[0].Check(file, nil, model.RuleConfig{})
	if len(violations) != 1 {
		t.Fatalf("violations len = %d, want 1", len(violations))
	}
	if violations[0].RuleID != "CUSTOM-no-fmt-print" {
		t.Fatalf("rule id = %q", violations[0].RuleID)
	}
}

func TestLoadRejectsDuplicateRuleIDs(t *testing.T) {
	tmp := t.TempDir()
	a := filepath.Join(tmp, "a.yml")
	b := filepath.Join(tmp, "b.yml")
	content := `id: CUSTOM-dup
category: custom
severity: error
check:
  must_not_contain:
    pattern: "bad"
    message: "bad"
`
	if err := os.WriteFile(a, []byte(content), 0o644); err != nil {
		t.Fatalf("write a.yml: %v", err)
	}
	if err := os.WriteFile(b, []byte(content), 0o644); err != nil {
		t.Fatalf("write b.yml: %v", err)
	}

	if _, err := Load([]string{a, b}); err == nil {
		t.Fatalf("expected duplicate ID error")
	}
}
