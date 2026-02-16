// generate-usecase-examples.go - Builds cross-domain lineage use-case examples.
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const (
	useCaseDriftBlocking       = "drift_blocking"
	useCaseExternalProvider    = "external_provider_drift"
	useCaseEscalation          = "escalation_chain"
	useCaseCompliance          = "compliance_traceability"
	useCaseMultilang           = "multilang_contract_parity"
	outputCatalogPath          = "tests/lineage/usecases/flows.json"
	outputRegistryPath         = "tests/lineage/usecases/systems.yml"
	outputFixtureDir           = "tests/fixtures/lineage-usecases"
	outputFakeAPIDataDir       = "tests/fake-apis/data"
	annotationSchemaVersion    = "1"
	defaultMinSupportedVersion = "v2026.01"
)

type flowCatalogEntry struct {
	ID          string         `json:"id"`
	Domain      string         `json:"domain"`
	Workflow    string         `json:"workflow"`
	Service     string         `json:"service"`
	Endpoint    string         `json:"endpoint"`
	Description string         `json:"description"`
	UseCases    []string       `json:"use_cases"`
	Languages   []string       `json:"languages"`
	Annotation  annotationSpec `json:"annotation"`
}

type annotationSpec struct {
	FieldID                   string   `json:"field_id"`
	Field                     string   `json:"field"`
	SourceSystem              string   `json:"source_system"`
	SourceVersion             string   `json:"source_version"`
	MinSupportedSourceVersion string   `json:"min_supported_source_version"`
	TransformType             string   `json:"transform_type"`
	MergeStrategy             string   `json:"merge_strategy"`
	BreakPolicy               string   `json:"break_policy"`
	Confidence                string   `json:"confidence"`
	DataClassification        string   `json:"data_classification"`
	Owner                     string   `json:"owner"`
	Escalation                string   `json:"escalation"`
	ContractTestID            string   `json:"contract_test_id"`
	IntroducedAt              string   `json:"introduced_at"`
	Sources                   []string `json:"sources"`
	Flow                      string   `json:"flow"`
	Note                      string   `json:"note"`
}

type fakeAPIRoute struct {
	ID          string   `json:"id"`
	Domain      string   `json:"domain"`
	Workflow    string   `json:"workflow"`
	Service     string   `json:"service"`
	Endpoint    string   `json:"endpoint"`
	FieldID     string   `json:"field_id"`
	UseCases    []string `json:"use_cases"`
	Languages   []string `json:"languages"`
	Description string   `json:"description"`
}

type domainDefinition struct {
	Name        string
	Service     string
	CoreSystem  string
	EventSystem string
	ProviderID  string
	Workflows   []string
}

func main() {
	domains := domainDefinitions()
	flows := buildFlows(domains)

	if err := writeCatalog(flows); err != nil {
		fail(err)
	}
	if err := writeSystemRegistry(domains); err != nil {
		fail(err)
	}
	if err := writeFixtureSources(domains, flows); err != nil {
		fail(err)
	}
	if err := writeFakeAPIData(domains, flows); err != nil {
		fail(err)
	}

	fmt.Printf("Generated %d use-case flows across %d domains.\n", len(flows), len(domains))
}

func fail(err error) {
	fmt.Fprintf(os.Stderr, "generate usecase examples: %v\n", err)
	os.Exit(1)
}

