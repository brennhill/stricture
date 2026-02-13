// no_circular_deps_test.go — Tests for ARCH-no-circular-deps.
package arch

import "testing"

func TestNoCircularDeps(t *testing.T) {
	assertRuleContract(t, &NoCircularDeps{})
}
