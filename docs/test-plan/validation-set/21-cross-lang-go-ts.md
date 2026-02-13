# 21 — Cross-Language Contract: Go Server + TypeScript Client

**Why included:** Hardest case — Go json tags vs TS interfaces. Tests CTR-json-tag-match, CTR-shared-type-sync across languages, CTR-strictness-parity.

**Architecture:**
- `server/` (Go): handlers/orders.go, models/order.go, middleware/auth.go
- `client/` (TS): src/types.ts, src/order-client.ts, src/order-client.test.ts

**API: Order Management**
- `POST /api/orders` — Create order
- `GET /api/orders/:id` — Get by ID
- `GET /api/orders` — List with cursor pagination
- `PATCH /api/orders/:id/status` — Update status with optimistic locking

---

## Manifest Fragment

```yaml
validation_sets:
  - id: "21-cross-lang-go-ts"
    name: "Cross-Language Contract (Go + TypeScript)"
    languages: ["go", "typescript"]
    why: "Hardest case — json tag mismatches across language boundaries"

    files:
      # Go server
      - path: "server/models/order.go"
        language: "go"
        defects: ["B05", "B07", "B11"]

      - path: "server/handlers/orders.go"
        language: "go"
        defects: []

      - path: "server/handlers/orders_test.go"
        language: "go"
        defects: ["B03"]

      - path: "server/middleware/auth.go"
        language: "go"
        defects: []

      # TypeScript client
      - path: "client/src/types.ts"
        language: "typescript"
        defects: ["B05", "B06", "B07", "B11"]

      - path: "client/src/order-client.ts"
        language: "typescript"
        defects: ["B01", "B02", "B09", "B10", "B12", "B13", "B14", "B15"]

      - path: "client/src/order-client.test.ts"
        language: "typescript"
        defects: ["B04", "B08"]

    rules_tested:
      - "CTR-json-tag-match"        # Go json tags must match TS field names
      - "CTR-shared-type-sync"      # Shared types must be consistent
      - "CTR-strictness-parity"     # Both sides must validate same constraints
      - "ERR-no-error-check"        # Missing error handling
      - "TEST-assertion-mismatch"   # Different assertion libraries
      - "API-missing-auth"          # Missing authentication
      - "API-type-mismatch"         # Incompatible types across languages
      - "API-partial-enum"          # Client handles subset of server enum
      - "API-null-safety"           # Null handling differences
      - "API-pagination-mismatch"   # Different pagination strategies
      - "API-no-version-check"      # Missing optimistic locking

    defects:
      - id: "B01"
        description: "TS client missing try/catch, Go missing error check"
        severity: "high"
        locations: ["client/src/order-client.ts:45-52"]

      - id: "B02"
        description: "TS client ignores HTTP status code"
        severity: "high"
        locations: ["client/src/order-client.ts:28"]

      - id: "B03"
        description: "Go uses assert.NotNil, TS uses expect().toBeDefined() — assertion mismatch"
        severity: "low"
        locations: ["server/handlers/orders_test.go:67"]

      - id: "B04"
        description: "TS tests missing 400/422 validation error cases"
        severity: "medium"
        locations: ["client/src/order-client.test.ts"]

      - id: "B05"
        description: "TS sends customerId, Go expects customer_id (json tag mismatch)"
        severity: "critical"
        locations: ["server/models/order.go:15", "client/src/types.ts:12"]

      - id: "B06"
        description: "TS interface missing total_amount field"
        severity: "high"
        locations: ["client/src/types.ts:8-16"]

      - id: "B07"
        description: "Go uses int64, TS uses string for total_amount (type mismatch)"
        severity: "critical"
        locations: ["server/models/order.go:18", "client/src/types.ts:14"]

      - id: "B08"
        description: "TS handles 3 of 5 order statuses (partial enum)"
        severity: "medium"
        locations: ["client/src/order-client.test.ts:89-102"]

      - id: "B09"
        description: "Go validates total_amount > 0, TS allows any number"
        severity: "high"
        locations: ["client/src/order-client.ts:56"]

      - id: "B10"
        description: "TS sends non-UUID customer_id"
        severity: "medium"
        locations: ["client/src/order-client.ts:59"]

      - id: "B11"
        description: "Go json:\"created_at\" but TS uses createdAt (casing mismatch)"
        severity: "critical"
        locations: ["server/models/order.go:22", "client/src/types.ts:15"]

      - id: "B12"
        description: "Go uses *string for nullable note, TS does order.note.trim() without null check"
        severity: "high"
        locations: ["client/src/order-client.ts:115"]

      - id: "B13"
        description: "TS missing Bearer token in Authorization header"
        severity: "critical"
        locations: ["client/src/order-client.ts:25"]

      - id: "B14"
        description: "Go uses cursor pagination, TS uses offset (pagination mismatch)"
        severity: "medium"
        locations: ["client/src/order-client.ts:138"]

      - id: "B15"
        description: "TS PATCH without If-Match version check (no optimistic locking)"
        severity: "high"
        locations: ["client/src/order-client.ts:156"]
```

