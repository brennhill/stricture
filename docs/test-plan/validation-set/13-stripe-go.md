# 13 — Stripe Payments API (Go)

**Why included:** Go struct json tags, `if err != nil`, table-driven tests, `testify/assert`. Exercises CTR-json-tag-match, error handling patterns, and API client architecture.

## Manifest Fragment

```yaml
contracts:
  stripe_payments:
    base_url: https://api.stripe.com/v1
    auth: bearer_token

    operations:
      create_charge:
        method: POST
        path: /charges
        request:
          amount: integer  # cents, min 50
          currency: string  # lowercase ISO 4217
          source: string  # tok_* or card_*
          description: string?
        response:
          id: string  # ch_ prefix
          amount: integer
          currency: string
          status: enum[succeeded, pending, failed]
          created: integer  # unix timestamp

      get_charge:
        method: GET
        path: /charges/{id}
        response:
          id: string
          amount: integer
          currency: string
          status: enum[succeeded, pending, failed]

      list_charges:
        method: GET
        path: /charges
        query:
          limit: integer?  # max 100, default 10
          starting_after: string?  # cursor
        response:
          object: "list"
          data: array[charge]
          has_more: boolean

      create_customer:
        method: POST
        path: /customers
        request:
          email: string?
          description: string?
          source: string?
        response:
          id: string  # cus_ prefix
          email: string?
          created: integer

      webhook_event:
        signature_header: Stripe-Signature
        signature_algo: HMAC-SHA256
        payload:
          id: string  # evt_ prefix
          type: string  # charge.succeeded, etc
          data:
            object: charge | customer
```

---

## PERFECT — Zero Violations Expected

### stripe/client.go

```go
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
```

### stripe/webhook.go

```go
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
```

### stripe/client_test.go

```go
// stripe/client_test.go — Table-driven tests for Stripe client.
package stripe

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateCharge(t *testing.T) {
	tests := []struct {
		name           string
		request        CreateChargeRequest
		mockStatus     int
		mockResponse   string
		expectedError  string
		expectedCharge *Charge
	}{
		{
			name: "successful charge",
			request: CreateChargeRequest{
				Amount:      1000,
				Currency:    "usd",
				Source:      "tok_visa",
				Description: "Test charge",
			},
			mockStatus: http.StatusOK,
			mockResponse: `{
				"id": "ch_1234567890",
				"amount": 1000,
				"currency": "usd",
				"status": "succeeded",
				"description": "Test charge",
				"created": 1234567890
			}`,
			expectedCharge: &Charge{
				ID:          "ch_1234567890",
				Amount:      1000,
				Currency:    "usd",
				Status:      "succeeded",
				Description: "Test charge",
				Created:     1234567890,
			},
		},
		{
			name: "amount below minimum",
			request: CreateChargeRequest{
				Amount:   49,
				Currency: "usd",
				Source:   "tok_visa",
			},
			expectedError: "amount must be at least 50 cents",
		},
		{
			name: "missing currency",
			request: CreateChargeRequest{
				Amount: 1000,
				Source: "tok_visa",
			},
			expectedError: "currency is required",
		},
		{
			name: "missing source",
			request: CreateChargeRequest{
				Amount:   1000,
				Currency: "usd",
			},
			expectedError: "source is required",
		},
		{
			name: "API error",
			request: CreateChargeRequest{
				Amount:   1000,
				Currency: "usd",
				Source:   "tok_visa",
			},
			mockStatus:    http.StatusPaymentRequired,
			mockResponse:  `{"error": "card_declined"}`,
			expectedError: "stripe API error: status 402",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "POST", r.Method)
				assert.Equal(t, "/charges", r.URL.Path)
				assert.Equal(t, "Bearer test_key", r.Header.Get("Authorization"))
				assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

				if tt.mockStatus != 0 {
					w.WriteHeader(tt.mockStatus)
					w.Write([]byte(tt.mockResponse))
				}
			}))
			defer server.Close()

			client := NewClient("test_key")
			client.baseURL = server.URL

			charge, err := client.CreateCharge(tt.request)

			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, charge)
			} else {
				require.NoError(t, err)
				require.NotNil(t, charge)
				assert.Equal(t, tt.expectedCharge.ID, charge.ID)
				assert.Equal(t, tt.expectedCharge.Amount, charge.Amount)
				assert.Equal(t, tt.expectedCharge.Currency, charge.Currency)
				assert.Equal(t, tt.expectedCharge.Status, charge.Status)
				assert.Equal(t, tt.expectedCharge.Description, charge.Description)
				assert.Equal(t, tt.expectedCharge.Created, charge.Created)
			}
		})
	}
}

```

