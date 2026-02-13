// layer_violation_test.go — Tests for ARCH-layer-violation.
package arch

import "testing"

func TestLayerViolation(t *testing.T) {
	assertRuleContract(t, &LayerViolation{})
}
