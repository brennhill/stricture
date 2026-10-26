// demo_overlay.go — Hand-authored demo overlays layered onto the generated pack.
package main

import (
	"sort"

	"github.com/stricture/stricture/internal/lineage"
)

func ensureDemoService(pack *demoPack, service demoService) {
	for _, existing := range pack.Services {
		if existing.ID == service.ID {
			return
		}
	}
	if len(service.Flows) == 0 {
		service.Flows = flowForDomain(service.Domain)
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
	sortEdges(pack.Edges)
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

// applyPromotionsOverlay makes payment failure causality explicit:
// PromotionsConfig -> PromotionsApplication -> CommerceGateway -> FintechGateway.
func applyPromotionsOverlay(pack *demoPack) {
	const promotionsFieldID = "response_ecommerce_cart_pricing_waterfall"
	services := []demoService{
		{ID: "promotionsconfig", Name: "PromotionsConfig Service", Owner: "team.promotions-config", Escalation: "pagerduty:promotions-config-oncall"},
		{ID: "promotionsconfig:api-service", Name: "PromotionsConfig API Service", Owner: "team.promotions-config", Escalation: "pagerduty:promotions-config-oncall"},
		{ID: "promotionsapp", Name: "PromotionsApplication Service", Owner: "team.promotions-app", Escalation: "pagerduty:promotions-app-oncall"},
		{ID: "promotionsapp:api-service", Name: "PromotionsApplication API Service", Owner: "team.promotions-app", Escalation: "pagerduty:promotions-app-oncall"},
	}
	for _, service := range services {
		service.Domain = "ecommerce"
		service.Kind = "internal"
		service.FlowCount = 1
		ensureDemoService(pack, service)
	}
	edges := []demoEdge{
		{ID: "e_demo_promotionsconfig_to_promotionsapp", From: "promotionsconfig:api-service", To: "promotionsapp:api-service", Label: "PromotionsConfig.GetPromotionType"},
		{ID: "e_demo_promotionsapp_to_commercegateway", From: "promotionsapp:api-service", To: "commercegateway:api-service", Label: "PromotionsApplication.ApplyPromotion"},
		{ID: "e_demo_commercegateway_to_fintechgateway", From: "commercegateway:api-service", To: "fintechgateway:api-service", Label: "CommerceGateway.ForwardCheckoutPricing"},
	}
	for _, edge := range edges {
		edge.FieldID = promotionsFieldID
		ensureDemoEdge(pack, edge)
	}
}

// addNumericWidenScenario illustrates cross-language width drift (Go uint8 -> JS Number -> Go uint16).
func addNumericWidenScenario(pack *demoPack) {
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
}
