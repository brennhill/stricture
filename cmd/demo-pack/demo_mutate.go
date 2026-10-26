// demo_mutate.go — Artifact mutations used to generate demo drift scenarios.
package main

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/stricture/stricture/internal/lineage"
)

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
