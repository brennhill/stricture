// Server side error handling (CORRECT)
func (h *OrderHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	var req models.CreateOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.store.Create(r.Context(), order); err != nil {
		http.Error(w, "failed to create order", http.StatusInternalServerError)
		return
	}
	// ...
}
