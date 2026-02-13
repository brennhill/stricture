// payment_processor_anti1_test.go (BAD)
package payment

import (
    "testing"
    "github.com/stretchr/testify/assert"
)

func TestProcessPayment_CarpetBomb(t *testing.T) {
    processor := NewPaymentProcessor()
    result, err := processor.Process(PaymentRequest{
        Amount:          100.00,
        Currency:        "USD",
        UserID:          "user_123",
        PaymentMethodID: "pm_456",
    })

    assert.NotNil(t, result)
    assert.NotNil(t, result.TransactionID)
    assert.NotNil(t, result.Amount)
    assert.NotNil(t, result.Currency)
    assert.NotNil(t, result.Status)
    assert.NotNil(t, result.Fee)
    assert.NotNil(t, result.NetAmount)
    assert.NotNil(t, result.ProcessedAt)
    assert.NotNil(t, result.Metadata)
    assert.NoError(t, err)
}

func TestRefund_Defined(t *testing.T) {
    processor := NewPaymentProcessor()
    result, err := processor.Refund("txn_123", 50.00)

    assert.NotNil(t, result)
    assert.NotNil(t, result.TransactionID)
    assert.NoError(t, err)
}
