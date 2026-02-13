package config

import (
	"testing"
)

func TestLoadFromBytes_MapRuleConfigSeverityAndOptions(t *testing.T) {
	data := []byte(`rules:
  CONV-file-header:
    severity: warn
    pattern: "// {filename} — {purpose}"
`)

	cfg, err := LoadFromBytes(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	rule := cfg.Rules["CONV-file-header"]
	if rule.Severity != "warn" {
		t.Fatalf("severity = %q, want warn", rule.Severity)
	}
	if got, ok := rule.Options["pattern"].(string); !ok || got == "" {
		t.Fatalf("expected pattern option, got %#v", rule.Options["pattern"])
	}
}

func TestLoadFromBytes_MapRuleConfigOptionsObject(t *testing.T) {
	data := []byte(`rules:
  ARCH-max-file-lines:
    severity: error
    options:
      max: 300
      labels:
        env: prod
`)

	cfg, err := LoadFromBytes(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	rule := cfg.Rules["ARCH-max-file-lines"]
	if rule.Severity != "error" {
		t.Fatalf("severity = %q, want error", rule.Severity)
	}
	if _, ok := rule.Options["max"]; !ok {
		t.Fatalf("expected max option")
	}
	labels, ok := rule.Options["labels"].(map[string]interface{})
	if !ok {
		t.Fatalf("labels option type = %T, want map[string]interface{}", rule.Options["labels"])
	}
	if labels["env"] != "prod" {
		t.Fatalf("labels.env = %v, want prod", labels["env"])
	}
}

func TestLoadFromBytes_RejectsMapRuleSeverityNotString(t *testing.T) {
	data := []byte(`rules:
  CONV-file-naming:
    severity: 123
`)

	if _, err := LoadFromBytes(data); err == nil {
		t.Fatalf("expected error for non-string severity")
	}
}

func TestLoadFromBytes_RejectsArrayRuleWithInvalidShape(t *testing.T) {
	data := []byte(`rules:
  CONV-file-naming: []
`)
	if _, err := LoadFromBytes(data); err == nil {
		t.Fatalf("expected error for empty array config")
	}

	data = []byte(`rules:
  CONV-file-naming: [123]
`)
	if _, err := LoadFromBytes(data); err == nil {
		t.Fatalf("expected error for non-string severity in array config")
	}
}

func TestNormalizeToStringMapAndNormalizeValue(t *testing.T) {
	if got := normalizeToStringMap("not-a-map"); len(got) != 0 {
		t.Fatalf("normalizeToStringMap(non-map) len = %d, want 0", len(got))
	}

	input := map[interface{}]interface{}{
		"a": 1,
		"nested": map[interface{}]interface{}{
			"b": true,
			1:   "ignored-non-string-key",
		},
	}

	normalized := normalizeValue(input)
	root, ok := normalized.(map[string]interface{})
	if !ok {
		t.Fatalf("normalized root type = %T, want map[string]interface{}", normalized)
	}

	nested, ok := root["nested"].(map[string]interface{})
	if !ok {
		t.Fatalf("nested type = %T, want map[string]interface{}", root["nested"])
	}
	if _, exists := nested["1"]; exists {
		t.Fatalf("unexpected normalized non-string key in nested map")
	}
	if nested["b"] != true {
		t.Fatalf("nested.b = %v, want true", nested["b"])
	}
}

func TestUnknownRuleIDsNilInputs(t *testing.T) {
	if got := UnknownRuleIDs(nil, nil); got != nil {
		t.Fatalf("UnknownRuleIDs(nil,nil) = %v, want nil", got)
	}
}

func TestLoadFromBytes_ParsesPluginsList(t *testing.T) {
	data := []byte(`plugins:
  - ./plugins/custom.yml
`)
	cfg, err := LoadFromBytes(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cfg.Plugins) != 1 || cfg.Plugins[0] != "./plugins/custom.yml" {
		t.Fatalf("plugins = %v, want [./plugins/custom.yml]", cfg.Plugins)
	}
}
