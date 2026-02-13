// strictness_parity_test.go — Tests for CTR-strictness-parity.
package ctr

import "testing"

func TestStrictnessParity(t *testing.T) {
	assertRuleContract(t, &StrictnessParity{})
}
