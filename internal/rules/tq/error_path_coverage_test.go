// error_path_coverage_test.go — Tests for TQ-error-path-coverage.
package tq

import "testing"

func TestErrorPathCoverage(t *testing.T) {
	assertRuleContract(t, &ErrorPathCoverage{})
}