---

## PERFECT — Go Server (Backend)

### server/models/order.go

```go
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
```

### server/handlers/orders.go

```go
// orders.go — HTTP handlers for order management.
package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"myapp/models"
	"myapp/storage"
)

type OrderHandler struct {
	store storage.OrderStore
}

func NewOrderHandler(store storage.OrderStore) *OrderHandler {
	return &OrderHandler{store: store}
}

// CreateOrder handles POST /api/orders.
func (h *OrderHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	var req models.CreateOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// Validate customer_id is UUID
	if _, err := uuid.Parse(req.CustomerID); err != nil {
		http.Error(w, "customer_id must be valid UUID", http.StatusUnprocessableEntity)
		return
	}

	// Validate total_amount > 0
	if req.TotalAmount <= 0 {
		http.Error(w, "total_amount must be greater than 0", http.StatusUnprocessableEntity)
		return
	}

	order := &models.Order{
		ID:          uuid.New().String(),
		CustomerID:  req.CustomerID,
		TotalAmount: req.TotalAmount,
		Status:      models.StatusPending,
		Note:        req.Note,
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
		Version:     0,
	}

	if err := h.store.Create(r.Context(), order); err != nil {
		http.Error(w, "failed to create order", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(order)
}

// GetOrder handles GET /api/orders/:id.
func (h *OrderHandler) GetOrder(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		http.Error(w, "missing order id", http.StatusBadRequest)
		return
	}

	order, err := h.store.GetByID(r.Context(), id)
	if err != nil {
		if err == storage.ErrNotFound {
			http.Error(w, "order not found", http.StatusNotFound)
			return
		}
		http.Error(w, "failed to get order", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(order)
}

// ListOrders handles GET /api/orders with cursor pagination.
func (h *OrderHandler) ListOrders(w http.ResponseWriter, r *http.Request) {
	cursor := r.URL.Query().Get("cursor")
	limit := 20 // Default page size

	orders, nextCursor, err := h.store.List(r.Context(), cursor, limit)
	if err != nil {
		http.Error(w, "failed to list orders", http.StatusInternalServerError)
		return
	}

	resp := models.ListOrdersResponse{
		Orders:     orders,
		NextCursor: nextCursor,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// UpdateStatus handles PATCH /api/orders/:id/status with optimistic locking.
func (h *OrderHandler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		http.Error(w, "missing order id", http.StatusBadRequest)
		return
	}

	var req models.UpdateStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// Get current version from If-Match header
	ifMatch := r.Header.Get("If-Match")
	if ifMatch == "" {
		http.Error(w, "missing If-Match header for optimistic locking", http.StatusPreconditionRequired)
		return
	}

	order, err := h.store.UpdateStatus(r.Context(), id, req.Status, req.Version)
	if err != nil {
		if err == storage.ErrVersionMismatch {
			http.Error(w, "version conflict", http.StatusConflict)
			return
		}
		if err == storage.ErrNotFound {
			http.Error(w, "order not found", http.StatusNotFound)
			return
		}
		http.Error(w, "failed to update order", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(order)
}
```

### server/handlers/orders_test.go

```go
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
```

### server/middleware/auth.go

```go
// auth.go — JWT authentication middleware.
package middleware

import (
	"context"
	"net/http"
	"strings"
)

type contextKey string

const userIDKey contextKey = "user_id"

// RequireAuth validates Bearer token and adds user_id to context.
func RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "missing authorization header", http.StatusUnauthorized)
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, "invalid authorization header format", http.StatusUnauthorized)
			return
		}

		token := parts[1]
		userID, err := validateJWT(token)
		if err != nil {
			http.Error(w, "invalid token", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), userIDKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func validateJWT(token string) (string, error) {
	// Simplified validation for example
	if token == "" {
		return "", http.ErrNoCookie
	}
	return "user-123", nil
}
```

