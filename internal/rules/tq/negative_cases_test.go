// negative_cases_test.go — Tests for TQ-negative-cases.
package tq

import "testing"

func TestNegativeCases(t *testing.T) {
	assertRuleContract(t, &NegativeCases{})
}
