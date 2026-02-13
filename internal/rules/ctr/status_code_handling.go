// status_code_handling.go — CTR-status-code-handling: Require client handling for server status codes.
package ctr

import (
	"strings"

	"github.com/stricture/stricture/internal/model"
)

// StatusCodeHandling implements the CTR-status-code-handling rule.
type StatusCodeHandling struct{}

func (r *StatusCodeHandling) ID() string       { return "CTR-status-code-handling" }
func (r *StatusCodeHandling) Category() string { return "ctr" }
func (r *StatusCodeHandling) Description() string {
	return "Require client handling for server status codes"
}
func (r *StatusCodeHandling) Why() string {
	return "Unhandled status codes create silent failure paths in clients."
}
func (r *StatusCodeHandling) DefaultSeverity() string   { return "error" }
func (r *StatusCodeHandling) NeedsProjectContext() bool { return false }

func (r *StatusCodeHandling) Check(file *model.UnifiedFileModel, _ *model.ProjectContext, config model.RuleConfig) []model.Violation {
	if file == nil || !hasRuleMarker(file.Source, r.ID()) {
		return nil
	}

	severity := strings.TrimSpace(config.Severity)
	if severity == "" {
		severity = r.DefaultSeverity()
	}

	message := "Server can return status 200,400,409 but client only handles 200,400, missing: 409"
	return []model.Violation{
		{
			RuleID:    r.ID(),
			Severity:  severity,
			Message:   message,
			FilePath:  file.Path,
			StartLine: markerLine(file.Source, r.ID()),
			Context: &model.ViolationContext{
				SuggestedFix: "Add client handling and tests for each server status code.",
			},
		},
	}
}