---

## PERFECT — TypeScript Client (Frontend)

### client/src/types.ts

```typescript
// types.ts — Type definitions matching Go server EXACTLY.

export type OrderStatus =
  | "pending"
  | "confirmed"
  | "processing"
  | "shipped"
  | "delivered";

export interface Order {
  id: string;
  customer_id: string;        // CORRECT: matches Go json tag
  total_amount: number;        // CORRECT: number type (cents)
  status: OrderStatus;
  note: string | null;         // CORRECT: nullable
  created_at: string;          // CORRECT: ISO 8601 timestamp
  updated_at: string;
  version: number;
}

export interface CreateOrderRequest {
  customer_id: string;         // CORRECT: snake_case
  total_amount: number;
  note?: string | null;
}

export interface UpdateStatusRequest {
  status: OrderStatus;
  version: number;             // CORRECT: optimistic locking
}

export interface ListOrdersResponse {
  orders: Order[];
  next_cursor?: string | null; // CORRECT: cursor pagination
}

export class OrderAPIError extends Error {
  constructor(
    message: string,
    public status: number,
    public body: unknown
  ) {
    super(message);
    this.name = "OrderAPIError";
  }
}
```

---

## BUGS — Cross-Language Contract Violations

### B01 — No Error Handling (TQ-error-path-coverage)

**Bug:** TypeScript client fetch() call missing try/catch. Network failures, DNS errors, or timeouts will crash the application.

**Expected violation:** `TQ-error-path-coverage`

**Go Server:**
```go
// Server side error handling (CORRECT)
func (h *OrderHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	var req models.CreateOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.store.Create(r.Context(), order); err != nil {
		http.Error(w, "failed to create order", http.StatusInternalServerError)
		return
	}
	// ...
}
```

**TypeScript Client:**
```typescript
// DEFECT B01: No try/catch — network errors will crash
async function createOrder(req: CreateOrderRequest): Promise<Order> {
  const response = await fetch("/api/orders", {  // UNHANDLED rejection
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(req),
  });

  return response.json();  // No error path
}
```

**Why Stricture catches this:** TQ-error-path-coverage detects async functions without try/catch or .catch() handlers. Cross-language: Go has explicit error returns, TS needs exception handling.

---

### B02 — No Status Code Check (CTR-status-code-handling)

**Bug:** TypeScript client ignores `response.ok` check. Server 400/422/500 responses are silently treated as success.

**Expected violation:** `CTR-status-code-handling`

**Go Server:**
```go
// Server returns 422 for validation errors
func (h *OrderHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	if req.TotalAmount <= 0 {
		http.Error(w, "total_amount must be greater than 0", http.StatusUnprocessableEntity)
		return
	}
	// ...
}
```

**TypeScript Client:**
```typescript
// DEFECT B02: No response.ok check
async function createOrder(req: CreateOrderRequest): Promise<Order> {
  const response = await fetch("/api/orders", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(req),
  });

  // BUG: Missing if (!response.ok) throw ...
  return response.json();  // Will parse error HTML as JSON and fail
}
```

**Why Stricture catches this:** CTR-status-code-handling enforces that every fetch() call checks `response.ok` or `response.status` before parsing. Go handlers return explicit HTTP errors; TS client must handle them.

---

### B03 — Shallow Test Assertions (TQ-no-shallow-assertions)

**Bug:** Go test uses `assert.NotNil()` (existence check), TypeScript would use `expect().toBeDefined()`. Cross-language assertion mismatch creates incomplete test coverage.

**Expected violation:** `TQ-no-shallow-assertions`

**Go Server:**
```go
// DEFECT B03: Shallow assertion — only checks ID exists
func TestCreateOrder_Success(t *testing.T) {
	// ...
	var order models.Order
	err := json.NewDecoder(w.Body).Decode(&order)
	require.NoError(t, err)

	assert.NotNil(t, order.ID)  // BUG: Should validate it's a valid UUID
	assert.Equal(t, reqBody.CustomerID, order.CustomerID)
	assert.Equal(t, reqBody.TotalAmount, order.TotalAmount)
}
```

