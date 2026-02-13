// load.go - Config loading and normalization.
package config

import (
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/stricture/stricture/internal/model"
	"gopkg.in/yaml.v3"
)

// Config is the normalized representation of .stricture.yml.
type Config struct {
	Version string
	Rules   map[string]model.RuleConfig
}

// Default returns an empty configuration with default schema version.
func Default() *Config {
	return &Config{
		Version: "1.0",
		Rules:   map[string]model.RuleConfig{},
	}
}

// Load reads and parses configuration from disk.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, model.ErrConfigNotFound
		}
		return nil, fmt.Errorf("%w: %v", model.ErrConfigInvalid, err)
	}
	return LoadFromBytes(data)
}

// LoadFromBytes parses configuration from YAML bytes.
func LoadFromBytes(data []byte) (*Config, error) {
	if strings.TrimSpace(string(data)) == "" {
		return Default(), nil
	}

	var raw struct {
		Version string                 `yaml:"version"`
		Rules   map[string]interface{} `yaml:"rules"`
	}
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("%w: %v", model.ErrConfigInvalid, err)
	}

	cfg := Default()
	if strings.TrimSpace(raw.Version) != "" {
		cfg.Version = strings.TrimSpace(raw.Version)
	}

	for ruleID, value := range raw.Rules {
		ruleCfg, err := parseRuleConfig(value)
		if err != nil {
			return nil, fmt.Errorf("%w: rule %s: %v", model.ErrConfigInvalid, ruleID, err)
		}
		cfg.Rules[ruleID] = ruleCfg
	}

	return cfg, nil
}

func parseRuleConfig(raw interface{}) (model.RuleConfig, error) {
	switch value := raw.(type) {
	case string:
		severity, err := normalizeSeverity(value)
		if err != nil {
			return model.RuleConfig{}, err
		}
		return model.RuleConfig{Severity: severity, Options: map[string]interface{}{}}, nil
	case []interface{}:
		if len(value) == 0 {
			return model.RuleConfig{}, fmt.Errorf("array config must include severity as first item")
		}
		severityRaw, ok := value[0].(string)
		if !ok {
			return model.RuleConfig{}, fmt.Errorf("first array item must be severity string")
		}
		severity, err := normalizeSeverity(severityRaw)
		if err != nil {
			return model.RuleConfig{}, err
		}

		ruleCfg := model.RuleConfig{Severity: severity, Options: map[string]interface{}{}}
		if len(value) > 1 {
			ruleCfg.Options = normalizeToStringMap(value[1])
		}
		return ruleCfg, nil
	case map[string]interface{}:
		return parseMapRuleConfig(value)
	case map[interface{}]interface{}:
		return parseMapRuleConfig(normalizeToStringMap(value))
	default:
		return model.RuleConfig{}, fmt.Errorf("unsupported rule config type %T", raw)
	}
}

func parseMapRuleConfig(value map[string]interface{}) (model.RuleConfig, error) {
	options := map[string]interface{}{}
	for k, v := range value {
		options[k] = normalizeValue(v)
	}

	ruleCfg := model.RuleConfig{Options: options}
	if rawSeverity, ok := value["severity"]; ok {
		severity, ok := rawSeverity.(string)
		if !ok {
			return model.RuleConfig{}, fmt.Errorf("severity must be a string")
		}
		normalized, err := normalizeSeverity(severity)
		if err != nil {
			return model.RuleConfig{}, err
		}
		ruleCfg.Severity = normalized
		delete(ruleCfg.Options, "severity")
	}

	if rawOptions, ok := value["options"]; ok {
		ruleCfg.Options = normalizeToStringMap(rawOptions)
	}

	if ruleCfg.Options == nil {
		ruleCfg.Options = map[string]interface{}{}
	}
	return ruleCfg, nil
}

func normalizeSeverity(raw string) (string, error) {
	severity := strings.ToLower(strings.TrimSpace(raw))
	switch severity {
	case "error", "warn", "off":
		return severity, nil
	default:
		return "", fmt.Errorf("invalid severity %q (valid: error|warn|off)", raw)
	}
}

func normalizeToStringMap(raw interface{}) map[string]interface{} {
	converted := normalizeValue(raw)
	if m, ok := converted.(map[string]interface{}); ok {
		return m
	}
	return map[string]interface{}{}
}

func normalizeValue(raw interface{}) interface{} {
	switch v := raw.(type) {
	case map[string]interface{}:
		out := make(map[string]interface{}, len(v))
		for k, value := range v {
			out[k] = normalizeValue(value)
		}
		return out
	case map[interface{}]interface{}:
		out := make(map[string]interface{}, len(v))
		for k, value := range v {
			ks, ok := k.(string)
			if !ok {
				continue
			}
			out[ks] = normalizeValue(value)
		}
		return out
	case []interface{}:
		out := make([]interface{}, len(v))
		for i := range v {
			out[i] = normalizeValue(v[i])
		}
		return out
	default:
		return raw
	}
}

// UnknownRuleIDs returns config rule IDs that are not registered.
func UnknownRuleIDs(cfg *Config, registry *model.RuleRegistry) []string {
	if cfg == nil || registry == nil {
		return nil
	}
	unknown := make([]string, 0)
	for ruleID := range cfg.Rules {
		if _, ok := registry.ByID(ruleID); !ok {
			unknown = append(unknown, ruleID)
		}
	}
	sort.Strings(unknown)
	return unknown
}
