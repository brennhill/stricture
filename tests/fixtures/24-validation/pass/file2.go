// models.go — Inventory domain models with correct JSON tags
package main

import (
	"time"
)

type ItemStatus string

const (
	StatusInStock      ItemStatus = "in_stock"
	StatusLowStock     ItemStatus = "low_stock"
	StatusOutOfStock   ItemStatus = "out_of_stock"
	StatusDiscontinued ItemStatus = "discontinued"
	StatusBackorder    ItemStatus = "backorder"
)

type Item struct {
	ID          string     `json:"id"`
	SKU         string     `json:"sku"`
	Name        string     `json:"name"`
	Description *string    `json:"description"` // Nullable
	StockCount  int64      `json:"stock_count"`
	WarehouseID string     `json:"warehouse_id"`
	Status      ItemStatus `json:"status"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	ETag        string     `json:"etag"`
}

type CreateItemRequest struct {
	SKU         string  `json:"sku"`
	Name        string  `json:"name"`
	Description *string `json:"description"`
	StockCount  int64   `json:"stock_count"`
	WarehouseID string  `json:"warehouse_id"`
}

type UpdateItemRequest struct {
	StockCount *int64      `json:"stock_count,omitempty"`
	Status     *ItemStatus `json:"status,omitempty"`
	Description *string    `json:"description,omitempty"`
}

type ListItemsResponse struct {
	Items      []Item     `json:"items"`
	Pagination Pagination `json:"pagination"`
}

type Pagination struct {
	Page    int  `json:"page"`
	Limit   int  `json:"limit"`
	Total   int  `json:"total"`
	HasNext bool `json:"has_next"`
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}
