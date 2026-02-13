package plugins

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stricture/stricture/internal/model"
	plugapi "github.com/stricture/stricture/pkg/rule"
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
	if rules[0].Category() != "custom" {
		t.Fatalf("category = %q, want custom", rules[0].Category())
	}
	if rules[0].Description() != "Disallow fmt.Print" {
		t.Fatalf("description = %q", rules[0].Description())
	}
	if rules[0].DefaultSeverity() != "warn" {
		t.Fatalf("severity = %q, want warn", rules[0].DefaultSeverity())
	}
	if rules[0].NeedsProjectContext() {
		t.Fatalf("yaml plugin rule should not require project context")
	}
	if rules[0].Why() != "Custom policy from plugin configuration." {
		t.Fatalf("why = %q", rules[0].Why())
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
	if violations[0].Severity != "warn" {
		t.Fatalf("severity = %q, want warn", violations[0].Severity)
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

func TestLoadRejectsUnsupportedPluginType(t *testing.T) {
	_, err := Load([]string{"custom.json"})
	if err == nil {
		t.Fatalf("expected unsupported type error")
	}
}

func TestLoadSkipsBlankEntriesAndSortsRuleIDs(t *testing.T) {
	tmp := t.TempDir()
	b := filepath.Join(tmp, "b.yml")
	a := filepath.Join(tmp, "a.yml")
	if err := os.WriteFile(b, []byte(`id: CUSTOM-b
check:
  must_not_contain:
    pattern: "bad"
`), 0o644); err != nil {
		t.Fatalf("write b.yml: %v", err)
	}
	if err := os.WriteFile(a, []byte(`id: CUSTOM-a
check:
  must_not_contain:
    pattern: "bad"
`), 0o644); err != nil {
		t.Fatalf("write a.yml: %v", err)
	}

	rules, err := Load([]string{"  ", b, "", a})
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if len(rules) != 2 {
		t.Fatalf("rules len = %d, want 2", len(rules))
	}
	if rules[0].ID() != "CUSTOM-a" || rules[1].ID() != "CUSTOM-b" {
		t.Fatalf("rules are not sorted by id: [%s, %s]", rules[0].ID(), rules[1].ID())
	}
}

func TestLoadYAMLRulesRequiresAtLeastOneRule(t *testing.T) {
	tmp := t.TempDir()
	pluginPath := filepath.Join(tmp, "empty.yml")
	if err := os.WriteFile(pluginPath, []byte("rules: []\n"), 0o644); err != nil {
		t.Fatalf("write plugin: %v", err)
	}

	_, err := loadYAMLRules(pluginPath)
	if err == nil {
		t.Fatalf("expected no-rules error")
	}
}

func TestNewYAMLRuleValidation(t *testing.T) {
	tests := []struct {
		name string
		rule yamlRule
	}{
		{
			name: "missing id",
			rule: yamlRule{
				Check: yamlCheckSpec{MustNotContain: yamlMustNotContain{Pattern: "bad"}},
			},
		},
		{
			name: "invalid severity",
			rule: yamlRule{
				ID:       "CUSTOM-x",
				Severity: "fatal",
				Check:    yamlCheckSpec{MustNotContain: yamlMustNotContain{Pattern: "bad"}},
			},
		},
		{
			name: "missing pattern",
			rule: yamlRule{
				ID:    "CUSTOM-x",
				Check: yamlCheckSpec{},
			},
		},
		{
			name: "invalid regex",
			rule: yamlRule{
				ID:    "CUSTOM-x",
				Check: yamlCheckSpec{MustNotContain: yamlMustNotContain{Pattern: "["}},
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			_, err := newYAMLRule(tc.rule)
			if err == nil {
				t.Fatalf("expected validation error")
			}
		})
	}
}

func TestYAMLRuleDefaultsAndCheckFilters(t *testing.T) {
	rule, err := newYAMLRule(yamlRule{
		ID: "CUSTOM-filter",
		Match: yamlMatch{
			Languages:    []string{"go"},
			PathPatterns: []string{"**/*.go"},
			ExcludePaths: []string{"vendor/**"},
		},
		Check: yamlCheckSpec{
			MustNotContain: yamlMustNotContain{
				Pattern: "fmt\\.Print",
			},
		},
	})
	if err != nil {
		t.Fatalf("newYAMLRule returned error: %v", err)
	}

	if rule.Category() != "custom" {
		t.Fatalf("category = %q, want custom", rule.Category())
	}
	if rule.DefaultSeverity() != "error" {
		t.Fatalf("severity = %q, want error", rule.DefaultSeverity())
	}
	if rule.Description() != "Custom YAML rule" {
		t.Fatalf("description = %q", rule.Description())
	}
	if rule.Why() != "Custom policy from plugin configuration." {
		t.Fatalf("why = %q", rule.Why())
	}
	if rule.NeedsProjectContext() {
		t.Fatalf("yamlLoadedRule should not require project context")
	}

	if got := rule.Check(nil, nil, model.RuleConfig{}); got != nil {
		t.Fatalf("expected nil violations for nil file")
	}
	if got := rule.Check(&model.UnifiedFileModel{
		Path:     "pkg/main.go",
		Language: "typescript",
		Source:   []byte("fmt.Print(\"x\")"),
	}, nil, model.RuleConfig{}); got != nil {
		t.Fatalf("language mismatch should not match")
	}
	if got := rule.Check(&model.UnifiedFileModel{
		Path:     "main.go",
		Language: "go",
		Source:   []byte("fmt.Print(\"x\")"),
	}, nil, model.RuleConfig{}); got != nil {
		t.Fatalf("include path mismatch should not match")
	}
	if got := rule.Check(&model.UnifiedFileModel{
		Path:     "vendor/pkg/main.go",
		Language: "go",
		Source:   []byte("fmt.Print(\"x\")"),
	}, nil, model.RuleConfig{}); got != nil {
		t.Fatalf("exclude path should not match")
	}
	if got := rule.Check(&model.UnifiedFileModel{
		Path:     "pkg/main.go",
		Language: "go",
		Source:   []byte("log.Println(\"x\")"),
	}, nil, model.RuleConfig{}); got != nil {
		t.Fatalf("source without pattern should not match")
	}

	match := rule.Check(&model.UnifiedFileModel{
		Path:     "pkg/main.go",
		Language: "go",
		Source:   []byte("package main\nfunc x(){fmt.Print(\"x\")}"),
	}, nil, model.RuleConfig{})
	if len(match) != 1 {
		t.Fatalf("violations len = %d, want 1", len(match))
	}
	if !strings.Contains(match[0].Message, "forbidden pattern") {
		t.Fatalf("expected default message fallback, got %q", match[0].Message)
	}
	if match[0].Severity != "error" {
		t.Fatalf("default severity = %q, want error", match[0].Severity)
	}

	override := rule.Check(&model.UnifiedFileModel{
		Path:     "pkg/main.go",
		Language: "go",
		Source:   []byte("fmt.Print(\"x\")"),
	}, nil, model.RuleConfig{Severity: "warn"})
	if len(override) != 1 || override[0].Severity != "warn" {
		t.Fatalf("config severity override not applied: %#v", override)
	}
}

func TestGoPluginRuleMetadataAndCheck(t *testing.T) {
	capturedFileCount := -1
	rule := &goPluginRule{
		definition: &plugapi.Definition{
			ID:                  "  CUSTOM-go  ",
			Category:            " platform ",
			Severity:            " WARN ",
			Description:         " checks plugin behavior ",
			Why:                 " keeps policy centralized ",
			NeedsProjectContext: true,
			Check: func(file *plugapi.UnifiedFileModel, context *plugapi.ProjectContext, options map[string]interface{}) []plugapi.Violation {
				if file.Path != "pkg/main.go" {
					t.Fatalf("unexpected file path: %q", file.Path)
				}
				if context == nil {
					t.Fatalf("expected context")
				}
				capturedFileCount = context.FileCount
				if options != nil && options["flag"] != "on" {
					t.Fatalf("unexpected options: %#v", options)
				}
				return []plugapi.Violation{
					{Message: "primary", StartLine: 2, SuggestedFix: "replace me"},
					{RuleID: "CUSTOM-alt", Severity: "off", Message: "secondary", StartLine: 3, EndLine: 4},
				}
			},
		},
	}

	if rule.ID() != "CUSTOM-go" {
		t.Fatalf("id = %q, want CUSTOM-go", rule.ID())
	}
	if rule.Category() != "platform" {
		t.Fatalf("category = %q, want platform", rule.Category())
	}
	if rule.Description() != "checks plugin behavior" {
		t.Fatalf("description = %q", rule.Description())
	}
	if rule.DefaultSeverity() != "warn" {
		t.Fatalf("severity = %q, want warn", rule.DefaultSeverity())
	}
	if !rule.NeedsProjectContext() {
		t.Fatalf("expected project context requirement")
	}
	if rule.Why() != "keeps policy centralized" {
		t.Fatalf("why = %q", rule.Why())
	}

	file := &model.UnifiedFileModel{
		Path:      "pkg/main.go",
		Language:  "go",
		Source:    []byte("package main"),
		LineCount: 1,
	}
	ctx := &model.ProjectContext{
		Files: map[string]*model.UnifiedFileModel{
			"a.go": {},
			"b.go": {},
		},
	}
	violations := rule.Check(file, ctx, model.RuleConfig{
		Options: map[string]interface{}{"flag": "on"},
	})
	if capturedFileCount != 2 {
		t.Fatalf("captured file count = %d, want 2", capturedFileCount)
	}
	if len(violations) != 2 {
		t.Fatalf("violations len = %d, want 2", len(violations))
	}
	if violations[0].RuleID != "CUSTOM-go" {
		t.Fatalf("rule id fallback = %q, want CUSTOM-go", violations[0].RuleID)
	}
	if violations[0].Severity != "warn" {
		t.Fatalf("severity fallback = %q, want warn", violations[0].Severity)
	}
	if violations[0].Context == nil || violations[0].Context.SuggestedFix != "replace me" {
		t.Fatalf("missing suggested fix context: %#v", violations[0].Context)
	}
	if violations[1].RuleID != "CUSTOM-alt" {
		t.Fatalf("explicit rule id not preserved: %q", violations[1].RuleID)
	}
	if violations[1].Severity != "off" {
		t.Fatalf("explicit severity not preserved: %q", violations[1].Severity)
	}

	override := rule.Check(file, nil, model.RuleConfig{Severity: "error"})
	if len(override) == 0 || override[0].Severity != "error" || override[1].Severity != "error" {
		t.Fatalf("config severity override not applied: %#v", override)
	}
}

func TestGoPluginRuleFallbacksAndNilCheck(t *testing.T) {
	rule := &goPluginRule{
		definition: &plugapi.Definition{
			ID:       "  CUSTOM-empty ",
			Severity: "fatal",
		},
	}
	if rule.Category() != "custom" {
		t.Fatalf("category = %q, want custom", rule.Category())
	}
	if rule.Description() != "Custom Go plugin rule" {
		t.Fatalf("description = %q", rule.Description())
	}
	if rule.DefaultSeverity() != "error" {
		t.Fatalf("severity = %q, want error", rule.DefaultSeverity())
	}
	if rule.NeedsProjectContext() {
		t.Fatalf("needs project context should default to false")
	}
	if rule.Why() != "Custom policy from Go plugin." {
		t.Fatalf("why = %q", rule.Why())
	}
	if got := rule.Check(nil, nil, model.RuleConfig{}); got != nil {
		t.Fatalf("nil check function should return nil violations")
	}
}

func TestLoadGoPluginPathErrors(t *testing.T) {
	pathValue := filepath.Join(t.TempDir(), "missing.so")
	if _, err := loadGoPluginRules(pathValue); err == nil {
		t.Fatalf("expected loadGoPluginRules error for missing plugin")
	}
	if _, err := Load([]string{pathValue}); err == nil {
		t.Fatalf("expected Load error for missing .so plugin")
	}
}
