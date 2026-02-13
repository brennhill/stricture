// stripe/buggy_idempotency_b15.go — Missing idempotency key for charge creation.
package stripe

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type CreateChargeRequest struct {
	Amount   int64  `json:"amount"`
	Currency string `json:"currency"`
	Source   string `json:"source"`
}

type Charge struct {
	ID       string `json:"id"`
	Amount   int64  `json:"amount"`
	Currency string `json:"currency"`
	Status   string `json:"status"`
	Created  int64  `json:"created"`
}

// BUG: No idempotency key - network timeout causes duplicate charges
func (c *Client) CreateCharge(req CreateChargeRequest) (*Charge, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequest("POST", c.baseURL+"/charges", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)
	httpReq.Header.Set("Content-Type", "application/json")
	// BUG: Missing Idempotency-Key header
	// Should set: httpReq.Header.Set("Idempotency-Key", uniqueKey)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("stripe API error: status %d", resp.StatusCode)
	}

	var charge Charge
	if err := json.NewDecoder(resp.Body).Decode(&charge); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &charge, nil
}

// BUG: Retry logic without idempotency key causes duplicate charges
func (c *Client) CreateChargeWithRetry(req CreateChargeRequest) (*Charge, error) {
	var lastErr error
	for attempt := 0; attempt < 3; attempt++ {
		charge, err := c.CreateCharge(req)
		if err == nil {
			return charge, nil
		}
		lastErr = err
		time.Sleep(time.Second * time.Duration(attempt+1))
	}
	return nil, fmt.Errorf("all retries failed: %w", lastErr)
}
