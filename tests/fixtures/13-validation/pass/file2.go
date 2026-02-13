// stripe/webhook.go — Webhook signature verification using HMAC-SHA256.
package stripe

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// WebhookVerifier handles webhook signature verification.
type WebhookVerifier struct {
	secret string
}

// NewWebhookVerifier creates a webhook verifier with the given secret.
func NewWebhookVerifier(secret string) *WebhookVerifier {
	return &WebhookVerifier{secret: secret}
}

// VerifySignature verifies the Stripe-Signature header.
func (v *WebhookVerifier) VerifySignature(payload []byte, header string) error {
	if header == "" {
		return fmt.Errorf("missing Stripe-Signature header")
	}

	parts := strings.Split(header, ",")
	var timestamp int64
	var signature string

	for _, part := range parts {
		kv := strings.SplitN(part, "=", 2)
		if len(kv) != 2 {
			continue
		}
		key := strings.TrimSpace(kv[0])
		value := strings.TrimSpace(kv[1])

		switch key {
		case "t":
			ts, err := strconv.ParseInt(value, 10, 64)
			if err != nil {
				return fmt.Errorf("invalid timestamp: %w", err)
			}
			timestamp = ts
		case "v1":
			signature = value
		}
	}

	if timestamp == 0 {
		return fmt.Errorf("missing timestamp in signature header")
	}
	if signature == "" {
		return fmt.Errorf("missing signature in header")
	}

	// Check timestamp tolerance (5 minutes)
	now := time.Now().Unix()
	if now-timestamp > 300 {
		return fmt.Errorf("timestamp too old: %d seconds", now-timestamp)
	}

	// Compute expected signature
	signedPayload := fmt.Sprintf("%d.%s", timestamp, string(payload))
	mac := hmac.New(sha256.New, []byte(v.secret))
	mac.Write([]byte(signedPayload))
	expected := hex.EncodeToString(mac.Sum(nil))

	if !hmac.Equal([]byte(signature), []byte(expected)) {
		return fmt.Errorf("signature mismatch")
	}

	return nil
}

// ParseEvent parses and verifies a webhook event.
func (v *WebhookVerifier) ParseEvent(payload []byte, header string) (*WebhookEvent, error) {
	if err := v.VerifySignature(payload, header); err != nil {
		return nil, fmt.Errorf("verify signature: %w", err)
	}

	var event WebhookEvent
	if err := json.Unmarshal(payload, &event); err != nil {
		return nil, fmt.Errorf("unmarshal event: %w", err)
	}

	if event.ID == "" || event.ID[:4] != "evt_" {
		return nil, fmt.Errorf("invalid event ID: %s", event.ID)
	}

	return &event, nil
}
