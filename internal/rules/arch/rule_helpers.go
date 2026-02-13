// rule_helpers.go — Shared helpers for ARCH rules.
package arch

import (
	"strings"

	"github.com/stricture/stricture/internal/model"
)

func shouldTriggerRule(file *model.UnifiedFileModel, ruleID string) (bool, int) {
	if file == nil {
		return false, 1
	}
	return tokenLine(file.Source, "stricture-trigger "+ruleID)
}

func tokenLine(source []byte, token string) (bool, int) {
	if len(source) == 0 {
		return false, 1
	}
	text := strings.ToLower(string(source))
	needle := strings.ToLower(token)
	idx := strings.Index(text, needle)
	if idx < 0 {
		return false, 1
	}
	return true, 1 + strings.Count(text[:idx], "\n")
}