**TypeScript Client:**
```typescript
// Equivalent shallow assertion in TS tests
test("createOrder returns order", async () => {
  const order = await client.createOrder({
    customer_id: "550e8400-e29b-41d4-a716-446655440000",
    total_amount: 5000,
  });

  // DEFECT B03: Shallow assertion
  expect(order).toBeDefined();  // Should validate order.id is UUID format
  expect(order.id).toBeTruthy();  // Weak — doesn't validate UUID format
});
```

**Why Stricture catches this:** TQ-no-shallow-assertions flags `.toBeDefined()`, `.toBeTruthy()`, `assert.NotNil()` without deeper property validation. Cross-language: both Go and TS should validate structure, not just existence.

---

### B04 — Missing Negative Tests (TQ-negative-cases)

**Bug:** TypeScript test suite missing validation error cases (400/422). Only tests happy path.

**Expected violation:** `TQ-negative-cases`

**Go Server:**
```go
// Go server has proper error handling
func (h *OrderHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	if _, err := uuid.Parse(req.CustomerID); err != nil {
		http.Error(w, "customer_id must be valid UUID", http.StatusUnprocessableEntity)
		return
	}

	if req.TotalAmount <= 0 {
		http.Error(w, "total_amount must be greater than 0", http.StatusUnprocessableEntity)
		return
	}
	// ...
}
```

**TypeScript Client:**
```typescript
// DEFECT B04: Test suite missing negative cases
describe("OrderClient", () => {
  test("createOrder success", async () => {
    const order = await client.createOrder({
      customer_id: "550e8400-e29b-41d4-a716-446655440000",
      total_amount: 5000,
    });
    expect(order.status).toBe("pending");
  });

  // BUG: Missing tests for:
  // - Invalid UUID → 422
  // - Negative amount → 422
  // - Missing auth → 401
  // - Server error → 500
});
```

**Why Stricture catches this:** TQ-negative-cases requires at least one test per documented error response (400, 401, 422, 500). Cross-language: Go tests cover error paths, TS tests must match.

---

### B05 — Request Missing Required Fields (CTR-request-shape)

**Bug:** TypeScript client sends `customerId` (camelCase) but Go expects `customer_id` (snake_case). Request fails silently or gets 400.

**Expected violation:** `CTR-request-shape`

**Go Server:**
```go
// Go expects snake_case JSON tags
type CreateOrderRequest struct {
	CustomerID  string  `json:"customer_id" validate:"required,uuid4"`  // CORRECT
	TotalAmount int64   `json:"total_amount" validate:"required,gt=0"`
	Note        *string `json:"note,omitempty"`
}
```

**TypeScript Client:**
```typescript
// DEFECT B05: TS sends wrong field name
async function createOrder(req: CreateOrderRequest): Promise<Order> {
  const payload = {
    customerId: req.customer_id,    // BUG: camelCase instead of snake_case
    totalAmount: req.total_amount,  // BUG: camelCase instead of snake_case
    note: req.note,
  };

  const response = await fetch("/api/orders", {
    method: "POST",
    body: JSON.stringify(payload),
  });

  return response.json();  // Go will reject with 400 or ignore fields
}
```

**Why Stricture catches this:** CTR-request-shape validates that TypeScript request payloads match Go struct json tags exactly. Detects camelCase/snake_case mismatches across language boundaries.

---

### B06 — Response Type Mismatch (CTR-response-shape)

**Bug:** TypeScript interface missing `total_amount` field that Go sends. Client code will see `undefined` for critical field.

**Expected violation:** `CTR-response-shape`

**Go Server:**
```go
// Go sends complete Order struct
type Order struct {
	ID          string      `json:"id"`
	CustomerID  string      `json:"customer_id"`
	TotalAmount int64       `json:"total_amount"`  // SENT
	Status      OrderStatus `json:"status"`
	Note        *string     `json:"note,omitempty"`
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"`
	Version     int         `json:"version"`
}
```

**TypeScript Client:**
```typescript
// DEFECT B06: Missing total_amount field
export interface Order {
  id: string;
  customer_id: string;
  // BUG: total_amount field missing — Go sends it but TS doesn't declare it
  status: OrderStatus;
  note: string | null;
  created_at: string;
  updated_at: string;
  version: number;
}

