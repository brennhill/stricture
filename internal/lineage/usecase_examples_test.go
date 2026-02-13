package lineage

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

const (
	useCaseCatalogPath  = "../../tests/lineage/usecases/flows.json"
	useCaseFixturePath  = "../../tests/fixtures/lineage-usecases"
	useCaseSystemsPath  = "../../tests/lineage/usecases/systems.yml"
	useCaseDrift        = "drift_blocking"
	useCaseExternal     = "external_provider_drift"
	useCaseEscalation   = "escalation_chain"
	useCaseCompliance   = "compliance_traceability"
	useCaseMultilang    = "multilang_contract_parity"
	minUseCaseFlowCount = 50
)

type useCaseFlow struct {
	ID         string   `json:"id"`
	Domain     string   `json:"domain"`
	Service    string   `json:"service"`
	UseCases   []string `json:"use_cases"`
	Languages  []string `json:"languages"`
	Annotation struct {
		FieldID string   `json:"field_id"`
		Sources []string `json:"sources"`
	} `json:"annotation"`
}

func TestUseCaseCatalog_HasRequiredBreadthAndCombination(t *testing.T) {
	flows := loadUseCaseCatalog(t)
	if len(flows) < minUseCaseFlowCount {
		t.Fatalf("flow count = %d, want at least %d", len(flows), minUseCaseFlowCount)
	}

	expectedDomains := map[string]bool{
		"logistics":  true,
		"fintech":    true,
		"media":      true,
		"ecommerce":  true,
		"governance": true,
	}
	domainCounts := map[string]int{}
	externalByDomain := map[string]int{}

	for _, flow := range flows {
		domainCounts[flow.Domain]++

		if len(flow.UseCases) < 4 {
			t.Fatalf("flow %s has %d use cases, want >= 4", flow.ID, len(flow.UseCases))
		}
		if !contains(flow.UseCases, useCaseDrift) {
			t.Fatalf("flow %s missing use case %s", flow.ID, useCaseDrift)
		}
		if !contains(flow.UseCases, useCaseEscalation) {
			t.Fatalf("flow %s missing use case %s", flow.ID, useCaseEscalation)
		}
		if !contains(flow.UseCases, useCaseCompliance) {
			t.Fatalf("flow %s missing use case %s", flow.ID, useCaseCompliance)
		}
		if !contains(flow.UseCases, useCaseMultilang) {
			t.Fatalf("flow %s missing use case %s", flow.ID, useCaseMultilang)
		}

		if len(flow.Languages) < 2 {
			t.Fatalf("flow %s has %d languages, want >= 2", flow.ID, len(flow.Languages))
		}
		if !contains(flow.Languages, "go") {
			t.Fatalf("flow %s missing go in languages", flow.ID)
		}

		hasExternal := contains(flow.UseCases, useCaseExternal)
		if hasExternal {
			externalByDomain[flow.Domain]++
		}
	}

	if len(domainCounts) != len(expectedDomains) {
		t.Fatalf("domain coverage = %d domains, want %d", len(domainCounts), len(expectedDomains))
	}
	for domain := range expectedDomains {
		if domainCounts[domain] < 10 {
			t.Fatalf("domain %s has %d flows, want >= 10", domain, domainCounts[domain])
		}
		if externalByDomain[domain] < 1 {
			t.Fatalf("domain %s has no external-provider use case flow", domain)
		}
	}
}

func TestUseCaseFixtures_CollectsAllCatalogFieldIDs(t *testing.T) {
	flows := loadUseCaseCatalog(t)
	expectedFieldIDs := map[string]bool{}
	for _, flow := range flows {
		expectedFieldIDs[flow.Annotation.FieldID] = true
	}

	artifact, parseErrs, err := Collect([]string{fixturePath(t, useCaseFixturePath)})
	if err != nil {
		t.Fatalf("collect fixture artifact: %v", err)
	}
	if len(parseErrs) != 0 {
		t.Fatalf("parseErrs len = %d, want 0 (first: %+v)", len(parseErrs), parseErrs[0])
	}
	if len(artifact.Fields) != len(expectedFieldIDs) {
		t.Fatalf("artifact fields = %d, want %d", len(artifact.Fields), len(expectedFieldIDs))
	}

	for _, field := range artifact.Fields {
		if !expectedFieldIDs[field.FieldID] {
			t.Fatalf("unexpected field_id in artifact: %s", field.FieldID)
		}
	}
}

