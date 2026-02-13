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
