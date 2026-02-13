// loader.go — Custom plugin rule loaders (YAML and Go plugins).
package plugins

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/stricture/stricture/internal/model"
	plugapi "github.com/stricture/stricture/pkg/rule"
	"gopkg.in/yaml.v3"
)

// Load loads custom rules from plugin paths.
func Load(paths []string) ([]model.Rule, error) {
	loaded := make([]model.Rule, 0)
	seen := map[string]bool{}

	for _, raw := range paths {
		pathValue := strings.TrimSpace(raw)
		if pathValue == "" {
			continue
		}

		ext := strings.ToLower(filepath.Ext(pathValue))
		var rules []model.Rule
		var err error

		switch ext {
		case ".yml", ".yaml":
			rules, err = loadYAMLRules(pathValue)
		case ".so":
			rules, err = loadGoPluginRules(pathValue)
		default:
			err = fmt.Errorf("unsupported plugin type %q for %s", ext, pathValue)
		}
		if err != nil {
			return nil, err
		}

		for _, r := range rules {
			if seen[r.ID()] {
				return nil, fmt.Errorf("duplicate plugin rule ID %q", r.ID())
			}
			seen[r.ID()] = true
			loaded = append(loaded, r)
		}
	}

	sort.Slice(loaded, func(i, j int) bool { return loaded[i].ID() < loaded[j].ID() })
	return loaded, nil
}

type yamlPluginFile struct {
	Rules []yamlRule `yaml:"rules"`
}

type yamlRule struct {
	ID          string        `yaml:"id"`
	Category    string        `yaml:"category"`
	Severity    string        `yaml:"severity"`
	Description string        `yaml:"description"`
	Why         string        `yaml:"why"`
	Match       yamlMatch     `yaml:"match"`
	Check       yamlCheckSpec `yaml:"check"`
}

type yamlMatch struct {
	Languages    []string `yaml:"languages"`
	PathPatterns []string `yaml:"paths"`
	ExcludePaths []string `yaml:"exclude_paths"`
}

type yamlCheckSpec struct {
	MustNotContain yamlMustNotContain `yaml:"must_not_contain"`
}

type yamlMustNotContain struct {
	Pattern string `yaml:"pattern"`
	Message string `yaml:"message"`
}

func loadYAMLRules(pathValue string) ([]model.Rule, error) {
	data, err := os.ReadFile(pathValue)
	if err != nil {
		return nil, fmt.Errorf("read plugin file %s: %w", pathValue, err)
	}

	var doc yamlPluginFile
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return nil, fmt.Errorf("parse plugin yaml %s: %w", pathValue, err)
	}
	if len(doc.Rules) == 0 {
		var single yamlRule
		if err := yaml.Unmarshal(data, &single); err == nil && strings.TrimSpace(single.ID) != "" {
			doc.Rules = []yamlRule{single}
		}
	}
	if len(doc.Rules) == 0 {
		return nil, fmt.Errorf("plugin yaml %s has no rules", pathValue)
	}

	out := make([]model.Rule, 0, len(doc.Rules))
	for _, raw := range doc.Rules {
		r, err := newYAMLRule(raw)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", pathValue, err)
		}
		out = append(out, r)
	}
	return out, nil
}

type yamlLoadedRule struct {
	id                  string
	category            string
	severity            string
	description         string
	why                 string
	needsProjectContext bool
	languages           map[string]bool
	includePathRegex    []*regexp.Regexp
	excludePathRegex    []*regexp.Regexp
	pattern             *regexp.Regexp
	message             string
}

func newYAMLRule(raw yamlRule) (*yamlLoadedRule, error) {
	id := strings.TrimSpace(raw.ID)
	if id == "" {
		return nil, fmt.Errorf("rule id is required")
	}
	category := strings.TrimSpace(raw.Category)
	if category == "" {
		category = "custom"
	}
	severity := strings.ToLower(strings.TrimSpace(raw.Severity))
	if severity == "" {
		severity = "error"
	}
	switch severity {
	case "error", "warn", "off":
	default:
		return nil, fmt.Errorf("rule %s has invalid severity %q", id, raw.Severity)
	}

	patternRaw := strings.TrimSpace(raw.Check.MustNotContain.Pattern)
	if patternRaw == "" {
		return nil, fmt.Errorf("rule %s must define check.must_not_contain.pattern", id)
	}
	pattern, err := regexp.Compile(patternRaw)
	if err != nil {
		return nil, fmt.Errorf("rule %s pattern compile failed: %w", id, err)
	}

	languages := map[string]bool{}
	for _, lang := range raw.Match.Languages {
		n := strings.ToLower(strings.TrimSpace(lang))
		if n == "" {
			continue
		}
		languages[n] = true
	}

	include, err := compileGlobList(raw.Match.PathPatterns)
	if err != nil {
		return nil, fmt.Errorf("rule %s include path pattern: %w", id, err)
	}
	exclude, err := compileGlobList(raw.Match.ExcludePaths)
	if err != nil {
		return nil, fmt.Errorf("rule %s exclude path pattern: %w", id, err)
	}

	message := strings.TrimSpace(raw.Check.MustNotContain.Message)
	if message == "" {
		message = fmt.Sprintf("Rule %s matched forbidden pattern %q", id, patternRaw)
	}

	desc := strings.TrimSpace(raw.Description)
	if desc == "" {
		desc = "Custom YAML rule"
	}
	why := strings.TrimSpace(raw.Why)
	if why == "" {
		why = "Custom policy from plugin configuration."
	}

	return &yamlLoadedRule{
		id:               id,
		category:         category,
		severity:         severity,
		description:      desc,
		why:              why,
		languages:        languages,
		includePathRegex: include,
		excludePathRegex: exclude,
		pattern:          pattern,
		message:          message,
	}, nil
}