---

## Bug Cases (B01-B15)

### B01 — No Error Handling (TQ-error-path-coverage)

**Bug:** Missing `if err != nil` check after `http.Client.Do()` call, allowing nil pointer dereference on failed requests.
**Expected violation:** `TQ-error-path-coverage` on line 24

```go
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
```

**Why Stricture catches this:** The `http.Client.Do()` call can return an error (network failure, timeout, DNS resolution failure), but the code ignores it. Stricture's `TQ-error-path-coverage` rule detects function calls that return `(T, error)` where the error is not checked before using the returned value.

---

### B02 — No Status Code Check (CTR-status-code-handling)

**Bug:** Doesn't check `resp.StatusCode` before decoding, accepting 4xx/5xx errors as successful responses.
**Expected violation:** `CTR-status-code-handling` on line 31

```go
// stripe/buggy_client_b02.go — No HTTP status code validation.
package stripe

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

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

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("execute request: %w", err)
	}
	defer resp.Body.Close()

	// BUG: No status code check - decodes 402, 404, 500 as success
	var charge Charge
	if err := json.NewDecoder(resp.Body).Decode(&charge); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &charge, nil
}
```

**Why Stricture catches this:** Per the manifest, `/charges` can return status codes `[200, 400, 401, 402, 404, 429, 500]`. Stricture's `CTR-status-code-handling` rule verifies that after `http.Client.Do()`, the code checks `resp.StatusCode` against the allowed values before processing the body.

---

### B03 — Shallow Test Assertions (TQ-no-shallow-assertions)

**Bug:** Test only checks `assert.NotNil(t, charge)` instead of validating individual fields.
**Expected violation:** `TQ-no-shallow-assertions` on line 18

```go
// stripe/buggy_client_test_b03.go — Shallow assertion in test.
package stripe

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateCharge_Shallow(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":"ch_123","amount":1000,"currency":"usd","status":"succeeded","created":1234567890}`))
	}))
	defer server.Close()

	client := NewClient("test_key")
	client.baseURL = server.URL

	req := CreateChargeRequest{Amount: 1000, Currency: "usd", Source: "tok_visa"}
	charge, err := client.CreateCharge(req)

	require.NoError(t, err)
	assert.NotNil(t, charge) // BUG: Shallow assertion - doesn't validate fields

	// Missing: charge.ID == "ch_123", charge.Amount == 1000, etc.
}
```

**Why Stricture catches this:** The manifest specifies exact field requirements for the charge response (id format "ch_*", amount range, currency enum, status enum). Stricture's `TQ-no-shallow-assertions` rule requires tests to validate at least the required fields from the contract, not just check for non-nil.

---

### B04 — Missing Negative Tests (TQ-negative-cases)

**Bug:** Test suite only covers happy path, no tests for 402 card_declined, 404 not found, missing fields.
**Expected violation:** `TQ-negative-cases` on test suite

```go
// stripe/buggy_client_test_b04.go — Only happy path tests.
package stripe

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// BUG: No negative test cases for:
// - 402 Payment Required (card declined)
// - 404 Not Found
// - 400 Invalid Request (missing currency)
// - 429 Rate Limit
// - Network timeout
// - Malformed JSON response

