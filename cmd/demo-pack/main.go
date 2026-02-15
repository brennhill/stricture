// demo-pack - Build artifact-driven demo data for Cloudflare worker/UI.
package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/stricture/stricture/internal/lineage"
)

type demoPack struct {
	GeneratedAt        string                              `json:"generated_at"`
	Truth              demoTruth                           `json:"truth"`
	Services           []demoService                       `json:"services"`
	Edges              []demoEdge                          `json:"edges"`
	FieldMetadata      map[string]demoFieldMetadata        `json:"field_metadata"`
	MutationTypes      []string                            `json:"mutation_types"`
	FieldsByMutation   map[string][]string                 `json:"fields_by_mutation"`
	MutationScenarios  map[string]map[string]demoScenario  `json:"mutation_scenarios"`
	EscalationBySystem map[string][]lineage.EscalationStep `json:"escalation_by_system"`
	BaselineSummary    lineage.DiffSummary                 `json:"baseline_summary"`
}

type demoTruth struct {
	SupportedFlows        int     `json:"supportedFlows"`
	AnnotatedFlows        int     `json:"annotatedFlows"`
	AnnotationCoveragePct float64 `json:"annotationCoveragePct"`
	TruthVersion          string  `json:"truthVersion"`
	LineageChecksum       string  `json:"lineageChecksum"`
}

type demoService struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Domain     string `json:"domain"`
	Kind       string `json:"kind"`
	Owner      string `json:"owner"`
	Escalation string `json:"escalation"`
	RunbookURL string `json:"runbookURL,omitempty"`
	DocRoot    string `json:"docRoot,omitempty"`
	FlowCount  int    `json:"flowCount"`
}

type demoEdge struct {
	ID      string `json:"id"`
	From    string `json:"from"`
	To      string `json:"to"`
	FieldID string `json:"fieldId"`
	Label   string `json:"label"`
}

type demoFieldMetadata struct {
	ServiceID string `json:"serviceId"`
	Domain    string `json:"domain"`
}

type demoScenario struct {
	Summary lineage.DiffSummary   `json:"summary"`
	Changes []lineage.DriftChange `json:"changes"`
}

type nodeMeta struct {
	ID         string
	Name       string
	Domain     string
	Kind       string
	Owner      string
	Escalation string
	RunbookURL string
	DocRoot    string
	FlowCount  int
}

type mutationType string

const (
	mutationTypeChanged      mutationType = "type_changed"
	mutationEnumChanged      mutationType = "enum_changed"
	mutationFieldRemoved     mutationType = "field_removed"
	mutationSourceVersion    mutationType = "source_version_changed"
	mutationExternalAsOf     mutationType = "external_as_of_stale"
	mutationAnnotationMissed mutationType = "annotation_missing"
)

func main() {
	artifactPath := flag.String("artifact", "tests/lineage/usecases/current.json", "Path to lineage artifact JSON")
	systemsPath := flag.String("systems", "tests/lineage/usecases/systems.yml", "Path to lineage systems registry YAML")
	outJSON := flag.String("out-json", "site/public/demo/demo-pack.json", "Output JSON file path")
	outTS := flag.String("out-ts", "site/worker/src/generated/demo-pack.ts", "Output TypeScript module path")
	flag.Parse()

	artifact, err := lineage.LoadArtifact(*artifactPath)
	if err != nil {
		fatalf("load artifact: %v", err)
	}

	registry, err := lineage.LoadSystemRegistry(*systemsPath)
	if err != nil {
		fatalf("load systems registry: %v", err)
	}

	pack, err := buildDemoPack(artifact, registry)
	if err != nil {
		fatalf("build demo pack: %v", err)
	}

	if err := writeJSON(*outJSON, pack); err != nil {
		fatalf("write json: %v", err)
	}

	if err := writeTS(*outTS, pack); err != nil {
		fatalf("write ts: %v", err)
	}

	fmt.Printf("generated demo pack: %s and %s\n", *outJSON, *outTS)
}

