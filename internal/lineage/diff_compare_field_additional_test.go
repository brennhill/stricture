package lineage

import "testing"

func requireChangeSeverity(t *testing.T, changes []DriftChange, changeType string, severity Severity) {
	t.Helper()

	for _, change := range changes {
		if change.ChangeType != changeType {
			continue
		}
		if change.Severity != severity {
			t.Fatalf("change %s severity = %s, want %s", changeType, change.Severity, severity)
		}
		return
	}

	t.Fatalf("missing change type %s", changeType)
}

func TestDiffArtifacts_EmitsComprehensiveFieldAndSourceDrift(t *testing.T) {
	base := mkField("response_user_id")
	base.BreakPolicy = "strict"
	base.MergeStrategy = "single_source"
	base.Sources = []SourceRef{
		{
			Kind:           "api",
			Target:         "spotify.GetTrack",
			Path:           "response.track",
			Scope:          "external",
			AsOf:           "2026-02-13",
			ProviderID:     "spotify",
			UpstreamSystem: "spotify",
			ContractRef:    "https://example.com/v1",
			Raw:            "api:spotify.GetTrack#response.track@external!2026-02-13?provider_id=spotify&contract_ref=https://example.com/v1&upstream_system=spotify",
		},
		{
			Kind:        "db",
			Target:      "users.profile",
			Path:        "user_id",
			Scope:       "internal",
			ContractRef: "internal://db/users.profile",
			Raw:         "db:users.profile#user_id@internal?contract_ref=internal://db/users.profile",
		},
	}

	head := base
	head.Field = "response.user_primary_id"
	head.SourceSystem = "IdentityV2"
	head.SourceVersion = "v2"
	head.MinSupportedSourceVersion = "v2"
	head.TransformType = "enrich"
	head.MergeStrategy = "priority"
	head.BreakPolicy = "additive_only"
	head.Confidence = "inferred"
	head.DataClassification = "public"
	head.Owner = "team.platform"
	head.Escalation = "pagerduty:platform"
	head.ContractTestID = "ci://contracts/identity-user-primary-id"
	head.SunsetAt = "2027-01-01"
	head.Flow = "from @IdentityV2 enriched @self"
	head.Note = "migration in progress"
	head.Sources = []SourceRef{
		{
			Kind:           "api",
			Target:         "spotify.GetTrack",
			Path:           "response.track",
			Scope:          "external",
			AsOf:           "2026-02-14",
			ProviderID:     "spotify-enterprise",
			UpstreamSystem: "spotify-v2",
			ContractRef:    "https://example.com/v2",
			Raw:            "api:spotify.GetTrack#response.track@external!2026-02-14?provider_id=spotify-enterprise&contract_ref=https://example.com/v2&upstream_system=spotify-v2",
		},
		{
			Kind:        "event",
			Target:      "catalog.TrackUpdated",
			Path:        "payload.track",
			Scope:       "cross_repo",
			ContractRef: "git+https://github.com/acme/catalog//events.yaml@e1",
			Raw:         "event:catalog.TrackUpdated#payload.track@cross_repo?contract_ref=git+https://github.com/acme/catalog//events.yaml@e1",
		},
	}

	result := DiffArtifacts(
		Artifact{SchemaVersion: "1", Fields: []Annotation{base}},
		Artifact{SchemaVersion: "1", Fields: []Annotation{head}},
	)

	requireChangeSeverity(t, result.Changes, "field_path_changed", SeverityMedium)
	requireChangeSeverity(t, result.Changes, "source_system_changed", SeverityHigh)
	requireChangeSeverity(t, result.Changes, "source_version_changed", SeverityMedium)
	requireChangeSeverity(t, result.Changes, "min_supported_source_version_changed", SeverityHigh)
	requireChangeSeverity(t, result.Changes, "transform_type_changed", SeverityMedium)
	requireChangeSeverity(t, result.Changes, "merge_strategy_changed", SeverityMedium)
	requireChangeSeverity(t, result.Changes, "break_policy_changed", SeverityHigh)
	requireChangeSeverity(t, result.Changes, "confidence_changed", SeverityMedium)
	requireChangeSeverity(t, result.Changes, "classification_relaxed", SeverityHigh)
	requireChangeSeverity(t, result.Changes, "owner_changed", SeverityLow)
	requireChangeSeverity(t, result.Changes, "escalation_changed", SeverityLow)
	requireChangeSeverity(t, result.Changes, "contract_test_id_changed", SeverityMedium)
	requireChangeSeverity(t, result.Changes, "sunset_changed", SeverityMedium)
	requireChangeSeverity(t, result.Changes, "flow_changed", SeverityLow)
	requireChangeSeverity(t, result.Changes, "note_changed", SeverityInfo)
	requireChangeSeverity(t, result.Changes, "source_contract_ref_changed", SeverityMedium)
	requireChangeSeverity(t, result.Changes, "source_provider_changed", SeverityMedium)
	requireChangeSeverity(t, result.Changes, "source_upstream_system_changed", SeverityMedium)
	requireChangeSeverity(t, result.Changes, "external_as_of_advanced", SeverityLow)
	requireChangeSeverity(t, result.Changes, "source_removed", SeverityHigh)
	requireChangeSeverity(t, result.Changes, "source_added", SeverityMedium)
}

func TestDiffArtifacts_ConfidenceAndClassificationTighteningAreLow(t *testing.T) {
	base := mkField("response_user_id")
	base.Confidence = "inferred"
	base.DataClassification = "internal"

	head := base
	head.Confidence = "declared"
	head.DataClassification = "regulated"

	result := DiffArtifacts(
		Artifact{SchemaVersion: "1", Fields: []Annotation{base}},
		Artifact{SchemaVersion: "1", Fields: []Annotation{head}},
	)

	requireChangeSeverity(t, result.Changes, "confidence_changed", SeverityLow)
	requireChangeSeverity(t, result.Changes, "classification_tightened", SeverityLow)
}