func TestCreateCharge_HappyPath(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":"ch_123","amount":1000,"currency":"usd","status":"succeeded","created":1234567890}`))
	}))
	defer server.Close()

	client := NewClient("test_key")
	client.baseURL = server.URL

	req := CreateChargeRequest{Amount: 1000, Currency: "usd", Source: "tok_visa"}
	charge, err := client.CreateCharge(req)

	require.NoError(t, err)
	require.NotNil(t, charge)
	assert.Equal(t, "ch_123", charge.ID)
	assert.Equal(t, int64(1000), charge.Amount)
	assert.Equal(t, "usd", charge.Currency)
	assert.Equal(t, "succeeded", charge.Status)
}

func TestGetCharge_HappyPath(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":"ch_123","amount":1000,"currency":"usd","status":"succeeded","created":1234567890}`))
	}))
	defer server.Close()

	client := NewClient("test_key")
	client.baseURL = server.URL

	charge, err := client.GetCharge("ch_123")

	require.NoError(t, err)
	assert.Equal(t, "ch_123", charge.ID)
}
```

**Why Stricture catches this:** The manifest lists 7 possible status codes `[200, 400, 401, 402, 404, 429, 500]`. Stricture's `TQ-negative-cases` rule requires tests to cover at least one error status code from the contract's error space to prove error handling works.

---

### B05 — Request Missing Required Fields (CTR-request-shape)

**Bug:** `CreateChargeRequest` struct omits the required `currency` field, violating the manifest.
**Expected violation:** `CTR-request-shape` on line 8

```go
// stripe/buggy_client_b05.go — Missing required field in request struct.
package stripe

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

// BUG: Missing `currency` field, which is required per manifest
type CreateChargeRequest struct {
	Amount int64  `json:"amount"`
	Source string `json:"source"`
	// currency field is MISSING (required in manifest)
}

type Charge struct {
	ID       string `json:"id"`
	Amount   int64  `json:"amount"`
	Currency string `json:"currency"`
	Status   string `json:"status"`
	Created  int64  `json:"created"`
}

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
```

**Why Stricture catches this:** The manifest specifies `currency: { type: enum, required: true }` for the create_charge operation. Stricture's `CTR-request-shape` rule verifies that all required fields from the manifest appear in the request struct with correct json tags.

---

### B06 — Response Type Mismatch (CTR-response-shape)

**Bug:** `Charge` struct has `Refunded bool` instead of `amount_refunded` integer field from the manifest.
**Expected violation:** `CTR-response-shape` on line 11

```go
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
```

**Why Stricture catches this:** The manifest defines the response shape with `amount_refunded: { type: integer, range: [0, 99999999], required: true }`. Stricture's `CTR-response-shape` rule verifies that the Go struct's json tags match the manifest's field names and types.

---

### B07 — Wrong Field Types (CTR-manifest-conformance)

**Bug:** `Amount` is `float64` instead of `int64`, causing precision loss for currency cents.
**Expected violation:** `CTR-manifest-conformance` on line 10

```go
// stripe/buggy_client_b07.go — Wrong type for monetary amount.
package stripe

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// BUG: amount should be int64 (cents), not float64
type Charge struct {
	ID       string  `json:"id"`
	Amount   float64 `json:"amount"` // BUG: Manifest specifies integer
	Currency string  `json:"currency"`
	Status   string  `json:"status"`
	Created  int64   `json:"created"`
}

type CreateChargeRequest struct {
	Amount   float64 `json:"amount"` // BUG: Also wrong here
	Currency string  `json:"currency"`
	Source   string  `json:"source"`
}

func (c *Client) CreateCharge(req CreateChargeRequest) (*Charge, error) {
	// Validation will fail because 10.50 dollars becomes 1050 cents incorrectly
	if req.Amount < 0.50 {
		return nil, fmt.Errorf("amount must be at least $0.50")
	}

	// Rest of implementation...
	return &Charge{
		ID:       "ch_123",
		Amount:   req.Amount,
		Currency: req.Currency,
		Status:   "succeeded",
		Created:  1234567890,
	}, nil
}
```

**Why Stricture catches this:** The manifest specifies `amount: { type: integer, range: [50, 99999999] }` indicating cents as an integer. Using `float64` violates `CTR-manifest-conformance`, which ensures field types match the contract (integer, string, boolean, enum).

---

### B08 — Incomplete Enum Handling (CTR-strictness-parity)

**Bug:** Switch statement handles only 2 of 3 charge statuses (missing "pending"), no default case.
**Expected violation:** `CTR-strictness-parity` on line 15

```go
// stripe/buggy_client_b08.go — Incomplete enum handling in business logic.
package stripe

