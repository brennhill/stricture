// response_shape_test.go — Tests for CTR-response-shape.
package ctr

import "testing"

func TestResponseShape(t *testing.T) {
	assertRuleContract(t, &ResponseShape{})
}