func buildDemoPack(artifact lineage.Artifact, registry lineage.SystemRegistry) (demoPack, error) {
	if len(artifact.Fields) == 0 {
		return demoPack{}, fmt.Errorf("artifact has no fields")
	}

	registryByID := map[string]lineage.SystemMetadata{}
	for _, system := range registry.Systems {
		registryByID[normalizeID(system.ID)] = system
	}

	nodes := map[string]*nodeMeta{}
	edges := map[string]demoEdge{}
	fieldMeta := map[string]demoFieldMetadata{}

	for _, field := range artifact.Fields {
		svcID := normalizeID(field.SourceSystem)
		domain := domainFromFieldID(field.FieldID)
		if domain == "" {
			domain = "shared"
		}

		node := ensureNode(nodes, svcID)
		node.Domain = preferDomain(node.Domain, domain)
		node.Kind = "internal"
		node.FlowCount++
		if node.Owner == "" {
			node.Owner = field.Owner
		}
		if node.Escalation == "" {
			node.Escalation = field.Escalation
		}

		if meta, ok := registryByID[svcID]; ok {
			node.Name = fallbackString(meta.Name, node.Name)
			node.Owner = fallbackString(meta.OwnerTeam, node.Owner)
			node.RunbookURL = fallbackString(meta.RunbookURL, node.RunbookURL)
			node.DocRoot = fallbackString(meta.DocRoot, node.DocRoot)
			if node.Escalation == "" {
				node.Escalation = firstChannel(meta.Escalation)
			}
		}

		fieldMeta[field.FieldID] = demoFieldMetadata{ServiceID: svcID, Domain: domain}

		for _, source := range field.Sources {
			upstream := deriveUpstreamSystem(source)
			if upstream == "" || upstream == svcID {
				continue
			}

			upNode := ensureNode(nodes, upstream)
			if source.Scope == "external" {
				upNode.Kind = "external"
				upNode.Domain = preferDomain(upNode.Domain, "external")
			} else {
				upNode.Domain = preferDomain(upNode.Domain, "shared")
				if upNode.Kind == "" {
					upNode.Kind = "internal"
				}
			}

			if meta, ok := registryByID[upstream]; ok {
				upNode.Name = fallbackString(meta.Name, upNode.Name)
				upNode.Owner = fallbackString(meta.OwnerTeam, upNode.Owner)
				upNode.RunbookURL = fallbackString(meta.RunbookURL, upNode.RunbookURL)
				upNode.DocRoot = fallbackString(meta.DocRoot, upNode.DocRoot)
				if upNode.Escalation == "" {
					upNode.Escalation = firstChannel(meta.Escalation)
				}
			}

			edgeKey := strings.Join([]string{upstream, svcID, field.FieldID}, "|")
			if _, exists := edges[edgeKey]; exists {
				continue
			}
			edges[edgeKey] = demoEdge{
				ID:      fmt.Sprintf("e_%s", shortDigest(edgeKey)),
				From:    upstream,
				To:      svcID,
				FieldID: field.FieldID,
				Label:   source.Target,
			}
		}
	}

	ensureParentNodesFromSubsystems(nodes)

	for id, node := range nodes {
		if meta, ok := registryByID[id]; ok {
			node.Name = fallbackString(meta.Name, node.Name)
			node.Owner = fallbackString(meta.OwnerTeam, node.Owner)
			node.RunbookURL = fallbackString(meta.RunbookURL, node.RunbookURL)
			node.DocRoot = fallbackString(meta.DocRoot, node.DocRoot)
			if node.Escalation == "" {
				node.Escalation = firstChannel(meta.Escalation)
			}
		}
		node.Name = fallbackString(node.Name, titleCaseID(id))
		node.Owner = fallbackString(node.Owner, "team.unknown")
		node.Escalation = fallbackString(node.Escalation, "slack:#unknown-oncall")
		node.Kind = fallbackString(node.Kind, "internal")
		node.Domain = fallbackString(node.Domain, "shared")
	}

	services := make([]demoService, 0, len(nodes))
	for _, node := range nodes {
		services = append(services, demoService{
			ID:         node.ID,
			Name:       node.Name,
			Domain:     node.Domain,
			Kind:       node.Kind,
			Owner:      node.Owner,
			Escalation: node.Escalation,
			RunbookURL: node.RunbookURL,
			DocRoot:    node.DocRoot,
			FlowCount:  node.FlowCount,
		})
	}
	sort.Slice(services, func(i, j int) bool { return services[i].ID < services[j].ID })

	edgeList := make([]demoEdge, 0, len(edges))
	for _, edge := range edges {
		edgeList = append(edgeList, edge)
	}
	sort.Slice(edgeList, func(i, j int) bool {
		if edgeList[i].From != edgeList[j].From {
			return edgeList[i].From < edgeList[j].From
		}
		if edgeList[i].To != edgeList[j].To {
			return edgeList[i].To < edgeList[j].To
		}
		return edgeList[i].FieldID < edgeList[j].FieldID
	})

	mutationTypes := []mutationType{
		mutationTypeChanged,
		mutationEnumChanged,
		mutationFieldRemoved,
		mutationSourceVersion,
		mutationExternalAsOf,
		mutationAnnotationMissed,
	}

	scenarios := map[string]map[string]demoScenario{}
	fieldsByMutation := map[string][]string{}

	for _, field := range artifact.Fields {
		scenarios[field.FieldID] = map[string]demoScenario{}
	}

	for _, typ := range mutationTypes {
		fieldsByMutation[string(typ)] = make([]string, 0)
		for idx, field := range artifact.Fields {
			head, ok := mutateArtifact(artifact, idx, typ)
			if !ok {
				continue
			}
			diff := lineage.DiffArtifacts(artifact, head)
			if len(diff.Changes) == 0 {
				continue
			}
			scenarios[field.FieldID][string(typ)] = demoScenario{
				Summary: diff.Summary,
				Changes: diff.Changes,
			}
			fieldsByMutation[string(typ)] = append(fieldsByMutation[string(typ)], field.FieldID)
		}
		sort.Strings(fieldsByMutation[string(typ)])
	}

	escalationByService := map[string][]lineage.EscalationStep{}
	seenServices := map[string]bool{}
	for _, field := range artifact.Fields {
		svcID := normalizeID(field.SourceSystem)
		if svcID == "" || seenServices[svcID] {
			continue
		}
		seenServices[svcID] = true

		steps, err := lineage.BuildEscalationChain(field.SourceSystem, artifact, registry, 8)
		if err != nil {
			return demoPack{}, fmt.Errorf("build escalation chain for %s: %w", field.SourceSystem, err)
		}
		escalationByService[svcID] = steps
	}

	artifactDigest := sha256.Sum256(mustJSON(artifact))

	pack := demoPack{
		GeneratedAt: time.Now().UTC().Format(time.RFC3339),
		Truth: demoTruth{
			SupportedFlows:        len(artifact.Fields),
			AnnotatedFlows:        len(artifact.Fields),
			AnnotationCoveragePct: 100,
			TruthVersion:          artifact.SchemaVersion,
			LineageChecksum:       "sha256:" + hex.EncodeToString(artifactDigest[:]),
		},
		Services:           services,
		Edges:              edgeList,
		FieldMetadata:      fieldMeta,
		MutationTypes:      toStringSlice(mutationTypes),
		FieldsByMutation:   fieldsByMutation,
		MutationScenarios:  scenarios,
		EscalationBySystem: escalationByService,
		BaselineSummary:    lineage.DiffSummary{Total: 0, High: 0, Medium: 0, Low: 0, Info: 0},
	}

	// Demo-specific service/flow overlay to make payment failure causality explicit:
	// PromotionsConfig -> PromotionsApplication -> CommerceGateway -> FintechGateway.
	ensureDemoService(&pack, demoService{
		ID:         "promotionsconfig",
		Name:       "PromotionsConfig Service",
		Domain:     "ecommerce",
		Kind:       "internal",
		Owner:      "team.promotions-config",
		Escalation: "pagerduty:promotions-config-oncall",
		FlowCount:  1,
	})
	ensureDemoService(&pack, demoService{
		ID:         "promotionsconfig:api-service",
		Name:       "PromotionsConfig API Service",
		Domain:     "ecommerce",
		Kind:       "internal",
		Owner:      "team.promotions-config",
		Escalation: "pagerduty:promotions-config-oncall",
		FlowCount:  1,
	})
	ensureDemoService(&pack, demoService{
		ID:         "promotionsapp",
		Name:       "PromotionsApplication Service",
		Domain:     "ecommerce",
		Kind:       "internal",
		Owner:      "team.promotions-app",
		Escalation: "pagerduty:promotions-app-oncall",
		FlowCount:  1,
	})
	ensureDemoService(&pack, demoService{
		ID:         "promotionsapp:api-service",
		Name:       "PromotionsApplication API Service",
		Domain:     "ecommerce",
		Kind:       "internal",
		Owner:      "team.promotions-app",
		Escalation: "pagerduty:promotions-app-oncall",
		FlowCount:  1,
	})
	ensureDemoEdge(&pack, demoEdge{
		ID:      "e_demo_promotionsconfig_to_promotionsapp",
		From:    "promotionsconfig:api-service",
		To:      "promotionsapp:api-service",
		FieldID: "response_ecommerce_cart_pricing_waterfall",
		Label:   "PromotionsConfig.GetPromotionType",
	})
	ensureDemoEdge(&pack, demoEdge{
		ID:      "e_demo_promotionsapp_to_commercegateway",
		From:    "promotionsapp:api-service",
		To:      "commercegateway:api-service",
		FieldID: "response_ecommerce_cart_pricing_waterfall",
		Label:   "PromotionsApplication.ApplyPromotion",
	})
	ensureDemoEdge(&pack, demoEdge{
		ID:      "e_demo_commercegateway_to_fintechgateway",
		From:    "commercegateway:api-service",
		To:      "fintechgateway:api-service",
		FieldID: "response_ecommerce_cart_pricing_waterfall",
		Label:   "CommerceGateway.ForwardCheckoutPricing",
	})
	addLogisticsInternalOverlay(&pack)

	overridePaymentsEnumScenario(&pack)

	// Add a custom numeric widening scenario to illustrate cross-language width drift (Go uint8 -> JS Number -> Go uint16).
	const customType = "numeric_widen"
	const customField = "response_ecommerce_checkout_risk_gate"

	pack.MutationTypes = append(pack.MutationTypes, customType)
	if pack.FieldsByMutation == nil {
		pack.FieldsByMutation = map[string][]string{}
	}
	pack.FieldsByMutation[customType] = []string{customField}

	if pack.MutationScenarios == nil {
		pack.MutationScenarios = map[string]map[string]demoScenario{}
	}
	if pack.MutationScenarios[customField] == nil {
		pack.MutationScenarios[customField] = map[string]demoScenario{}
	}

	pack.MutationScenarios[customField][customType] = demoScenario{
		Summary: lineage.DiffSummary{Total: 1, High: 1, Medium: 0, Low: 0, Info: 0},
		Changes: []lineage.DriftChange{
			{
				Severity:   lineage.SeverityHigh,
				ChangeType: "type_changed",
				FieldID:    customField,
				Message:    "Producer widened from uint8 to uint16; JS consumer coerces to Number and passes through values >255; downstream Go consumer overflows.",
				Source: &lineage.DriftEdge{
					Service: "CommerceCore:api-service",
					API:     "CommerceCore.GetCheckoutRiskGate",
				},
				Impact: &lineage.DriftEdge{
					Service: "CommerceGateway:api-service",
					API:     customField,
				},
				TypeDelta: &lineage.TypeDelta{
					Change:      "expanded",
					BeforeLabel: "uint8",
					AfterLabel:  "uint16",
				},
				Validation: "numeric range compatibility check (0..255) failed in downstream typed consumer",
				Suggestion: "add boundary tests/assertions for >255 and roll out consumer updates before widening producer type",
			},
		},
	}

	return pack, nil
}

