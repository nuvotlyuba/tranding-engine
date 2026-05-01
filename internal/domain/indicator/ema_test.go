package indicator

import (
	"testing"

	"github.com/nuvotlyuba/trading-engine/internal/domain/candle"
	domain_candle "github.com/nuvotlyuba/trading-engine/internal/domain/candle"
	"github.com/shopspring/decimal"
)

/*
1. первая свеча → EMA = Close первой свечи
2. несколько свечей с одинаковой ценой → EMA = этой цене
3. растущие цены → EMA растёт но отстаёт от цены
4. IsReady() = false пока count < n, true после
*/

func makeCandle(closePrice int64) domain_candle.Candle {
	return domain_candle.Candle{
		Symbol: "BTCUSDT",
		Period: domain_candle.Period1m,
		Close:  decimal.NewFromInt(closePrice),
	}
}

// вспомогательная функция - прогоняет срез цен через индикатор
// возвращет последнее значение
func feedEMA(ema *EMA, prices []int64) decimal.Decimal {
	var last decimal.Decimal
	for _, p := range prices {
		val, _ := ema.Update(makeCandle(p))
		if v, ok := val.Data["ema"]; ok {
			last = v
		}
	}
	return last
}

// 1. первая свеча → EMA = Close первой свечи
func TestEMA_FirstCandle_EqualToClose(t *testing.T) {
	ema := NewEMA("BTCUSDT", candle.Period1m, 3)

	val, err := ema.Update(makeCandle(100))

	if err != nil {
		t.Fatalf("Update() error = %v, want nil", err)
	}

	if !val.Data["ema"].Equal(decimal.NewFromInt(100)) {
		t.Errorf("EMA after first candle = %s, want 100", val.Data["ema"])
	}
}

// 2. несколько свечей с одинаковой ценой → EMA = этой цене
func TestEMA_SamePrice_EMAUnchanged(t *testing.T) {
	ema := NewEMA("BTCUSDT", candle.Period1m, 3)

	prices := []int64{100, 100, 100, 100, 100}
	last := feedEMA(ema, prices)

	if !last.Equal(decimal.NewFromInt(100)) {
		t.Errorf("EMA with conctsnt price = %s, want 100", last)
	}
}

// 4. IsReady() = false пока count < n, true после
func TestEMA_IsReady(t *testing.T) {
	ema := NewEMA("BTCUSDT", candle.Period1m, 3)

	ema.Update(makeCandle(100))
	ema.Update(makeCandle(100))
	if ema.IsReady() {
		t.Errorf("IsReady() = true after 2 candle with period 3, want false")
	}

	ema.Update(makeCandle(100))
	if !ema.IsReady() {
		t.Errorf("IsReady() = false after 3 candles with period 3, want true")
	}
}

// 3. растущие цены → EMA растёт но отстаёт от цены
func TestEMA_RisingPrices_EMALagsPrice(t *testing.T) {
	ema := NewEMA("BTCUSDT", candle.Period1m, 3)

	prices := []int64{100, 110, 120, 130, 140}
	last := feedEMA(ema, prices)

	lastPrice := decimal.NewFromInt(140)
	if !last.LessThan(lastPrice) {
		t.Errorf("EMA = %s should be less than last price %s (EMA lags)", last, lastPrice)
	}
}