func domainDefinitions() []domainDefinition {
	return []domainDefinition{
		{
			Name:        "logistics",
			Service:     "LogisticsGateway",
			CoreSystem:  "LogisticsCore",
			EventSystem: "LogisticsEvents",
			ProviderID:  "fedex",
			Workflows: []string{
				"shipment_eta_projection",
				"route_risk_assessment",
				"customs_clearance_status",
				"delivery_exception_triage",
				"carrier_capacity_balancing",
				"warehouse_slot_optimization",
				"cold_chain_compliance",
				"proof_of_delivery_digest",
				"reverse_logistics_refund",
				"fleet_maintenance_window",
			},
		},
		{
			Name:        "fintech",
			Service:     "FintechGateway",
			CoreSystem:  "LedgerCore",
			EventSystem: "RiskEvents",
			ProviderID:  "stripe",
			Workflows: []string{
				"payment_authorization_decision",
				"fraud_score_explanation",
				"settlement_reconciliation",
				"chargeback_case_resolution",
				"sanctions_screening_signal",
				"liquidity_buffer_projection",
				"kyc_profile_consistency",
				"invoice_factoring_eligibility",
				"treasury_exposure_rollup",
				"card_token_lifecycle",
			},
		},
		{
			Name:        "media",
			Service:     "MediaGateway",
			CoreSystem:  "MediaCore",
			EventSystem: "AudienceEvents",
			ProviderID:  "spotify",
			Workflows: []string{
				"track_metadata_unification",
				"stream_quality_signal",
				"ad_break_targeting",
				"royalty_statement_projection",
				"creator_payout_rollup",
				"rights_window_enforcement",
				"audience_segment_refresh",
				"recommendation_ranking_trace",
				"moderation_flag_resolution",
				"partner_feed_alignment",
			},
		},
		{
			Name:        "ecommerce",
			Service:     "CommerceGateway",
			CoreSystem:  "CommerceCore",
			EventSystem: "OrderEvents",
			ProviderID:  "shopify",
			Workflows: []string{
				"cart_pricing_waterfall",
				"checkout_risk_gate",
				"inventory_reservation_health",
				"shipping_quote_selection",
				"return_policy_eligibility",
				"catalog_attribute_enrichment",
				"promotion_eligibility_resolution",
				"loyalty_balance_reconciliation",
				"seller_policy_attestation",
				"tax_estimate_trace",
			},
		},
		{
			Name:        "governance",
			Service:     "GovernanceHub",
			CoreSystem:  "GovernanceCore",
			EventSystem: "AuditEvents",
			ProviderID:  "sec",
			Workflows: []string{
				"board_vote_tally",
				"policy_acknowledgement_state",
				"control_attestation_snapshot",
				"audit_finding_status",
				"risk_register_residual_score",
				"procurement_approval_chain",
				"insider_window_enforcement",
				"entity_hierarchy_alignment",
				"disclosure_readiness_signal",
				"whistleblower_case_tracking",
			},
		},
	}
}

