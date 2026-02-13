// no_shallow_assertions_test.go — Tests for TQ-no-shallow-assertions.
package tq

import "testing"

func TestNoShallowAssertions(t *testing.T) {
	assertRuleContract(t, &NoShallowAssertions{})
}
