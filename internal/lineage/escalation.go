// escalation.go - System registry and escalation-chain resolver.
package lineage

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

// Contact is an emergency contact entry for a system.
type Contact struct {
	Role    string `yaml:"role" json:"role"`
	Name    string `yaml:"name" json:"name"`
	Channel string `yaml:"channel" json:"channel"`
}

// SystemMetadata captures contact ownership for one system.
type SystemMetadata struct {
	ID         string    `yaml:"id" json:"id"`
	Name       string    `yaml:"name" json:"name"`
	OwnerTeam  string    `yaml:"owner_team" json:"owner_team"`
	Escalation []Contact `yaml:"escalation" json:"escalation"`
}

// SystemRegistry lists all known systems.
type SystemRegistry struct {
	Systems []SystemMetadata `yaml:"systems" json:"systems"`
}

// EscalationStep is one hop in the upstream escalation chain.
type EscalationStep struct {
	Depth    int       `json:"depth"`
	SystemID string    `json:"system_id"`
	Name     string    `json:"name,omitempty"`
	Owner    string    `json:"owner,omitempty"`
	Contacts []Contact `json:"contacts"`
	Reason   string    `json:"reason"`
}

// LoadSystemRegistry loads a YAML system registry file.
func LoadSystemRegistry(path string) (SystemRegistry, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return SystemRegistry{}, fmt.Errorf("read system registry: %w", err)
	}
	var registry SystemRegistry
	if err := yaml.Unmarshal(data, &registry); err != nil {
		return SystemRegistry{}, fmt.Errorf("parse system registry: %w", err)
	}

	seen := map[string]bool{}
	for _, system := range registry.Systems {
		id := normalizeSystemID(system.ID)
		if id == "" {
			return SystemRegistry{}, fmt.Errorf("system registry contains empty id")
		}
		if seen[id] {
			return SystemRegistry{}, fmt.Errorf("system registry contains duplicate id %q", system.ID)
		}
		seen[id] = true
	}

	return registry, nil
}

// BuildEscalationChain resolves emergency contacts working backwards from a service.
func BuildEscalationChain(serviceID string, artifact Artifact, registry SystemRegistry, maxDepth int) ([]EscalationStep, error) {
	if strings.TrimSpace(serviceID) == "" {
		return nil, fmt.Errorf("service_id cannot be empty")
	}
	if maxDepth <= 0 {
		maxDepth = 8
	}

	start := normalizeSystemID(serviceID)
	graph := buildUpstreamGraph(artifact)
	registryByID := mapRegistry(registry)
	fallbackBySystem := buildFallbackContacts(artifact)

	type queueItem struct {
		System string
		Depth  int
		Reason string
	}

	queue := []queueItem{{System: start, Depth: 0, Reason: "reported_bad_data"}}
	visited := map[string]bool{}
	steps := make([]EscalationStep, 0)

	for len(queue) > 0 {
		item := queue[0]
		queue = queue[1:]

		if visited[item.System] {
			continue
		}
		visited[item.System] = true

		step := EscalationStep{Depth: item.Depth, SystemID: item.System, Reason: item.Reason}
		if system, ok := registryByID[item.System]; ok {
			step.Name = system.Name
			step.Owner = system.OwnerTeam
			step.Contacts = append(step.Contacts, system.Escalation...)
		}
		if len(step.Contacts) == 0 {
			step.Contacts = append(step.Contacts, fallbackBySystem[item.System]...)
		}
		steps = append(steps, step)

		if item.Depth >= maxDepth {
			continue
		}

		next := append([]string{}, graph[item.System]...)
		sort.Strings(next)
		for _, upstream := range next {
			if visited[upstream] {
				continue
			}
			queue = append(queue, queueItem{
				System: upstream,
				Depth:  item.Depth + 1,
				Reason: fmt.Sprintf("upstream_of:%s", item.System),
			})
		}
	}

	if len(steps) == 0 {
		return nil, fmt.Errorf("no systems resolved for service %q", serviceID)
	}

	sort.SliceStable(steps, func(i, j int) bool {
		if steps[i].Depth != steps[j].Depth {
			return steps[i].Depth < steps[j].Depth
		}
		return steps[i].SystemID < steps[j].SystemID
	})

	return steps, nil
}

func buildUpstreamGraph(artifact Artifact) map[string][]string {
	edges := map[string]map[string]bool{}
	addEdge := func(from string, to string) {
		if from == "" || to == "" || from == to {
			return
		}
		if _, ok := edges[from]; !ok {
			edges[from] = map[string]bool{}
		}
		edges[from][to] = true
	}

	for _, field := range artifact.Fields {
		from := normalizeSystemID(field.SourceSystem)
		if from == "" {
			continue
		}

		for _, source := range field.Sources {
			upstream := deriveUpstreamSystem(source)
			if upstream == "" || upstream == from {
				continue
			}
			addEdge(from, upstream)
			addEdge(topologyRootSystemID(from), topologyRootSystemID(upstream))
		}
	}

	graph := map[string][]string{}
	for from, toSet := range edges {
		for to := range toSet {
			graph[from] = append(graph[from], to)
		}
		sort.Strings(graph[from])
	}
	return graph
}

func topologyRootSystemID(value string) string {
	value = normalizeSystemID(value)
	if value == "" {
		return ""
	}
	if cut := strings.Index(value, ":"); cut > 0 {
		return value[:cut]
	}
	return value
}

func deriveUpstreamSystem(source SourceRef) string {
	if source.UpstreamSystem != "" {
		return normalizeSystemID(source.UpstreamSystem)
	}
	if source.Scope == "external" && source.ProviderID != "" {
		return normalizeSystemID(source.ProviderID)
	}
	if source.Kind == "api" || source.Kind == "event" {
		target := strings.TrimSpace(source.Target)
		if target == "" {
			return ""
		}
		if dot := strings.Index(target, "."); dot > 0 {
			return normalizeSystemID(target[:dot])
		}
		if slash := strings.Index(target, "/"); slash > 0 {
			return normalizeSystemID(target[:slash])
		}
		return normalizeSystemID(target)
	}
	return ""
}

func mapRegistry(registry SystemRegistry) map[string]SystemMetadata {
	result := map[string]SystemMetadata{}
	for _, system := range registry.Systems {
		result[normalizeSystemID(system.ID)] = system
	}
	return result
}

func buildFallbackContacts(artifact Artifact) map[string][]Contact {
	fallback := map[string][]Contact{}
	seen := map[string]map[string]bool{}
	for _, field := range artifact.Fields {
		system := normalizeSystemID(field.SourceSystem)
		if system == "" {
			continue
		}
		if _, ok := seen[system]; !ok {
			seen[system] = map[string]bool{}
		}

		ownerRef := "owner:" + field.Owner
		if field.Owner != "" && !seen[system][ownerRef] {
			fallback[system] = append(fallback[system], Contact{Role: "owner", Name: field.Owner, Channel: ""})
			seen[system][ownerRef] = true
		}

		escalationRef := "escalation:" + field.Escalation
		if field.Escalation != "" && !seen[system][escalationRef] {
			fallback[system] = append(fallback[system], Contact{Role: "escalation", Name: "", Channel: field.Escalation})
			seen[system][escalationRef] = true
		}
	}
	return fallback
}

func normalizeSystemID(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}
