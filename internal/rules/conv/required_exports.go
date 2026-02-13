// required_exports.go - CONV-required-exports: Require configured exports for matching modules.
package conv

import (
	"fmt"
	"path"
	"sort"
	"strings"

	"github.com/stricture/stricture/internal/model"
)

// RequiredExports enforces required exported symbols per configured file patterns.
type RequiredExports struct{}

func (r *RequiredExports) ID() string       { return "CONV-required-exports" }
func (r *RequiredExports) Category() string { return "conv" }
func (r *RequiredExports) Description() string {
	return "Require configured exports for matching modules"
}
func (r *RequiredExports) DefaultSeverity() string   { return "error" }
func (r *RequiredExports) NeedsProjectContext() bool { return false }
func (r *RequiredExports) Why() string {
	return "Missing required exports break module contracts and cause integration failures."
}

func (r *RequiredExports) Check(file *model.UnifiedFileModel, _ *model.ProjectContext, config model.RuleConfig) []model.Violation {
	if file == nil {
		return nil
	}

	patternRules := resolveRequiredExportPatterns(config.Options)
	if len(patternRules) == 0 {
		return nil
	}

	pathValue := filepathToSlash(file.Path)
	matched := make([]requiredExportPattern, 0)
	for _, p := range patternRules {
		ok, err := path.Match(p.Pattern, pathValue)
		if err != nil {
			continue
		}
		if ok {
			matched = append(matched, p)
		}
	}
	if len(matched) == 0 {
		return nil
	}

	severity := config.Severity
	if severity == "" {
		severity = r.DefaultSeverity()
	}

	symbols := scanExportedSymbols(file)
	exports := make([]string, 0, len(symbols))
	seen := map[string]bool{}
	for _, symbol := range symbols {
		name := strings.TrimSpace(symbol.Name)
		if name == "" || seen[name] {
			continue
		}
		seen[name] = true
		exports = append(exports, name)
	}

	violations := make([]model.Violation, 0)
	for _, rule := range matched {
		for _, required := range rule.Required {
			if requiredExportSatisfied(exports, required) {
				continue
			}
			violations = append(violations, model.Violation{
				RuleID:    r.ID(),
				Severity:  severity,
				Message:   fmt.Sprintf("Module '%s' missing required export '%s', expected in configured modules", path.Base(pathValue), required),
				FilePath:  file.Path,
				StartLine: 1,
				Context: &model.ViolationContext{
					SuggestedFix: fmt.Sprintf("Add export for '%s' in %s.", required, path.Base(pathValue)),
				},
			})
		}
	}

	return violations
}

type requiredExportPattern struct {
	Pattern  string
	Required []string
}

func resolveRequiredExportPatterns(options map[string]interface{}) []requiredExportPattern {
	if options == nil {
		return nil
	}
	raw, ok := options["patterns"]
	if !ok {
		return nil
	}

	patternsMap, ok := toStringMap(raw)
	if !ok {
		return nil
	}

	resolved := make([]requiredExportPattern, 0, len(patternsMap))
	for pattern, val := range patternsMap {
		entry, ok := toStringMap(val)
		if !ok {
			continue
		}
		required := toStringSlice(entry["required"])
		if len(required) == 0 {
			continue
		}
		resolved = append(resolved, requiredExportPattern{Pattern: pattern, Required: required})
	}

	sort.Slice(resolved, func(i, j int) bool {
		return resolved[i].Pattern < resolved[j].Pattern
	})

	return resolved
}

func requiredExportSatisfied(exports []string, required string) bool {
	required = strings.TrimSpace(required)
	if required == "" {
		return true
	}

	if strings.Contains(required, "*") {
		for _, exp := range exports {
			matched, err := path.Match(required, exp)
			if err == nil && matched {
				return true
			}
		}
		return false
	}

	for _, exp := range exports {
		if exp == required {
			return true
		}
	}
	return false
}
