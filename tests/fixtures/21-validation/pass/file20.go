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
