// handler/order_handler_cycle.go — Part 1 of cycle.
package handler

import (
	"net/http"
	"super-lint-test/internal/service"
	"super-lint-test/pkg/logging"
)

type OrderHandler struct {
	service *service.OrderService
	logger  logging.Logger
}

func (h *OrderHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	// Handler calls service (normal flow)
	h.service.CreateOrder(r.Context())
}
