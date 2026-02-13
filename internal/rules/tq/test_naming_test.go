// test_naming_test.go — Tests for TQ-test-naming.
package tq

import "testing"

func TestTestNaming(t *testing.T) {
	assertRuleContract(t, &TestNaming{})
}