func TestUseCaseFakeAPIData_AlignedWithCatalog(t *testing.T) {
	flows := loadUseCaseCatalog(t)
	catalogByDomain := map[string]map[string]bool{}
	for _, flow := range flows {
		if _, ok := catalogByDomain[flow.Domain]; !ok {
			catalogByDomain[flow.Domain] = map[string]bool{}
		}
		catalogByDomain[flow.Domain][flow.ID] = true
	}

	for domain, ids := range catalogByDomain {
		path := fixturePath(t, filepath.Join("../../tests/fake-apis/data", domain+".json"))
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read fake api data %s: %v", path, err)
		}

		var fakeFlows []struct {
			ID string `json:"id"`
		}
		if err := json.Unmarshal(data, &fakeFlows); err != nil {
			t.Fatalf("parse fake api data %s: %v", path, err)
		}
		if len(fakeFlows) != 10 {
			t.Fatalf("fake api data %s has %d flows, want 10", path, len(fakeFlows))
		}
		for _, fakeFlow := range fakeFlows {
			if !ids[fakeFlow.ID] {
				t.Fatalf("fake api flow id %s not found in catalog domain %s", fakeFlow.ID, domain)
			}
		}
	}
}

func TestUseCaseExamples_DriftExternalAndEscalationBehavior(t *testing.T) {
	base, parseErrs, err := Collect([]string{fixturePath(t, useCaseFixturePath)})
	if err != nil {
		t.Fatalf("collect fixture artifact: %v", err)
	}
	if len(parseErrs) != 0 {
		t.Fatalf("parseErrs len = %d, want 0", len(parseErrs))
	}

	registry, err := LoadSystemRegistry(fixturePath(t, useCaseSystemsPath))
	if err != nil {
		t.Fatalf("load systems registry: %v", err)
	}

	chain, err := BuildEscalationChain("LogisticsGateway", base, registry, 4)
	if err != nil {
		t.Fatalf("build escalation chain: %v", err)
	}
	if len(chain) < 2 {
		t.Fatalf("escalation chain length = %d, want >= 2", len(chain))
	}
	if chain[0].SystemID != "logisticsgateway" {
		t.Fatalf("chain[0].system_id = %s, want logisticsgateway", chain[0].SystemID)
	}
	if len(chain[0].Contacts) == 0 {
		t.Fatalf("chain[0] expected at least one contact")
	}

	removedFieldID := base.Fields[0].FieldID
	head := cloneArtifact(t, base)
	head.Fields = append([]Annotation{}, head.Fields[1:]...)
	head.Fields[0].SourceVersion = "v2099.01"

	result := DiffArtifacts(base, head)
	if !ShouldFailAtThreshold(result, SeverityHigh) {
		t.Fatalf("expected high-severity failure without override")
	}

	head.Overrides = []Override{
		{
			FieldID:    removedFieldID,
			ChangeType: "field_removed",
			Expires:    time.Now().UTC().AddDate(0, 1, 0).Format("2006-01-02"),
			Reason:     "temporary migration",
			Ticket:     "INC-9001",
		},
	}
	overrideResult := DiffArtifacts(base, head)
	if ShouldFailAtThreshold(overrideResult, SeverityHigh) {
		t.Fatalf("expected high threshold to pass with active field_removed override")
	}
	if !ShouldFailAtThreshold(overrideResult, SeverityMedium) {
		t.Fatalf("expected medium threshold failure from source_version drift")
	}

	externalFieldIndex := -1
	externalSourceIndex := -1
	for i, field := range base.Fields {
		for j, source := range field.Sources {
			if source.Scope == "external" {
				externalFieldIndex = i
				externalSourceIndex = j
				break
			}
		}
		if externalFieldIndex >= 0 {
			break
		}
	}
	if externalFieldIndex < 0 {
		t.Fatalf("expected at least one external source in base artifact")
	}

	headExternal := cloneArtifact(t, base)
	headExternal.Fields[externalFieldIndex].Sources[externalSourceIndex].AsOf = "2026-01-01"
	resultExternal := DiffArtifacts(base, headExternal)
	if !hasChange(resultExternal, "external_as_of_rollback", SeverityHigh) {
		t.Fatalf("expected external_as_of_rollback high-severity change")
	}
}

func loadUseCaseCatalog(t *testing.T) []useCaseFlow {
	t.Helper()

	data, err := os.ReadFile(fixturePath(t, useCaseCatalogPath))
	if err != nil {
		t.Fatalf("read use-case catalog: %v", err)
	}

	var flows []useCaseFlow
	if err := json.Unmarshal(data, &flows); err != nil {
		t.Fatalf("parse use-case catalog: %v", err)
	}
	return flows
}

func fixturePath(t *testing.T, rel string) string {
	t.Helper()
	return filepath.Clean(rel)
}

func contains(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func hasChange(result DiffResult, changeType string, severity Severity) bool {
	for _, change := range result.Changes {
		if change.ChangeType == changeType && change.Severity == severity {
			return true
		}
	}
	return false
}

func cloneArtifact(t *testing.T, artifact Artifact) Artifact {
	t.Helper()

	data, err := json.Marshal(artifact)
	if err != nil {
		t.Fatalf("marshal artifact clone: %v", err)
	}
	var clone Artifact
	if err := json.Unmarshal(data, &clone); err != nil {
		t.Fatalf("unmarshal artifact clone: %v", err)
	}
	return clone
}
