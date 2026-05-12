package jobs

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/nuvotlyuba/trading-engine/internal/config"
	"github.com/nuvotlyuba/trading-engine/internal/domain/candle"
	"github.com/nuvotlyuba/trading-engine/internal/domain/indicator"
	"github.com/nuvotlyuba/trading-engine/internal/infra/clients/binance"
)

type CandleSeeder struct {
	cache      *candle.Cache
	indicators map[string][]indicator.Indicator
	cli        *binance.Client
	symbols    []string
	periods    []seedConfig
	logger     *slog.Logger
}

type seedConfig struct {
	period candle.Period
	limit  int // сколько свечей загрузить
}

func NewCandleSeeder(cache *candle.Cache, cfg config.Config, logger *slog.Logger, client *binance.Client) *CandleSeeder {
	return &CandleSeeder{
		cache:      cache,
		indicators: cfg.SeederJob.Indicators, // спросить у клода
		symbols:    cfg.SeederJob.Symbols,
		logger:     logger,
		cli:        client,
		periods: []seedConfig{
			{candle.Period1m, 500}, // 500 минутных свечей = ~8 часов истории
			{candle.Period5m, 500}, // 500 пятиминутных = ~41 час
			{candle.Period1h, 200}, // 200 часовых = ~8 дней
		},
	}
}

func (s *CandleSeeder) Run(ctx context.Context) error {
	s.logger.Info("candle seeder started", "symbol", s.symbols)
	start := time.Now()

	for _, symbol := range s.symbols {
		for _, cfg := range s.periods {
			if err := s.seedOne(ctx, symbol, cfg); err != nil {
				// не падаем при ошибке одного символа — продолжаем
				s.logger.Error("failed to seed candles",
					"symbol", symbol,
					"period", cfg.period,
					"error", err,
				)
				continue
			}

			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(100 * time.Millisecond):
			}
		}
	}
	s.logger.Info("candle seeder finished",
		"duration", time.Since(start),
		"symbols", s.symbols,
	)
	return nil
}

func (s *CandleSeeder) seedOne(ctx context.Context, symbol string, cfg seedConfig) error {
	candles, err := s.cli.GetKlines(ctx, symbol, cfg.period, cfg.limit)
	if err != nil {
		return fmt.Errorf("fetch klines: %w", err)
	}

	for _, c := range candles {
		s.cache.Push(c)
	}

	key := fmt.Sprintf("%s_%s, symbol, cfg.Period")
	if inds, ok := s.indicators[key]; ok {
		for _, c := range candles {
			for _, ind := range inds {
				ind.Update(c) // игнорируем ошибки при прогреве
			}
		}
		s.logger.Info("indicators warmed up",
			"symbol", symbol,
			"period", cfg.period,
			"candles", len(candles),
		)
	}
	return nil
}
