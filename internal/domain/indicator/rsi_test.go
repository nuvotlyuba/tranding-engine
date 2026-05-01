package indicator

import (
	"testing"

	"github.com/nuvotlyuba/trading-engine/internal/domain/candle"
	"github.com/shopspring/decimal"
)

/*
1. все цены растут → RSI стремится к
2. все цены падают → RSI стремится к 0
3. avgLoss = 0 → RSI = 100 (нет потерь)
*/

func feedRSI(rsi *RSI, prices []int64) decimal.Decimal {
	var last decimal.Decimal
	for _, p := range prices {
		val, _ := rsi.Update(makeCandle(p))
		if v, ok := val.Data["rsi"]; ok {
			last = v
		}
	}

	return last
}

// 1. все цены растут → RSI стремится к
func TestRSI_AllGains_RSIAbove50(t *testing.T) {
	rsi := NewRSI("BTCUSDT", candle.Period1m, 14)

	prices := make([]int64, 20)
	for i := range prices {
		prices[i] = int64(100 + i*2)
	}
	last := feedRSI(rsi, prices)

	if !last.GreaterThan(decimal.NewFromInt(50)) {
		t.Errorf("RSI with all gains = %s, want > 50", last)
	}
}

func TestRSI_AllLosses_RSIBelow50(t *testing.T) {
	rsi := NewRSI("BTCUSDT", candle.Period1m, 14)

	// только падающие цены → RSI должен быть ниже 50
	prices := make([]int64, 20)
	for i := range prices {
		prices[i] = int64(200 - i*2)
	}
	last := feedRSI(rsi, prices)

	if !last.LessThan(decimal.NewFromInt(50)) {
		t.Errorf("RSI with all losses = %s, want < 50", last)
	}
}

func TestRSI_NoLosses_RSI100(t *testing.T) {
	rsi := NewRSI("BTCUSDT", candle.Period1m, 14)

	// avgLoss = 0 → RSI = 100
	prices := make([]int64, 20)
	for i := range prices {
		prices[i] = int64(100 + i)
	}
	last := feedRSI(rsi, prices)

	if !last.Equal(decimal.NewFromInt(100)) {
		t.Errorf("RSI with no losses = %s, want 100", last)
	}
}

// 3. avgLoss = 0 → RSI = 100 (нет потерь)
func TestRSI_Range_0_to_100(t *testing.T) {
	rsi := NewRSI("BTCUSDT", candle.Period1m, 14)

	// RSI всегда в диапазоне [0, 100]
	prices := []int64{100, 90, 110, 80, 120, 70, 130, 60, 140, 50, 150, 40, 160, 30, 170, 20}
	for _, p := range prices {
		val, _ := rsi.Update(makeCandle(p))
		v := val.Data["rsi"]
		if v.LessThan(decimal.Zero) || v.GreaterThan(decimal.NewFromInt(100)) {
			t.Errorf("RSI = %s out of range [0, 100]", v)
		}
	}
}
