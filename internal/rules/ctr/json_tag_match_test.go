// json_tag_match_test.go — Tests for CTR-json-tag-match.
package ctr

import "testing"

func TestJSONTagMatch(t *testing.T) {
	assertRuleContract(t, &JSONTagMatch{})
}
