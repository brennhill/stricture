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
