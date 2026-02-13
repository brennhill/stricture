// module_boundary_test.go — Tests for ARCH-module-boundary.
package arch

import "testing"

func TestModuleBoundary(t *testing.T) {
	assertRuleContract(t, &ModuleBoundary{})
}
