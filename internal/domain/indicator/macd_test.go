package indicator

/*
1. быстрая EMA > медленной → MACD line положительный
2. histogram = macd - signal
*/

// domain/indicator/macd_test.go

import (
	"testing"

	"github.com/nuvotlyuba/trading-engine/internal/domain/candle"
	"github.com/shopspring/decimal"
)

func feedMACD(macd *MACD, prices []int64) (macdLine, signal, histogram decimal.Decimal) {
	for _, p := range prices {
		val, _ := macd.Update(makeCandle(p))
		macdLine = val.Data["macd"]
		signal = val.Data["signal"]
		histogram = val.Data["histogram"]
	}
	return
}

func TestMACD_Histogram_EqualsMACDMinusSignal(t *testing.T) {
	macd := NewMACD("BTCUSDT", candle.Period1m)

	prices := make([]int64, 40)
	for i := range prices {
		prices[i] = int64(100 + i)
	}

	macdLine, signal, histogram := feedMACD(macd, prices)

	want := macdLine.Sub(signal)
	if !histogram.Equal(want) {
		t.Errorf("histogram = %s, want macd-signal = %s", histogram, want)
	}
}

func TestMACD_RisingPrices_PositiveMACDLine(t *testing.T) {
	macd := NewMACD("BTCUSDT", candle.Period1m)

	// при росте цен быстрая EMA(12) > медленной EMA(26) → MACD > 0
	prices := make([]int64, 40)
	for i := range prices {
		prices[i] = int64(100 + i*3)
	}

	macdLine, _, _ := feedMACD(macd, prices)

	if !macdLine.IsPositive() {
		t.Errorf("MACD line with rising prices = %s, want > 0", macdLine)
	}
}

func TestMACD_FallingPrices_NegativeMACDLine(t *testing.T) {
	macd := NewMACD("BTCUSDT", candle.Period1m)

	// при падении цен быстрая EMA(12) < медленной EMA(26) → MACD < 0
	prices := make([]int64, 40)
	for i := range prices {
		prices[i] = int64(500 - i*3)
	}

	macdLine, _, _ := feedMACD(macd, prices)

	if !macdLine.IsNegative() {
		t.Errorf("MACD line with falling prices = %s, want < 0", macdLine)
	}
}

func TestMACD_IsReady(t *testing.T) {
	macd := NewMACD("BTCUSDT", candle.Period1m)

	// WarmUp = 26 + 9 = 35
	for i := range 34 {
		macd.Update(makeCandle(int64(100 + i)))
		if macd.IsReady() {
			t.Errorf("IsReady() = true after %d candles, want false", i+1)
		}
	}

	macd.Update(makeCandle(135))
	if !macd.IsReady() {
		t.Error("IsReady() = false after 35 candles, want true")
	}
}
