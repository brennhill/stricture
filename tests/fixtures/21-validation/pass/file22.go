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
