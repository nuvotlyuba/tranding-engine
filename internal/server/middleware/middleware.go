package middleware

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

func Logger(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/health" {
				next.ServeHTTP(w, r)
				return
			}

			start := time.Now()
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			defer func() {
				status := ww.Status()

				attrs := []any{
					"method", r.Method,
					"path", r.URL.Path,
					"status", status,
					"bytes", ww.BytesWritten(),
					"duration", time.Since(start),
					"request_id", middleware.GetReqID(r.Context()),
					"remote_addr", r.RemoteAddr,
					"user_agent", r.UserAgent(),
				}

				switch {
				case status >= 500:
					logger.Error("HTTP request", attrs...)
				case status >= 400:
					logger.Warn("HTTP request", attrs...)
				default:
					logger.Info("HTTP request", attrs...)
				}
			}()

			next.ServeHTTP(ww, r)
		})
	}
}

func CORS() func(http.Handler) http.Handler {
	return cors.Handler(cors.Options{
		// AllowedOrigins: s.opts.corsOrigins,
		AllowedMethods: []string{
			http.MethodGet,
			http.MethodPost,
			http.MethodDelete,
			http.MethodOptions,
		},
		AllowedHeaders: []string{
			"Authorization",
			"Content-Type",
			"X-Request-ID",
		},
		// разрешаем куки и Authorization заголовки
		AllowCredentials: true,
		// браузер кеширует preflight на 5 минут
		MaxAge: 300,
	})
}