func buildFlows(domains []domainDefinition) []flowCatalogEntry {
	flows := make([]flowCatalogEntry, 0, len(domains)*10)

	transformTypes := []string{"normalize", "derive", "aggregate", "join", "passthrough", "mask"}
	mergeStrategies := []string{"priority", "first_non_null", "union", "custom"}
	breakPolicies := []string{"strict", "additive_only", "opaque"}
	confidences := []string{"declared", "inferred"}
	classifications := []string{"internal", "sensitive", "regulated", "public"}

	baseDate := time.Date(2026, time.January, 10, 0, 0, 0, 0, time.UTC)

	for dIndex, domain := range domains {
		for wIndex, workflow := range domain.Workflows {
			slug := sanitizeToken(workflow)
			fieldID := fmt.Sprintf("response_%s_%s", domain.Name, slug)
			fieldPath := fmt.Sprintf("response.%s.%s", domain.Name, slug)
			flowID := fmt.Sprintf("%s_%02d_%s", domain.Name, wIndex+1, slug)
			sourceVersion := fmt.Sprintf("v2026.%02d", (wIndex%12)+1)
			introducedAt := baseDate.AddDate(0, 0, dIndex*12+wIndex).Format("2006-01-02")

			hasExternal := wIndex%2 == 0
			languages := []string{"go", "typescript"}
			if wIndex%3 == 0 {
				languages = append(languages, "python")
			} else {
				languages = append(languages, "java")
			}

			useCases := []string{
				useCaseDriftBlocking,
				useCaseEscalation,
				useCaseCompliance,
				useCaseMultilang,
			}
			if hasExternal {
				useCases = append(useCases, useCaseExternalProvider)
			}

			producerSubsystem := subsystemID(domain.Service, "api-service")
			coreSubsystem := subsystemID(domain.CoreSystem, "api-service")
			eventSubsystem := subsystemID(domain.EventSystem, "stream-service")
			uiSubsystem := subsystemID(domain.Service, "ui-service")
			loggingSubsystem := subsystemID(domain.Service, "logging-sidecar")
			providerSubsystem := subsystemID(domain.ProviderID, "api-service")

			sources := []string{
				fmt.Sprintf(
					"event:%s.Ui%sObserved#payload.%s@internal?contract_ref=internal://events/%s/%s_ui_observed&upstream_system=%s",
					domain.Service,
					camelCase(slug),
					slug,
					domain.Name,
					slug,
					uiSubsystem,
				),
				fmt.Sprintf(
					"api:%s.Get%s#response.%s@cross_repo?contract_ref=git+https://github.com/acme/%s-core//openapi.yaml@r%02d",
					domain.CoreSystem,
					camelCase(slug),
					slug,
					domain.Name,
					wIndex+1,
				),
				fmt.Sprintf(
					"event:%s.%sChanged#payload.%s@internal?contract_ref=internal://events/%s/%s_changed&upstream_system=%s",
					domain.EventSystem,
					camelCase(slug),
					slug,
					domain.Name,
					slug,
					eventSubsystem,
				),
			}
			sources[1] = sources[1] + "&upstream_system=" + coreSubsystem
			if wIndex%3 == 0 {
				sources = append(sources, fmt.Sprintf(
					"event:%s.Trace%sCaptured#payload.%s_trace@internal?contract_ref=internal://events/%s/%s_trace_captured&upstream_system=%s",
					domain.Service,
					camelCase(slug),
					slug,
					domain.Name,
					slug,
					loggingSubsystem,
				))
			}
			if hasExternal {
				asOf := time.Date(2026, time.February, 10+(wIndex%18), 0, 0, 0, 0, time.UTC).Format("2006-01-02")
				sources = append(sources, fmt.Sprintf(
					"api:%s.Get%s#response.%s@external!%s?provider_id=%s&contract_ref=https://api.%s.example.com/%s&upstream_system=%s",
					titleCase(domain.ProviderID),
					camelCase(slug),
					slug,
					asOf,
					domain.ProviderID,
					domain.ProviderID,
					slug,
					providerSubsystem,
				))
			}

			verb := "normalized"
			if hasExternal {
				verb = "mapped"
			}

			flow := flowCatalogEntry{
				ID:          flowID,
				Domain:      domain.Name,
				Workflow:    workflow,
				Service:     domain.Service,
				Endpoint:    fmt.Sprintf("GET /api/v1/%s/%s", domain.Name, strings.ReplaceAll(slug, "_", "-")),
				Description: fmt.Sprintf("%s workflow for %s with contract lineage, drift checks, and escalation context.", titleCase(strings.ReplaceAll(workflow, "_", " ")), domain.Name),
				UseCases:    useCases,
				Languages:   languages,
				Annotation: annotationSpec{
					FieldID:                   fieldID,
					Field:                     fieldPath,
					SourceSystem:              producerSubsystem,
					SourceVersion:             sourceVersion,
					MinSupportedSourceVersion: defaultMinSupportedVersion,
					TransformType:             transformTypes[(dIndex+wIndex)%len(transformTypes)],
					MergeStrategy:             mergeStrategies[(dIndex+wIndex)%len(mergeStrategies)],
					BreakPolicy:               breakPolicies[(dIndex+wIndex)%len(breakPolicies)],
					Confidence:                confidences[(dIndex+wIndex)%len(confidences)],
					DataClassification:        classifications[(dIndex+wIndex)%len(classifications)],
					Owner:                     fmt.Sprintf("team.%s", domain.Name),
					Escalation:                fmt.Sprintf("pagerduty:%s-oncall", domain.Name),
					ContractTestID:            fmt.Sprintf("ci://contracts/%s/%s", domain.Name, slug),
					IntroducedAt:              introducedAt,
					Sources:                   sources,
					Flow:                      fmt.Sprintf("from @%s %s @self", coreSubsystem, verb),
					Note: fmt.Sprintf(
						"Combines %s flow inputs with governed schema checks; reference=https://specs.example.com/%s/%s",
						domain.Name,
						domain.Name,
						slug,
					),
				},
			}
			flows = append(flows, flow)
		}
	}

	sort.Slice(flows, func(i, j int) bool {
		return flows[i].ID < flows[j].ID
	})
	return flows
}

