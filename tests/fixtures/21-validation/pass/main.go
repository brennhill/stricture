// order.go — Order domain model with proper json tags.
package models

import (
	"time"
)

// OrderStatus represents valid order states.
type OrderStatus string

const (
	StatusPending    OrderStatus = "pending"
	StatusConfirmed  OrderStatus = "confirmed"
	StatusProcessing OrderStatus = "processing"
	StatusShipped    OrderStatus = "shipped"
	StatusDelivered  OrderStatus = "delivered"
)

// Order represents a customer order.
type Order struct {
	ID          string      `json:"id"`
	CustomerID  string      `json:"customer_id"` // CORRECT: snake_case
	TotalAmount int64       `json:"total_amount"` // CORRECT: cents as int64
	Status      OrderStatus `json:"status"`
	Note        *string     `json:"note,omitempty"` // CORRECT: nullable
	CreatedAt   time.Time   `json:"created_at"`     // CORRECT: snake_case
	UpdatedAt   time.Time   `json:"updated_at"`
	Version     int         `json:"version"` // For optimistic locking
}

// CreateOrderRequest is the API request for POST /api/orders.
type CreateOrderRequest struct {
	CustomerID  string  `json:"customer_id" validate:"required,uuid4"`
	TotalAmount int64   `json:"total_amount" validate:"required,gt=0"`
	Note        *string `json:"note,omitempty"`
}

// UpdateStatusRequest is the API request for PATCH /api/orders/:id/status.
type UpdateStatusRequest struct {
	Status  OrderStatus `json:"status" validate:"required,oneof=pending confirmed processing shipped delivered"`
	Version int         `json:"version" validate:"required,gte=0"` // Optimistic locking
}

// ListOrdersResponse is the cursor-based pagination response.
type ListOrdersResponse struct {
	Orders     []Order `json:"orders"`
	NextCursor *string `json:"next_cursor,omitempty"`
}
