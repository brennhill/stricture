// orders_test.go — Unit tests for order handlers.
package handlers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"myapp/handlers"
	"myapp/models"
	"myapp/storage/mocks"
)

func TestCreateOrder_Success(t *testing.T) {
	mockStore := mocks.NewOrderStore()
	handler := handlers.NewOrderHandler(mockStore)

	reqBody := models.CreateOrderRequest{
		CustomerID:  "550e8400-e29b-41d4-a716-446655440000",
		TotalAmount: 5000, // $50.00 in cents
		Note:        stringPtr("Rush delivery"),
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/orders", bytes.NewReader(body))
	w := httptest.NewRecorder()

	handler.CreateOrder(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var order models.Order
	err := json.NewDecoder(w.Body).Decode(&order)
	require.NoError(t, err)

	// DEFECT B03: Go uses assert.NotNil, TS would use expect().toBeDefined()
	// This is an assertion library mismatch across languages
	assert.NotNil(t, order.ID)
	assert.Equal(t, reqBody.CustomerID, order.CustomerID)
	assert.Equal(t, reqBody.TotalAmount, order.TotalAmount)
	assert.Equal(t, models.StatusPending, order.Status)
}

func TestCreateOrder_InvalidCustomerID(t *testing.T) {
	mockStore := mocks.NewOrderStore()
	handler := handlers.NewOrderHandler(mockStore)

	reqBody := models.CreateOrderRequest{
		CustomerID:  "not-a-uuid", // INVALID
		TotalAmount: 5000,
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/orders", bytes.NewReader(body))
	w := httptest.NewRecorder()

	handler.CreateOrder(w, req)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
}

func TestCreateOrder_NegativeAmount(t *testing.T) {
	mockStore := mocks.NewOrderStore()
	handler := handlers.NewOrderHandler(mockStore)

	reqBody := models.CreateOrderRequest{
		CustomerID:  "550e8400-e29b-41d4-a716-446655440000",
		TotalAmount: -100, // INVALID
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/orders", bytes.NewReader(body))
	w := httptest.NewRecorder()

	handler.CreateOrder(w, req)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
}

func TestUpdateStatus_VersionConflict(t *testing.T) {
	mockStore := mocks.NewOrderStore()
	mockStore.SetVersionConflict(true) // Simulate version mismatch
	handler := handlers.NewOrderHandler(mockStore)

	reqBody := models.UpdateStatusRequest{
		Status:  models.StatusConfirmed,
		Version: 0,
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPatch, "/api/orders/123/status", bytes.NewReader(body))
	req.Header.Set("If-Match", "0")
	w := httptest.NewRecorder()

	handler.UpdateStatus(w, req)

	assert.Equal(t, http.StatusConflict, w.Code)
}

func stringPtr(s string) *string {
	return &s
}
