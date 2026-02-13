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
