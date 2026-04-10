package http

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	castom_middleware "github.com/nuvotlyuba/trading-engine/internal/server/middleware"
)

type Router interface {
	Bind(r chi.Router)
}

type Controller struct {
	logger  *slog.Logger
	routers []Router
}

func New(logger *slog.Logger, routers ...Router) *Controller {
	return &Controller{
		logger:  logger,
		routers: routers,
	}
}

func (c *Controller) Build() *chi.Mux {
	r := chi.NewRouter()

	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(middleware.Timeout(60 * time.Second))
	r.Use(castom_middleware.Logger(c.logger))

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	r.Route("/api/v1", func(r chi.Router) {
		for _, router := range c.routers {
			router.Bind(r)
		}
	})

	return r
}