import (
	"fmt"
)

type Charge struct {
	ID       string `json:"id"`
	Amount   int64  `json:"amount"`
	Currency string `json:"currency"`
	Status   string `json:"status"` // Enum: succeeded, pending, failed
	Created  int64  `json:"created"`
}

// BUG: Only handles 2 of 3 status values from manifest
func (c *Charge) IsComplete() bool {
	switch c.Status {
	case "succeeded":
		return true
	case "failed":
		return true
	// BUG: Missing "pending" case
	// BUG: No default case for unknown values
	}
	return false
}

func (c *Charge) GetDisplayStatus() string {
	switch c.Status {
	case "succeeded":
		return "Payment successful"
	case "failed":
		return "Payment failed"
	// BUG: "pending" returns empty string
	}
	return ""
}

func ProcessCharge(charge *Charge) error {
	if charge.Status == "succeeded" {
		fmt.Println("Charge succeeded, send confirmation")
		return nil
	}
	if charge.Status == "failed" {
		fmt.Println("Charge failed, notify user")
		return fmt.Errorf("charge failed")
	}
	// BUG: "pending" status falls through with no handling
	return nil
}
```

**Why Stricture catches this:** The manifest defines `status: enum[succeeded, pending, failed]` with exactly 3 values. Stricture's `CTR-strictness-parity` rule verifies that enum handling covers all possible values from the contract, typically via exhaustive switch statements with default cases.

---

### B09 — Missing Range Validation (CTR-strictness-parity)

**Bug:** No bounds check on amount field, allowing values below 50 cents minimum.
**Expected violation:** `CTR-strictness-parity` on line 23

```go
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
```

**Why Stricture catches this:** The manifest specifies `amount: { type: integer, range: [50, 99999999], required: true }`. Stricture's `CTR-strictness-parity` rule requires client-side validation of range constraints before making the API call.

---

### B10 — Format Not Validated (CTR-strictness-parity)

**Bug:** Accepts any string for charge ID, no "ch_" prefix validation.
**Expected violation:** `CTR-strictness-parity` on line 15

```go
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
```

**Why Stricture catches this:** The manifest specifies `id: { type: string, format: "ch_*", required: true }`. Stricture's `CTR-strictness-parity` rule enforces format constraints (regex/prefix patterns) on both request parameters and response validation.

---

### B11 — Precision Loss on Currency (CTR-strictness-parity)

**Bug:** Uses float64 for intermediate calculations on monetary amounts, causing rounding errors.
**Expected violation:** `CTR-strictness-parity` on line 18

```go
// stripe/buggy_client_b11.go — Precision loss in currency calculations.
package stripe

import (
	"fmt"
)

type Charge struct {
	ID       string `json:"id"`
	Amount   int64  `json:"amount"` // Cents as integer
	Currency string `json:"currency"`
	Status   string `json:"status"`
	Created  int64  `json:"created"`
}

// BUG: Converts to float64 for calculations, loses precision
func CalculateRefund(charge *Charge, percentage float64) int64 {
	// BUG: Intermediate float64 calculation causes rounding errors
	amountFloat := float64(charge.Amount)
	refundFloat := amountFloat * percentage
	refundCents := int64(refundFloat)
	return refundCents
}

func ApplyDiscount(amount int64, discountPercent float64) int64 {
	// BUG: Converting to float for discount calculation
	discounted := float64(amount) * (1.0 - discountPercent)
	return int64(discounted) // Truncation, not proper rounding
}

