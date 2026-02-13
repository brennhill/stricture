// assert.go — Violation assertion helpers for rule tests.
package testutil

import (
	"testing"

	"github.com/stricture/stricture/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// AssertViolationCount checks that exactly n violations were produced.
func AssertViolationCount(t *testing.T, violations []model.Violation, want int) {
	t.Helper()
	require.Len(t, violations, want, "violation count mismatch")
}

// AssertNoViolations checks that no violations were produced.
func AssertNoViolations(t *testing.T, violations []model.Violation) {
	t.Helper()
	assert.Empty(t, violations, "expected no violations")
}

// AssertHasViolation checks that at least one violation has the given rule ID.
func AssertHasViolation(t *testing.T, violations []model.Violation, ruleID string) {
	t.Helper()
	for _, v := range violations {
		if v.RuleID == ruleID {
			return
		}
	}
	t.Errorf("expected violation with rule ID %q, got %v", ruleID, ruleIDs(violations))
}

// AssertViolationFields checks that a violation has the expected fields.
func AssertViolationFields(t *testing.T, v model.Violation, ruleID string, line int) {
	t.Helper()
	assert.Equal(t, ruleID, v.RuleID, "rule ID mismatch")
	assert.Equal(t, line, v.StartLine, "start line mismatch")
	assert.NotEmpty(t, v.Message, "violation message must not be empty")
	assert.NotEmpty(t, v.Severity, "violation severity must not be empty")
	assert.NotEmpty(t, v.FilePath, "violation file path must not be empty")
}

// AssertViolationMessage checks that the first violation's message contains a substring.
func AssertViolationMessage(t *testing.T, violations []model.Violation, contains string) {
	t.Helper()
	require.NotEmpty(t, violations, "expected at least one violation")
	assert.Contains(t, violations[0].Message, contains, "violation message mismatch")
}

// AssertAllViolationsHaveRule checks that all violations belong to the same rule.
func AssertAllViolationsHaveRule(t *testing.T, violations []model.Violation, ruleID string) {
	t.Helper()
	for i, v := range violations {
		assert.Equal(t, ruleID, v.RuleID, "violation[%d] rule ID mismatch", i)
	}
}

// AssertViolationHasSuggestion checks that a violation has a non-empty suggested fix.
func AssertViolationHasSuggestion(t *testing.T, v model.Violation) {
	t.Helper()
	require.NotNil(t, v.Context, "violation context must not be nil")
	assert.NotEmpty(t, v.Context.SuggestedFix, "violation must have a suggested fix")
}

// AssertRuleInterface validates that a rule properly implements the Rule interface.
func AssertRuleInterface(t *testing.T, rule model.Rule, wantID, wantCategory, wantSeverity string) {
	t.Helper()
	assert.Equal(t, wantID, rule.ID(), "rule ID")
	assert.Equal(t, wantCategory, rule.Category(), "rule category")
	assert.Equal(t, wantSeverity, rule.DefaultSeverity(), "default severity")
	assert.NotEmpty(t, rule.Description(), "description must not be empty")
	assert.NotEmpty(t, rule.Why(), "why must not be empty")
}

func ruleIDs(violations []model.Violation) []string {
	ids := make([]string, len(violations))
	for i, v := range violations {
		ids[i] = v.RuleID
	}
	return ids
}
