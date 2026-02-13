// Go server has proper error handling
func (h *OrderHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	if _, err := uuid.Parse(req.CustomerID); err != nil {
		http.Error(w, "customer_id must be valid UUID", http.StatusUnprocessableEntity)
		return
	}

	if req.TotalAmount <= 0 {
		http.Error(w, "total_amount must be greater than 0", http.StatusUnprocessableEntity)
		return
	}
	// ...
}
