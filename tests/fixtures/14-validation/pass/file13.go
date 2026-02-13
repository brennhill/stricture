// github_client_b13.go — Missing webhook signature verification
package github

import (
	"encoding/json"
	"io"
	"net/http"
)

type WebhookPayload struct {
	Action     string     `json:"action"`
	Issue      *Issue     `json:"issue,omitempty"`
	Repository Repository `json:"repository"`
}

func HandleWebhook(w http.ResponseWriter, r *http.Request) {
	// BUG: No signature verification - anyone can send fake webhooks

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read body", http.StatusBadRequest)
		return
	}

	var payload WebhookPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	// BUG: Processing webhook without verifying it came from GitHub
	// Attacker can send POST to webhook endpoint with fake data

	w.WriteHeader(http.StatusOK)
}

// Correct implementation requires:
// 1. Read X-Hub-Signature-256 header
// 2. Compute HMAC-SHA256 of body using webhook secret
// 3. Compare computed signature with header value
// 4. Reject if signatures don't match

// Example attack:
// curl -X POST https://myapp.com/webhook \
//   -H "Content-Type: application/json" \
//   -d '{"action":"closed","issue":{"id":999,"title":"Fake"}}'
//
// Without signature verification, this fake webhook is accepted
