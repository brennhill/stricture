// dual_test_test.go — Tests for CTR-dual-test.
package ctr

import "testing"

func TestDualTest(t *testing.T) {
	assertRuleContract(t, &DualTest{})
}
