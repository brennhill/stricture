// stripe/buggy_client_b06.go — Response struct doesn't match manifest.
package stripe

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// BUG: Manifest specifies `amount_refunded: integer`, this has wrong field
type Charge struct {
	ID       string `json:"id"`
	Amount   int64  `json:"amount"`
	Currency string `json:"currency"`
	Status   string `json:"status"`
	Refunded bool   `json:"refunded"` // BUG: Should be amount_refunded int64
	Created  int64  `json:"created"`
}

func (c *Client) GetCharge(id string) (*Charge, error) {
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

	return &charge, nil
}
