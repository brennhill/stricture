// export_naming.go - CONV-export-naming: Enforce naming conventions for exported symbols.
package conv

import (
	"fmt"
	"strings"

	"github.com/stricture/stricture/internal/model"
)

// ExportNaming enforces naming conventions for exported/public symbols.
type ExportNaming struct{}

func (r *ExportNaming) ID() string       { return "CONV-export-naming" }
func (r *ExportNaming) Category() string { return "conv" }
func (r *ExportNaming) Description() string {
	return "Enforce naming conventions for exported/public symbols"
}
func (r *ExportNaming) DefaultSeverity() string   { return "error" }
func (r *ExportNaming) NeedsProjectContext() bool { return false }
func (r *ExportNaming) Why() string {
	return "Inconsistent export names make imports confusing and break IDE autocomplete."
}

func (r *ExportNaming) Check(file *model.UnifiedFileModel, _ *model.ProjectContext, config model.RuleConfig) []model.Violation {
	if file == nil {
		return nil
	}

	severity := config.Severity
	if severity == "" {
		severity = r.DefaultSeverity()
	}

	conventions := resolveExportNamingConventions(file.Language, config.Options)
	symbols := scanExportedSymbols(file)
	violations := make([]model.Violation, 0)

	for _, symbol := range symbols {
		expected := expectedConventionForKind(symbol.Kind, conventions)
		if expected == "" {
			continue
		}
		if matchesNameConvention(symbol.Name, expected) {
			continue
		}

		suggested := suggestNameForConvention(symbol.Name, expected)
		violations = append(violations, model.Violation{
			RuleID:    r.ID(),
			Severity:  severity,
			Message:   fmt.Sprintf("Export '%s' does not follow convention '%s', should be '%s'", symbol.Name, expected, suggested),
			FilePath:  file.Path,
			StartLine: symbol.Line,
			Context: &model.ViolationContext{
				SuggestedFix: fmt.Sprintf("Rename to '%s' following %s.", suggested, expected),
			},
		})
	}

	return violations
}

func expectedConventionForKind(kind string, conventions map[string]string) string {
	switch normalizeSymbolKind(kind) {
	case "function":
		return conventions["exportedFunctions"]
	case "class":
		return conventions["exportedClasses"]
	case "type":
		return conventions["exportedTypes"]
	case "constant", "value":
		return conventions["exportedConstants"]
	default:
		return ""
	}
}

func resolveExportNamingConventions(language string, options map[string]interface{}) map[string]string {
	conventions := defaultExportNamingConventions(language)
	applyNamingConventionOverrides(conventions, options)

	if options != nil {
		langKey := strings.ToLower(strings.TrimSpace(language))
		switch langKey {
		case "javascript", "jsx", "tsx":
			langKey = "typescript"
		case "golang":
			langKey = "go"
		}
		if nested, ok := options[langKey]; ok {
			if nestedMap, ok := toStringMap(nested); ok {
				applyNamingConventionOverrides(conventions, nestedMap)
			}
		}
	}

	return conventions
}

func defaultExportNamingConventions(language string) map[string]string {
	lang := strings.ToLower(strings.TrimSpace(language))
	if lang == "go" || lang == "golang" {
		return map[string]string{
			"exportedFunctions": namingPascalCase,
			"exportedTypes":     namingPascalCase,
			"exportedConstants": namingPascalCase,
		}
	}
	return map[string]string{
		"exportedFunctions": namingCamelCase,
		"exportedClasses":   namingPascalCase,
		"exportedConstants": namingUpperSnakeCase,
		"exportedTypes":     namingPascalCase,
	}
}

func applyNamingConventionOverrides(conventions map[string]string, values map[string]interface{}) {
	if values == nil {
		return
	}

	for _, key := range []string{"exportedFunctions", "exportedClasses", "exportedConstants", "exportedTypes"} {
		raw, ok := values[key]
		if !ok {
			continue
		}
		name, ok := raw.(string)
		if !ok {
			continue
		}
		normalized := normalizeConventionName(name)
		if normalized == "" {
			continue
		}
		conventions[key] = normalized
	}
}
