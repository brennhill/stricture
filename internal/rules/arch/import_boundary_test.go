// import_boundary_test.go — Tests for ARCH-import-boundary.
package arch

import "testing"

func TestImportBoundary(t *testing.T) {
	assertRuleContract(t, &ImportBoundary{})
}
