package binance

import (
	"time"

	"github.com/nuvotlyuba/trading-engine/internal/domain/candle"
)

func convertToBinanceInterval(p candle.Period) string {
	switch p {
	case candle.Period1m:
		return "1m"

	case candle.Period5m:
		return "5m"

	case candle.Period15m:
		return "15m"

	case candle.Period1h:
		return "1h"

	default:
		return "1m"
	}
}

func converterPeriodToDuration(p candle.Period) time.Duration {
	switch p {
	case candle.Period1m:
		return time.Minute
	case candle.Period5m:
		return 5 * time.Minute
	case candle.Period15m:
		return 15 * time.Minute
	case candle.Period1h:
		return time.Hour
	default:
		return time.Minute
	}
}