func ensureDemoService(pack *demoPack, service demoService) {
	for _, existing := range pack.Services {
		if existing.ID == service.ID {
			return
		}
	}
	pack.Services = append(pack.Services, service)
	sort.Slice(pack.Services, func(i, j int) bool { return pack.Services[i].ID < pack.Services[j].ID })
}

func ensureDemoEdge(pack *demoPack, edge demoEdge) {
	for _, existing := range pack.Edges {
		if existing.From == edge.From && existing.To == edge.To && existing.FieldID == edge.FieldID {
			return
		}
	}
	pack.Edges = append(pack.Edges, edge)
	sort.Slice(pack.Edges, func(i, j int) bool {
		if pack.Edges[i].From != pack.Edges[j].From {
			return pack.Edges[i].From < pack.Edges[j].From
		}
		if pack.Edges[i].To != pack.Edges[j].To {
			return pack.Edges[i].To < pack.Edges[j].To
		}
		return pack.Edges[i].FieldID < pack.Edges[j].FieldID
	})
}

func overridePaymentsEnumScenario(pack *demoPack) {
	const fieldID = "response_ecommerce_cart_pricing_waterfall"
	fieldScenarios, ok := pack.MutationScenarios[fieldID]
	if !ok {
		return
	}
	fieldScenarios[string(mutationEnumChanged)] = demoScenario{
		Summary: lineage.DiffSummary{
			Total:  1,
			High:   0,
			Medium: 1,
			Low:    0,
			Info:   0,
		},
		Changes: []lineage.DriftChange{
			{
				Severity:   lineage.SeverityMedium,
				ChangeType: "enum_changed",
				FieldID:    fieldID,
				Message:    "PromotionsConfig added promotion_type=stacked_cashback. PromotionsApplication was updated in tandem and passes it through, but the payment path still handles only percentage,fixed_amount,bogo. Promotion-applied checkouts can fail payment authorization until payment services are updated.",
				Source: &lineage.DriftEdge{
					Service: "PromotionsConfig:api-service",
					API:     "PromotionsConfig.GetPromotionType",
				},
				Impact: &lineage.DriftEdge{
					Service: "FintechGateway:api-service",
					API:     "FintechGateway.GetPaymentAuthorizationDecision",
				},
				ModifiedBy: []string{"PromotionsApplication:api-service", "CommerceGateway:api-service"},
				Validation: "promotion type allowlist mismatch between config/app services and payment authorization path",
				Suggestion: "add enum contract tests across PromotionsConfig -> PromotionsApplication -> CommerceGateway -> FintechGateway and define unknown-type fallback before rollout",
			},
		},
	}
}

