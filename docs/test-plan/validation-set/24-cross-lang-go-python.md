# 24 — Cross-Language Contract: Go Server + Python Client

**Category:** Cross-Language Contract Testing
**Languages:** Go (server) + Python (client)
**Framework:** Go chi router + Python httpx + Pydantic
**API Domain:** Inventory Management (items with SKU, stock levels, warehouse)
**CTR Rules Tested:** B01-B15 (boundary violations, type mismatches, null handling, pagination, conditional updates)

---

## Overview

This validation set tests cross-language contract adherence between a Go HTTP server and Python HTTP client. The API manages inventory items with SKU codes, stock counts, and warehouse assignments. Tests focus on common cross-language pitfalls:

- JSON field naming conventions (snake_case vs camelCase)
- Type system differences (Go int64 vs Python float, nullable types)
- Enum exhaustiveness across languages
- Null/None handling differences
- Pagination contract consistency
- Conditional update mechanisms (ETags)

**PERFECT Implementation:**
- Go server with `chi` router, correct `json` struct tags
- Python client with Pydantic models matching server schema
- Comprehensive test suites in both languages

**B01-B15 Violations:**
- B05: Python uses `stockCount` instead of `stock_count`
- B06: Python model missing required `warehouse_id` field
- B07: Go returns `int64`, Python model expects `float` for stock count
- B08: `ItemStatus` enum incomplete in Python (missing states)
- B11: JSON naming convention mismatch (snake_case vs camelCase)
- B12: Go returns `*string` (nullable), Python calls `.upper()` without null check
- B14: Pagination response structure mismatch between languages
- B15: PATCH endpoint requires ETag but Python client doesn't send it

---

## Directory Structure

```
24-cross-lang-go-python/
├── PERFECT/
│   ├── go-server/
│   │   ├── main.go                    # Chi server with routes
│   │   ├── models.go                  # Inventory models with correct json tags
│   │   ├── handlers.go                # HTTP handlers
│   │   ├── store.go                   # In-memory store
│   │   ├── main_test.go               # Go integration tests
│   │   ├── go.mod
│   │   └── go.sum
│   ├── python-client/
│   │   ├── client.py                  # Httpx client implementation
│   │   ├── models.py                  # Pydantic models matching Go schema
│   │   ├── test_client.py             # Python integration tests
│   │   ├── requirements.txt           # httpx, pydantic, pytest
│   │   └── pyproject.toml
│   └── README.md                      # Setup and test instructions
│
├── B05-field-naming-mismatch/
│   └── python-client/
│       └── models.py                  # Uses stockCount instead of stock_count
│
├── B06-missing-required-field/
│   └── python-client/
│       └── models.py                  # Missing warehouse_id field
│
├── B07-type-mismatch-int-float/
│   └── python-client/
│       └── models.py                  # stock_count: float instead of int
│
├── B08-incomplete-enum/
│   └── python-client/
│       └── models.py                  # ItemStatus missing states
│
├── B11-json-convention-violation/
│   └── go-server/
│       └── models.go                  # Uses camelCase json tags
│
├── B12-null-safety-violation/
│   └── python-client/
│       └── client.py                  # Calls .upper() on nullable field
│
├── B14-pagination-mismatch/
│   └── python-client/
│       └── models.py                  # Wrong pagination response structure
│
├── B15-missing-etag/
│   └── python-client/
│       └── client.py                  # PATCH without If-Match header
│
└── README.md                          # Validation set overview
```

---

## API Specification

### Base URL
```
http://localhost:8080/api/v1
```

### Endpoints

#### GET /items
List inventory items with pagination.

**Query Parameters:**
- `page` (int, optional, default=1): Page number
- `limit` (int, optional, default=20): Items per page
- `status` (string, optional): Filter by status

**Response:**
```json
{
  "items": [
    {
      "id": "item-123",
      "sku": "SKU-WIDGET-001",
      "name": "Blue Widget",
      "description": "Premium blue widget",
      "stock_count": 150,
      "warehouse_id": "warehouse-west",
      "status": "in_stock",
      "created_at": "2026-02-10T10:00:00Z",
      "updated_at": "2026-02-13T14:30:00Z",
      "etag": "v1-abc123"
    }
  ],
  "pagination": {
    "page": 1,
    "limit": 20,
    "total": 45,
    "has_next": true
  }
}
```

#### GET /items/:id
Get single item by ID.

**Response:**
```json
{
  "id": "item-123",
  "sku": "SKU-WIDGET-001",
  "name": "Blue Widget",
  "description": "Premium blue widget",
  "stock_count": 150,
  "warehouse_id": "warehouse-west",
  "status": "in_stock",
  "created_at": "2026-02-10T10:00:00Z",
  "updated_at": "2026-02-13T14:30:00Z",
  "etag": "v1-abc123"
}
```

#### POST /items
Create new item.

**Request:**
```json
{
  "sku": "SKU-WIDGET-002",
  "name": "Red Widget",
  "description": "Premium red widget",
  "stock_count": 100,
  "warehouse_id": "warehouse-east"
}
```

**Response:** 201 Created with item object

#### PATCH /items/:id
Update item (requires ETag for optimistic locking).

**Headers:**
- `If-Match: "v1-abc123"` (required)

**Request:**
```json
{
  "stock_count": 175,
  "status": "low_stock"
}
```

