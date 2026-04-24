package indicator

import (
	domain_candle "github.com/nuvotlyuba/trading-engine/internal/domain/candle"
	"github.com/shopspring/decimal"
)

type RSI struct {
	symbol    string
	period    domain_candle.Period
	n         int
	avgGain   decimal.Decimal
	avgLoss   decimal.Decimal
	prevClose decimal.Decimal
	count     int
}

func NewRSI(symbol string, period domain_candle.Period, n int) *RSI {
	return &RSI{
		symbol: symbol,
		period: period,
		n:      n,
	}
}

func (r *RSI) Update(c domain_candle.Candle) (Value, error) {
	r.count++

	if r.count == 1 {
		r.prevClose = c.Close
		return Value{
			Symbol:    r.symbol,
			Period:    r.period,
			Timestamp: c.CloseTime,
			Data:      map[string]decimal.Decimal{"rsi": decimal.Zero},
		}, nil
	}

	delta := c.Close.Sub(r.prevClose)
	r.prevClose = c.Close

	var gain, loss decimal.Decimal
	if delta.IsPositive() {
		gain = delta
	} else {
		loss = delta.Abs()
	}

	k := decimal.NewFromInt(2).Div(decimal.NewFromInt(int64(r.n + 1)))
	one := decimal.NewFromInt(1)

	r.avgGain = gain.Mul(k).Add(r.avgGain.Mul(one.Sub(k)))
	r.avgLoss = loss.Mul(k).Add(r.avgLoss.Mul(one.Sub(k)))

	var rsi decimal.Decimal
	if r.avgLoss.IsZero() {
		rsi = decimal.NewFromInt(100)
	} else {
		rs := r.avgGain.Div(r.avgLoss)
		rsi = decimal.NewFromInt(100).Sub(
			decimal.NewFromInt(100).Div(decimal.NewFromInt(1).Add(rs)),
		)
	}
	return Value{
		Symbol: r.symbol, Period: r.period,
		Timestamp: c.CloseTime,
		Data:      map[string]decimal.Decimal{"rsi": rsi},
	}, nil
}

func (r *RSI) Name() string                 { return "RSI" }
func (r *RSI) Symbol() string               { return r.symbol }
func (r *RSI) Period() domain_candle.Period { return r.period }
func (r *RSI) WarmUp() int                  { return r.n }
func (r *RSI) IsReady() bool                { return r.count >= r.n }