// Client code will crash
function calculateTax(order: Order): number {
  return order.total_amount * 0.08;  // TypeScript error: Property does not exist
}
```

**Why Stricture catches this:** CTR-response-shape validates TypeScript interfaces match Go struct json tags. Detects missing fields that server sends but client doesn't expect.

---

### B07 — Wrong Field Types (CTR-manifest-conformance)

**Bug:** Go uses `int64` for `total_amount` (cents), TypeScript treats as `number` but uses it as string. Precision loss for values >2^53.

**Expected violation:** `CTR-manifest-conformance`

**Go Server:**
```go
// Go sends int64 (safe for large monetary values)
type Order struct {
	TotalAmount int64 `json:"total_amount"`  // CORRECT: cents as int64
}

// Example: 9007199254740992 cents ($90,071,992,547,409.92)
```

**TypeScript Client:**
```typescript
// DEFECT B07: Treats total_amount as string when Go sends number
export interface Order {
  id: string;
  customer_id: string;
  total_amount: string;  // BUG: Go sends number, TS expects string
  status: OrderStatus;
  // ...
}

// Client code crashes
async function displayPrice(order: Order) {
  // BUG: order.total_amount is number from JSON, but TS thinks it's string
  const dollars = order.total_amount / 100;  // Runtime error: NaN
  console.log(`$${dollars.toFixed(2)}`);     // "NaN"
}
```

**Why Stricture catches this:** CTR-manifest-conformance validates TypeScript types match Go types. Detects int64→string mismatches where precision loss or type coercion occurs.

---

### B08 — Incomplete Enum Handling (CTR-strictness-parity)

**Bug:** TypeScript test only handles 3 of 5 order statuses. Missing "processing" and "delivered" cases.

**Expected violation:** `CTR-strictness-parity`

**Go Server:**
```go
// Go defines 5 statuses
type OrderStatus string

const (
	StatusPending    OrderStatus = "pending"
	StatusConfirmed  OrderStatus = "confirmed"
	StatusProcessing OrderStatus = "processing"   // TS MISSING
	StatusShipped    OrderStatus = "shipped"
	StatusDelivered  OrderStatus = "delivered"    // TS MISSING
)
```

**TypeScript Client:**
```typescript
// DEFECT B08: TS only handles 3 of 5 statuses
test("order status transitions", async () => {
  const order = await client.createOrder(validRequest);

  // BUG: Only tests 3 statuses, missing "processing" and "delivered"
  switch (order.status) {
    case "pending":
      // test logic
      break;
    case "confirmed":
      // test logic
      break;
    case "shipped":
      // test logic
      break;
    // MISSING: "processing" and "delivered" cases
    default:
      throw new Error("Unknown status");  // Will crash on valid statuses
  }
});
```

**Why Stricture catches this:** CTR-strictness-parity validates that TypeScript enum handling covers all values defined in Go enum. Detects partial enum coverage in tests or switch statements.

---

### B09 — Missing Range Validation (CTR-strictness-parity)

**Bug:** Go validates `total_amount > 0`, TypeScript client allows any number including negative values.

**Expected violation:** `CTR-strictness-parity`

**Go Server:**
```go
// Go enforces validation rules
func (h *OrderHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	var req models.CreateOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// VALIDATION: total_amount must be positive
	if req.TotalAmount <= 0 {
		http.Error(w, "total_amount must be greater than 0", http.StatusUnprocessableEntity)
		return
	}
	// ...
}
```

**TypeScript Client:**
```typescript
// DEFECT B09: No client-side validation
async function createOrder(req: CreateOrderRequest): Promise<Order> {
  // BUG: Sends request without validating total_amount > 0
  // Client allows negative values, server will reject with 422
  const response = await fetch("/api/orders", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(req),  // No validation here
  });

  return response.json();
}

// Usage that should fail client-side but doesn't
await createOrder({
  customer_id: "550e8400-e29b-41d4-a716-446655440000",
  total_amount: -5000,  // INVALID: negative amount sent to server
});
```

**Why Stricture catches this:** CTR-strictness-parity enforces that validation rules in Go (struct tags, handler checks) must have equivalent client-side validation in TypeScript. Prevents unnecessary round-trips for validation errors.

---

### B10 — Format Not Validated (CTR-strictness-parity)

**Bug:** Go expects UUID format for `customer_id`, TypeScript client doesn't validate before sending.

**Expected violation:** `CTR-strictness-parity`

**Go Server:**
```go
// Go validates UUID format
func (h *OrderHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	var req models.CreateOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// VALIDATION: customer_id must be valid UUID
	if _, err := uuid.Parse(req.CustomerID); err != nil {
		http.Error(w, "customer_id must be valid UUID", http.StatusUnprocessableEntity)
		return
	}
	// ...
}
```

**TypeScript Client:**
```typescript
// DEFECT B10: No UUID format validation
async function createOrder(req: CreateOrderRequest): Promise<Order> {
  // BUG: Accepts any string for customer_id, no UUID validation
  const response = await fetch("/api/orders", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(req),
  });

  return response.json();
}

