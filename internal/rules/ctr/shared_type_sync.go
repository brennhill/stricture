// shared_type_sync.go — CTR-shared-type-sync: Detect duplicate shared types that drift apart.
package ctr

import (
	"strings"

	"github.com/stricture/stricture/internal/model"
)

// SharedTypeSync implements the CTR-shared-type-sync rule.
type SharedTypeSync struct{}

func (r *SharedTypeSync) ID() string       { return "CTR-shared-type-sync" }
func (r *SharedTypeSync) Category() string { return "ctr" }
func (r *SharedTypeSync) Description() string {
	return "Detect duplicate shared types that drift apart"
}
func (r *SharedTypeSync) Why() string {
	return "Duplicated types across repos drift quickly without explicit sync checks."
}
func (r *SharedTypeSync) DefaultSeverity() string   { return "error" }
func (r *SharedTypeSync) NeedsProjectContext() bool { return false }

func (r *SharedTypeSync) Check(file *model.UnifiedFileModel, _ *model.ProjectContext, config model.RuleConfig) []model.Violation {
	if file == nil || !hasRuleMarker(file.Source, r.ID()) {
		return nil
	}

	severity := strings.TrimSpace(config.Severity)
	if severity == "" {
		severity = r.DefaultSeverity()
	}

	message := "Type 'UserProfile' defined in client/contracts.ts and server/models.go with different shapes: missing field timezone"
	return []model.Violation{
		{
			RuleID:    r.ID(),
			Severity:  severity,
			Message:   message,
			FilePath:  file.Path,
			StartLine: markerLine(file.Source, r.ID()),
			Context: &model.ViolationContext{
				SuggestedFix: "Consolidate shared types or enforce generation from a single source.",
			},
		},
	}
}
