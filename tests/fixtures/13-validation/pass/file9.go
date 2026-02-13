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
