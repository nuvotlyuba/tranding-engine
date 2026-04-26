package indicator

import (
	domain_candle "github.com/nuvotlyuba/trading-engine/internal/domain/candle"
	"github.com/shopspring/decimal"
)

type MACD struct {
	symbol string
	period domain_candle.Period
	fast   *EMA // период 12
	slow   *EMA // период 26
	signal *EMA // период 9, считается от MACD line
	count  int
}

func NewMACD(symbol string, period domain_candle.Period) *MACD {
	return &MACD{
		symbol: symbol,
		period: period,
		fast:   NewEMA(symbol, period, 12),
		slow:   NewEMA(symbol, period, 26),
		signal: NewEMA(symbol, period, 9),
	}
}

func (m *MACD) Update(c domain_candle.Candle) (Value, error) {
	m.count++
	m.fast.Update(c)
	m.slow.Update(c)

	macdLine := m.fast.Current().Sub(m.slow.Current())

	// создае фиктивную свечу для signal EMA - она считается от MACD line
	signalCandle := domain_candle.Candle{Close: macdLine, CloseTime: c.CloseTime}
	m.signal.Update(signalCandle)

	signalLine := m.signal.Current()
	histogram := macdLine.Sub(signalLine)

	return Value{
		Symbol:    m.symbol,
		Period:    m.period,
		Timestamp: c.CloseTime,
		Data: map[string]decimal.Decimal{
			"macd":      macdLine,
			"signal":    signalLine,
			"histogram": histogram,
		},
	}, nil
}
