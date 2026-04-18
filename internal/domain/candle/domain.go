package candle

import (
	"time"

	"github.com/shopspring/decimal"
)

type Candle struct {
	Symbol    string
	Period    Period
	OpenTime  time.Time
	CloseTime time.Time
	Open      decimal.Decimal
	High      decimal.Decimal
	Low       decimal.Decimal
	Close     decimal.Decimal
	Volume    decimal.Decimal
	Trades    int
}

type Period string

const (
	Period1m  Period = "1m"
	Period5m  Period = "5m"
	Period15m Period = "15m"
	Period1h  Period = "1h"
)
