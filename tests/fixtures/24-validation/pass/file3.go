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
