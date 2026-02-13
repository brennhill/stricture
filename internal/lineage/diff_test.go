// diff_test.go - Tests for lineage artifact drift classification.
package lineage

import (
	"testing"
	"time"
)

func mkField(id string) Annotation {
	return Annotation{
		AnnotationSchemaVersion:   "1",
		FieldID:                   id,
		Field:                     "response.user_id",
		SourceSystem:              "Identity",
		SourceVersion:             "v1",
		MinSupportedSourceVersion: "v1",
		TransformType:             "normalize",
		MergeStrategy:             "single_source",
		BreakPolicy:               "additive_only",
		Confidence:                "declared",
		DataClassification:        "internal",
		Owner:                     "team.identity",
		Escalation:                "slack:#identity-oncall",
		ContractTestID:            "ci://contracts/identity-user-id",
		IntroducedAt:              "2026-01-01",
		Sources: []SourceRef{{
			Kind:        "api",
			Target:      "identity.GetUser",
			Path:        "response.id",
			Scope:       "cross_repo",
			ContractRef: "git+https://github.com/acme/identity//openapi.yaml@a1",
			Raw:         "api:identity.GetUser#response.id@cross_repo?contract_ref=git+https://github.com/acme/identity//openapi.yaml@a1",
		}},
		Flow: "from @Identity normalized @self",
		Note: "normalized",
	}
}

func TestDiffArtifacts_RemovedFieldIsHighSeverity(t *testing.T) {
	base := Artifact{SchemaVersion: "1", Fields: []Annotation{mkField("response_user_id")}}
	head := Artifact{SchemaVersion: "1", Fields: []Annotation{}}

	result := DiffArtifacts(base, head)
	if len(result.Changes) == 0 {
		t.Fatalf("expected changes")
	}
	if result.Changes[0].Severity != SeverityHigh {
		t.Fatalf("first severity = %q, want high", result.Changes[0].Severity)
	}
}

func TestDiffArtifacts_RenameTrackedWithoutRemoval(t *testing.T) {
	base := Artifact{SchemaVersion: "1", Fields: []Annotation{mkField("response_user_id")}}
	renamed := mkField("response_user_primary_id")
	renamed.RenamedFrom = "response_user_id"
	head := Artifact{SchemaVersion: "1", Fields: []Annotation{renamed}}

	result := DiffArtifacts(base, head)
	for _, c := range result.Changes {
		if c.ChangeType == "field_removed" {
			t.Fatalf("did not expect field_removed when renamed_from is provided")
		}
	}
}

func TestDiffArtifacts_ClassificationRelaxationIsHigh(t *testing.T) {
	baseField := mkField("response_user_id")
	baseField.DataClassification = "regulated"
	headField := mkField("response_user_id")
	headField.DataClassification = "public"

	base := Artifact{SchemaVersion: "1", Fields: []Annotation{baseField}}
	head := Artifact{SchemaVersion: "1", Fields: []Annotation{headField}}
	result := DiffArtifacts(base, head)

	found := false
	for _, c := range result.Changes {
		if c.ChangeType == "classification_relaxed" {
			found = true
			if c.Severity != SeverityHigh {
				t.Fatalf("classification_relaxed severity = %q, want high", c.Severity)
			}
		}
	}
	if !found {
		t.Fatalf("expected classification_relaxed change")
	}
}

func TestDiffArtifacts_ExternalAsOfRollbackIsHigh(t *testing.T) {
	baseField := mkField("response_song")
	baseField.Sources = []SourceRef{{
		Kind:        "api",
		Target:      "spotify.GetTrack",
		Path:        "response.track",
		Scope:       "external",
		AsOf:        "2026-02-14",
		ProviderID:  "spotify",
		ContractRef: "https://developer.spotify.com/reference/get-track",
		Raw:         "api:spotify.GetTrack#response.track@external!2026-02-14?provider_id=spotify&contract_ref=https://developer.spotify.com/reference/get-track",
	}}
	headField := baseField
	headField.Sources = []SourceRef{{
		Kind:        "api",
		Target:      "spotify.GetTrack",
		Path:        "response.track",
		Scope:       "external",
		AsOf:        "2026-02-13",
		ProviderID:  "spotify",
		ContractRef: "https://developer.spotify.com/reference/get-track",
		Raw:         "api:spotify.GetTrack#response.track@external!2026-02-13?provider_id=spotify&contract_ref=https://developer.spotify.com/reference/get-track",
	}}

	base := Artifact{SchemaVersion: "1", Fields: []Annotation{baseField}}
	head := Artifact{SchemaVersion: "1", Fields: []Annotation{headField}}
	result := DiffArtifacts(base, head)

	found := false
	for _, c := range result.Changes {
		if c.ChangeType == "external_as_of_rollback" {
			found = true
			if c.Severity != SeverityHigh {
				t.Fatalf("external_as_of_rollback severity = %q, want high", c.Severity)
			}
		}
	}
	if !found {
		t.Fatalf("expected external_as_of_rollback change")
	}
}