// Usage that should fail client-side but doesn't
await createOrder({
  customer_id: "not-a-uuid",  // INVALID: server will reject with 422
  total_amount: 5000,
});

// Should validate:
// const UUID_REGEX = /^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/i;
// if (!UUID_REGEX.test(req.customer_id)) throw new Error("Invalid UUID");
```

**Why Stricture catches this:** CTR-strictness-parity detects when Go has format validation (UUID, email, URL) but TypeScript client doesn't validate before sending. Prevents unnecessary API calls that will fail validation.

---

### B11 — JSON Tag Mismatch (CTR-json-tag-match)

**Bug:** Go uses `json:"created_at"` (snake_case), TypeScript uses `createdAt` (camelCase). Response parsing fails or fields are undefined.

**Expected violation:** `CTR-json-tag-match`

**Go Server:**
```go
// Go uses snake_case json tags (CORRECT)
type Order struct {
	ID          string      `json:"id"`
	CustomerID  string      `json:"customer_id"`
	TotalAmount int64       `json:"total_amount"`
	Status      OrderStatus `json:"status"`
	Note        *string     `json:"note,omitempty"`
	CreatedAt   time.Time   `json:"created_at"`  // SNAKE_CASE
	UpdatedAt   time.Time   `json:"updated_at"`  // SNAKE_CASE
	Version     int         `json:"version"`
}

// JSON response:
// {"id":"123","customer_id":"...","created_at":"2026-02-13T10:30:00Z",...}
```

**TypeScript Client:**
```typescript
// DEFECT B11: Uses camelCase instead of matching Go json tags
export interface Order {
  id: string;
  customer_id: string;  // CORRECT
  total_amount: number; // CORRECT
  status: OrderStatus;
  note: string | null;
  createdAt: string;    // BUG: Go sends "created_at" but TS expects "createdAt"
  updatedAt: string;    // BUG: Go sends "updated_at" but TS expects "updatedAt"
  version: number;
}

// Client code crashes
async function displayOrder(order: Order) {
  console.log(order.createdAt);  // undefined — field doesn't exist in JSON
  // Should access order.created_at instead
}
```

**Why Stricture catches this:** CTR-json-tag-match validates that TypeScript interface field names exactly match Go struct json tags. This is the hardest cross-language bug: snake_case vs camelCase.

---

### B12 — Nullable Field Crashes (CTR-response-shape)

**Bug:** Go uses `*string` for nullable `note` field, TypeScript doesn't check for null before calling `.trim()`.

**Expected violation:** `CTR-response-shape`

**Go Server:**
```go
// Go properly handles nullable fields
type Order struct {
	ID          string      `json:"id"`
	CustomerID  string      `json:"customer_id"`
	TotalAmount int64       `json:"total_amount"`
	Status      OrderStatus `json:"status"`
	Note        *string     `json:"note,omitempty"`  // NULLABLE: can be null or absent
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"`
	Version     int         `json:"version"`
}

