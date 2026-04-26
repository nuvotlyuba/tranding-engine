package indicator

import (
	"math"

	domain_candle "github.com/nuvotlyuba/trading-engine/internal/domain/candle"
	"github.com/shopspring/decimal"
)

type BollingerBands struct {
	symbol string
	period domain_candle.Period
	n      int               // период, обычно 20
	window []decimal.Decimal // полсдение n значений Close
	mult   decimal.Decimal   // множитель, обычно 2
}

func NewBollongerbands(symbol string, period domain_candle.Period, n int, mult decimal.Decimal) *BollingerBands {
	return &BollingerBands{
		symbol: symbol,
		period: period,
		n:      n,
		window: make([]decimal.Decimal, 0, n),
		mult:   mult,
	}
}

func (b *BollingerBands) Update(c domain_candle.Candle) (Value, error) {
	b.window = append(b.window, c.Close)
	if len(b.window) > b.n {
		b.window = b.window[1:]
	}
	if len(b.window) < b.n {
		return Value{
			Symbol:    b.symbol,
			Period:    b.period,
			Timestamp: c.CloseTime,
			Data:      map[string]decimal.Decimal{},
		}, nil
	}

	//SMA
	sum := decimal.Zero
	for _, v := range b.window {
		sum = sum.Add(v)
	}
	sma := sum.Div(decimal.NewFromInt(int64(b.n)))

	// стандартное отклонение
	variance := decimal.Zero
	for _, v := range b.window {
		diff := v.Sub(sma)
		variance = variance.Add(diff.Mul(diff))
	}
	variance = variance.Div(decimal.NewFromInt(int64(b.n)))
	stddev, _ := variance.Float64()
	stddevDec := decimal.NewFromFloat(math.Sqrt(stddev))

	upper := sma.Add(b.mult.Mul(stddevDec))
	lower := sma.Sub(b.mult.Mul(stddevDec))

	return Value{
		Symbol: b.symbol, Period: b.period, Timestamp: c.CloseTime,
		Data: map[string]decimal.Decimal{
			"upper":  upper,
			"middle": sma,
			"lower":  lower,
		},
	}, nil
}

func (b *BollingerBands) Name() string                 { return "BB" }
func (b *BollingerBands) Symbol() string               { return b.symbol }
func (b *BollingerBands) Period() domain_candle.Period { return b.period }
func (b *BollingerBands) WarmUp() int                  { return b.n }
func (b *BollingerBands) IsReady() bool                { return len(b.window) >= b.n }
