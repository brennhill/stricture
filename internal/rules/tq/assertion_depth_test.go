// assertion_depth_test.go — Tests for TQ-assertion-depth.
package tq

import "testing"

func TestAssertionDepth(t *testing.T) {
	assertRuleContract(t, &AssertionDepth{})
}
