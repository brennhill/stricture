// max_file_lines_test.go — Tests for ARCH-max-file-lines.
package arch

import "testing"

func TestMaxFileLines(t *testing.T) {
	assertRuleContract(t, &MaxFileLines{})
}
