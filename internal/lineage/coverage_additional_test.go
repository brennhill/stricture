package lineage

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseErrorErrorFormatting(t *testing.T) {
	errNoFile := ParseError{Line: 7, Message: "bad annotation"}
	if got := errNoFile.Error(); got != "line 7: bad annotation" {
		t.Fatalf("error without file = %q", got)
	}

	errWithFile := ParseError{FilePath: "svc/user.go", Line: 9, Message: "bad annotation"}
	if got := errWithFile.Error(); got != "svc/user.go:9: bad annotation" {
		t.Fatalf("error with file = %q", got)
	}
}

func TestCommentTextVariants(t *testing.T) {
	cases := []struct {
		line string
		want string
		ok   bool
	}{
		{line: "// hello", want: "hello", ok: true},
		{line: "# hello", want: "hello", ok: true},
		{line: "/* hello */", want: "hello", ok: true},
		{line: "* hello", want: "hello", ok: true},
		{line: "let x = 1;", want: "", ok: false},
	}

	for _, tc := range cases {
		got, ok := commentText(tc.line)
		if ok != tc.ok {
			t.Fatalf("commentText(%q) ok=%v, want %v", tc.line, ok, tc.ok)
		}
		if got != tc.want {
			t.Fatalf("commentText(%q)=%q, want %q", tc.line, got, tc.want)
		}
	}
}

func TestLoadAndWriteArtifactRoundTrip(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "lineage", "artifact.json")

	artifact := Artifact{
		Fields: []Annotation{
			{
				AnnotationSchemaVersion: "1",
				FieldID:                 "response_user_id",
				Field:                   "response.user_id",
				SourceSystem:            "Identity",
				SourceVersion:           "v1",
				Line:                    3,
			},
		},
	}

	if err := WriteArtifact(path, artifact); err != nil {
		t.Fatalf("write artifact: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read written artifact: %v", err)
	}
	if !strings.HasSuffix(string(data), "\n") {
		t.Fatalf("expected trailing newline in written artifact")
	}

	loaded, err := LoadArtifact(path)
	if err != nil {
		t.Fatalf("load artifact: %v", err)
	}
	if loaded.SchemaVersion != "1" {
		t.Fatalf("schema_version = %q, want 1", loaded.SchemaVersion)
	}
	if len(loaded.Fields) != 1 || loaded.Fields[0].FieldID != "response_user_id" {
		t.Fatalf("unexpected loaded fields: %+v", loaded.Fields)
	}
}

func TestLoadArtifactDefaultsMissingSchemaVersion(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "artifact.json")
	content := `{"fields":[{"field_id":"x"}]}`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write artifact: %v", err)
	}

	loaded, err := LoadArtifact(path)
	if err != nil {
		t.Fatalf("load artifact: %v", err)
	}
	if loaded.SchemaVersion != "1" {
		t.Fatalf("schema_version = %q, want 1", loaded.SchemaVersion)
	}
}

func TestLoadArtifactInvalidJSON(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "artifact.json")
	if err := os.WriteFile(path, []byte("{"), 0o644); err != nil {
		t.Fatalf("write artifact: %v", err)
	}
	if _, err := LoadArtifact(path); err == nil {
		t.Fatalf("expected parse error")
	}
}

func TestCollectMissingPathReturnsError(t *testing.T) {
	if _, _, err := Collect([]string{"/definitely/not/found/lineage"}); err == nil {
		t.Fatalf("expected stat error")
	}
}

func TestParseSeverityValues(t *testing.T) {
	cases := []struct {
		raw     string
		want    Severity
		wantErr bool
	}{
		{raw: "high", want: SeverityHigh},
		{raw: "medium", want: SeverityMedium},
		{raw: "low", want: SeverityLow},
		{raw: "info", want: SeverityInfo},
		{raw: "none", want: Severity("none")},
		{raw: "invalid", wantErr: true},
	}

	for _, tc := range cases {
		got, err := ParseSeverity(tc.raw)
		if tc.wantErr {
			if err == nil {
				t.Fatalf("ParseSeverity(%q) expected error", tc.raw)
			}
			continue
		}
		if err != nil {
			t.Fatalf("ParseSeverity(%q) unexpected error: %v", tc.raw, err)
		}
		if got != tc.want {
			t.Fatalf("ParseSeverity(%q)=%q, want %q", tc.raw, got, tc.want)
		}
	}
}

func TestCompareSourcesInvalidAsOfFallsBackToMedium(t *testing.T) {
	base := []SourceRef{{
		Kind:        "api",
		Target:      "spotify.GetTrack",
		Path:        "response.track",
		Scope:       "external",
		AsOf:        "bad-date",
		ProviderID:  "spotify",
		ContractRef: "https://example.com",
		Raw:         "base",
	}}
	head := []SourceRef{{
		Kind:        "api",
		Target:      "spotify.GetTrack",
		Path:        "response.track",
		Scope:       "external",
		AsOf:        "2026-02-13",
		ProviderID:  "spotify",
		ContractRef: "https://example.com",
		Raw:         "head",
	}}

	changes := compareSources("field_id", "consumer", "", base, head)
	if len(changes) != 1 {
		t.Fatalf("changes len = %d, want 1", len(changes))
	}
	if changes[0].ChangeType != "external_as_of_changed" || changes[0].Severity != SeverityMedium {
		t.Fatalf("unexpected change: %+v", changes[0])
	}
}

