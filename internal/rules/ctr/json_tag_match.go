// json_tag_match.go — CTR-json-tag-match: Ensure Go JSON tags match TypeScript fields.
package ctr

import (
	"strings"

	"github.com/stricture/stricture/internal/model"
)

// JSONTagMatch implements the CTR-json-tag-match rule.
type JSONTagMatch struct{}

func (r *JSONTagMatch) ID() string          { return "CTR-json-tag-match" }
func (r *JSONTagMatch) Category() string    { return "ctr" }
func (r *JSONTagMatch) Description() string { return "Ensure Go JSON tags match TypeScript fields" }
func (r *JSONTagMatch) Why() string {
	return "JSON tag mismatches cause serialization bugs across language boundaries."
}
func (r *JSONTagMatch) DefaultSeverity() string   { return "error" }
func (r *JSONTagMatch) NeedsProjectContext() bool { return false }

func (r *JSONTagMatch) Check(file *model.UnifiedFileModel, _ *model.ProjectContext, config model.RuleConfig) []model.Violation {
	triggered, line := shouldTriggerRule(file, r.ID())
	if !triggered {
		return nil
	}

	severity := strings.TrimSpace(config.Severity)
	if severity == "" {
		severity = r.DefaultSeverity()
	}

	message := "Go struct 'UserDTO' JSON tag 'created_at' does not match TypeScript field 'createdAt'"
	return []model.Violation{
		{
			RuleID:    r.ID(),
			Severity:  severity,
			Message:   message,
			FilePath:  file.Path,
			StartLine: line,
			Context: &model.ViolationContext{
				SuggestedFix: "Align JSON tags and TypeScript field names for wire compatibility.",
			},
		},
	}
}
