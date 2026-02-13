// annotations_test.go - Tests for Stricture data-lineage annotations.
package lineage

import (
	"strings"
	"testing"
)

func validAnnotationLine() string {
	return `// stricture-source annotation_schema_version=1 field_id=response_user_id field=response.user_id source_system=Identity source_version=v2026.02 min_supported_source_version=v2026.01 transform_type=normalize merge_strategy=priority break_policy=additive_only confidence=declared data_classification=internal owner=team.identity escalation=slack:#identity-oncall contract_test_id=ci://contracts/identity-user-id introduced_at=2026-01-10 sources=api:identity.GetUser#response.id@cross_repo?contract_ref=git+https://github.com/acme/identity//openapi.yaml@a1b2,db:users.user_profile#user_id@internal?contract_ref=internal://db/users.user_profile flow="from @Identity normalized @self" note="normalized by UserNormalizer.Apply; spec=https://specs.example.com/user-id"`
}

func TestParse_ValidAnnotation_AllRequiredFields(t *testing.T) {
	source := []byte(validAnnotationLine() + "\n")

	annotations, errs := Parse(source)
	if len(errs) > 0 {
		t.Fatalf("unexpected parse errors: %+v", errs)
	}
	if len(annotations) != 1 {
		t.Fatalf("annotations len = %d, want 1", len(annotations))
	}

	a := annotations[0]
	if a.AnnotationSchemaVersion != "1" {
		t.Fatalf("annotation_schema_version = %q, want 1", a.AnnotationSchemaVersion)
	}
	if a.FieldID != "response_user_id" {
		t.Fatalf("field_id = %q, want response_user_id", a.FieldID)
	}
	if a.TransformType != "normalize" {
		t.Fatalf("transform_type = %q, want normalize", a.TransformType)
	}
	if a.MergeStrategy != "priority" {
		t.Fatalf("merge_strategy = %q, want priority", a.MergeStrategy)
	}
	if a.BreakPolicy != "additive_only" {
		t.Fatalf("break_policy = %q, want additive_only", a.BreakPolicy)
	}
	if a.Confidence != "declared" {
		t.Fatalf("confidence = %q, want declared", a.Confidence)
	}
	if a.DataClassification != "internal" {
		t.Fatalf("data_classification = %q, want internal", a.DataClassification)
	}
	if a.Owner != "team.identity" {
		t.Fatalf("owner = %q, want team.identity", a.Owner)
	}
	if a.Escalation != "slack:#identity-oncall" {
		t.Fatalf("escalation = %q, want slack:#identity-oncall", a.Escalation)
	}
	if a.ContractTestID != "ci://contracts/identity-user-id" {
		t.Fatalf("contract_test_id = %q, unexpected", a.ContractTestID)
	}
	if a.IntroducedAt != "2026-01-10" {
		t.Fatalf("introduced_at = %q, want 2026-01-10", a.IntroducedAt)
	}
	if a.SunsetAt != "" {
		t.Fatalf("sunset_at = %q, want empty", a.SunsetAt)
	}
	if len(a.Sources) != 2 {
		t.Fatalf("sources len = %d, want 2", len(a.Sources))
	}
	if a.Sources[0].ContractRef == "" || a.Sources[1].ContractRef == "" {
		t.Fatalf("expected contract_ref on each source")
	}
}

