package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
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
