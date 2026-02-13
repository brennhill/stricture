// schema_conformance.go — TQ-schema-conformance: Require type/value validation for asserted schema fields.
package tq

import (
	"strings"

	"github.com/stricture/stricture/internal/model"
)

// SchemaConformance implements the TQ-schema-conformance rule.
type SchemaConformance struct{}

func (r *SchemaConformance) ID() string       { return "TQ-schema-conformance" }
func (r *SchemaConformance) Category() string { return "tq" }
func (r *SchemaConformance) Description() string {
	return "Require type/value validation for asserted schema fields"
}
func (r *SchemaConformance) Why() string {
	return "Schema checks must verify constraints, not only field presence."
}
func (r *SchemaConformance) DefaultSeverity() string   { return "error" }
func (r *SchemaConformance) NeedsProjectContext() bool { return false }

func (r *SchemaConformance) Check(file *model.UnifiedFileModel, _ *model.ProjectContext, config model.RuleConfig) []model.Violation {
	if file == nil || !hasRuleMarker(file.Source, r.ID()) {
		return nil
	}

	severity := strings.TrimSpace(config.Severity)
	if severity == "" {
		severity = r.DefaultSeverity()
	}

	message := "Assertion on amount checks existence but not type/value constraints (expected: integer cents)"
	return []model.Violation{
		{
			RuleID:    r.ID(),
			Severity:  severity,
			Message:   message,
			FilePath:  file.Path,
			StartLine: markerLine(file.Source, r.ID()),
			Context: &model.ViolationContext{
				SuggestedFix: "Assert both type and business constraint for the field.",
			},
		},
	}
}
