package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestHealthz(t *testing.T) {
	handler, err := NewHandler(Config{DataDir: t.TempDir()})
	if err != nil {
		t.Fatalf("NewHandler() error = %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var body map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("json decode response: %v", err)
	}
	if body["status"] != "ok" {
		t.Fatalf("expected status=ok, got %q", body["status"])
	}
}

func TestArtifactsIngestPersistsPayload(t *testing.T) {
	dataDir := t.TempDir()
	handler, err := NewHandler(Config{DataDir: dataDir})
	if err != nil {
		t.Fatalf("NewHandler() error = %v", err)
	}

	reqBody := `{
		"organization": "Acme Inc",
		"project": "Checkout",
		"service": "Gateway",
		"artifact": {"field":"payment.status","source":"PaymentGateway"}
	}`
	req := httptest.NewRequest(http.MethodPost, "/v1/artifacts", bytes.NewBufferString(reqBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusAccepted {
		t.Fatalf("expected 202, got %d body=%s", rec.Code, rec.Body.String())
	}

	var resp ingestResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if !resp.Accepted {
		t.Fatal("expected accepted=true")
	}
	if resp.RunID == "" {
		t.Fatal("expected run_id in response")
	}
	if !strings.Contains(resp.Location, "/v1/artifacts/acme-inc/checkout/gateway/") {
		t.Fatalf("unexpected location %q", resp.Location)
	}

	payloadPath := filepath.Join(dataDir, "acme-inc", "checkout", "gateway", resp.RunID, "payload.json")
	content, err := os.ReadFile(payloadPath)
	if err != nil {
		t.Fatalf("read persisted payload: %v", err)
	}
	var persisted map[string]any
	if err := json.Unmarshal(content, &persisted); err != nil {
		t.Fatalf("decode persisted payload: %v", err)
	}
	if persisted["organization"] != "acme-inc" {
		t.Fatalf("expected sanitized organization in payload, got %v", persisted["organization"])
	}
}

func TestArtifactsIngestRequiresBearerTokenWhenConfigured(t *testing.T) {
	handler, err := NewHandler(Config{
		DataDir:     t.TempDir(),
		IngestToken: "secret-token",
		AuthMode:    "token",
	})
	if err != nil {
		t.Fatalf("NewHandler() error = %v", err)
	}

	body := `{"organization":"acme","project":"checkout","service":"gateway","artifact":{"field":"f"}}`

	reqNoAuth := httptest.NewRequest(http.MethodPost, "/v1/artifacts", bytes.NewBufferString(body))
	recNoAuth := httptest.NewRecorder()
	handler.ServeHTTP(recNoAuth, reqNoAuth)
	if recNoAuth.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 without token, got %d", recNoAuth.Code)
	}

	reqAuth := httptest.NewRequest(http.MethodPost, "/v1/artifacts", bytes.NewBufferString(body))
	reqAuth.Header.Set("Authorization", "Bearer secret-token")
	recAuth := httptest.NewRecorder()
	handler.ServeHTTP(recAuth, reqAuth)
	if recAuth.Code != http.StatusAccepted {
		t.Fatalf("expected 202 with token, got %d body=%s", recAuth.Code, recAuth.Body.String())
	}
}

func TestNewHandlerRejectsUnsupportedStorageDriver(t *testing.T) {
	_, err := NewHandler(Config{
		DataDir:       t.TempDir(),
		StorageDriver: "r2",
	})
	if err == nil {
		t.Fatal("expected error for unsupported storage driver")
	}
}

func TestNewHandlerRejectsTokenModeWithoutToken(t *testing.T) {
	_, err := NewHandler(Config{
		DataDir:  t.TempDir(),
		AuthMode: "token",
	})
	if err == nil {
		t.Fatal("expected error for token auth without token")
	}
}
