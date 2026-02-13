// Go enforces optimistic locking with If-Match header
func (h *OrderHandler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var req models.UpdateStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// REQUIRES If-Match header for version check
	ifMatch := r.Header.Get("If-Match")
	if ifMatch == "" {
		http.Error(w, "missing If-Match header for optimistic locking", http.StatusPreconditionRequired)
		return
	}

	order, err := h.store.UpdateStatus(r.Context(), id, req.Status, req.Version)
	if err == storage.ErrVersionMismatch {
		http.Error(w, "version conflict", http.StatusConflict)
		return
	}
	// ...
}
