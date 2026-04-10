package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/nuvotlyuba/trading-engine/internal/config"
)

type Server struct {
	httpServer *http.Server
	logger     *slog.Logger
	config     config.HTTP
}

func New(logger *slog.Logger, cfg config.HTTP, handler http.Handler) *Server {

	httpServer := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Handler:      handler,
		ReadTimeout:  cfg.Timeout.ReadTimeout,
		WriteTimeout: cfg.Timeout.WriteTimeout,
		IdleTimeout:  cfg.Timeout.IdleTimeout,
	}

	return &Server{
		httpServer: httpServer,
		logger:     logger,
		config:     cfg,
	}
}

func (s *Server) Run() error {
	// Канал для ошибок сервера
	serverErr := make(chan error, 1)

	go func() {
		s.logger.Info("Starting HTTP server",
			"addr", s.httpServer.Addr,
			"read_timeout", s.config.Timeout.ReadTimeout.String(), // "5s"
			"write_timeout", s.config.Timeout.WriteTimeout.String(), // "10s"
		)

		if err := s.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}
	}()

	// Ожидание сигнала или ошибки
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-serverErr:
		return fmt.Errorf("server error: %w", err)
	case sig := <-quit:
		s.logger.Info("Received shutdown signal", "signal", sig)
	}

	return s.shutdown()
}

func (s *Server) shutdown() error {
	s.logger.Info("Shutting down server gracefully...")

	ctx, cancel := context.WithTimeout(context.Background(), s.config.Timeout.ShutdownTimeout)
	defer cancel()

	if err := s.httpServer.Shutdown(ctx); err != nil {
		s.logger.Error("Graceful shutdown failed, forcing close", "error", err)
		// Принудительное закрытие как fallback
		if closeErr := s.httpServer.Close(); closeErr != nil {
			s.logger.Error("Force close error", "error", closeErr)
		}
		return fmt.Errorf("shutdown error: %w", err)
	}

	s.logger.Info("Server exited properly")
	return nil
}
