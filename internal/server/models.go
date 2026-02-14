package server

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// ArtifactIngestRequest is the v0 ingest envelope for lineage runs.
type ArtifactIngestRequest struct {
	Organization string            `json:"organization"`
	Project      string            `json:"project"`
	Service      string            `json:"service"`
	RunID        string            `json:"run_id,omitempty"`
	CommitSHA    string            `json:"commit_sha,omitempty"`
	GeneratedAt  string            `json:"generated_at,omitempty"`
	Artifact     json.RawMessage   `json:"artifact"`
	Diff         json.RawMessage   `json:"diff,omitempty"`
	Summary      string            `json:"summary,omitempty"`
	Metadata     map[string]string `json:"metadata,omitempty"`
}

type ingestResponse struct {
	Accepted bool   `json:"accepted"`
	RunID    string `json:"run_id"`
	Location string `json:"location"`
}

func normalizeAndValidateIngest(req ArtifactIngestRequest) (ArtifactIngestRequest, error) {
	req.Organization = sanitizePathToken(req.Organization)
	req.Project = sanitizePathToken(req.Project)
	req.Service = sanitizePathToken(req.Service)
	req.RunID = sanitizePathToken(req.RunID)
	req.CommitSHA = strings.TrimSpace(req.CommitSHA)
	req.GeneratedAt = strings.TrimSpace(req.GeneratedAt)

	if req.Organization == "" {
		return req, fmt.Errorf("organization is required")
	}
	if req.Project == "" {
		return req, fmt.Errorf("project is required")
	}
	if req.Service == "" {
		return req, fmt.Errorf("service is required")
	}
	if len(req.Artifact) == 0 {
		return req, fmt.Errorf("artifact is required")
	}
	if req.GeneratedAt != "" {
		if _, err := time.Parse(time.RFC3339, req.GeneratedAt); err != nil {
			return req, fmt.Errorf("generated_at must be RFC3339: %w", err)
		}
	}

	if req.RunID == "" {
		req.RunID = fmt.Sprintf("run-%d", time.Now().UTC().UnixNano())
	}
	return req, nil
}

func sanitizePathToken(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return ""
	}

	var b strings.Builder
	lastDash := false
	for _, r := range trimmed {
		switch {
		case (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9'):
			if r >= 'A' && r <= 'Z' {
				r = r - 'A' + 'a'
			}
			b.WriteRune(r)
			lastDash = false
		case r == '-' || r == '_' || r == '.':
			if !lastDash {
				b.WriteRune('-')
				lastDash = true
			}
		default:
			if !lastDash {
				b.WriteRune('-')
				lastDash = true
			}
		}
	}
	return strings.Trim(b.String(), "-")
}
