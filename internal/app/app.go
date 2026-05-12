package app

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/nuvotlyuba/trading-engine/internal/config"
	"github.com/nuvotlyuba/trading-engine/internal/domain/candle"
	"github.com/nuvotlyuba/trading-engine/internal/infra/clients/binance"
	"github.com/nuvotlyuba/trading-engine/internal/infra/postgres"
	"github.com/nuvotlyuba/trading-engine/internal/jobs"
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

	if err := postgres.RunMigrations(cfg.PostgresDB, "migrations"); err != nil {
		return fmt.Errorf("run migrations: %w", err)
	}

	db, err := postgres.New(ctx, cfg.PostgresDB)
	if err != nil {
		return fmt.Errorf("connect postgres: %w", err)
	}
	defer db.Close()
	logger.Info("postgres connected",
		"host", cfg.PostgresDB.Host,
		"port", cfg.PostgresDB.Port,
		"db", cfg.PostgresDB.DB,
	)

	cache := candle.NewCache(cfg.CacheSize)
	binanceClient := binance.NewClient(cfg, logger)

	seeder := jobs.NewCandleSeeder(
		cache,
		cfg,
		logger,
		binanceClient,
	)

	err = seeder.Run(ctx)
	if err != nil {
		return fmt.Errorf("run seeder: %w", err)
	}

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
