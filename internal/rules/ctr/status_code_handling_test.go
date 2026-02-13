// status_code_handling_test.go — Tests for CTR-status-code-handling.
package ctr

import "testing"

func TestStatusCodeHandling(t *testing.T) {
	assertRuleContract(t, &StatusCodeHandling{})
}