**Response:** 200 OK with updated item object

#### DELETE /items/:id
Delete item.

**Response:** 204 No Content

### Data Types

**ItemStatus Enum:**
- `in_stock`: Normal stock levels
- `low_stock`: Below reorder threshold
- `out_of_stock`: No units available
- `discontinued`: No longer sold
- `backorder`: Available for order, no current stock

**Field Types:**
- `id`: string (UUID format)
- `sku`: string (alphanumeric with dashes)
- `name`: string (max 200 chars)
- `description`: string, nullable (max 1000 chars)
- `stock_count`: int64 (non-negative)
- `warehouse_id`: string (required, FK to warehouse)
- `status`: ItemStatus enum
- `created_at`: RFC3339 timestamp
- `updated_at`: RFC3339 timestamp
- `etag`: string (version tag)

---

## PERFECT Implementation

### Go Server (go-server/)

#### main.go
```go
// main.go — Chi server for inventory management API
package main

import (
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(30 * time.Second))

	// Initialize store
	store := NewStore()

	// Routes
	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/items", GetItems(store))
		r.Post("/items", CreateItem(store))
		r.Get("/items/{id}", GetItem(store))
		r.Patch("/items/{id}", UpdateItem(store))
		r.Delete("/items/{id}", DeleteItem(store))
	})

	log.Println("Server starting on :8080")
	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatal(err)
	}
}
```

#### models.go
```go
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
```

#### handlers.go
```go
// handlers.go — HTTP handlers for inventory endpoints
package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
)

func GetItems(store *Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Parse query params
		page, _ := strconv.Atoi(r.URL.Query().Get("page"))
		if page < 1 {
			page = 1
		}

		limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
		if limit < 1 || limit > 100 {
			limit = 20
		}

		statusFilter := r.URL.Query().Get("status")

		items, total := store.ListItems(page, limit, statusFilter)

		resp := ListItemsResponse{
			Items: items,
			Pagination: Pagination{
				Page:    page,
				Limit:   limit,
				Total:   total,
				HasNext: page*limit < total,
			},
		}

		respondJSON(w, http.StatusOK, resp)
	}
}

func GetItem(store *Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")

		item, exists := store.GetItem(id)
		if !exists {
			respondError(w, http.StatusNotFound, "ITEM_NOT_FOUND", "Item not found")
			return
		}

		respondJSON(w, http.StatusOK, item)
	}
}

func CreateItem(store *Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req CreateItemRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondError(w, http.StatusBadRequest, "INVALID_JSON", err.Error())
			return
		}

		// Validation
		if req.SKU == "" || req.Name == "" || req.WarehouseID == "" {
			respondError(w, http.StatusBadRequest, "MISSING_FIELDS", "SKU, name, and warehouse_id are required")
			return
		}

		if req.StockCount < 0 {
			respondError(w, http.StatusBadRequest, "INVALID_STOCK", "Stock count cannot be negative")
			return
		}

		item := store.CreateItem(req)
		respondJSON(w, http.StatusCreated, item)
	}
}

func UpdateItem(store *Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")

		// Check ETag
		ifMatch := r.Header.Get("If-Match")
		if ifMatch == "" {
			respondError(w, http.StatusPreconditionRequired, "ETAG_REQUIRED", "If-Match header is required")
			return
		}

		// Remove quotes from ETag
		ifMatch = strings.Trim(ifMatch, `"`)

		var req UpdateItemRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondError(w, http.StatusBadRequest, "INVALID_JSON", err.Error())
			return
		}

		item, err := store.UpdateItem(id, ifMatch, req)
		if err != nil {
			if err == ErrItemNotFound {
				respondError(w, http.StatusNotFound, "ITEM_NOT_FOUND", "Item not found")
			} else if err == ErrETagMismatch {
				respondError(w, http.StatusPreconditionFailed, "ETAG_MISMATCH", "ETag does not match")
			} else {
				respondError(w, http.StatusBadRequest, "UPDATE_FAILED", err.Error())
			}
			return
		}

		respondJSON(w, http.StatusOK, item)
	}
}

