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
	c := &compiler{
		policy: &Policy{
			fileRules: map[string]bool{},
			lineAll:   map[int]bool{},
			lineRules: map[int]map[string]bool{},
		},
		activeRules: map[string]bool{},
	}
	for i, line := range strings.Split(string(source), "\n") {
		lineNo := i + 1
		c.markActive(lineNo)
		dir, rules, all := parseDirective(line)
		c.apply(dir, lineNo, rules, all)
	}
	return c.policy
}

// compiler tracks the block-level (disable ... enable) state while scanning lines.
type compiler struct {
	policy      *Policy
	activeAll   bool
	activeRules map[string]bool
}

// markActive applies any open disable block to lineNo.
func (c *compiler) markActive(lineNo int) {
	if c.activeAll {
		c.policy.lineAll[lineNo] = true
	}
	for ruleID := range c.activeRules {
		addLineRule(c.policy.lineRules, lineNo, ruleID)
	}
}

func (c *compiler) apply(dir string, lineNo int, rules []string, all bool) {
	switch dir {
	case "disable-file":
		c.disableFile(rules, all)
	case "disable-next-line":
		c.disableLine(lineNo+1, rules, all)
	case "disable":
		c.openBlock(rules, all)
	case "enable":
		c.closeBlock(rules, all)
	}
}

func (c *compiler) disableFile(rules []string, all bool) {
	if all {
		c.policy.fileAll = true
		return
	}
	for _, ruleID := range rules {
		c.policy.fileRules[ruleID] = true
	}
}

func (c *compiler) disableLine(lineNo int, rules []string, all bool) {
	if all {
		c.policy.lineAll[lineNo] = true
		return
	}
	for _, ruleID := range rules {
		addLineRule(c.policy.lineRules, lineNo, ruleID)
	}
}

func (c *compiler) openBlock(rules []string, all bool) {
	if all {
		c.activeAll = true
		return
	}
	for _, ruleID := range rules {
		c.activeRules[ruleID] = true
	}
}

func (c *compiler) closeBlock(rules []string, all bool) {
	if all {
		c.activeAll = false
		c.activeRules = map[string]bool{}
		return
	}
	for _, ruleID := range rules {
		delete(c.activeRules, ruleID)
	}
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
