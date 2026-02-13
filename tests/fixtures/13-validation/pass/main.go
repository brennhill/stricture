// stripe/client.go — Stripe API client with proper error handling and struct tags.
package stripe

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client handles Stripe API requests.
type Client struct {
	apiKey     string
	httpClient *http.Client
	baseURL    string
}

// NewClient creates a Stripe client with the given API key.
func NewClient(apiKey string) *Client {
	return &Client{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL: "https://api.stripe.com/v1",
	}
}

// Charge represents a Stripe charge object.
type Charge struct {
	ID          string `json:"id"`
	Amount      int64  `json:"amount"`
	Currency    string `json:"currency"`
	Status      string `json:"status"`
	Description string `json:"description,omitempty"`
	Created     int64  `json:"created"`
}

// Customer represents a Stripe customer object.
type Customer struct {
	ID          string `json:"id"`
	Email       string `json:"email,omitempty"`
	Description string `json:"description,omitempty"`
	Created     int64  `json:"created"`
}

// ChargeList represents a paginated list of charges.
type ChargeList struct {
	Object  string   `json:"object"`
	Data    []Charge `json:"data"`
	HasMore bool     `json:"has_more"`
}

// CreateChargeRequest represents the request to create a charge.
type CreateChargeRequest struct {
	Amount      int64  `json:"amount"`
	Currency    string `json:"currency"`
	Source      string `json:"source"`
	Description string `json:"description,omitempty"`
}

// CreateCustomerRequest represents the request to create a customer.
type CreateCustomerRequest struct {
	Email       string `json:"email,omitempty"`
	Description string `json:"description,omitempty"`
	Source      string `json:"source,omitempty"`
}

// WebhookEvent represents a Stripe webhook event.
type WebhookEvent struct {
	ID   string          `json:"id"`
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

// CreateCharge creates a new charge.
func (c *Client) CreateCharge(req CreateChargeRequest) (*Charge, error) {
	if req.Amount < 50 {
		return nil, fmt.Errorf("amount must be at least 50 cents")
	}
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
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("stripe API error: status %d, body: %s", resp.StatusCode, string(body))
	}

	var charge Charge
	if err := json.NewDecoder(resp.Body).Decode(&charge); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	if charge.ID == "" || charge.ID[:3] != "ch_" {
		return nil, fmt.Errorf("invalid charge ID: %s", charge.ID)
	}

	return &charge, nil
}

// GetCharge retrieves a charge by ID.
func (c *Client) GetCharge(id string) (*Charge, error) {
	if id == "" {
		return nil, fmt.Errorf("charge ID is required")
	}
	if len(id) < 4 || id[:3] != "ch_" {
		return nil, fmt.Errorf("invalid charge ID format: %s", id)
	}

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
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("stripe API error: status %d, body: %s", resp.StatusCode, string(body))
	}

	var charge Charge
	if err := json.NewDecoder(resp.Body).Decode(&charge); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &charge, nil
}

// ListCharges retrieves a paginated list of charges.
func (c *Client) ListCharges(limit int, startingAfter string) (*ChargeList, error) {
	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		return nil, fmt.Errorf("limit cannot exceed 100")
	}

	url := fmt.Sprintf("%s/charges?limit=%d", c.baseURL, limit)
	if startingAfter != "" {
		url += "&starting_after=" + startingAfter
	}

	httpReq, err := http.NewRequest("GET", url, nil)
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
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("stripe API error: status %d, body: %s", resp.StatusCode, string(body))
	}

	var list ChargeList
	if err := json.NewDecoder(resp.Body).Decode(&list); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	if list.Object != "list" {
		return nil, fmt.Errorf("unexpected object type: %s", list.Object)
	}

	return &list, nil
}

// CreateCustomer creates a new customer.
func (c *Client) CreateCustomer(req CreateCustomerRequest) (*Customer, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequest("POST", c.baseURL+"/customers", bytes.NewReader(body))
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
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("stripe API error: status %d, body: %s", resp.StatusCode, string(body))
	}

	var customer Customer
	if err := json.NewDecoder(resp.Body).Decode(&customer); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	if customer.ID == "" || customer.ID[:4] != "cus_" {
		return nil, fmt.Errorf("invalid customer ID: %s", customer.ID)
	}

	return &customer, nil
}