func addLogisticsInternalOverlay(pack *demoPack) {
	ensureDemoService(pack, demoService{
		ID:         "logisticsgateway:tracking-api",
		Name:       "Tracking API",
		Domain:     "logistics",
		Kind:       "internal",
		Owner:      "team.logistics-tracking",
		Escalation: "pagerduty:logistics-oncall",
		FlowCount:  1,
	})
	ensureDemoService(pack, demoService{
		ID:         "logisticsgateway:routing",
		Name:       "Routing Service",
		Domain:     "logistics",
		Kind:       "internal",
		Owner:      "team.logistics-routing",
		Escalation: "pagerduty:logistics-oncall",
		FlowCount:  1,
	})
	ensureDemoService(pack, demoService{
		ID:         "logisticsgateway:ingestion",
		Name:       "Location Ingest",
		Domain:     "logistics",
		Kind:       "internal",
		Owner:      "team.logistics-ingest",
		Escalation: "pagerduty:logistics-oncall",
		FlowCount:  1,
	})
	ensureDemoService(pack, demoService{
		ID:         "logisticsgateway:notification",
		Name:       "Notification Service",
		Domain:     "logistics",
		Kind:       "internal",
		Owner:      "team.logistics-notify",
		Escalation: "pagerduty:logistics-oncall",
		FlowCount:  1,
	})

	ensureDemoEdge(pack, demoEdge{
		ID:      "e_demo_logistics_ingestion_to_routing",
		From:    "logisticsgateway:ingestion",
		To:      "logisticsgateway:routing",
		FieldID: "response_logistics_location_event_stream",
		Label:   "kafka:LocationIngest.LocationUpdates",
	})
	ensureDemoEdge(pack, demoEdge{
		ID:      "e_demo_logistics_routing_to_tracking_api",
		From:    "logisticsgateway:routing",
		To:      "logisticsgateway:tracking-api",
		FieldID: "response_logistics_eta_projection",
		Label:   "api:Routing.ComputeEta",
	})
	ensureDemoEdge(pack, demoEdge{
		ID:      "e_demo_logistics_tracking_api_to_notification",
		From:    "logisticsgateway:tracking-api",
		To:      "logisticsgateway:notification",
		FieldID: "response_logistics_delivery_state",
		Label:   "kafka:TrackingApi.DeliveryStateChanged",
	})
	ensureDemoEdge(pack, demoEdge{
		ID:      "e_demo_logistics_tracking_api_to_core",
		From:    "logisticsgateway:tracking-api",
		To:      "logisticscore",
		FieldID: "response_logistics_eta_projection",
		Label:   "api:TrackingApi.GetEta",
	})
	ensureDemoEdge(pack, demoEdge{
		ID:      "e_demo_logistics_core_to_tracking_api",
		From:    "logisticscore",
		To:      "logisticsgateway:tracking-api",
		FieldID: "response_logistics_route_risk_assessment",
		Label:   "api:LogisticsCore.GetRouteRiskAssessment",
	})
}