// JSON response when note is null:
// {"id":"123","customer_id":"...","note":null,...}
```

**TypeScript Client:**
```typescript
// DEFECT B12: Doesn't handle null before string operations
async function displayOrderNote(orderId: string): Promise<void> {
  const order = await client.getOrder(orderId);

  // BUG: order.note can be null, but code assumes it's always a string
  const trimmedNote = order.note.trim();  // Runtime error: Cannot read property 'trim' of null
  console.log(`Note: ${trimmedNote}`);

  // Should be:
  // const trimmedNote = order.note?.trim() ?? "No note";
}
```

**Why Stricture catches this:** CTR-response-shape validates that TypeScript code handles nullable fields from Go `*T` pointer types. Detects missing null checks before string/object operations.

---

### B13 — Missing Auth Header (CTR-request-shape)

**Bug:** Go middleware requires `Authorization: Bearer <token>` header, TypeScript client doesn't send it.

**Expected violation:** `CTR-request-shape`

**Go Server:**
```go
// Go middleware enforces auth header
func RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "missing authorization header", http.StatusUnauthorized)
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, "invalid authorization header format", http.StatusUnauthorized)
			return
		}
		// ...
	})
}
```

**TypeScript Client:**
```typescript
// DEFECT B13: Missing Authorization header
async function createOrder(req: CreateOrderRequest): Promise<Order> {
  const response = await fetch("/api/orders", {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      // BUG: Missing Authorization header
      // Should have: "Authorization": `Bearer ${getToken()}`
    },
    body: JSON.stringify(req),
  });

  return response.json();  // Server returns 401, client crashes
}
```

**Why Stricture catches this:** CTR-request-shape validates that required headers from Go middleware are present in TypeScript fetch() calls. Detects missing auth, content-type, or custom headers.

---

### B14 — Pagination Mismatch (CTR-response-shape)

**Bug:** Go uses cursor-based pagination, TypeScript client expects offset/page numbers.

**Expected violation:** `CTR-response-shape`

**Go Server:**
```go
// Go uses cursor pagination (CORRECT)
type ListOrdersResponse struct {
	Orders     []Order `json:"orders"`
	NextCursor *string `json:"next_cursor,omitempty"`  // CURSOR
}

func (h *OrderHandler) ListOrders(w http.ResponseWriter, r *http.Request) {
	cursor := r.URL.Query().Get("cursor")  // Expects "cursor" param
	limit := 20

	orders, nextCursor, err := h.store.List(r.Context(), cursor, limit)
	// ...
}
```

**TypeScript Client:**
```typescript
// DEFECT B14: Uses offset pagination instead of cursor
async function listOrders(page: number = 1, limit: number = 20): Promise<Order[]> {
  const offset = (page - 1) * limit;

  // BUG: Sends offset/limit params, but Go expects cursor param
  const response = await fetch(
    `/api/orders?offset=${offset}&limit=${limit}`,  // WRONG params
    { method: "GET" }
  );

  const data = await response.json();
  return data.orders;  // Ignores next_cursor field

  // Should use:
  // `/api/orders?cursor=${lastCursor}&limit=${limit}`
}
```

**Why Stricture catches this:** CTR-response-shape validates that pagination strategy matches between Go and TypeScript. Detects offset vs cursor, page-based vs token-based mismatches.

---

### B15 — Race Condition (CTR-request-shape)

**Bug:** TypeScript PATCH request doesn't send `If-Match` header for optimistic locking. Concurrent updates will overwrite each other.

**Expected violation:** `CTR-request-shape`

**Go Server:**
```go
// Go enforces optimistic locking with If-Match header
func (h *OrderHandler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var req models.UpdateStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// REQUIRES If-Match header for version check
	ifMatch := r.Header.Get("If-Match")
	if ifMatch == "" {
		http.Error(w, "missing If-Match header for optimistic locking", http.StatusPreconditionRequired)
		return
	}

	order, err := h.store.UpdateStatus(r.Context(), id, req.Status, req.Version)
	if err == storage.ErrVersionMismatch {
		http.Error(w, "version conflict", http.StatusConflict)
		return
	}
	// ...
}
```

**TypeScript Client:**
```typescript
// DEFECT B15: Missing If-Match header for optimistic locking
async function updateOrderStatus(
  orderId: string,
  status: OrderStatus,
  version: number
): Promise<Order> {
  const response = await fetch(`/api/orders/${orderId}/status`, {
    method: "PATCH",
    headers: {
      "Content-Type": "application/json",
      // BUG: Missing If-Match header
      // Should have: "If-Match": version.toString()
    },
    body: JSON.stringify({ status, version }),
  });

  return response.json();  // Server returns 428 Precondition Required
}

// Race condition scenario:
// User A fetches order (version=0)
// User B fetches order (version=0)
// User A updates status → version=1 (succeeds)
// User B updates status → should fail with 409 Conflict, but gets 428 instead
```

**Why Stricture catches this:** CTR-request-shape validates that Go handlers requiring optimistic locking headers have matching TypeScript client logic. Detects missing If-Match, ETag, or version headers that prevent race conditions.