func SplitCharge(charge *Charge, numParts int) []int64 {
	parts := make([]int64, numParts)
	// BUG: Division via float64 causes precision loss
	perPart := float64(charge.Amount) / float64(numParts)
	for i := 0; i < numParts; i++ {
		parts[i] = int64(perPart)
	}
	return parts
}

func FormatAmount(cents int64, currency string) string {
	// BUG: Uses float division for display
	dollars := float64(cents) / 100.0
	return fmt.Sprintf("%.2f %s", dollars, currency)
}
```

**Why Stricture catches this:** The manifest specifies `amount: { type: integer }` for monetary values (cents). Stricture's `CTR-strictness-parity` rule flags float64 arithmetic on monetary fields, as financial calculations must use integer arithmetic or decimal types to avoid precision loss.

---

### B12 — Nullable Field Crashes (CTR-response-shape)

**Bug:** Dereferences `*string` nullable field without nil check, causing panics.
**Expected violation:** `CTR-response-shape` on line 28

```go
// stripe/buggy_client_b12.go — Dereferencing nullable fields without nil check.
package stripe

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// Properly models nullable fields as pointers
type Charge struct {
	ID              string  `json:"id"`
	Amount          int64   `json:"amount"`
	Currency        string  `json:"currency"`
	Status          string  `json:"status"`
	FailureCode     *string `json:"failure_code"`     // nullable
	FailureMessage  *string `json:"failure_message"`  // nullable
	Created         int64   `json:"created"`
}

func (c *Client) GetCharge(id string) (*Charge, error) {
	httpReq, _ := http.NewRequest("GET", c.baseURL+"/charges/"+id, nil)
	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, _ := c.httpClient.Do(httpReq)
	defer resp.Body.Close()

	var charge Charge
	json.NewDecoder(resp.Body).Decode(&charge)

	// BUG: Dereferences nullable fields without nil check
	fmt.Printf("Failure: %s - %s\n", *charge.FailureCode, *charge.FailureMessage)

	return &charge, nil
}

func LogChargeFailure(charge *Charge) error {
	if charge.Status == "failed" {
		// BUG: Panic if failure_code or failure_message are null
		return fmt.Errorf("charge failed: %s - %s", *charge.FailureCode, *charge.FailureMessage)
	}
	return nil
}
```

**Why Stricture catches this:** The manifest marks fields as `nullable: true` (e.g., `failure_code: { type: string, required: false, nullable: true }`). Stricture's `CTR-response-shape` rule requires nil checks before dereferencing pointer fields that map to nullable contract fields.

---

### B13 — Missing Webhook Verification (CTR-request-shape)

**Bug:** No HMAC signature verification on Stripe webhook payload, accepting forged events.
**Expected violation:** `CTR-request-shape` on line 18

```go
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
```

**Why Stricture catches this:** The manifest specifies `webhook_event: { signature_header: Stripe-Signature, signature_algo: HMAC-SHA256 }`. Stricture's `CTR-request-shape` rule requires webhook handlers to verify signatures using the specified algorithm before processing payloads.

---

### B14 — Pagination Terminated Early (CTR-response-shape)

**Bug:** Ignores `has_more` flag, returns only first page of results instead of all charges.
**Expected violation:** `CTR-response-shape` on line 32

```go
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
```

**Why Stricture catches this:** The manifest defines `list_charges` response with `has_more: boolean` field. Stricture's `CTR-response-shape` rule detects when paginated endpoints are called but the `has_more` field is not used in a loop to fetch all pages.

---

### B15 — Race Condition (CTR-request-shape)

**Bug:** No idempotency key on charge creation, allowing duplicate charges on retry/timeout.
**Expected violation:** `CTR-request-shape` on line 24

```go
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
```

**Why Stricture catches this:** The manifest specifies `Idempotency-Key: { type: string, required: false }` in the request headers. Stricture's `CTR-request-shape` rule requires POST operations that mutate state (create charge) to include idempotency keys to prevent duplicate operations on network failures or retries.