func (r *yamlLoadedRule) ID() string                { return r.id }
func (r *yamlLoadedRule) Category() string          { return r.category }
func (r *yamlLoadedRule) Description() string       { return r.description }
func (r *yamlLoadedRule) DefaultSeverity() string   { return r.severity }
func (r *yamlLoadedRule) NeedsProjectContext() bool { return r.needsProjectContext }
func (r *yamlLoadedRule) Why() string               { return r.why }

func (r *yamlLoadedRule) Check(file *model.UnifiedFileModel, _ *model.ProjectContext, config model.RuleConfig) []model.Violation {
	if file == nil {
		return nil
	}
	if len(r.languages) > 0 && !r.languages[strings.ToLower(strings.TrimSpace(file.Language))] {
		return nil
	}

	pathValue := filepath.ToSlash(file.Path)
	if !matchAnyPath(r.includePathRegex, pathValue, true) {
		return nil
	}
	if matchAnyPath(r.excludePathRegex, pathValue, false) {
		return nil
	}

	index := r.pattern.FindIndex(file.Source)
	if index == nil {
		return nil
	}
	line := 1 + strings.Count(string(file.Source[:index[0]]), "\n")

	severity := config.Severity
	if strings.TrimSpace(severity) == "" {
		severity = r.severity
	}

	return []model.Violation{
		{
			RuleID:    r.id,
			Severity:  severity,
			Message:   r.message,
			FilePath:  file.Path,
			StartLine: line,
		},
	}
}

func compileGlobList(patterns []string) ([]*regexp.Regexp, error) {
	out := make([]*regexp.Regexp, 0, len(patterns))
	for _, pattern := range patterns {
		p := strings.TrimSpace(pattern)
		if p == "" {
			continue
		}
		re, err := globToRegex(p)
		if err != nil {
			return nil, err
		}
		out = append(out, re)
	}
	return out, nil
}

func globToRegex(pattern string) (*regexp.Regexp, error) {
	replaced := regexp.QuoteMeta(filepath.ToSlash(pattern))
	replaced = strings.ReplaceAll(replaced, `\*\*`, `.*`)
	replaced = strings.ReplaceAll(replaced, `\*`, `[^/]*`)
	replaced = strings.ReplaceAll(replaced, `\?`, `.`)
	return regexp.Compile("^" + replaced + "$")
}

func matchAnyPath(patterns []*regexp.Regexp, pathValue string, defaultWhenEmpty bool) bool {
	if len(patterns) == 0 {
		return defaultWhenEmpty
	}
	for _, re := range patterns {
		if re.MatchString(pathValue) {
			return true
		}
	}
	return false
}

type goPluginRule struct {
	definition *plugapi.Definition
}

func (r *goPluginRule) ID() string {
	return strings.TrimSpace(r.definition.ID)
}

func (r *goPluginRule) Category() string {
	category := strings.TrimSpace(r.definition.Category)
	if category == "" {
		return "custom"
	}
	return category
}

func (r *goPluginRule) Description() string {
	desc := strings.TrimSpace(r.definition.Description)
	if desc == "" {
		return "Custom Go plugin rule"
	}
	return desc
}

func (r *goPluginRule) DefaultSeverity() string {
	severity := strings.ToLower(strings.TrimSpace(r.definition.Severity))
	switch severity {
	case "error", "warn", "off":
		return severity
	default:
		return "error"
	}
}

func (r *goPluginRule) NeedsProjectContext() bool {
	return r.definition.NeedsProjectContext
}

func (r *goPluginRule) Why() string {
	why := strings.TrimSpace(r.definition.Why)
	if why == "" {
		return "Custom policy from Go plugin."
	}
	return why
}

func (r *goPluginRule) Check(file *model.UnifiedFileModel, ctx *model.ProjectContext, config model.RuleConfig) []model.Violation {
	if r.definition.Check == nil {
		return nil
	}
	fileIn := &plugapi.UnifiedFileModel{
		Path:       file.Path,
		Language:   file.Language,
		IsTestFile: file.IsTestFile,
		Source:     string(file.Source),
		LineCount:  file.LineCount,
	}
	ctxIn := &plugapi.ProjectContext{}
	if ctx != nil {
		ctxIn.FileCount = len(ctx.Files)
	}

	out := r.definition.Check(fileIn, ctxIn, config.Options)
	converted := make([]model.Violation, 0, len(out))
	for _, v := range out {
		ruleID := strings.TrimSpace(v.RuleID)
		if ruleID == "" {
			ruleID = r.ID()
		}
		severity := strings.TrimSpace(v.Severity)
		if severity == "" {
			severity = r.DefaultSeverity()
		}
		if strings.TrimSpace(config.Severity) != "" {
			severity = config.Severity
		}
		converted = append(converted, model.Violation{
			RuleID:    ruleID,
			Severity:  severity,
			Message:   v.Message,
			FilePath:  file.Path,
			StartLine: v.StartLine,
			EndLine:   v.EndLine,
			Context: &model.ViolationContext{
				SuggestedFix: v.SuggestedFix,
			},
		})
	}
	return converted
}
