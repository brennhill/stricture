// demo_types.go — Data model and domain flow defaults for the generated demo pack.
package main

import (
	"strings"

	"github.com/stricture/stricture/internal/lineage"
)

type demoPack struct {
	GeneratedAt        string                              `json:"generated_at"`
	Truth              demoTruth                           `json:"truth"`
	Flows              []demoFlow                          `json:"flows"`
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
	ID         string   `json:"id"`
	Name       string   `json:"name"`
	Domain     string   `json:"domain"`
	Kind       string   `json:"kind"`
	Owner      string   `json:"owner"`
	Escalation string   `json:"escalation"`
	RunbookURL string   `json:"runbookURL,omitempty"`
	DocRoot    string   `json:"docRoot,omitempty"`
	FlowCount  int      `json:"flowCount"`
	Flows      []string `json:"flows,omitempty"`
}

type demoFlow struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Level int    `json:"level"`
	Owner string `json:"owner"`
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
	Flows      []string
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

func defaultFlows() []demoFlow {
	return []demoFlow{
		{ID: "checkout", Name: "Checkout", Level: 1, Owner: "team.ecommerce"},
		{ID: "payments", Name: "Payments", Level: 1, Owner: "team.fintech"},
		{ID: "delivery", Name: "Delivery", Level: 2, Owner: "team.logistics"},
		{ID: "streaming", Name: "Streaming", Level: 2, Owner: "team.media"},
		{ID: "compliance", Name: "Compliance", Level: 1, Owner: "team.governance"},
		{ID: "platform", Name: "Platform", Level: 3, Owner: "team.platform"},
	}
}

func flowForDomain(domain string) []string {
	switch strings.ToLower(domain) {
	case "ecommerce":
		return []string{"checkout"}
	case "fintech":
		return []string{"payments"}
	case "logistics":
		return []string{"delivery"}
	case "media":
		return []string{"streaming"}
	case "governance", "corporate":
		return []string{"compliance"}
	case "shared":
		return []string{"platform"}
	default:
		return nil
	}
}