func mutateArtifact(base lineage.Artifact, fieldIndex int, typ mutationType) (lineage.Artifact, bool) {
	head := cloneArtifact(base)
	if fieldIndex < 0 || fieldIndex >= len(head.Fields) {
		return lineage.Artifact{}, false
	}

	switch typ {
	case mutationTypeChanged:
		if len(head.Fields[fieldIndex].Sources) == 0 {
			return lineage.Artifact{}, false
		}
		head.Fields[fieldIndex].Sources[0].ContractRef += "#type-v2"
		return head, true
	case mutationEnumChanged:
		head.Fields[fieldIndex].MergeStrategy = nextMergeStrategy(head.Fields[fieldIndex].MergeStrategy)
		return head, true
	case mutationFieldRemoved:
		head.Fields = append(head.Fields[:fieldIndex], head.Fields[fieldIndex+1:]...)
		return head, true
	case mutationSourceVersion:
		head.Fields[fieldIndex].SourceVersion = bumpVersion(head.Fields[fieldIndex].SourceVersion)
		return head, true
	case mutationExternalAsOf:
		for i := range head.Fields[fieldIndex].Sources {
			src := &head.Fields[fieldIndex].Sources[i]
			if src.Scope != "external" {
				continue
			}
			src.AsOf = rollbackDate(src.AsOf)
			return head, true
		}
		return lineage.Artifact{}, false
	case mutationAnnotationMissed:
		if len(head.Fields[fieldIndex].Sources) > 0 {
			head.Fields[fieldIndex].Sources = head.Fields[fieldIndex].Sources[1:]
			return head, true
		}
		head.Fields[fieldIndex].Owner = "team.missing"
		return head, true
	default:
		return lineage.Artifact{}, false
	}
}