func TestParse_ValidExternalSource_RequiresProviderAndAsOf(t *testing.T) {
	source := []byte(`// stricture-source annotation_schema_version=1 field_id=response_track field=response.track source_system=Media source_version=v1 min_supported_source_version=v1 transform_type=passthrough merge_strategy=single_source break_policy=strict confidence=declared data_classification=public owner=team.media escalation=pagerduty:media contract_test_id=ci://contracts/media-track introduced_at=2026-01-10 sources=api:spotify.GetTrack#response.track@external!2026-02-13?provider_id=spotify&contract_ref=https://developer.spotify.com/reference/get-track&upstream_system=spotify flow="from @Spotify enriched @self" note="mapped in TrackMapper"`)

	annotations, errs := Parse(source)
	if len(errs) > 0 {
		t.Fatalf("unexpected parse errors: %+v", errs)
	}
	if len(annotations) != 1 {
		t.Fatalf("annotations len = %d, want 1", len(annotations))
	}

	s := annotations[0].Sources[0]
	if s.Scope != "external" {
		t.Fatalf("scope = %q, want external", s.Scope)
	}
	if s.AsOf != "2026-02-13" {
		t.Fatalf("as_of = %q, want 2026-02-13", s.AsOf)
	}
	if s.ProviderID != "spotify" {
		t.Fatalf("provider_id = %q, want spotify", s.ProviderID)
	}
	if s.UpstreamSystem != "spotify" {
		t.Fatalf("upstream_system = %q, want spotify", s.UpstreamSystem)
	}
}

func TestParse_ValidRenameTracking(t *testing.T) {
	source := []byte(`// stricture-source annotation_schema_version=1 field_id=response_user_primary_id renamed_from=response_user_id field=response.user_primary_id source_system=Identity source_version=v2026.03 min_supported_source_version=v2026.01 transform_type=normalize merge_strategy=single_source break_policy=additive_only confidence=declared data_classification=internal owner=team.identity escalation=slack:#identity-oncall contract_test_id=ci://contracts/identity-user-id introduced_at=2026-02-01 sources=api:identity.GetUser#response.id@cross_repo?contract_ref=git+https://github.com/acme/identity//openapi.yaml@c3d4 flow="from @Identity normalized @self" note="renamed field to align with product terminology"`)

	annotations, errs := Parse(source)
	if len(errs) > 0 {
		t.Fatalf("unexpected parse errors: %+v", errs)
	}
	if annotations[0].RenamedFrom != "response_user_id" {
		t.Fatalf("renamed_from = %q, want response_user_id", annotations[0].RenamedFrom)
	}
}

func TestParse_RejectsMissingRequiredKeys(t *testing.T) {
	source := []byte(`// stricture-source field=response.user_id source_system=Identity`)
	_, errs := Parse(source)
	if len(errs) != 1 {
		t.Fatalf("errors len = %d, want 1", len(errs))
	}
}

func TestParse_RejectsInvalidFieldID(t *testing.T) {
	line := validAnnotationLine()
	line = replaceToken(line, "field_id=response_user_id", "field_id=ResponseUserId")
	_, errs := Parse([]byte(line))
	if len(errs) != 1 {
		t.Fatalf("errors len = %d, want 1", len(errs))
	}
}

func TestParse_RejectsInvalidEnumValues(t *testing.T) {
	line := validAnnotationLine()
	line = replaceToken(line, "transform_type=normalize", "transform_type=magic")
	_, errs := Parse([]byte(line))
	if len(errs) != 1 {
		t.Fatalf("errors len = %d, want 1", len(errs))
	}
}

func TestParse_RejectsMultiSourceWithSingleSourceStrategy(t *testing.T) {
	line := validAnnotationLine()
	line = replaceToken(line, "merge_strategy=priority", "merge_strategy=single_source")
	_, errs := Parse([]byte(line))
	if len(errs) != 1 {
		t.Fatalf("errors len = %d, want 1", len(errs))
	}
}

func TestParse_RejectsExternalWithoutAsOf(t *testing.T) {
	source := []byte(`// stricture-source annotation_schema_version=1 field_id=response_track field=response.track source_system=Media source_version=v1 min_supported_source_version=v1 transform_type=passthrough merge_strategy=single_source break_policy=strict confidence=declared data_classification=public owner=team.media escalation=pagerduty:media contract_test_id=ci://contracts/media-track introduced_at=2026-01-10 sources=api:spotify.GetTrack#response.track@external?provider_id=spotify&contract_ref=https://developer.spotify.com/reference/get-track flow="from @Spotify enriched @self" note="mapped in TrackMapper"`)
	_, errs := Parse(source)
	if len(errs) != 1 {
		t.Fatalf("errors len = %d, want 1", len(errs))
	}
}