func writeCatalog(flows []flowCatalogEntry) error {
	if err := os.MkdirAll(filepath.Dir(outputCatalogPath), 0o755); err != nil {
		return fmt.Errorf("mkdir catalog dir: %w", err)
	}

	data, err := json.MarshalIndent(flows, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal catalog: %w", err)
	}

	return os.WriteFile(outputCatalogPath, append(data, '\n'), 0o644)
}

func writeSystemRegistry(domains []domainDefinition) error {
	if err := os.MkdirAll(filepath.Dir(outputRegistryPath), 0o755); err != nil {
		return fmt.Errorf("mkdir registry dir: %w", err)
	}

	var b strings.Builder
	b.WriteString("systems:\n")

	for _, domain := range domains {
		writeSystemRegistryEntry(&b, domain.Service, domain.Service+" Service", "team."+domain.Name, []registryContact{
			{Role: "primary", Name: titleCase(domain.Name) + " Oncall", Channel: "pagerduty:" + domain.Name + "-oncall"},
			{Role: "backup", Name: titleCase(domain.Name) + " Platform", Channel: "slack:#" + domain.Name + "-platform"},
		})
		writeSystemRegistryEntry(&b, subsystemID(domain.Service, "api-service"), domain.Service+" API Service", "team."+domain.Name, []registryContact{
			{Role: "primary", Name: titleCase(domain.Name) + " API Oncall", Channel: "pagerduty:" + domain.Name + "-api"},
		})
		writeSystemRegistryEntry(&b, subsystemID(domain.Service, "ui-service"), domain.Service+" UI Service", "team."+domain.Name, []registryContact{
			{Role: "primary", Name: titleCase(domain.Name) + " UI Oncall", Channel: "slack:#" + domain.Name + "-ui-oncall"},
		})
		writeSystemRegistryEntry(&b, subsystemID(domain.Service, "logging-sidecar"), domain.Service+" Logging Sidecar", "team."+domain.Name, []registryContact{
			{Role: "primary", Name: titleCase(domain.Name) + " Observability", Channel: "slack:#" + domain.Name + "-observability"},
		})
	}

	for _, domain := range domains {
		writeSystemRegistryEntry(&b, domain.CoreSystem, domain.CoreSystem+" Core", "team."+domain.Name+"-core", []registryContact{
			{Role: "primary", Name: titleCase(domain.Name) + " Core Oncall", Channel: "pagerduty:" + domain.Name + "-core"},
		})
		writeSystemRegistryEntry(&b, subsystemID(domain.CoreSystem, "api-service"), domain.CoreSystem+" API Service", "team."+domain.Name+"-core", []registryContact{
			{Role: "primary", Name: titleCase(domain.Name) + " Core API", Channel: "pagerduty:" + domain.Name + "-core-api"},
		})
	}

	for _, domain := range domains {
		writeSystemRegistryEntry(&b, domain.EventSystem, domain.EventSystem+" Stream", "team."+domain.Name+"-events", []registryContact{
			{Role: "primary", Name: titleCase(domain.Name) + " Events Oncall", Channel: "slack:#" + domain.Name + "-events"},
		})
		writeSystemRegistryEntry(&b, subsystemID(domain.EventSystem, "stream-service"), domain.EventSystem+" Stream Service", "team."+domain.Name+"-events", []registryContact{
			{Role: "primary", Name: titleCase(domain.Name) + " Stream Oncall", Channel: "slack:#" + domain.Name + "-stream"},
		})
	}

	seenProviders := map[string]bool{}
	for _, domain := range domains {
		provider := domain.ProviderID
		if seenProviders[provider] {
			continue
		}
		seenProviders[provider] = true
		writeSystemRegistryEntry(&b, provider, titleCase(provider)+" External Provider", "team.external-integrations", []registryContact{
			{Role: "provider", Name: titleCase(provider) + " Provider Ops", Channel: "slack:#provider-" + provider},
		})
		writeSystemRegistryEntry(&b, subsystemID(provider, "api-service"), titleCase(provider)+" Provider API", "team.external-integrations", []registryContact{
			{Role: "provider", Name: titleCase(provider) + " API Ops", Channel: "slack:#provider-" + provider + "-api"},
		})
	}

	return os.WriteFile(outputRegistryPath, []byte(b.String()), 0o644)
}

