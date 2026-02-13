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
