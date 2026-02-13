// stripe/buggy_client_b10.go — Missing format validation on ID fields.
package stripe

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type Charge struct {
	ID       string `json:"id"` // Manifest: format "ch_*"
	Amount   int64  `json:"amount"`
	Currency string `json:"currency"`
	Status   string `json:"status"`
	Created  int64  `json:"created"`
}

// BUG: No validation of ID format (should start with "ch_")
func (c *Client) GetCharge(id string) (*Charge, error) {
	if id == "" {
		return nil, fmt.Errorf("charge ID is required")
	}
	// BUG: Missing format validation - manifest requires "ch_*" prefix

	httpReq, err := http.NewRequest("GET", c.baseURL+"/charges/"+id, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)

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

	// BUG: Also missing validation on response ID
	return &charge, nil
}