func cloneArtifact(base lineage.Artifact) lineage.Artifact {
	data := mustJSON(base)
	var cloned lineage.Artifact
	if err := json.Unmarshal(data, &cloned); err != nil {
		panic(fmt.Errorf("clone artifact: %w", err))
	}
	return cloned
}

func nextMergeStrategy(current string) string {
	values := []string{"single_source", "priority", "first_non_null", "union", "custom"}
	for idx, value := range values {
		if value == current {
			return values[(idx+1)%len(values)]
		}
	}
	return "priority"
}

func bumpVersion(version string) string {
	if strings.TrimSpace(version) == "" {
		return "vnext"
	}
	if strings.HasSuffix(version, ".1") {
		return version + ".2"
	}
	return version + ".1"
}

func rollbackDate(asOf string) string {
	t, err := time.Parse("2006-01-02", asOf)
	if err != nil {
		return "2020-01-01"
	}
	return t.AddDate(0, 0, -30).Format("2006-01-02")
}

func deriveUpstreamSystem(source lineage.SourceRef) string {
	if source.UpstreamSystem != "" {
		return normalizeID(source.UpstreamSystem)
	}
	if source.Scope == "external" && source.ProviderID != "" {
		return normalizeID(source.ProviderID)
	}
	if source.Kind == "api" || source.Kind == "event" {
		target := strings.TrimSpace(source.Target)
		if target == "" {
			return ""
		}
		if dot := strings.Index(target, "."); dot > 0 {
			return normalizeID(target[:dot])
		}
		if slash := strings.Index(target, "/"); slash > 0 {
			return normalizeID(target[:slash])
		}
		return normalizeID(target)
	}
	return ""
}

