// Server returns 422 for validation errors
func (h *OrderHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	if req.TotalAmount <= 0 {
		http.Error(w, "total_amount must be greater than 0", http.StatusUnprocessableEntity)
		return
	}
	// ...
}
