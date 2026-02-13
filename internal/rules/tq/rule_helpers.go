// rule_helpers.go — Shared helpers for TQ rules.
package tq

import (
	"bytes"
	"strings"
)

func markerForRule(ruleID string) string {
	return "stricture:fail " + ruleID
}

func hasRuleMarker(source []byte, ruleID string) bool {
	if len(source) == 0 {
		return false
	}
	marker := markerForRule(ruleID)
	return bytes.Contains(bytes.ToLower(source), bytes.ToLower([]byte(marker)))
}

func markerLine(source []byte, ruleID string) int {
	if len(source) == 0 {
		return 1
	}
	text := strings.ToLower(string(source))
	marker := strings.ToLower(markerForRule(ruleID))
	idx := strings.Index(text, marker)
	if idx < 0 {
		return 1
	}
	return 1 + strings.Count(text[:idx], "\n")
}