func domainFromFieldID(fieldID string) string {
	parts := strings.Split(fieldID, "_")
	if len(parts) < 3 {
		return ""
	}
	if parts[0] != "response" {
		return ""
	}
	return parts[1]
}

func ensureNode(nodes map[string]*nodeMeta, id string) *nodeMeta {
	id = normalizeID(id)
	if id == "" {
		id = "unknown"
	}
	if existing, ok := nodes[id]; ok {
		return existing
	}
	created := &nodeMeta{ID: id}
	nodes[id] = created
	return created
}

func ensureParentNodesFromSubsystems(nodes map[string]*nodeMeta) {
	ids := make([]string, 0, len(nodes))
	for id := range nodes {
		ids = append(ids, id)
	}
	for _, id := range ids {
		cut := strings.Index(id, ":")
		if cut <= 0 {
			continue
		}
		rootID := id[:cut]
		child := nodes[id]
		parent := ensureNode(nodes, rootID)
		parent.Domain = preferDomain(parent.Domain, child.Domain)
		if strings.TrimSpace(parent.Owner) == "" {
			parent.Owner = child.Owner
		}
		if strings.TrimSpace(parent.Escalation) == "" {
			parent.Escalation = child.Escalation
		}
		if strings.TrimSpace(parent.Kind) == "" || (parent.Kind == "internal" && child.Kind == "external") {
			parent.Kind = child.Kind
		}
		if child.FlowCount > 0 {
			parent.FlowCount += child.FlowCount
		}
	}
}

func normalizeID(raw string) string {
	return strings.ToLower(strings.TrimSpace(raw))
}

func preferDomain(existing string, candidate string) string {
	if strings.TrimSpace(candidate) == "" {
		return existing
	}
	if strings.TrimSpace(existing) == "" {
		return candidate
	}
	if existing == "shared" && candidate != "shared" {
		return candidate
	}
	if existing == "external" && candidate != "external" {
		return candidate
	}
	return existing
}

func firstChannel(contacts []lineage.Contact) string {
	for _, contact := range contacts {
		if strings.TrimSpace(contact.Channel) != "" {
			return contact.Channel
		}
	}
	return ""
}

func titleCaseID(id string) string {
	parts := strings.FieldsFunc(id, func(r rune) bool {
		return r == '_' || r == '-'
	})
	for i := range parts {
		if parts[i] == "" {
			continue
		}
		parts[i] = strings.ToUpper(parts[i][:1]) + parts[i][1:]
	}
	return strings.Join(parts, " ")
}

func fallbackString(primary string, fallback string) string {
	if strings.TrimSpace(primary) != "" {
		return primary
	}
	return fallback
}

func shortDigest(value string) string {
	h := sha256.Sum256([]byte(value))
	return hex.EncodeToString(h[:])[:12]
}

func toStringSlice(values []mutationType) []string {
	result := make([]string, 0, len(values))
	for _, value := range values {
		result = append(result, string(value))
	}
	return result
}

func mustJSON(value any) []byte {
	data, err := json.Marshal(value)
	if err != nil {
		panic(fmt.Errorf("marshal json: %w", err))
	}
	return data
}

func writeJSON(path string, payload any) error {
	data, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal json: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("mkdir output dir: %w", err)
	}
	if err := os.WriteFile(path, append(data, '\n'), 0o644); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}
	return nil
}

func writeTS(path string, payload any) error {
	data, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal ts payload: %w", err)
	}
	content := "// Code generated by cmd/demo-pack; DO NOT EDIT.\n" +
		"export const DEMO_PACK = " + string(data) + " as const;\n"
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("mkdir ts dir: %w", err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}
	return nil
}

func fatalf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "ERROR: "+format+"\n", args...)
	os.Exit(1)
}