func DeleteItem(store *Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")

		if !store.DeleteItem(id) {
			respondError(w, http.StatusNotFound, "ITEM_NOT_FOUND", "Item not found")
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

// Helper functions

func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func respondError(w http.ResponseWriter, status int, code, message string) {
	respondJSON(w, status, ErrorResponse{
		Error:   code,
		Message: message,
	})
}
```

#### store.go
```go
// store.go — In-memory inventory store
package main

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

var (
	ErrItemNotFound = errors.New("item not found")
	ErrETagMismatch = errors.New("etag mismatch")
)

type Store struct {
	mu    sync.RWMutex
	items map[string]Item
}

func NewStore() *Store {
	s := &Store{
		items: make(map[string]Item),
	}

	// Seed with sample data
	s.seedData()
	return s
}

func (s *Store) seedData() {
	desc1 := "Premium blue widget for industrial use"
	desc2 := "Standard green widget"

	items := []Item{
		{
			ID:          "item-001",
			SKU:         "SKU-WIDGET-001",
			Name:        "Blue Widget",
			Description: &desc1,
			StockCount:  150,
			WarehouseID: "warehouse-west",
			Status:      StatusInStock,
			CreatedAt:   time.Now().Add(-48 * time.Hour),
			UpdatedAt:   time.Now().Add(-1 * time.Hour),
		},
		{
			ID:          "item-002",
			SKU:         "SKU-WIDGET-002",
			Name:        "Green Widget",
			Description: &desc2,
			StockCount:  25,
			WarehouseID: "warehouse-east",
			Status:      StatusLowStock,
			CreatedAt:   time.Now().Add(-24 * time.Hour),
			UpdatedAt:   time.Now().Add(-30 * time.Minute),
		},
		{
			ID:          "item-003",
			SKU:         "SKU-GADGET-003",
			Name:        "Red Gadget",
			Description: nil, // Nullable field
			StockCount:  0,
			WarehouseID: "warehouse-central",
			Status:      StatusOutOfStock,
			CreatedAt:   time.Now().Add(-72 * time.Hour),
			UpdatedAt:   time.Now(),
		},
	}

	for _, item := range items {
		item.ETag = s.generateETag(item)
		s.items[item.ID] = item
	}
}

func (s *Store) ListItems(page, limit int, statusFilter string) ([]Item, int) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Filter items
	filtered := []Item{}
	for _, item := range s.items {
		if statusFilter == "" || string(item.Status) == statusFilter {
			filtered = append(filtered, item)
		}
	}

	total := len(filtered)

	// Paginate
	start := (page - 1) * limit
	end := start + limit

	if start >= total {
		return []Item{}, total
	}

	if end > total {
		end = total
	}

	return filtered[start:end], total
}

func (s *Store) GetItem(id string) (Item, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	item, exists := s.items[id]
	return item, exists
}

func (s *Store) CreateItem(req CreateItemRequest) Item {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	item := Item{
		ID:          uuid.New().String(),
		SKU:         req.SKU,
		Name:        req.Name,
		Description: req.Description,
		StockCount:  req.StockCount,
		WarehouseID: req.WarehouseID,
		Status:      s.calculateStatus(req.StockCount),
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	item.ETag = s.generateETag(item)
	s.items[item.ID] = item

	return item
}

func (s *Store) UpdateItem(id, etag string, req UpdateItemRequest) (Item, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	item, exists := s.items[id]
	if !exists {
		return Item{}, ErrItemNotFound
	}

	// Verify ETag
	if item.ETag != etag {
		return Item{}, ErrETagMismatch
	}

	// Apply updates
	if req.StockCount != nil {
		if *req.StockCount < 0 {
			return Item{}, errors.New("stock count cannot be negative")
		}
		item.StockCount = *req.StockCount
		// Recalculate status if not explicitly set
		if req.Status == nil {
			item.Status = s.calculateStatus(item.StockCount)
		}
	}

	if req.Status != nil {
		item.Status = *req.Status
	}

	if req.Description != nil {
		item.Description = req.Description
	}

	item.UpdatedAt = time.Now()
	item.ETag = s.generateETag(item)

	s.items[id] = item
	return item, nil
}

func (s *Store) DeleteItem(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, exists := s.items[id]
	if !exists {
		return false
	}

	delete(s.items, id)
	return true
}

func (s *Store) generateETag(item Item) string {
	data := fmt.Sprintf("%s-%d-%s-%s", item.ID, item.StockCount, item.Status, item.UpdatedAt.Format(time.RFC3339))
	hash := sha256.Sum256([]byte(data))
	return "v1-" + hex.EncodeToString(hash[:8])
}

func (s *Store) calculateStatus(stockCount int64) ItemStatus {
	if stockCount == 0 {
		return StatusOutOfStock
	} else if stockCount < 50 {
		return StatusLowStock
	}
	return StatusInStock
}
```

#### main_test.go
```go
// main_test.go — Integration tests for inventory API
package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupRouter() *chi.Mux {
	r := chi.NewRouter()
	store := NewStore()

	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/items", GetItems(store))
		r.Post("/items", CreateItem(store))
		r.Get("/items/{id}", GetItem(store))
		r.Patch("/items/{id}", UpdateItem(store))
		r.Delete("/items/{id}", DeleteItem(store))
	})

	return r
}

func TestListItems(t *testing.T) {
	router := setupRouter()

	req := httptest.NewRequest("GET", "/api/v1/items?page=1&limit=10", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp ListItemsResponse
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)

	assert.NotEmpty(t, resp.Items)
	assert.Equal(t, 1, resp.Pagination.Page)
	assert.Equal(t, 10, resp.Pagination.Limit)
	assert.GreaterOrEqual(t, resp.Pagination.Total, len(resp.Items))

	// Verify JSON field naming (snake_case)
	for _, item := range resp.Items {
		assert.NotEmpty(t, item.ID)
		assert.NotEmpty(t, item.SKU)
		assert.NotEmpty(t, item.WarehouseID)
		assert.GreaterOrEqual(t, item.StockCount, int64(0))
	}
}

