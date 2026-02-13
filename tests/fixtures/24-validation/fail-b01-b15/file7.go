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
