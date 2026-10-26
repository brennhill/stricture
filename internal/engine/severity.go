// severity.go — Severity ranking and threshold filtering.
package engine

import (
	"strings"

	"github.com/stricture/stricture/internal/model"
)

// FilterBySeverity drops violations below minSeverity ("warn" or "error"); an empty threshold keeps all.
func FilterBySeverity(violations []model.Violation, minSeverity string) []model.Violation {
	threshold := strings.ToLower(strings.TrimSpace(minSeverity))
	if threshold == "" {
		return violations
	}
	minRank := severityRank(threshold)
	filtered := make([]model.Violation, 0, len(violations))
	for _, v := range violations {
		if severityRank(v.Severity) >= minRank {
			filtered = append(filtered, v)
		}
	}
	return filtered
}

func severityRank(severity string) int {
	switch strings.ToLower(strings.TrimSpace(severity)) {
	case "warn", "warning":
		return 1
	case "error":
		return 2
	default:
		// Unknown severities are treated as errors to avoid accidental suppression.
		return 2
	}
}