func TestGetItemByID(t *testing.T) {
	router := setupRouter()

	req := httptest.NewRequest("GET", "/api/v1/items/item-001", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var item Item
	err := json.NewDecoder(w.Body).Decode(&item)
	require.NoError(t, err)

	assert.Equal(t, "item-001", item.ID)
	assert.NotEmpty(t, item.SKU)
	assert.NotEmpty(t, item.ETag)
}

func TestGetItemNotFound(t *testing.T) {
	router := setupRouter()

	req := httptest.NewRequest("GET", "/api/v1/items/nonexistent", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var errResp ErrorResponse
	err := json.NewDecoder(w.Body).Decode(&errResp)
	require.NoError(t, err)

	assert.Equal(t, "ITEM_NOT_FOUND", errResp.Error)
}

func TestCreateItem(t *testing.T) {
	router := setupRouter()

	desc := "Test description"
	reqBody := CreateItemRequest{
		SKU:         "SKU-TEST-999",
		Name:        "Test Item",
		Description: &desc,
		StockCount:  100,
		WarehouseID: "warehouse-test",
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v1/items", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var item Item
	err := json.NewDecoder(w.Body).Decode(&item)
	require.NoError(t, err)

	assert.NotEmpty(t, item.ID)
	assert.Equal(t, "SKU-TEST-999", item.SKU)
	assert.Equal(t, "Test Item", item.Name)
	assert.Equal(t, int64(100), item.StockCount)
	assert.Equal(t, "warehouse-test", item.WarehouseID)
	assert.NotEmpty(t, item.ETag)
	assert.Equal(t, StatusInStock, item.Status)
}

func TestCreateItemMissingFields(t *testing.T) {
	router := setupRouter()

	reqBody := CreateItemRequest{
		SKU:  "SKU-TEST-999",
		Name: "Test Item",
		// Missing WarehouseID
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v1/items", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var errResp ErrorResponse
	err := json.NewDecoder(w.Body).Decode(&errResp)
	require.NoError(t, err)

	assert.Equal(t, "MISSING_FIELDS", errResp.Error)
}

func TestUpdateItem(t *testing.T) {
	router := setupRouter()

	// First get the item to retrieve ETag
	getReq := httptest.NewRequest("GET", "/api/v1/items/item-001", nil)
	getW := httptest.NewRecorder()
	router.ServeHTTP(getW, getReq)

	var existingItem Item
	json.NewDecoder(getW.Body).Decode(&existingItem)

	// Now update with correct ETag
	newStock := int64(200)
	newStatus := StatusInStock
	updateReq := UpdateItemRequest{
		StockCount: &newStock,
		Status:     &newStatus,
	}

	body, _ := json.Marshal(updateReq)
	req := httptest.NewRequest("PATCH", "/api/v1/items/item-001", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("If-Match", `"`+existingItem.ETag+`"`)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var updatedItem Item
	err := json.NewDecoder(w.Body).Decode(&updatedItem)
	require.NoError(t, err)

	assert.Equal(t, int64(200), updatedItem.StockCount)
	assert.Equal(t, StatusInStock, updatedItem.Status)
	assert.NotEqual(t, existingItem.ETag, updatedItem.ETag) // ETag should change
}

func TestUpdateItemWithoutETag(t *testing.T) {
	router := setupRouter()

	newStock := int64(200)
	updateReq := UpdateItemRequest{
		StockCount: &newStock,
	}

	body, _ := json.Marshal(updateReq)
	req := httptest.NewRequest("PATCH", "/api/v1/items/item-001", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	// Missing If-Match header
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusPreconditionRequired, w.Code)

	var errResp ErrorResponse
	err := json.NewDecoder(w.Body).Decode(&errResp)
	require.NoError(t, err)

	assert.Equal(t, "ETAG_REQUIRED", errResp.Error)
}

func TestUpdateItemETagMismatch(t *testing.T) {
	router := setupRouter()

	newStock := int64(200)
	updateReq := UpdateItemRequest{
		StockCount: &newStock,
	}

	body, _ := json.Marshal(updateReq)
	req := httptest.NewRequest("PATCH", "/api/v1/items/item-001", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("If-Match", `"wrong-etag"`)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusPreconditionFailed, w.Code)

	var errResp ErrorResponse
	err := json.NewDecoder(w.Body).Decode(&errResp)
	require.NoError(t, err)

	assert.Equal(t, "ETAG_MISMATCH", errResp.Error)
}

func TestDeleteItem(t *testing.T) {
	router := setupRouter()

	req := httptest.NewRequest("DELETE", "/api/v1/items/item-001", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)

	// Verify item is deleted
	getReq := httptest.NewRequest("GET", "/api/v1/items/item-001", nil)
	getW := httptest.NewRecorder()
	router.ServeHTTP(getW, getReq)

	assert.Equal(t, http.StatusNotFound, getW.Code)
}

func TestNullableDescriptionField(t *testing.T) {
	router := setupRouter()

	// Create item with null description
	reqBody := CreateItemRequest{
		SKU:         "SKU-TEST-NULL",
		Name:        "Null Description Test",
		Description: nil,
		StockCount:  50,
		WarehouseID: "warehouse-test",
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v1/items", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var item Item
	err := json.NewDecoder(w.Body).Decode(&item)
	require.NoError(t, err)

	assert.Nil(t, item.Description)
}
```

#### go.mod
```go
module inventory-api

go 1.22

require (
	github.com/go-chi/chi/v5 v5.0.12
	github.com/google/uuid v1.6.0
	github.com/stretchr/testify v1.9.0
)
```

---

### Python Client (python-client/)

#### models.py
```python
"""models.py — Pydantic models matching Go server schema"""
from datetime import datetime
from enum import Enum
from typing import Optional

from pydantic import BaseModel, Field


class ItemStatus(str, Enum):
    """Item availability status enum"""
    IN_STOCK = "in_stock"
    LOW_STOCK = "low_stock"
    OUT_OF_STOCK = "out_of_stock"
    DISCONTINUED = "discontinued"
    BACKORDER = "backorder"


class Item(BaseModel):
    """Inventory item model"""
    id: str
    sku: str
    name: str
    description: Optional[str] = None
    stock_count: int = Field(ge=0)  # Non-negative integer
    warehouse_id: str
    status: ItemStatus
    created_at: datetime
    updated_at: datetime
    etag: str

    class Config:
        use_enum_values = True


class CreateItemRequest(BaseModel):
    """Request model for creating a new item"""
    sku: str
    name: str
    description: Optional[str] = None
    stock_count: int = Field(ge=0)
    warehouse_id: str


class UpdateItemRequest(BaseModel):
    """Request model for updating an item"""
    stock_count: Optional[int] = Field(None, ge=0)
    status: Optional[ItemStatus] = None
    description: Optional[str] = None

    class Config:
        use_enum_values = True


class Pagination(BaseModel):
    """Pagination metadata"""
    page: int
    limit: int
    total: int
    has_next: bool


class ListItemsResponse(BaseModel):
    """Response model for listing items"""
    items: list[Item]
    pagination: Pagination


class ErrorResponse(BaseModel):
    """Error response model"""
    error: str
    message: str
```

#### client.py
```python
"""client.py — Httpx client for inventory API"""
import httpx
from typing import Optional

from models import (
    Item,
    CreateItemRequest,
    UpdateItemRequest,
    ListItemsResponse,
    ErrorResponse,
    ItemStatus,
)


class InventoryClient:
    """HTTP client for inventory management API"""

    def __init__(self, base_url: str = "http://localhost:8080/api/v1"):
        self.base_url = base_url
        self.client = httpx.Client(base_url=base_url, timeout=30.0)

    def close(self):
        """Close the HTTP client"""
        self.client.close()

    def __enter__(self):
        return self

    def __exit__(self, exc_type, exc_val, exc_tb):
        self.close()

    def list_items(
        self,
        page: int = 1,
        limit: int = 20,
        status: Optional[ItemStatus] = None,
    ) -> ListItemsResponse:
        """List items with pagination and optional status filter"""
        params = {"page": page, "limit": limit}
        if status:
            params["status"] = status.value

        response = self.client.get("/items", params=params)
        response.raise_for_status()

        return ListItemsResponse(**response.json())

    def get_item(self, item_id: str) -> Item:
        """Get a single item by ID"""
        response = self.client.get(f"/items/{item_id}")
        response.raise_for_status()

        return Item(**response.json())

    def create_item(self, request: CreateItemRequest) -> Item:
        """Create a new item"""
        response = self.client.post(
            "/items",
            json=request.model_dump(exclude_none=True),
        )
        response.raise_for_status()

        return Item(**response.json())

    def update_item(
        self,
        item_id: str,
        etag: str,
        request: UpdateItemRequest,
    ) -> Item:
        """Update an existing item (requires ETag)"""
        response = self.client.patch(
            f"/items/{item_id}",
            json=request.model_dump(exclude_none=True),
            headers={"If-Match": f'"{etag}"'},
        )
        response.raise_for_status()

        return Item(**response.json())

    def delete_item(self, item_id: str) -> None:
        """Delete an item"""
        response = self.client.delete(f"/items/{item_id}")
        response.raise_for_status()

    def handle_error(self, response: httpx.Response) -> ErrorResponse:
        """Parse error response"""
        try:
            return ErrorResponse(**response.json())
        except Exception:
            return ErrorResponse(
                error="UNKNOWN_ERROR",
                message=f"HTTP {response.status_code}: {response.text}",
            )
```

#### test_client.py
```python
"""test_client.py — Integration tests for Python client"""
import pytest
from datetime import datetime

from client import InventoryClient
from models import (
    CreateItemRequest,
    UpdateItemRequest,
    ItemStatus,
)


@pytest.fixture
def client():
    """Create client instance"""
    with InventoryClient() as c:
        yield c


def test_list_items(client):
    """Test listing items with pagination"""
    response = client.list_items(page=1, limit=10)

    assert len(response.items) > 0
    assert response.pagination.page == 1
    assert response.pagination.limit == 10
    assert response.pagination.total >= len(response.items)

    # Verify item structure
    for item in response.items:
        assert item.id
        assert item.sku
        assert item.name
        assert item.warehouse_id
        assert item.stock_count >= 0
        assert isinstance(item.status, ItemStatus)
        assert item.etag


def test_get_item_by_id(client):
    """Test getting a single item"""
    # First list to get a valid ID
    items = client.list_items(limit=1)
    item_id = items.items[0].id

    # Get specific item
    item = client.get_item(item_id)

    assert item.id == item_id
    assert item.sku
    assert item.etag


def test_get_item_not_found(client):
    """Test getting non-existent item"""
    with pytest.raises(Exception) as exc_info:
        client.get_item("nonexistent-id")

    assert exc_info.value.response.status_code == 404


def test_create_item(client):
    """Test creating a new item"""
    request = CreateItemRequest(
        sku="SKU-PYTEST-001",
        name="Python Test Item",
        description="Created by pytest",
        stock_count=100,
        warehouse_id="warehouse-test",
    )

    item = client.create_item(request)

    assert item.id
    assert item.sku == "SKU-PYTEST-001"
    assert item.name == "Python Test Item"
    assert item.description == "Created by pytest"
    assert item.stock_count == 100
    assert item.warehouse_id == "warehouse-test"
    assert item.status == ItemStatus.IN_STOCK
    assert item.etag
    assert isinstance(item.created_at, datetime)
    assert isinstance(item.updated_at, datetime)


def test_create_item_with_null_description(client):
    """Test creating item with null description"""
    request = CreateItemRequest(
        sku="SKU-PYTEST-NULL",
        name="Null Description Test",
        description=None,
        stock_count=50,
        warehouse_id="warehouse-test",
    )

    item = client.create_item(request)

    assert item.id
    assert item.description is None


def test_create_item_missing_required_field(client):
    """Test creating item without required field"""
    request = CreateItemRequest(
        sku="SKU-INCOMPLETE",
        name="Incomplete Item",
        stock_count=10,
        warehouse_id="",  # Invalid empty string
    )

    with pytest.raises(Exception) as exc_info:
        client.create_item(request)

    assert exc_info.value.response.status_code == 400


def test_update_item_with_etag(client):
    """Test updating an item with correct ETag"""
    # Create item
    create_req = CreateItemRequest(
        sku="SKU-UPDATE-TEST",
        name="Update Test",
        stock_count=100,
        warehouse_id="warehouse-test",
    )
    created = client.create_item(create_req)

    # Update item
    update_req = UpdateItemRequest(
        stock_count=150,
        status=ItemStatus.IN_STOCK,
    )
    updated = client.update_item(created.id, created.etag, update_req)

    assert updated.id == created.id
    assert updated.stock_count == 150
    assert updated.status == ItemStatus.IN_STOCK
    assert updated.etag != created.etag  # ETag should change


def test_update_item_without_etag(client):
    """Test updating item without ETag header"""
    # Create item first
    create_req = CreateItemRequest(
        sku="SKU-ETAG-TEST",
        name="ETag Test",
        stock_count=50,
        warehouse_id="warehouse-test",
    )
    created = client.create_item(create_req)

    # Try to update without ETag (by modifying client temporarily)
    update_req = UpdateItemRequest(stock_count=75)

    # This should fail with 428 Precondition Required
    with pytest.raises(Exception) as exc_info:
        # Pass empty string as etag
        client.client.patch(
            f"/items/{created.id}",
            json=update_req.model_dump(exclude_none=True),
            # Missing If-Match header
        )

    assert exc_info.value.response.status_code == 428


def test_update_item_etag_mismatch(client):
    """Test updating item with wrong ETag"""
    # Create item
    create_req = CreateItemRequest(
        sku="SKU-MISMATCH",
        name="Mismatch Test",
        stock_count=100,
        warehouse_id="warehouse-test",
    )
    created = client.create_item(create_req)

    # Update with wrong ETag
    update_req = UpdateItemRequest(stock_count=200)

    with pytest.raises(Exception) as exc_info:
        client.update_item(created.id, "wrong-etag", update_req)

    assert exc_info.value.response.status_code == 412  # Precondition Failed


def test_delete_item(client):
    """Test deleting an item"""
    # Create item
    create_req = CreateItemRequest(
        sku="SKU-DELETE",
        name="Delete Test",
        stock_count=10,
        warehouse_id="warehouse-test",
    )
    created = client.create_item(create_req)

    # Delete item
    client.delete_item(created.id)

    # Verify deletion
    with pytest.raises(Exception) as exc_info:
        client.get_item(created.id)

    assert exc_info.value.response.status_code == 404


def test_pagination(client):
    """Test pagination functionality"""
    # Get first page
    page1 = client.list_items(page=1, limit=2)

    assert len(page1.items) <= 2
    assert page1.pagination.page == 1
    assert page1.pagination.limit == 2

    # If there are more items, test second page
    if page1.pagination.has_next:
        page2 = client.list_items(page=2, limit=2)
        assert page2.pagination.page == 2
        assert page2.items[0].id != page1.items[0].id


def test_status_filter(client):
    """Test filtering by status"""
    response = client.list_items(status=ItemStatus.IN_STOCK)

    for item in response.items:
        assert item.status == ItemStatus.IN_STOCK


def test_nullable_field_handling(client):
    """Test proper handling of nullable description field"""
    # Get item with null description
    items = client.list_items(limit=10)

    # Find item with null description (item-003 from seed data)
    null_desc_items = [item for item in items.items if item.description is None]

    if null_desc_items:
        item = null_desc_items[0]
        assert item.description is None

        # Verify we can safely handle null description
        # This should NOT raise an error
        desc_upper = item.description.upper() if item.description else None
        assert desc_upper is None
```

#### requirements.txt
```
httpx>=0.27.0
pydantic>=2.0.0
pytest>=8.0.0
pytest-asyncio>=0.23.0
```

---

## B01-B15 Violation Examples

### B05: Field Naming Mismatch (python-client/models.py)

**Issue:** Python model uses `stockCount` (camelCase) instead of `stock_count` (snake_case) to match Go JSON tags.

```python
# B05-field-naming-mismatch/python-client/models.py
class Item(BaseModel):
    """Inventory item model with WRONG field naming"""
    id: str
    sku: str
    name: str
    description: Optional[str] = None
    stockCount: int = Field(ge=0)  # ❌ Should be stock_count
    warehouse_id: str
    status: ItemStatus
    created_at: datetime
    updated_at: datetime
    etag: str
```

**Result:** Pydantic validation error when parsing Go JSON response:
```
pydantic.ValidationError: 1 validation error for Item
  stockCount
    Field required [type=missing, input_value={'id': 'item-001', 'sku'...}]
```

---

### B06: Missing Required Field (python-client/models.py)

**Issue:** Python model missing required `warehouse_id` field.

```python
# B06-missing-required-field/python-client/models.py
class Item(BaseModel):
    """Inventory item model MISSING warehouse_id"""
    id: str
    sku: str
    name: str
    description: Optional[str] = None
    stock_count: int = Field(ge=0)
    # ❌ warehouse_id field is missing
    status: ItemStatus
    created_at: datetime
    updated_at: datetime
    etag: str
```

**Result:** Pydantic ignores `warehouse_id` from Go response, silently drops data. Client code expecting `warehouse_id` will fail:
```python
item = client.get_item("item-001")
print(item.warehouse_id)  # ❌ AttributeError: 'Item' object has no attribute 'warehouse_id'
```

---

### B07: Type Mismatch (int64 vs float) (python-client/models.py)

**Issue:** Python model uses `float` for `stock_count` instead of `int`.

```python
# B07-type-mismatch-int-float/python-client/models.py
class Item(BaseModel):
    """Inventory item with WRONG type for stock_count"""
    id: str
    sku: str
    name: str
    description: Optional[str] = None
    stock_count: float = Field(ge=0)  # ❌ Should be int to match Go int64
    warehouse_id: str
    status: ItemStatus
    created_at: datetime
    updated_at: datetime
    etag: str
```

**Result:** Stock count values lose precision semantics. Go expects integers, Python sends floats:
```python
# Python sends:
{"stock_count": 150.0}

# Go receives and stores as int64(150)
# On next read, Python gets:
{"stock_count": 150}

# Can cause validation issues if constraints differ
```

---

### B08: Incomplete Enum (python-client/models.py)

**Issue:** Python `ItemStatus` enum missing `BACKORDER` state.

```python
# B08-incomplete-enum/python-client/models.py
class ItemStatus(str, Enum):
    """INCOMPLETE item status enum"""
    IN_STOCK = "in_stock"
    LOW_STOCK = "low_stock"
    OUT_OF_STOCK = "out_of_stock"
    DISCONTINUED = "discontinued"
    # ❌ Missing BACKORDER = "backorder"
```

**Result:** Pydantic validation error when Go returns `backorder` status:
```
pydantic.ValidationError: 1 validation error for Item
  status
    Input should be 'in_stock', 'low_stock', 'out_of_stock' or 'discontinued' [type=enum, input_value='backorder']
```

---

### B11: JSON Convention Violation (go-server/models.go)

**Issue:** Go struct uses camelCase JSON tags instead of snake_case.

```go
// B11-json-convention-violation/go-server/models.go
type Item struct {
	ID          string     `json:"id"`
	SKU         string     `json:"sku"`
	Name        string     `json:"name"`
	Description *string    `json:"description"`
	StockCount  int64      `json:"stockCount"`     // ❌ Should be stock_count
	WarehouseID string     `json:"warehouseId"`    // ❌ Should be warehouse_id
	Status      ItemStatus `json:"status"`
	CreatedAt   time.Time  `json:"createdAt"`      // ❌ Should be created_at
	UpdatedAt   time.Time  `json:"updatedAt"`      // ❌ Should be updated_at
	ETag        string     `json:"etag"`
}
```

**Result:** Python client with correct snake_case models fails to parse response:
```python
# Python expects:
{"stock_count": 150, "warehouse_id": "warehouse-west", "created_at": "..."}

# Go sends:
{"stockCount": 150, "warehouseId": "warehouse-west", "createdAt": "..."}

# Pydantic validation error:
# Field required: stock_count, warehouse_id, created_at
```

---

### B12: Null Safety Violation (python-client/client.py)

**Issue:** Python client calls `.upper()` on nullable `description` field without null check.

```python
# B12-null-safety-violation/python-client/client.py
def get_item_description_upper(self, item_id: str) -> str:
    """Get item description in uppercase"""
    item = self.get_item(item_id)

    # ❌ description can be None, will raise AttributeError
    return item.description.upper()
```

**Result:** Runtime error when description is null:
```python
client.get_item_description_upper("item-003")  # Has null description

# ❌ AttributeError: 'NoneType' object has no attribute 'upper'
```

**Fix:**
```python
return item.description.upper() if item.description else None
```

---

### B14: Pagination Mismatch (python-client/models.py)

**Issue:** Python pagination model has different structure than Go response.

```python
# B14-pagination-mismatch/python-client/models.py
class Pagination(BaseModel):
    """WRONG pagination structure"""
    current_page: int     # ❌ Go uses "page"
    per_page: int         # ❌ Go uses "limit"
    total_items: int      # ❌ Go uses "total"
    next_page: bool       # ❌ Go uses "has_next"
```

**Result:** Pydantic validation error:
```
pydantic.ValidationError: 4 validation errors for Pagination
  current_page
    Field required [type=missing]
  per_page
    Field required [type=missing]
  total_items
    Field required [type=missing]
  next_page
    Field required [type=missing]
```

---

### B15: Missing ETag (python-client/client.py)

**Issue:** PATCH request doesn't send required `If-Match` header with ETag.

```python
# B15-missing-etag/python-client/client.py
def update_item(
    self,
    item_id: str,
    request: UpdateItemRequest,
) -> Item:
    """Update item WITHOUT ETag (wrong signature)"""
    response = self.client.patch(
        f"/items/{item_id}",
        json=request.model_dump(exclude_none=True),
        # ❌ Missing If-Match header with ETag
    )
    response.raise_for_status()
    return Item(**response.json())
```

**Result:** Go server returns `428 Precondition Required`:
```json
{
  "error": "ETAG_REQUIRED",
  "message": "If-Match header is required"
}
```

**Fix:**
```python
def update_item(
    self,
    item_id: str,
    etag: str,  # Add ETag parameter
    request: UpdateItemRequest,
) -> Item:
    response = self.client.patch(
        f"/items/{item_id}",
        json=request.model_dump(exclude_none=True),
        headers={"If-Match": f'"{etag}"'},  # ✅ Include ETag
    )
    response.raise_for_status()
    return Item(**response.json())
```

---

## Test Execution

### Running Go Server Tests

```bash
cd PERFECT/go-server
go test -v ./...
```

**Expected output:**
```
=== RUN   TestListItems
--- PASS: TestListItems (0.01s)
=== RUN   TestGetItemByID
--- PASS: TestGetItemByID (0.00s)
=== RUN   TestCreateItem
--- PASS: TestCreateItem (0.00s)
=== RUN   TestUpdateItem
--- PASS: TestUpdateItem (0.01s)
=== RUN   TestUpdateItemWithoutETag
--- PASS: TestUpdateItemWithoutETag (0.00s)
=== RUN   TestUpdateItemETagMismatch
--- PASS: TestUpdateItemETagMismatch (0.00s)
PASS
ok      inventory-api   0.123s
```

### Running Python Client Tests

```bash
cd PERFECT/python-client

# Start Go server first
cd ../go-server
go run . &
SERVER_PID=$!

# Run Python tests
cd ../python-client
pytest test_client.py -v

# Stop server
kill $SERVER_PID
```

**Expected output:**
```
test_client.py::test_list_items PASSED
test_client.py::test_get_item_by_id PASSED
test_client.py::test_create_item PASSED
test_client.py::test_update_item_with_etag PASSED
test_client.py::test_update_item_without_etag PASSED
test_client.py::test_update_item_etag_mismatch PASSED
test_client.py::test_delete_item PASSED
test_client.py::test_pagination PASSED
test_client.py::test_status_filter PASSED
test_client.py::test_nullable_field_handling PASSED

========== 10 passed in 0.45s ==========
```

### Testing Violation Examples

Each B01-B15 violation directory can be tested by replacing the relevant file in PERFECT and observing the failure:

```bash
# Test B05 (field naming mismatch)
cp B05-field-naming-mismatch/python-client/models.py PERFECT/python-client/
pytest PERFECT/python-client/test_client.py -v
# Expected: ValidationError on stock_count field

# Test B08 (incomplete enum)
cp B08-incomplete-enum/python-client/models.py PERFECT/python-client/
pytest PERFECT/python-client/test_client.py -v
# Expected: ValidationError when backorder status returned

# Test B15 (missing ETag)
cp B15-missing-etag/python-client/client.py PERFECT/python-client/
pytest PERFECT/python-client/test_client.py::test_update_item_with_etag -v
# Expected: 428 Precondition Required error
```

---

## Linter Detection Rules

### CTR Rules Applied

| Rule | Violation | Detection Method |
|------|-----------|------------------|
| B05 | Field naming mismatch | Compare JSON field names in Go struct tags vs Python Pydantic Field aliases |
| B06 | Missing required field | Cross-validate required fields between Go struct and Python model |
| B07 | Type mismatch (int vs float) | Compare Go type declarations (int64) vs Python type hints (float) |
| B08 | Incomplete enum | Verify all Go enum constants exist in Python Enum class |
| B11 | JSON convention violation | Enforce snake_case in Go JSON tags |
| B12 | Null safety violation | Detect method calls on Optional fields without null checks |
| B14 | Pagination structure mismatch | Compare pagination field names across languages |
| B15 | Missing ETag header | Verify PATCH/PUT methods include conditional headers |

---

## Summary

This validation set demonstrates cross-language contract testing between Go (server) and Python (client) for an inventory management API. The PERFECT implementation shows correct JSON serialization, type mapping, null handling, and HTTP semantics. The B01-B15 violations illustrate common pitfalls when integrating strongly-typed languages with different conventions.

**Key Testing Areas:**
1. JSON field naming conventions (snake_case consistency)
2. Type system mapping (Go int64 ↔ Python int, nullable types)
3. Enum exhaustiveness across languages
4. Null/None safety in dynamically-typed code
5. Pagination contract consistency
6. Conditional update mechanisms (ETags, optimistic locking)
7. Error handling and status code interpretation

**Total Files:** ~25
**Total Lines:** ~1,100
**Languages:** Go, Python
**Frameworks:** chi, httpx, Pydantic, pytest