func TestMatchOverrideWildcard(t *testing.T) {
	change := DriftChange{
		FieldID:    "response_user_id",
		ChangeType: "field_removed",
	}
	overrides := []Override{
		{FieldID: "response_user_id", ChangeType: "*", Expires: "2099-12-31", Reason: "wildcard"},
	}
	match := matchOverride(change, overrides)
	if match == nil {
		t.Fatalf("expected wildcard override to match")
	}
}

func TestLoadSystemRegistryValidation(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "systems.yml")
	valid := `systems:
  - id: ServiceA
    name: Service A
    owner_team: team.a
    runbook_url: https://runbooks.example.com/service-a
    doc_root: https://docs.example.com/service-a
    escalation:
      - role: primary
        name: A Oncall
        channel: pagerduty:a
`
	if err := os.WriteFile(path, []byte(valid), 0o644); err != nil {
		t.Fatalf("write registry: %v", err)
	}
	registry, err := LoadSystemRegistry(path)
	if err != nil {
		t.Fatalf("load valid registry: %v", err)
	}
	if len(registry.Systems) != 1 {
		t.Fatalf("systems len = %d, want 1", len(registry.Systems))
	}
	if registry.Systems[0].RunbookURL != "https://runbooks.example.com/service-a" {
		t.Fatalf("runbook_url = %q, want https://runbooks.example.com/service-a", registry.Systems[0].RunbookURL)
	}
	if registry.Systems[0].DocRoot != "https://docs.example.com/service-a" {
		t.Fatalf("doc_root = %q, want https://docs.example.com/service-a", registry.Systems[0].DocRoot)
	}

	dup := `systems:
  - id: ServiceA
  - id: servicea
`
	if err := os.WriteFile(path, []byte(dup), 0o644); err != nil {
		t.Fatalf("write registry: %v", err)
	}
	if _, err := LoadSystemRegistry(path); err == nil {
		t.Fatalf("expected duplicate id error")
	}

	empty := `systems:
  - id: "  "
`
	if err := os.WriteFile(path, []byte(empty), 0o644); err != nil {
		t.Fatalf("write registry: %v", err)
	}
	if _, err := LoadSystemRegistry(path); err == nil {
		t.Fatalf("expected empty id error")
	}
}

func TestDeriveUpstreamSystemVariants(t *testing.T) {
	if got := deriveUpstreamSystem(SourceRef{UpstreamSystem: "ServiceX"}); got != "servicex" {
		t.Fatalf("upstream_system variant = %q, want servicex", got)
	}
	if got := deriveUpstreamSystem(SourceRef{Scope: "external", ProviderID: "spotify"}); got != "spotify" {
		t.Fatalf("external provider variant = %q, want spotify", got)
	}
	if got := deriveUpstreamSystem(SourceRef{Kind: "api", Target: "Identity.GetUser"}); got != "identity" {
		t.Fatalf("api dot target variant = %q, want identity", got)
	}
	if got := deriveUpstreamSystem(SourceRef{Kind: "event", Target: "Billing/InvoiceCreated"}); got != "billing" {
		t.Fatalf("event slash target variant = %q, want billing", got)
	}
	if got := deriveUpstreamSystem(SourceRef{Kind: "api", Target: "simple"}); got != "simple" {
		t.Fatalf("api simple target variant = %q, want simple", got)
	}
	if got := deriveUpstreamSystem(SourceRef{Kind: "db", Target: "users"}); got != "" {
		t.Fatalf("non api/event kind variant = %q, want empty", got)
	}
}

func TestRankHelpersCoverAllBranches(t *testing.T) {
	if severityRank(SeverityHigh) != 4 ||
		severityRank(SeverityMedium) != 3 ||
		severityRank(SeverityLow) != 2 ||
		severityRank(SeverityInfo) != 1 ||
		severityRank(Severity("unknown")) != 0 {
		t.Fatalf("unexpected severity ranks")
	}

	if classificationRank("public") != 1 ||
		classificationRank("internal") != 2 ||
		classificationRank("sensitive") != 3 ||
		classificationRank("regulated") != 4 ||
		classificationRank("unknown") != 0 {
		t.Fatalf("unexpected classification ranks")
	}
}

func TestDiffArtifactsReportsFieldAdded(t *testing.T) {
	base := Artifact{SchemaVersion: "1", Fields: []Annotation{}}
	head := Artifact{SchemaVersion: "1", Fields: []Annotation{mkField("response_user_id")}}

	result := DiffArtifacts(base, head)
	if len(result.Changes) != 1 {
		t.Fatalf("changes len = %d, want 1", len(result.Changes))
	}
	if result.Changes[0].ChangeType != "field_added" {
		t.Fatalf("change type = %q, want field_added", result.Changes[0].ChangeType)
	}
}
