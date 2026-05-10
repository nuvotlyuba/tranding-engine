package http

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
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
	r.Use(middleware.RequestID)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Compress(5, "application/json"))
	// r.Use(middleware.Timeout(cfg.Timeout.RequestTimeout)) // обычно 10s
	r.Use(castom_middleware.Logger(c.logger))

	r.Use(cors.Handler(cors.Options{
		// AllowedOrigins:   cfg.CORS.AllowedOrigins, // из конфига, не "*"
		AllowedMethods:   []string{"GET", "POST", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Authorization", "Content-Type", "X-Request-ID"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	r.Route("/api/v1", func(r chi.Router) {
		for _, router := range c.routers {
			router.Bind(r)
		}
	})

	chi.Walk(r, func(method, route string, _ http.Handler, _ ...func(http.Handler) http.Handler) error {
		c.logger.Debug("endpoint registered", "method", method, "path", route)
		return nil
	})

	// r.Mount("/", r)

	return r
}
