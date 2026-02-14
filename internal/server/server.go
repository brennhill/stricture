package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

const maxIngestBodyBytes = 10 << 20 // 10MB

// IngestStore persists normalized ingest payloads.
type IngestStore interface {
	Save(req ArtifactIngestRequest) (string, error)
}

// App handles the HTTP API for stricture-server.
type App struct {
	cfg   Config
	store IngestStore
}

// New constructs the production HTTP server.
func New(cfg Config) (*http.Server, error) {
	handler, err := NewHandler(cfg)
	if err != nil {
		return nil, err
	}
	return &http.Server{
		Addr:    cfg.Addr,
		Handler: handler,
	}, nil
}

// NewHandler constructs the HTTP handler for tests and local embedding.
func NewHandler(cfg Config) (http.Handler, error) {
	if err := validateConfig(cfg); err != nil {
		return nil, err
	}

	store, err := newStore(cfg)
	if err != nil {
		return nil, err
	}

	app := &App{
		cfg:   cfg,
		store: store,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", app.handleHealthz)
	mux.HandleFunc("POST /v1/artifacts", app.handleArtifactsIngest)
	return mux, nil
}

func validateConfig(cfg Config) error {
	switch cfg.AuthMode {
	case "", "none":
	case "token":
		if strings.TrimSpace(cfg.IngestToken) == "" {
			return fmt.Errorf("STRICTURE_SERVER_AUTH_MODE=token requires STRICTURE_SERVER_INGEST_TOKEN")
		}
	default:
		return fmt.Errorf("unsupported auth mode %q", cfg.AuthMode)
	}
	return nil
}

func newStore(cfg Config) (IngestStore, error) {
	switch cfg.StorageDriver {
	case "", "fs":
		return NewFileStore(cfg.DataDir)
	case "s3", "r2":
		return nil, fmt.Errorf("storage driver %q is not implemented in this build; use STRICTURE_SERVER_STORAGE_DRIVER=fs and configure external replication", cfg.StorageDriver)
	default:
		return nil, fmt.Errorf("unsupported storage driver %q", cfg.StorageDriver)
	}
}

func (a *App) handleHealthz(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (a *App) handleArtifactsIngest(w http.ResponseWriter, r *http.Request) {
	if !a.isAuthorized(r) {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxIngestBodyBytes)

	var req ArtifactIngestRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": fmt.Sprintf("invalid request body: %v", err)})
		return
	}
	if err := decoder.Decode(&struct{}{}); err == nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "request body must contain a single JSON object"})
		return
	}

	normalized, err := normalizeAndValidateIngest(req)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	location, err := a.store.Save(normalized)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": fmt.Sprintf("persist ingest: %v", err)})
		return
	}

	writeJSON(w, http.StatusAccepted, ingestResponse{
		Accepted: true,
		RunID:    normalized.RunID,
		Location: location,
	})
}

func (a *App) isAuthorized(r *http.Request) bool {
	switch a.cfg.AuthMode {
	case "", "none":
		return true
	case "token":
		auth := strings.TrimSpace(r.Header.Get("Authorization"))
		const prefix = "Bearer "
		if !strings.HasPrefix(auth, prefix) {
			return false
		}
		return strings.TrimSpace(strings.TrimPrefix(auth, prefix)) == a.cfg.IngestToken
	default:
		return false
	}
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"encode response: %v"}`, err), http.StatusInternalServerError)
	}
}
