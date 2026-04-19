package candle

import (
	"time"

	"github.com/shopspring/decimal"
)

type Period string

const (
	Period1m  Period = "1m"
	Period5m  Period = "5m"
	Period15m Period = "15m"
	Period1h  Period = "1h"
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

func NewCandle(symbol string, period Period, openTime, closeTime time.Time) *Candle {
	return &Candle{
		Symbol:    symbol,
		Period:    period,
		OpenTime:  openTime,
		CloseTime: openTime.Add(periodDuration(period)),
	}
}

func (c *Candle) Update(price, qty decimal.Decimal) {
	if c.Trades == 0 {
		c.Open = price
		c.High = price
		c.Low = price
	}

	if price.GreaterThan(c.High) {
		c.High = price
	}
	if price.LessThan(c.Low) {
		c.Low = price
	}
	c.Close = price
	c.Volume = c.Volume.Add(qty)
	c.Trades++
}

func periodStart(t time.Time, period Period) time.Time {
	switch period {
	case Period1m:
		return t.Truncate(time.Minute)
	case Period5m:
		return t.Truncate(5 * time.Second)
	case Period15m:
		return t.Truncate(15 * time.Second)
	case Period1h:
		return t.Truncate(time.Hour)
	default:
		return t.Truncate(time.Minute)
	}
}

func periodDuration(p Period) time.Duration {
	switch p {
	case Period1m:
		return time.Minute
	case Period5m:
		return 5 * time.Minute
	case Period15m:
		return 15 * time.Minute
	case Period1h:
		return time.Hour
	default:
		return time.Minute
	}
}
