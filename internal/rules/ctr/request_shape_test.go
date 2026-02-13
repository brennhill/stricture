// request_shape_test.go — Tests for CTR-request-shape.
package ctr

import "testing"

func TestRequestShape(t *testing.T) {
	assertRuleContract(t, &RequestShape{})
}
