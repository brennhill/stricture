// Go expects snake_case JSON tags
type CreateOrderRequest struct {
	CustomerID  string  `json:"customer_id" validate:"required,uuid4"`  // CORRECT
	TotalAmount int64   `json:"total_amount" validate:"required,gt=0"`
	Note        *string `json:"note,omitempty"`
}
