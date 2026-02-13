// schema_conformance_test.go — Tests for TQ-schema-conformance.
package tq

import "testing"

func TestSchemaConformance(t *testing.T) {
	assertRuleContract(t, &SchemaConformance{})
}
