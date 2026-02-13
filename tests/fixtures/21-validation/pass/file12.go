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
