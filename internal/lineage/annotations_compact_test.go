// annotations_compact_test.go — Tests compact-form payload parsing and chunk handling.
package lineage

import (
	"reflect"
	"testing"
)

func TestParseCompactPayload(t *testing.T) {
	cases := []struct {
		name    string
		payload string
		want    map[string]string
		wantOK  bool
	}{
		{"source only", " Identity ", map[string]string{"source_system": "Identity"}, true},
		{"leading colon and from prefix", ": from Identity", map[string]string{"source_system": "Identity"}, true},
		{
			"known chunks map to canonical keys",
			"Identity, source_version v2, transform normalize, merge single_source, classification internal, break_policy additive_only, scope cross_repo, provider idp",
			map[string]string{
				"source_system": "Identity", "source_version": "v2", "transform_type": "normalize",
				"merge_strategy": "single_source", "data_classification": "internal", "break_policy": "additive_only",
				"__scope": "cross_repo", "__provider_id": "idp",
			},
			true,
		},
		{"note is unquoted", `Identity, note "hello world"`, map[string]string{"source_system": "Identity", "note": "hello world"}, true},
		{"unknown, valueless, and empty chunks are ignored", "Identity, bogus x, transform, , merge   ", map[string]string{"source_system": "Identity"}, true},
		{"key=value payloads are not compact form", "source_system=Identity", nil, false},
		{"blank payload", "   ", nil, false},
		{"bare from (trimmed, so no prefix) is the source system", "from   ", map[string]string{"source_system": "from"}, true},
		{"empty source system", " , transform normalize", nil, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := parseCompactPayload(tc.payload)
			if ok != tc.wantOK || !reflect.DeepEqual(got, tc.want) {
				t.Errorf("parseCompactPayload(%q) = %v, %v; want %v, %v", tc.payload, got, ok, tc.want, tc.wantOK)
			}
		})
	}
}