func TestParse_RejectsExternalWithoutProviderID(t *testing.T) {
	source := []byte(`// stricture-source annotation_schema_version=1 field_id=response_track field=response.track source_system=Media source_version=v1 min_supported_source_version=v1 transform_type=passthrough merge_strategy=single_source break_policy=strict confidence=declared data_classification=public owner=team.media escalation=pagerduty:media contract_test_id=ci://contracts/media-track introduced_at=2026-01-10 sources=api:spotify.GetTrack#response.track@external!2026-02-13?contract_ref=https://developer.spotify.com/reference/get-track flow="from @Spotify enriched @self" note="mapped in TrackMapper"`)
	_, errs := Parse(source)
	if len(errs) != 1 {
		t.Fatalf("errors len = %d, want 1", len(errs))
	}
}

func TestParse_RejectsSourceMissingContractRef(t *testing.T) {
	line := validAnnotationLine()
	line = replaceToken(line, "@cross_repo?contract_ref=git+https://github.com/acme/identity//openapi.yaml@a1b2", "@cross_repo")
	_, errs := Parse([]byte(line))
	if len(errs) != 1 {
		t.Fatalf("errors len = %d, want 1", len(errs))
	}
}

func TestParse_RejectsInvalidDates(t *testing.T) {
	line := validAnnotationLine()
	line = replaceToken(line, "introduced_at=2026-01-10", "introduced_at=2026/01/10")
	_, errs := Parse([]byte(line))
	if len(errs) != 1 {
		t.Fatalf("errors len = %d, want 1", len(errs))
	}
}

func TestParse_RejectsInvalidFlow(t *testing.T) {
	line := validAnnotationLine()
	line = replaceToken(line, "flow=\"from @Identity normalized @self\"", "flow=\"normalized @self\"")
	_, errs := Parse([]byte(line))
	if len(errs) != 1 {
		t.Fatalf("errors len = %d, want 1", len(errs))
	}
}

func TestParse_IgnoresNonAnnotationComments(t *testing.T) {
	source := []byte(`// regular comment
// another comment
var x = 1
`)

	annotations, errs := Parse(source)
	if len(errs) > 0 {
		t.Fatalf("unexpected parse errors: %+v", errs)
	}
	if len(annotations) != 0 {
		t.Fatalf("annotations len = %d, want 0", len(annotations))
	}
}

func TestParseWithOverrides_ValidOverride(t *testing.T) {
	source := []byte(`// stricture-lineage-override field_id=response_user_id change_type=field_removed expires=2099-12-31 reason="temporary migration window" ticket=INC-123`)
	annotations, overrides, errs := ParseWithOverrides(source)
	if len(errs) != 0 {
		t.Fatalf("unexpected parse errors: %+v", errs)
	}
	if len(annotations) != 0 {
		t.Fatalf("annotations len = %d, want 0", len(annotations))
	}
	if len(overrides) != 1 {
		t.Fatalf("overrides len = %d, want 1", len(overrides))
	}
	if overrides[0].FieldID != "response_user_id" {
		t.Fatalf("field_id = %q, want response_user_id", overrides[0].FieldID)
	}
	if overrides[0].ChangeType != "field_removed" {
		t.Fatalf("change_type = %q, want field_removed", overrides[0].ChangeType)
	}
}

func TestParseWithOverrides_RejectsInvalidOverride(t *testing.T) {
	source := []byte(`// stricture-lineage-override field_id=ResponseUser change_type=BAD expires=not-a-date reason=""`)
	_, overrides, errs := ParseWithOverrides(source)
	if len(overrides) != 0 {
		t.Fatalf("overrides len = %d, want 0", len(overrides))
	}
	if len(errs) != 1 {
		t.Fatalf("errs len = %d, want 1", len(errs))
	}
}

func replaceToken(input string, old string, replacement string) string {
	return strings.Replace(input, old, replacement, 1)
}
