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
