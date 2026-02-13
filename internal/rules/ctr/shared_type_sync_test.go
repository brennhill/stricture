// shared_type_sync_test.go — Tests for CTR-shared-type-sync.
package ctr

import "testing"

func TestSharedTypeSync(t *testing.T) {
	assertRuleContract(t, &SharedTypeSync{})
}
