// rule_helpers.go — Shared helpers for CTR rules.
package ctr

import (
	"regexp"
	"strings"

	"github.com/stricture/stricture/internal/model"
)

var ctrPathBugPattern = regexp.MustCompile(`(?i)(?:^|[\\/._-])(b(?:0[1-9]|1[0-5]))(?:[\\/._-]|$)`)

var ctrBugRuleMap = map[string]string{
	"B02": "CTR-status-code-handling",
	"B05": "CTR-request-shape",
	"B06": "CTR-response-shape",
	"B07": "CTR-manifest-conformance",
	"B08": "CTR-strictness-parity",
	"B09": "CTR-strictness-parity",
	"B10": "CTR-strictness-parity",
	"B11": "CTR-strictness-parity",
	"B12": "CTR-response-shape",
	"B13": "CTR-request-shape",
	"B14": "CTR-response-shape",
	"B15": "CTR-request-shape",
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
	expectedRule, ok := ctrBugRuleMap[bugID]
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
		m := ctrPathBugPattern.FindStringSubmatch(path)
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
