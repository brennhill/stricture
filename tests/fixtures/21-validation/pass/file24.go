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
