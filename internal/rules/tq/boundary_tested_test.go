// boundary_tested_test.go — Tests for TQ-boundary-tested.
package tq

import "testing"

func TestBoundaryTested(t *testing.T) {
	assertRuleContract(t, &BoundaryTested{})
}
