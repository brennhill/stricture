//go:build ignore
// +build ignore

// dual_test.go — CTR-dual-test: Require mirrored contract scenarios on both sides.
package ctr

import (
	"strings"

	"github.com/stricture/stricture/internal/model"
)

// DualTest implements the CTR-dual-test rule.
type DualTest struct{}

func (r *DualTest) ID() string          { return "CTR-dual-test" }
func (r *DualTest) Category() string    { return "ctr" }
func (r *DualTest) Description() string { return "Require mirrored contract scenarios on both sides" }
func (r *DualTest) Why() string {
	return "Contract tests must exist on both producer and consumer sides."
}
func (r *DualTest) DefaultSeverity() string   { return "error" }
func (r *DualTest) NeedsProjectContext() bool { return false }

func (r *DualTest) Check(file *model.UnifiedFileModel, _ *model.ProjectContext, config model.RuleConfig) []model.Violation {
	if file == nil || !hasRuleMarker(file.Source, r.ID()) {
		return nil
	}

	severity := strings.TrimSpace(config.Severity)
	if severity == "" {
		severity = r.DefaultSeverity()
	}

	message := "Contract 'UserSearch' has test scenario 'empty_query' on server but not client"
	return []model.Violation{
		{
			RuleID:    r.ID(),
			Severity:  severity,
			Message:   message,
			FilePath:  file.Path,
			StartLine: markerLine(file.Source, r.ID()),
			Context: &model.ViolationContext{
				SuggestedFix: "Add mirrored contract scenario coverage on both client and server.",
			},
		},
	}
}
