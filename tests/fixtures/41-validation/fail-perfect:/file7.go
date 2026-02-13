// payment_processor_anti3_test.go (BAD)
package payment

import (
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
)

// GLOBAL MOCK - shared across all tests
type MockGateway struct {
    mock.Mock
}

func (m *MockGateway) Charge(req ChargeRequest) (*ChargeResponse, error) {
    args := m.Called(req)
    return args.Get(0).(*ChargeResponse), args.Error(1)
}

var globalGateway = &MockGateway{}

func TestProcessPayment_Success(t *testing.T) {
    globalGateway.On("Charge", mock.Anything).Return(&ChargeResponse{Success: true}, nil)
    processor := NewPaymentProcessor(globalGateway)

    result, err := processor.Process(PaymentRequest{
        Amount:   100.00,
        Currency: "USD",
    })

    assert.NoError(t, err)
    assert.Equal(t, "succeeded", result.Status)
}

func TestProcessPayment_LargeAmount(t *testing.T) {
    // Uses same globalGateway with stale mock expectations
    processor := NewPaymentProcessor(globalGateway)

    result, err := processor.Process(PaymentRequest{
        Amount:   5000.00,
        Currency: "USD",
    })

    assert.NoError(t, err)
    assert.Equal(t, "succeeded", result.Status)
}

func TestProcessPayment_Failure(t *testing.T) {
    // Trying to override global mock behavior - will conflict
    globalGateway.On("Charge", mock.Anything).Return(nil, errors.New("gateway error"))
    processor := NewPaymentProcessor(globalGateway)

    _, err := processor.Process(PaymentRequest{
        Amount:   100.00,
        Currency: "USD",
    })

    assert.Error(t, err)
}
