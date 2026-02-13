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
