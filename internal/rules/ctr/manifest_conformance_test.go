// manifest_conformance_test.go — Tests for CTR-manifest-conformance.
package ctr

import "testing"

func TestManifestConformance(t *testing.T) {
	assertRuleContract(t, &ManifestConformance{})
}