func TestDiffArtifacts_ActiveOverrideSuppressesFailure(t *testing.T) {
	base := Artifact{SchemaVersion: "1", Fields: []Annotation{mkField("response_user_id")}}
	head := Artifact{
		SchemaVersion: "1",
		Fields:        []Annotation{},
		Overrides: []Override{
			{
				FieldID:    "response_user_id",
				ChangeType: "field_removed",
				Expires:    "2099-12-31",
				Reason:     "temporary migration window",
				Ticket:     "INC-123",
			},
		},
	}

	result := DiffArtifacts(base, head)
	if len(result.Changes) == 0 {
		t.Fatalf("expected changes")
	}
	if !result.Changes[0].Overridden {
		t.Fatalf("expected change to be overridden")
	}
	if ShouldFailAtThreshold(result, SeverityHigh) {
		t.Fatalf("expected no failure at high threshold because change is overridden")
	}
}

func TestDiffArtifacts_ExpiredOverrideDoesNotSuppressFailure(t *testing.T) {
	base := Artifact{SchemaVersion: "1", Fields: []Annotation{mkField("response_user_id")}}
	head := Artifact{
		SchemaVersion: "1",
		Fields:        []Annotation{},
		Overrides: []Override{
			{
				FieldID:    "response_user_id",
				ChangeType: "field_removed",
				Expires:    "2000-01-01",
				Reason:     "expired",
			},
		},
	}

	result := DiffArtifacts(base, head)
	if len(result.Changes) == 0 {
		t.Fatalf("expected changes")
	}
	if result.Changes[0].Overridden {
		t.Fatalf("did not expect override to apply when expired")
	}
	if !ShouldFailAtThreshold(result, SeverityHigh) {
		t.Fatalf("expected failure at high threshold for non-overridden removal")
	}
}

func TestParseEnforcementMode_ValidValues(t *testing.T) {
	mode, err := ParseEnforcementMode("block")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mode != ModeBlock {
		t.Fatalf("mode = %q, want %q", mode, ModeBlock)
	}

	mode, err = ParseEnforcementMode("warn")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mode != ModeWarn {
		t.Fatalf("mode = %q, want %q", mode, ModeWarn)
	}
}

func TestParseEnforcementMode_InvalidValue(t *testing.T) {
	_, err := ParseEnforcementMode("soft")
	if err == nil {
		t.Fatalf("expected parse error")
	}
}

func TestShouldFailAtThresholdWithMode_WarnNeverFails(t *testing.T) {
	base := Artifact{SchemaVersion: "1", Fields: []Annotation{mkField("response_user_id")}}
	head := Artifact{SchemaVersion: "1", Fields: []Annotation{}}
	result := DiffArtifacts(base, head)

	if !ShouldFailAtThreshold(result, SeverityHigh) {
		t.Fatalf("expected baseline failure in block behavior")
	}
	if ShouldFailAtThresholdWithMode(result, SeverityHigh, ModeWarn) {
		t.Fatalf("did not expect failure in warn mode")
	}
	if !ShouldFailAtThresholdWithMode(result, SeverityHigh, ModeBlock) {
		t.Fatalf("expected failure in block mode")
	}
}

func TestDiffArtifacts_OverrideExpiresTodayIsActive(t *testing.T) {
	base := Artifact{SchemaVersion: "1", Fields: []Annotation{mkField("response_user_id")}}
	head := Artifact{
		SchemaVersion: "1",
		Fields:        []Annotation{},
		Overrides: []Override{
			{
				FieldID:    "response_user_id",
				ChangeType: "field_removed",
				Expires:    time.Now().UTC().Format("2006-01-02"),
				Reason:     "valid through end of day",
			},
		},
	}

	result := DiffArtifacts(base, head)
	if len(result.Changes) == 0 {
		t.Fatalf("expected changes")
	}
	if !result.Changes[0].Overridden {
		t.Fatalf("expected override expiring today to remain active")
	}
}
