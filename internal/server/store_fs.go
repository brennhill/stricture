package server

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
)

// FileStore writes ingest records to a deterministic filesystem layout.
type FileStore struct {
	dataDir string
}

type storedIngestRecord struct {
	ArtifactIngestRequest
	ReceivedAt string `json:"received_at"`
}

// NewFileStore creates a filesystem-backed ingest store.
func NewFileStore(dataDir string) (*FileStore, error) {
	root := strings.TrimSpace(dataDir)
	if root == "" {
		return nil, fmt.Errorf("data dir is required")
	}
	if err := os.MkdirAll(root, 0o755); err != nil {
		return nil, fmt.Errorf("create data dir: %w", err)
	}
	return &FileStore{dataDir: root}, nil
}

// Save stores a single normalized ingest payload.
func (s *FileStore) Save(req ArtifactIngestRequest) (string, error) {
	dir := filepath.Join(s.dataDir, req.Organization, req.Project, req.Service, req.RunID)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("create run dir: %w", err)
	}

	record := storedIngestRecord{
		ArtifactIngestRequest: req,
		ReceivedAt:            time.Now().UTC().Format(time.RFC3339),
	}
	body, err := json.MarshalIndent(record, "", "  ")
	if err != nil {
		return "", fmt.Errorf("marshal payload: %w", err)
	}

	target := filepath.Join(dir, "payload.json")
	temp := target + ".tmp"
	if err := os.WriteFile(temp, append(body, '\n'), 0o644); err != nil {
		return "", fmt.Errorf("write payload temp file: %w", err)
	}
	if err := os.Rename(temp, target); err != nil {
		return "", fmt.Errorf("move payload into place: %w", err)
	}

	location := "/" + path.Join("v1", "artifacts", req.Organization, req.Project, req.Service, req.RunID)
	return location, nil
}