type registryContact struct {
	Role    string
	Name    string
	Channel string
}

func writeSystemRegistryEntry(b *strings.Builder, id string, name string, ownerTeam string, contacts []registryContact) {
	fmt.Fprintf(b, "  - id: %s\n", id)
	fmt.Fprintf(b, "    name: %s\n", name)
	fmt.Fprintf(b, "    owner_team: %s\n", ownerTeam)
	b.WriteString("    escalation:\n")
	for _, contact := range contacts {
		fmt.Fprintf(b, "      - role: %s\n", contact.Role)
		fmt.Fprintf(b, "        name: %s\n", contact.Name)
		fmt.Fprintf(b, "        channel: %s\n", contact.Channel)
	}
}

func writeFixtureSources(domains []domainDefinition, flows []flowCatalogEntry) error {
	if err := os.RemoveAll(outputFixtureDir); err != nil {
		return fmt.Errorf("clear fixture dir: %w", err)
	}
	if err := os.MkdirAll(outputFixtureDir, 0o755); err != nil {
		return fmt.Errorf("mkdir fixture dir: %w", err)
	}

	flowsByDomain := map[string][]flowCatalogEntry{}
	for _, flow := range flows {
		flowsByDomain[flow.Domain] = append(flowsByDomain[flow.Domain], flow)
	}

	for _, domain := range domains {
		domainFlows := flowsByDomain[domain.Name]
		sort.Slice(domainFlows, func(i, j int) bool { return domainFlows[i].ID < domainFlows[j].ID })

		path := filepath.Join(outputFixtureDir, fmt.Sprintf("%s_flows.go", domain.Name))
		var b strings.Builder
		b.WriteString("// Code generated by scripts/generate-usecase-examples.go; DO NOT EDIT.\n")
		b.WriteString("// This file contains lineage annotation fixtures for combined real-world use-cases.\n")
		b.WriteString("package usecases\n\n")

		for _, flow := range domainFlows {
			annotation := formatAnnotationLine(flow)
			b.WriteString(annotation)
			b.WriteString("\n")
		}

		if err := os.WriteFile(path, []byte(b.String()), 0o644); err != nil {
			return fmt.Errorf("write fixture file %s: %w", path, err)
		}
	}

	return nil
}

