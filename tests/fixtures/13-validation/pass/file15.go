// stripe/buggy_webhook_b13.go — Webhook handler without signature verification.
package stripe

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type WebhookEvent struct {
	ID   string          `json:"id"`
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

// BUG: No signature verification - accepts forged webhooks
func HandleWebhook(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "read body failed", http.StatusBadRequest)
		return
	}

	// BUG: Missing Stripe-Signature header validation
	// signature := r.Header.Get("Stripe-Signature")
	// Should verify HMAC-SHA256 signature here

	var event WebhookEvent
	if err := json.Unmarshal(body, &event); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	// BUG: Processing event without verifying authenticity
	fmt.Printf("Processing event: %s (%s)\n", event.ID, event.Type)
	w.WriteHeader(http.StatusOK)
}

func ProcessWebhookEvent(payload []byte) (*WebhookEvent, error) {
	var event WebhookEvent
	if err := json.Unmarshal(payload, &event); err != nil {
		return nil, fmt.Errorf("unmarshal event: %w", err)
	}

	// BUG: No signature verification before processing
	if event.ID[:4] != "evt_" {
		return nil, fmt.Errorf("invalid event ID format")
	}

	return &event, nil
}
