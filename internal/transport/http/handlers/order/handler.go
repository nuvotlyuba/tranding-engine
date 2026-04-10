package order

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type orderHandler struct {
	// svc service.OrderService
	logger *slog.Logger
}

func New(logger *slog.Logger) *orderHandler {
	return &orderHandler{
		logger: logger,
	}
}

func (h *orderHandler) Bind(r chi.Router) {
	r.Route("/order", func(r chi.Router) {
		r.Post("/", h.CreateOrder)
	})
}

func (h *orderHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {

}
