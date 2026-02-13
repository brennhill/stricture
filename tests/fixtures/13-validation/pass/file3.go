// stripe/buggy_client_b01.go — Missing error check after HTTP request.
package stripe

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Client struct {
	apiKey     string
	httpClient *http.Client
	baseURL    string
}

func NewClient(apiKey string) *Client {
	return &Client{
		apiKey:     apiKey,
		httpClient: &http.Client{Timeout: 30 * time.Second},
		baseURL:    "https://api.stripe.com/v1",
	}
}

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

func (c *Client) CreateCharge(req CreateChargeRequest) (*Charge, error) {
	body, _ := json.Marshal(req)
	httpReq, _ := http.NewRequest("POST", c.baseURL+"/charges", bytes.NewReader(body))
	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, _ := c.httpClient.Do(httpReq) // BUG: No error check
	defer resp.Body.Close()              // Will panic if resp is nil

	var charge Charge
	json.NewDecoder(resp.Body).Decode(&charge)
	return &charge, nil
}
