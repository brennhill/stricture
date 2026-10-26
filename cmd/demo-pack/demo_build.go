// demo_build.go — Builds the demo pack (services, edges, scenarios, escalation) from a lineage artifact.
package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/stricture/stricture/internal/lineage"
)

var demoMutationTypes = []mutationType{
	mutationTypeChanged,
	mutationEnumChanged,
	mutationFieldRemoved,
	mutationSourceVersion,
	mutationExternalAsOf,
	mutationAnnotationMissed,
}

func buildDemoPack(artifact lineage.Artifact, registry lineage.SystemRegistry) (demoPack, error) {
	if len(artifact.Fields) == 0 {
		return demoPack{}, fmt.Errorf("artifact has no fields")
	}

	registryByID := map[string]lineage.SystemMetadata{}
	for _, system := range registry.Systems {
		registryByID[normalizeID(system.ID)] = system
	}

	nodes, edges, fieldMeta := buildGraph(artifact, registryByID)
	ensureParentNodesFromSubsystems(nodes)
	finalizeNodes(nodes, registryByID)

	scenarios, fieldsByMutation := buildMutationScenarios(artifact, demoMutationTypes)

	escalationByService, err := buildEscalationMap(artifact, registry)
	if err != nil {
		return demoPack{}, err
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
		Flows:              defaultFlows(),
		Services:           sortedServices(nodes),
		Edges:              sortedEdges(edges),
		FieldMetadata:      fieldMeta,
		MutationTypes:      toStringSlice(demoMutationTypes),
		FieldsByMutation:   fieldsByMutation,
		MutationScenarios:  scenarios,
		EscalationBySystem: escalationByService,
		BaselineSummary:    lineage.DiffSummary{Total: 0, High: 0, Medium: 0, Low: 0, Info: 0},
	}

	applyPromotionsOverlay(&pack)
	addLogisticsInternalOverlay(&pack)
	overridePaymentsEnumScenario(&pack)
	addNumericWidenScenario(&pack)

	return pack, nil
}

// buildGraph derives service nodes, upstream edges, and per-field metadata from artifact fields.
func buildGraph(artifact lineage.Artifact, registryByID map[string]lineage.SystemMetadata) (map[string]*nodeMeta, map[string]demoEdge, map[string]demoFieldMetadata) {
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
		applyRegistryMeta(node, registryByID, svcID)

		fieldMeta[field.FieldID] = demoFieldMetadata{ServiceID: svcID, Domain: domain}

		for _, source := range field.Sources {
			addUpstreamEdge(nodes, edges, registryByID, field, svcID, source)
		}
	}
	return nodes, edges, fieldMeta
}

// addUpstreamEdge registers the upstream node of one source and its edge into svcID.
func addUpstreamEdge(nodes map[string]*nodeMeta, edges map[string]demoEdge, registryByID map[string]lineage.SystemMetadata, field lineage.Annotation, svcID string, source lineage.SourceRef) {
	upstream := deriveUpstreamSystem(source)
	if upstream == "" || upstream == svcID {
		return
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
	applyRegistryMeta(upNode, registryByID, upstream)

	edgeKey := strings.Join([]string{upstream, svcID, field.FieldID}, "|")
	if _, exists := edges[edgeKey]; exists {
		return
	}
	edges[edgeKey] = demoEdge{
		ID:      fmt.Sprintf("e_%s", shortDigest(edgeKey)),
		From:    upstream,
		To:      svcID,
		FieldID: field.FieldID,
		Label:   source.Target,
	}
}

// applyRegistryMeta overlays systems-registry metadata onto a node when registered.
func applyRegistryMeta(node *nodeMeta, registryByID map[string]lineage.SystemMetadata, id string) {
	meta, ok := registryByID[id]
	if !ok {
		return
	}
	node.Name = fallbackString(meta.Name, node.Name)
	node.Owner = fallbackString(meta.OwnerTeam, node.Owner)
	node.RunbookURL = fallbackString(meta.RunbookURL, node.RunbookURL)
	node.DocRoot = fallbackString(meta.DocRoot, node.DocRoot)
	if node.Escalation == "" {
		node.Escalation = firstChannel(meta.Escalation)
	}
}

// finalizeNodes applies registry metadata and fills remaining defaults on every node.
func finalizeNodes(nodes map[string]*nodeMeta, registryByID map[string]lineage.SystemMetadata) {
	for id, node := range nodes {
		applyRegistryMeta(node, registryByID, id)
		node.Name = fallbackString(node.Name, titleCaseID(id))
		node.Owner = fallbackString(node.Owner, "team.unknown")
		node.Escalation = fallbackString(node.Escalation, "slack:#unknown-oncall")
		node.Kind = fallbackString(node.Kind, "internal")
		node.Domain = fallbackString(node.Domain, "shared")
		if len(node.Flows) == 0 {
			node.Flows = flowForDomain(node.Domain)
		}
	}
}

func sortedServices(nodes map[string]*nodeMeta) []demoService {
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
			Flows:      node.Flows,
		})
	}
	sort.Slice(services, func(i, j int) bool { return services[i].ID < services[j].ID })
	return services
}

func sortedEdges(edges map[string]demoEdge) []demoEdge {
	edgeList := make([]demoEdge, 0, len(edges))
	for _, edge := range edges {
		edgeList = append(edgeList, edge)
	}
	sortEdges(edgeList)
	return edgeList
}

func sortEdges(edges []demoEdge) {
	sort.Slice(edges, func(i, j int) bool {
		if edges[i].From != edges[j].From {
			return edges[i].From < edges[j].From
		}
		if edges[i].To != edges[j].To {
			return edges[i].To < edges[j].To
		}
		return edges[i].FieldID < edges[j].FieldID
	})
}

// buildMutationScenarios diffs each mutated field to produce per-field drift scenarios.
func buildMutationScenarios(artifact lineage.Artifact, mutationTypes []mutationType) (map[string]map[string]demoScenario, map[string][]string) {
	scenarios := map[string]map[string]demoScenario{}
	fieldsByMutation := map[string][]string{}

	for _, field := range artifact.Fields {
		scenarios[field.FieldID] = map[string]demoScenario{}
	}

	for _, typ := range mutationTypes {
		key := string(typ)
		fieldsByMutation[key] = make([]string, 0)
		for idx, field := range artifact.Fields {
			head, ok := mutateArtifact(artifact, idx, typ)
			if !ok {
				continue
			}
			diff := lineage.DiffArtifacts(artifact, head)
			if len(diff.Changes) == 0 {
				continue
			}
			scenarios[field.FieldID][key] = demoScenario{
				Summary: diff.Summary,
				Changes: diff.Changes,
			}
			fieldsByMutation[key] = append(fieldsByMutation[key], field.FieldID)
		}
		sort.Strings(fieldsByMutation[key])
	}
	return scenarios, fieldsByMutation
}

// buildEscalationMap builds one escalation chain per distinct source system.
func buildEscalationMap(artifact lineage.Artifact, registry lineage.SystemRegistry) (map[string][]lineage.EscalationStep, error) {
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
			return nil, fmt.Errorf("build escalation chain for %s: %w", field.SourceSystem, err)
		}
		escalationByService[svcID] = steps
	}
	return escalationByService, nil
}
