// policy.go — Inline suppression policy parser and matcher.
package suppression

import (
	"strings"
)

// Policy stores per-file suppression state by line and rule ID.
type Policy struct {
	fileAll   bool
	fileRules map[string]bool
	lineAll   map[int]bool
	lineRules map[int]map[string]bool
}

// Compile parses suppression directives from source and returns a query policy.
func Compile(source []byte) *Policy {
	p := &Policy{
		fileRules: map[string]bool{},
		lineAll:   map[int]bool{},
		lineRules: map[int]map[string]bool{},
	}

	lines := strings.Split(string(source), "\n")
	activeAll := false
	activeRules := map[string]bool{}

	for i, line := range lines {
		lineNo := i + 1

		if activeAll {
			p.lineAll[lineNo] = true
		}
		for ruleID := range activeRules {
			addLineRule(p.lineRules, lineNo, ruleID)
		}

		dir, rules, all := parseDirective(line)
		switch dir {
		case "disable-file":
			if all {
				p.fileAll = true
				continue
			}
			for _, ruleID := range rules {
				p.fileRules[ruleID] = true
			}
		case "disable-next-line":
			next := lineNo + 1
			if all {
				p.lineAll[next] = true
				continue
			}
			for _, ruleID := range rules {
				addLineRule(p.lineRules, next, ruleID)
			}
		case "disable":
			if all {
				activeAll = true
				continue
			}
			for _, ruleID := range rules {
				activeRules[ruleID] = true
			}
		case "enable":
			if all {
				activeAll = false
				activeRules = map[string]bool{}
				continue
			}
			for _, ruleID := range rules {
				delete(activeRules, ruleID)
			}
		}
	}

	return p
}

// Suppressed reports whether a violation at line for ruleID should be filtered.
func (p *Policy) Suppressed(ruleID string, line int) bool {
	if p == nil {
		return false
	}
	if p.fileAll || p.fileRules[ruleID] {
		return true
	}
	if p.lineAll[line] {
		return true
	}
	if byRule, ok := p.lineRules[line]; ok {
		return byRule[ruleID]
	}
	return false
}

func addLineRule(index map[int]map[string]bool, line int, ruleID string) {
	byRule := index[line]
	if byRule == nil {
		byRule = map[string]bool{}
		index[line] = byRule
	}
	byRule[ruleID] = true
}

func parseDirective(line string) (directive string, rules []string, all bool) {
	idx := strings.Index(line, "stricture-")
	if idx < 0 {
		return "", nil, false
	}

	fragment := strings.TrimSpace(line[idx:])
	candidates := []string{
		"stricture-disable-next-line",
		"stricture-disable-file",
		"stricture-disable",
		"stricture-enable",
	}
	for _, candidate := range candidates {
		if !strings.HasPrefix(fragment, candidate) {
			continue
		}

		remainder := strings.TrimSpace(strings.TrimPrefix(fragment, candidate))
		remainder = strings.TrimPrefix(remainder, ":")
		remainder = strings.TrimSpace(remainder)

		if reasonIdx := strings.Index(remainder, "--"); reasonIdx >= 0 {
			remainder = strings.TrimSpace(remainder[:reasonIdx])
		}

		if remainder == "" {
			return strings.TrimPrefix(candidate, "stricture-"), nil, true
		}

		remainder = strings.ReplaceAll(remainder, ",", " ")
		parts := strings.Fields(remainder)
		out := make([]string, 0, len(parts))
		for _, part := range parts {
			id := strings.TrimSpace(part)
			if id == "" {
				continue
			}
			out = append(out, id)
		}

		if len(out) == 0 {
			return strings.TrimPrefix(candidate, "stricture-"), nil, true
		}
		return strings.TrimPrefix(candidate, "stricture-"), out, false
	}

	return "", nil, false
}
