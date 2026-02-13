// test_isolation_test.go — Tests for TQ-test-isolation.
package tq

import "testing"

func TestTestIsolation(t *testing.T) {
	assertRuleContract(t, &TestIsolation{})
}
