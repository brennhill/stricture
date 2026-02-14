package server

import "testing"

func TestNormalizeAndValidateIngestSanitizesAndGeneratesRunID(t *testing.T) {
	in := ArtifactIngestRequest{
		Organization: "  ACME, Inc  ",
		Project:      " Checkout/API ",
		Service:      " Gateway V2 ",
		Artifact:     []byte(`{"field":"status"}`),
	}

	got, err := normalizeAndValidateIngest(in)
	if err != nil {
		t.Fatalf("normalizeAndValidateIngest() error = %v", err)
	}
	if got.Organization != "acme-inc" {
		t.Fatalf("expected sanitized organization, got %q", got.Organization)
	}
	if got.Project != "checkout-api" {
		t.Fatalf("expected sanitized project, got %q", got.Project)
	}
	if got.Service != "gateway-v2" {
		t.Fatalf("expected sanitized service, got %q", got.Service)
	}
	if got.RunID == "" {
		t.Fatal("expected generated run id")
	}
}

func TestNormalizeAndValidateIngestRequiresFields(t *testing.T) {
	_, err := normalizeAndValidateIngest(ArtifactIngestRequest{})
	if err == nil {
		t.Fatal("expected validation error")
	}
}

func TestNormalizeAndValidateIngestGeneratedAtMustBeRFC3339(t *testing.T) {
	_, err := normalizeAndValidateIngest(ArtifactIngestRequest{
		Organization: "acme",
		Project:      "checkout",
		Service:      "gateway",
		GeneratedAt:  "2026/01/01",
		Artifact:     []byte(`{"field":"status"}`),
	})
	if err == nil {
		t.Fatal("expected generated_at validation error")
	}
}
