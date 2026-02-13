// dependency_direction_test.go — Tests for ARCH-dependency-direction.
package arch

import "testing"

func TestDependencyDirection(t *testing.T) {
	assertRuleContract(t, &DependencyDirection{})
}
