package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func testServer() *apiServer {
	flows := []flow{
		{
			ID:       "logistics_01_shipment_eta_projection",
			Domain:   "logistics",
			Workflow: "shipment_eta_projection",
			Service:  "LogisticsGateway",
			Endpoint: "GET /api/v1/logistics/shipment-eta-projection",
			FieldID:  "response_logistics_shipment_eta_projection",
			UseCases: []string{"drift_blocking", "escalation_chain"},
			Languages: []string{
				"go", "typescript",
			},
			Description: "Sample logistics flow",
		},
	}

	return &apiServer{
		domain:    "logistics",
		service:   "LogisticsGateway",
		flows:     flows,
		flowsByID: indexFlows(flows),
		truth:     buildStrictureTruth(flows, "logistics", "LogisticsGateway", time.Date(2026, time.February, 14, 0, 0, 0, 0, time.UTC)),
	}
}

func TestHandleHealth(t *testing.T) {
	server := testServer()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rr := httptest.NewRecorder()

	server.handleHealth(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
	}

	var payload map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload["service"] != "LogisticsGateway" {
		t.Fatalf("service = %v, want LogisticsGateway", payload["service"])
	}
	if _, ok := payload["flowCount"]; ok {
		t.Fatalf("flowCount should not be present in /health payload")
	}
}

func TestHandleFlowByID(t *testing.T) {
	server := testServer()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/flows/logistics_01_shipment_eta_projection", nil)
	rr := httptest.NewRecorder()
	server.handleFlowByID(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
	}

	notFoundReq := httptest.NewRequest(http.MethodGet, "/api/v1/flows/unknown", nil)
	notFoundRR := httptest.NewRecorder()
	server.handleFlowByID(notFoundRR, notFoundReq)
	if notFoundRR.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", notFoundRR.Code, http.StatusNotFound)
	}
}

func TestHandleSimulate(t *testing.T) {
	server := testServer()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/simulate/logistics_01_shipment_eta_projection?drift=source_version_changed", nil)
	rr := httptest.NewRecorder()

	server.handleSimulate(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
	}

	var payload map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload["simulatedDrift"] != "source_version_changed" {
		t.Fatalf("simulatedDrift = %v, want source_version_changed", payload["simulatedDrift"])
	}
}

func TestHandleFlows(t *testing.T) {
	server := testServer()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/flows", nil)
	rr := httptest.NewRecorder()

	server.handleFlows(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
	}

	var payload struct {
		Service string `json:"service"`
		Domain  string `json:"domain"`
		Flows   []flow `json:"flows"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload.Service != "LogisticsGateway" {
		t.Fatalf("service = %q, want LogisticsGateway", payload.Service)
	}
	if payload.Domain != "logistics" {
		t.Fatalf("domain = %q, want logistics", payload.Domain)
	}
	if len(payload.Flows) != 1 {
		t.Fatalf("flows len = %d, want 1", len(payload.Flows))
	}
}

func TestHandleUseCases(t *testing.T) {
	server := testServer()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/use-cases", nil)
	rr := httptest.NewRecorder()

	server.handleUseCases(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
	}

	var payload struct {
		Summary []struct {
			UseCase string `json:"useCase"`
			Count   int    `json:"count"`
		} `json:"summary"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(payload.Summary) != 2 {
		t.Fatalf("summary len = %d, want 2", len(payload.Summary))
	}
	if payload.Summary[0].UseCase != "drift_blocking" || payload.Summary[0].Count != 1 {
		t.Fatalf("summary[0] = %+v, want drift_blocking=1", payload.Summary[0])
	}
	if payload.Summary[1].UseCase != "escalation_chain" || payload.Summary[1].Count != 1 {
		t.Fatalf("summary[1] = %+v, want escalation_chain=1", payload.Summary[1])
	}
}

func TestHandleStrictureTruth(t *testing.T) {
	server := testServer()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/stricture-truth", nil)
	rr := httptest.NewRecorder()

	server.handleStrictureTruth(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
	}

	var payload struct {
		TruthVersion          string   `json:"truthVersion"`
		GeneratedAt           string   `json:"generatedAt"`
		Service               string   `json:"service"`
		Domain                string   `json:"domain"`
		SupportedFlows        int      `json:"supportedFlows"`
		AnnotatedFlows        int      `json:"annotatedFlows"`
		AnnotationCoveragePct float64  `json:"annotationCoveragePct"`
		ExpectedUseCases      []string `json:"expectedUseCases"`
		ExpectedLanguages     []string `json:"expectedLanguages"`
		LineageChecksum       string   `json:"lineageChecksum"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if payload.TruthVersion != "1" {
		t.Fatalf("truthVersion = %q, want 1", payload.TruthVersion)
	}
	if payload.Service != "LogisticsGateway" {
		t.Fatalf("service = %q, want LogisticsGateway", payload.Service)
	}
	if payload.Domain != "logistics" {
		t.Fatalf("domain = %q, want logistics", payload.Domain)
	}
	if payload.SupportedFlows != 1 {
		t.Fatalf("supportedFlows = %d, want 1", payload.SupportedFlows)
	}
	if payload.AnnotatedFlows != 1 {
		t.Fatalf("annotatedFlows = %d, want 1", payload.AnnotatedFlows)
	}
	if payload.AnnotationCoveragePct != 100 {
		t.Fatalf("annotationCoveragePct = %v, want 100", payload.AnnotationCoveragePct)
	}
	if len(payload.ExpectedUseCases) != 2 {
		t.Fatalf("expectedUseCases len = %d, want 2", len(payload.ExpectedUseCases))
	}
	if len(payload.ExpectedLanguages) != 2 {
		t.Fatalf("expectedLanguages len = %d, want 2", len(payload.ExpectedLanguages))
	}
	if payload.GeneratedAt == "" {
		t.Fatalf("generatedAt should not be empty")
	}
	if !strings.HasPrefix(payload.LineageChecksum, "sha256:") {
		t.Fatalf("lineageChecksum = %q, want prefix sha256:", payload.LineageChecksum)
	}
}

func TestLoadFlows(t *testing.T) {
	t.Run("missing file", func(t *testing.T) {
		if _, err := loadFlows("does-not-exist.json"); err == nil {
			t.Fatalf("expected error for missing file")
		}
	})

	t.Run("invalid json", func(t *testing.T) {
		tmpDir := t.TempDir()
		path := filepath.Join(tmpDir, "flows.json")
		if err := os.WriteFile(path, []byte("{not-json"), 0o644); err != nil {
			t.Fatalf("write invalid json file: %v", err)
		}

		if _, err := loadFlows(path); err == nil {
			t.Fatalf("expected parse error")
		}
	})
}
