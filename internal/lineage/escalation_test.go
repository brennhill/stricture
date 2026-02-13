// escalation_test.go - Tests for lineage escalation-chain resolution.
package lineage

import "testing"

func TestBuildEscalationChain_WorksBackwards(t *testing.T) {
	artifact := Artifact{SchemaVersion: "1", Fields: []Annotation{
		{
			FieldID:      "y_user_id",
			SourceSystem: "ServiceY",
			Owner:        "team.servicey",
			Escalation:   "slack:#servicey-oncall",
			Sources: []SourceRef{{
				Kind:           "api",
				Target:         "ServiceX.GetUser",
				Path:           "response.id",
				Scope:          "cross_repo",
				ContractRef:    "git+https://github.com/acme/servicex//openapi.yaml@abc",
				UpstreamSystem: "ServiceX",
			}},
		},
		{
			FieldID:      "x_song",
			SourceSystem: "ServiceX",
			Owner:        "team.servicex",
			Escalation:   "slack:#servicex-oncall",
			Sources: []SourceRef{{
				Kind:        "api",
				Target:      "spotify.GetTrack",
				Path:        "response.track",
				Scope:       "external",
				AsOf:        "2026-02-13",
				ProviderID:  "spotify",
				ContractRef: "https://developer.spotify.com/reference/get-track",
			}},
		},
	}}

	registry := SystemRegistry{Systems: []SystemMetadata{
		{
			ID:        "servicey",
			Name:      "Service Y",
			OwnerTeam: "team.servicey",
			Escalation: []Contact{
				{Role: "primary", Name: "Y Oncall", Channel: "pagerduty:servicey"},
			},
		},
		{
			ID:        "servicex",
			Name:      "Service X",
			OwnerTeam: "team.servicex",
			Escalation: []Contact{
				{Role: "primary", Name: "X Oncall", Channel: "pagerduty:servicex"},
			},
		},
		{
			ID:        "spotify",
			Name:      "Spotify API",
			OwnerTeam: "vendor.spotify",
			Escalation: []Contact{
				{Role: "vendor", Name: "Spotify Support", Channel: "https://developer.spotify.com/contact"},
			},
		},
	}}

	steps, err := BuildEscalationChain("ServiceY", artifact, registry, 5)
	if err != nil {
		t.Fatalf("build chain error: %v", err)
	}
	if len(steps) != 3 {
		t.Fatalf("steps len = %d, want 3", len(steps))
	}
	if steps[0].SystemID != "servicey" || steps[1].SystemID != "servicex" || steps[2].SystemID != "spotify" {
		t.Fatalf("unexpected step order: %+v", steps)
	}
	if len(steps[0].Contacts) == 0 || len(steps[1].Contacts) == 0 || len(steps[2].Contacts) == 0 {
		t.Fatalf("expected contacts at every step")
	}
}

func TestBuildEscalationChain_UsesFallbackWhenRegistryMissing(t *testing.T) {
	artifact := Artifact{SchemaVersion: "1", Fields: []Annotation{{
		FieldID:      "a",
		SourceSystem: "Billing",
		Owner:        "team.billing",
		Escalation:   "slack:#billing-oncall",
		Sources:      []SourceRef{},
	}}}

	steps, err := BuildEscalationChain("Billing", artifact, SystemRegistry{}, 3)
	if err != nil {
		t.Fatalf("build chain error: %v", err)
	}
	if len(steps) != 1 {
		t.Fatalf("steps len = %d, want 1", len(steps))
	}
	if len(steps[0].Contacts) == 0 {
		t.Fatalf("expected fallback contacts")
	}
}
