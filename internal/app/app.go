package app

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/nuvotlyuba/trading-engine/internal/config"
	"github.com/nuvotlyuba/trading-engine/internal/server"
	controller "github.com/nuvotlyuba/trading-engine/internal/transport/http"
	orderHndr "github.com/nuvotlyuba/trading-engine/internal/transport/http/handlers/order"
)

func InitAndRun(ctx context.Context) error {

	cfg, err := config.Load(ctx)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	// 2. инфраструктура
	// db := postgres.New(cfg.DSN)
	// cache := cache.NewLRU(cfg.CacheSize)
	// bus := eventbus.New()

	// // 3. репозитории
	// orderRepo := postgres.NewOrderRepository(db)

	// // 4. сервисы
	// matchingSvc := app.NewMatchingService(orderRepo, bus, cache)

	// userService := user.NewService(...)
	orderHandler := orderHndr.New(logger)

	ctrl := controller.New(logger,
		orderHandler,
	)

	srv := server.New(logger, cfg.HTTP, ctrl.Build())
	if err := srv.Run(); err != nil {
		return fmt.Errorf("run server: %w", err)
	}
	return nil
}
