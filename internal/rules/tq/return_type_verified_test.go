// return_type_verified_test.go — Tests for TQ-return-type-verified.
package tq

import "testing"

func TestReturnTypeVerified(t *testing.T) {
	assertRuleContract(t, &ReturnTypeVerified{})
}
