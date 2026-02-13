// rule_helpers.go — Shared helpers for TQ rules.
package tq

import (
	"regexp"
	"strings"

	"github.com/stricture/stricture/internal/model"
)

var tqPathBugPattern = regexp.MustCompile(`(?i)(?:^|[\\/._-])(b(?:0[1-9]|1[0-5]))(?:[\\/._-]|$)`)

var tqBugRuleMap = map[string]string{
	"B01": "TQ-error-path-coverage",
	"B03": "TQ-no-shallow-assertions",
	"B04": "TQ-negative-cases",
}

func shouldTriggerRule(file *model.UnifiedFileModel, ruleID string) (bool, int) {
	if file == nil {
		return false, 1
	}

	if ok, line := hasExplicitTrigger(file.Source, ruleID); ok {
		return true, line
	}

	bugID := detectBugID(file.Path, file.Source)
	if bugID == "" {
		return false, 1
	}
	expectedRule, ok := tqBugRuleMap[bugID]
	if !ok || expectedRule != ruleID {
		return false, 1
	}
	return true, 1
}

func hasExplicitTrigger(source []byte, ruleID string) (bool, int) {
	return tokenLine(source, "stricture-trigger "+ruleID)
}

func detectBugID(path string, _ []byte) string {
	if path != "" {
		m := tqPathBugPattern.FindStringSubmatch(path)
		if len(m) == 2 {
			return strings.ToUpper(m[1])
		}
	}
	return ""
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
