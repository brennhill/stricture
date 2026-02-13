// load_test.go - Tests for config loader.
package config

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/stricture/stricture/internal/model"
)

func TestLoadFromBytes_EmptyConfig(t *testing.T) {
	cfg, err := LoadFromBytes([]byte(""))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Version != "1.0" {
		t.Fatalf("version = %q, want 1.0", cfg.Version)
	}
	if len(cfg.Rules) != 0 {
		t.Fatalf("rules len = %d, want 0", len(cfg.Rules))
	}
}

func TestLoadFromBytes_ParsesStringAndTupleRules(t *testing.T) {
	data := []byte(`version: "1.0"
rules:
  CONV-file-naming: error
  CONV-file-header: [warn, { pattern: "// {filename} — {purpose}" }]
`)

	cfg, err := LoadFromBytes(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	naming, ok := cfg.Rules["CONV-file-naming"]
	if !ok {
		t.Fatalf("missing CONV-file-naming")
	}
	if naming.Severity != "error" {
		t.Fatalf("severity = %q, want error", naming.Severity)
	}

	header, ok := cfg.Rules["CONV-file-header"]
	if !ok {
		t.Fatalf("missing CONV-file-header")
	}
	if header.Severity != "warn" {
		t.Fatalf("severity = %q, want warn", header.Severity)
	}
	if _, ok := header.Options["pattern"]; !ok {
		t.Fatalf("expected pattern option")
	}
}

func TestLoadFromBytes_RejectsInvalidSeverity(t *testing.T) {
	_, err := LoadFromBytes([]byte(`rules:
  CONV-file-naming: critical
`))
	if err == nil {
		t.Fatalf("expected error")
	}
	if !errors.Is(err, model.ErrConfigInvalid) {
		t.Fatalf("error must wrap ErrConfigInvalid, got %v", err)
	}
}

func TestLoad_NotFound(t *testing.T) {
	_, err := Load("/definitely/not/found/.stricture.yml")
	if !errors.Is(err, model.ErrConfigNotFound) {
		t.Fatalf("expected ErrConfigNotFound, got %v", err)
	}
}

func TestUnknownRuleIDs(t *testing.T) {
	cfg := &Config{Rules: map[string]model.RuleConfig{
		"CONV-file-naming": {Severity: "error"},
		"FAKE-rule":        {Severity: "error"},
	}}
	registry := model.NewRuleRegistry()
	registry.Register(&fakeRule{id: "CONV-file-naming"})

	unknown := UnknownRuleIDs(cfg, registry)
	if len(unknown) != 1 || unknown[0] != "FAKE-rule" {
		t.Fatalf("unknown = %v, want [FAKE-rule]", unknown)
	}
}

func TestLoad_ReadFileError(t *testing.T) {
	tmp := t.TempDir()
	dir := filepath.Join(tmp, "cfgdir")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	_, err := Load(dir)
	if err == nil {
		t.Fatalf("expected error")
	}
	if !errors.Is(err, model.ErrConfigInvalid) {
		t.Fatalf("error must wrap ErrConfigInvalid, got %v", err)
	}
}

type fakeRule struct{ id string }

func (f *fakeRule) ID() string                { return f.id }
func (f *fakeRule) Category() string          { return "conv" }
func (f *fakeRule) Description() string       { return "fake" }
func (f *fakeRule) DefaultSeverity() string   { return "error" }
func (f *fakeRule) NeedsProjectContext() bool { return false }
func (f *fakeRule) Why() string               { return "fake" }
func (f *fakeRule) Check(*model.UnifiedFileModel, *model.ProjectContext, model.RuleConfig) []model.Violation {
	return nil
}