func writeFakeAPIData(domains []domainDefinition, flows []flowCatalogEntry) error {
	if err := os.RemoveAll(outputFakeAPIDataDir); err != nil {
		return fmt.Errorf("clear fake api data dir: %w", err)
	}
	if err := os.MkdirAll(outputFakeAPIDataDir, 0o755); err != nil {
		return fmt.Errorf("mkdir fake api data dir: %w", err)
	}

	flowsByDomain := map[string][]fakeAPIRoute{}
	for _, flow := range flows {
		flowsByDomain[flow.Domain] = append(flowsByDomain[flow.Domain], fakeAPIRoute{
			ID:          flow.ID,
			Domain:      flow.Domain,
			Workflow:    flow.Workflow,
			Service:     flow.Service,
			Endpoint:    flow.Endpoint,
			FieldID:     flow.Annotation.FieldID,
			UseCases:    append([]string{}, flow.UseCases...),
			Languages:   append([]string{}, flow.Languages...),
			Description: flow.Description,
		})
	}

	for _, domain := range domains {
		path := filepath.Join(outputFakeAPIDataDir, fmt.Sprintf("%s.json", domain.Name))
		payload := flowsByDomain[domain.Name]
		sort.Slice(payload, func(i, j int) bool { return payload[i].ID < payload[j].ID })
		data, err := json.MarshalIndent(payload, "", "  ")
		if err != nil {
			return fmt.Errorf("marshal fake api data for %s: %w", domain.Name, err)
		}
		if err := os.WriteFile(path, append(data, '\n'), 0o644); err != nil {
			return fmt.Errorf("write fake api data %s: %w", path, err)
		}
	}

	return nil
}

func formatAnnotationLine(flow flowCatalogEntry) string {
	sources := strings.Join(flow.Annotation.Sources, ",")
	return fmt.Sprintf(
		"// strict-source annotation_schema_version=%s field_id=%s field=%s source_system=%s source_version=%s min_supported_source_version=%s transform_type=%s merge_strategy=%s break_policy=%s confidence=%s data_classification=%s owner=%s escalation=%s contract_test_id=%s introduced_at=%s sources=%s flow=%q note=%q",
		annotationSchemaVersion,
		flow.Annotation.FieldID,
		flow.Annotation.Field,
		flow.Annotation.SourceSystem,
		flow.Annotation.SourceVersion,
		flow.Annotation.MinSupportedSourceVersion,
		flow.Annotation.TransformType,
		flow.Annotation.MergeStrategy,
		flow.Annotation.BreakPolicy,
		flow.Annotation.Confidence,
		flow.Annotation.DataClassification,
		flow.Annotation.Owner,
		flow.Annotation.Escalation,
		flow.Annotation.ContractTestID,
		flow.Annotation.IntroducedAt,
		sources,
		flow.Annotation.Flow,
		flow.Annotation.Note,
	)
}

func sanitizeToken(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	value = strings.ReplaceAll(value, "-", "_")
	value = strings.ReplaceAll(value, " ", "_")

	var b strings.Builder
	for _, ch := range value {
		if (ch >= 'a' && ch <= 'z') || (ch >= '0' && ch <= '9') || ch == '_' {
			b.WriteRune(ch)
		}
	}

	token := strings.Trim(b.String(), "_")
	if token == "" {
		return "value"
	}
	return token
}

func camelCase(snake string) string {
	parts := strings.Split(snake, "_")
	for i, part := range parts {
		if part == "" {
			continue
		}
		parts[i] = titleCase(part)
	}
	return strings.Join(parts, "")
}

func subsystemID(system string, subsystem string) string {
	return fmt.Sprintf("%s:%s", system, subsystem)
}

func titleCase(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}

	var out strings.Builder
	words := strings.FieldsFunc(raw, func(r rune) bool { return r == '_' || r == '-' || r == ' ' })
	for i, word := range words {
		if word == "" {
			continue
		}
		if i > 0 {
			out.WriteRune(' ')
		}
		word = strings.ToLower(word)
		out.WriteString(strings.ToUpper(word[:1]))
		if len(word) > 1 {
			out.WriteString(word[1:])
		}
	}
	return out.String()
}
