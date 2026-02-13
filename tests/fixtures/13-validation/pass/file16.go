// stripe/buggy_pagination_b14.go — Incomplete pagination implementation.
package stripe

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type ChargeList struct {
	Object  string   `json:"object"`
	Data    []Charge `json:"data"`
	HasMore bool     `json:"has_more"` // Indicates more pages exist
}

type Charge struct {
	ID       string `json:"id"`
	Amount   int64  `json:"amount"`
	Currency string `json:"currency"`
	Status   string `json:"status"`
	Created  int64  `json:"created"`
}

// BUG: Fetches only first page, ignores has_more flag
func (c *Client) GetAllCharges() ([]Charge, error) {
	httpReq, err := http.NewRequest("GET", c.baseURL+"/charges?limit=10", nil)
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

	var list ChargeList
	if err := json.NewDecoder(resp.Body).Decode(&list); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	// BUG: Returns first page only, ignoring list.HasMore
	// Should loop while HasMore == true using starting_after cursor
	return list.Data, nil
}
