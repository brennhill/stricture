// stripe/buggy_client_b09.go — Missing range validation on amount.
package stripe

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
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

// BUG: No validation of amount range [50, 99999999]
func (c *Client) CreateCharge(req CreateChargeRequest) (*Charge, error) {
	// BUG: Missing validation - manifest requires amount >= 50
	if req.Currency == "" {
		return nil, fmt.Errorf("currency is required")
	}
	if req.Source == "" {
		return nil, fmt.Errorf("source is required")
	}

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
