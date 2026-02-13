// Go uses cursor pagination (CORRECT)
type ListOrdersResponse struct {
	Orders     []Order `json:"orders"`
	NextCursor *string `json:"next_cursor,omitempty"`  // CURSOR
}

func (h *OrderHandler) ListOrders(w http.ResponseWriter, r *http.Request) {
	cursor := r.URL.Query().Get("cursor")  // Expects "cursor" param
	limit := 20

	orders, nextCursor, err := h.store.List(r.Context(), cursor, limit)
	// ...
}
