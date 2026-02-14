package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"
)

type flow struct {
	ID          string   `json:"id"`
	Domain      string   `json:"domain"`
	Workflow    string   `json:"workflow"`
	Service     string   `json:"service"`
	Endpoint    string   `json:"endpoint"`
	FieldID     string   `json:"field_id"`
	UseCases    []string `json:"use_cases"`
	Languages   []string `json:"languages"`
	Description string   `json:"description"`
}

type strictureTruth struct {
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

type apiServer struct {
	domain    string
	service   string
	flows     []flow
	flowsByID map[string]flow
	truth     strictureTruth
}

func main() {
	flowFile := strings.TrimSpace(os.Getenv("FLOW_FILE"))
	if flowFile == "" {
		log.Fatal("FLOW_FILE is required")
	}

	domain := strings.TrimSpace(os.Getenv("DOMAIN"))
	if domain == "" {
		log.Fatal("DOMAIN is required")
	}

	service := strings.TrimSpace(os.Getenv("SERVICE_NAME"))
	if service == "" {
		service = titleCase(domain) + "Gateway"
	}

	port := strings.TrimSpace(os.Getenv("PORT"))
	if port == "" {
		port = "18080"
	}

	flows, err := loadFlows(flowFile)
	if err != nil {
		log.Fatalf("load flows: %v", err)
	}

	server := &apiServer{
		domain:    domain,
		service:   service,
		flows:     flows,
		flowsByID: indexFlows(flows),
		truth:     buildStrictureTruth(flows, domain, service, time.Now().UTC()),
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/health", server.handleHealth)
	mux.HandleFunc("/api/v1/flows", server.handleFlows)
	mux.HandleFunc("/api/v1/flows/", server.handleFlowByID)
	mux.HandleFunc("/api/v1/simulate/", server.handleSimulate)
	mux.HandleFunc("/api/v1/use-cases", server.handleUseCases)
	mux.HandleFunc("/api/v1/stricture-truth", server.handleStrictureTruth)

	addr := ":" + port
	log.Printf("fake-api domain=%s service=%s flow_count=%d listen=%s", domain, service, len(flows), addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("listen: %v", err)
	}
}

func loadFlows(path string) ([]flow, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}

	var flows []flow
	if err := json.Unmarshal(data, &flows); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	return flows, nil
}

func indexFlows(flows []flow) map[string]flow {
	index := make(map[string]flow, len(flows))
	for _, flow := range flows {
		index[flow.ID] = flow
	}
	return index
}

func (s *apiServer) handleHealth(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"status":  "ok",
		"service": s.service,
		"domain":  s.domain,
		"asOf":    time.Now().UTC().Format(time.RFC3339),
	})
}

func (s *apiServer) handleFlows(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"service": s.service,
		"domain":  s.domain,
		"flows":   s.flows,
	})
}

func (s *apiServer) handleFlowByID(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/v1/flows/")
	flow, ok := s.flowsByID[id]
	if !ok {
		writeJSON(w, http.StatusNotFound, map[string]any{
			"error": fmt.Sprintf("flow %q not found", id),
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"service": s.service,
		"domain":  s.domain,
		"flow":    flow,
	})
}

func (s *apiServer) handleSimulate(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/v1/simulate/")
	flow, ok := s.flowsByID[id]
	if !ok {
		writeJSON(w, http.StatusNotFound, map[string]any{
			"error": fmt.Sprintf("flow %q not found", id),
		})
		return
	}

	driftType := strings.TrimSpace(r.URL.Query().Get("drift"))
	if driftType == "" {
		driftType = "none"
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"service":        s.service,
		"domain":         s.domain,
		"flowId":         flow.ID,
		"workflow":       flow.Workflow,
		"fieldId":        flow.FieldID,
		"useCases":       flow.UseCases,
		"languages":      flow.Languages,
		"simulatedDrift": driftType,
		"escalationHint": fmt.Sprintf("pagerduty:%s-oncall", s.domain),
		"timestamp":      time.Now().UTC().Format(time.RFC3339),
	})
}

func (s *apiServer) handleUseCases(w http.ResponseWriter, _ *http.Request) {
	counts := map[string]int{}
	for _, flow := range s.flows {
		for _, useCase := range flow.UseCases {
			counts[useCase]++
		}
	}

	keys := make([]string, 0, len(counts))
	for key := range counts {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	summary := make([]map[string]any, 0, len(keys))
	for _, key := range keys {
		summary = append(summary, map[string]any{
			"useCase": key,
			"count":   counts[key],
		})
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"service": s.service,
		"domain":  s.domain,
		"summary": summary,
	})
}

func (s *apiServer) handleStrictureTruth(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, s.truth)
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		log.Printf("write json: %v", err)
	}
}

func titleCase(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	raw = strings.ToLower(raw)
	return strings.ToUpper(raw[:1]) + raw[1:]
}

func buildStrictureTruth(flows []flow, domain string, service string, generatedAt time.Time) strictureTruth {
	signatures := make([]string, 0, len(flows))
	useCaseSet := map[string]bool{}
	languageSet := map[string]bool{}
	annotatedFlows := 0

	for _, candidate := range flows {
		useCases := append([]string{}, candidate.UseCases...)
		languages := append([]string{}, candidate.Languages...)
		sort.Strings(useCases)
		sort.Strings(languages)

		if strings.TrimSpace(candidate.FieldID) != "" && len(useCases) > 0 {
			annotatedFlows++
		}

		for _, useCase := range useCases {
			useCaseSet[useCase] = true
		}
		for _, language := range languages {
			languageSet[language] = true
		}

		signatures = append(signatures,
			fmt.Sprintf("%s|%s|%s|%s|%s",
				candidate.ID,
				candidate.FieldID,
				candidate.Workflow,
				strings.Join(useCases, ","),
				strings.Join(languages, ",")),
		)
	}

	sort.Strings(signatures)
	digest := sha256.Sum256([]byte(strings.Join(signatures, "\n")))

	expectedUseCases := make([]string, 0, len(useCaseSet))
	for useCase := range useCaseSet {
		expectedUseCases = append(expectedUseCases, useCase)
	}
	sort.Strings(expectedUseCases)

	expectedLanguages := make([]string, 0, len(languageSet))
	for language := range languageSet {
		expectedLanguages = append(expectedLanguages, language)
	}
	sort.Strings(expectedLanguages)

	coverage := 100.0
	if len(flows) > 0 {
		coverage = (float64(annotatedFlows) / float64(len(flows))) * 100
	}

	return strictureTruth{
		TruthVersion:          "1",
		GeneratedAt:           generatedAt.Format(time.RFC3339),
		Service:               service,
		Domain:                domain,
		SupportedFlows:        len(flows),
		AnnotatedFlows:        annotatedFlows,
		AnnotationCoveragePct: coverage,
		ExpectedUseCases:      expectedUseCases,
		ExpectedLanguages:     expectedLanguages,
		LineageChecksum:       "sha256:" + hex.EncodeToString(digest[:]),
	}
}
