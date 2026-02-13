// mock_scope_test.go — Tests for TQ-mock-scope.
package tq

import "testing"

func TestMockScope(t *testing.T) {
	assertRuleContract(t, &MockScope{})
}
