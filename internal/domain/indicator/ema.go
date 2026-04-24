package indicator

import (
	domain_candle "github.com/nuvotlyuba/trading-engine/internal/domain/candle"
	"github.com/shopspring/decimal"
)

type EMA struct {
	symbol string
	period domain_candle.Period
	n      int             // период скользящей, например, 12 или 26
	k      decimal.Decimal // сглаживающий коэффициент = 2/(n+1)
	prev   decimal.Decimal // предыдущее значение EMA
	count  int             // сколько свечей обработано
}

func NewEMA(symbol string, period domain_candle.Period, n int) *EMA {
	k := decimal.NewFromInt(2).Div(decimal.NewFromInt(int64(n + 1)))
	return &EMA{
		symbol: symbol,
		period: period,
		n:      n,
		k:      k,
	}
}

func (e *EMA) Update(c domain_candle.Candle) (Value, error) {
	e.count++
	if e.count == 1 {
		e.prev = c.Close
	} else {
		e.prev = c.Close.Mul(e.k).Add(e.prev.Mul(decimal.NewFromInt(1).Sub(e.k)))
	}

	return Value{
		Symbol:    e.symbol,
		Period:    e.period,
		Timestamp: c.CloseTime,
		Data:      map[string]decimal.Decimal{"ema": e.prev},
	}, nil
}
func (e *EMA) Name() string                 { return "EMA" }
func (e *EMA) Symbol() string               { return e.symbol }
func (e *EMA) Period() domain_candle.Period { return e.period }
func (e *EMA) WarmUp() int                  { return e.n }
func (e *EMA) IsReady() bool                { return e.count >= e.n }
func (e *EMA) Current() decimal.Decimal     { return e.prev }
