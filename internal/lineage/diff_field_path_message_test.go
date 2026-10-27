// diff_field_path_message_test.go — Pins the field_path_changed drift message wording.
package lineage

import "testing"

func TestDiffArtifacts_FieldPathChangedMessage(t *testing.T) {
	base := mkField("response_user_id")
	base.Field = "response.user_id"
	head := mkField("response_user_id")
	head.Field = "response.user.identifier"

	result := DiffArtifacts(
		Artifact{SchemaVersion: "1", Fields: []Annotation{base}},
		Artifact{SchemaVersion: "1", Fields: []Annotation{head}},
	)

	want := "Producer Identity changed field path from response.user_id to response.user.identifier for response_user_id."
	for _, change := range result.Changes {
		if change.ChangeType == "field_path_changed" {
			if change.Message != want {
				t.Fatalf("message = %q, want %q", change.Message, want)
			}
			return
		}
	}
	t.Fatal("missing field_path_changed change")
}
