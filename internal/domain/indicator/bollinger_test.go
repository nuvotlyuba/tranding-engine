// domain/indicator/bollinger_test.go
package indicator

import (
	"testing"

	"github.com/nuvotlyuba/trading-engine/internal/domain/candle"
	"github.com/shopspring/decimal"
)

func newBB() *BollingerBands {
	return NewBollingerBands(
		"BTCUSDT",
		candle.Period1m,
		20,
		decimal.NewFromInt(2),
	)
}

func TestBB_NotEnoughCandles_EmptyData(t *testing.T) {
	bb := newBB()

	// меньше 20 свечей — данных нет
	for i := range 19 {
		val, err := bb.Update(makeCandle(int64(100 + i)))
		if err != nil {
			t.Fatalf("Update() error = %v", err)
		}
		if len(val.Data) != 0 {
			t.Errorf("after %d candles Data should be empty, got %v", i+1, val.Data)
		}
	}
}

func TestBB_ConstantPrice_ZeroStdDev(t *testing.T) {
	bb := newBB()

	// все цены одинаковые → stddev = 0 → upper = lower = middle
	for range 20 {
		bb.Update(makeCandle(100))
	}

	val, _ := bb.Update(makeCandle(100))

	upper := val.Data["upper"]
	middle := val.Data["middle"]
	lower := val.Data["lower"]

	if !upper.Equal(middle) {
		t.Errorf("upper = %s, want equal to middle %s", upper, middle)
	}
	if !lower.Equal(middle) {
		t.Errorf("lower = %s, want equal to middle %s", lower, middle)
	}
	if !middle.Equal(decimal.NewFromInt(100)) {
		t.Errorf("middle = %s, want 100", middle)
	}
}

func TestBB_UpperGreaterThanLower(t *testing.T) {
	bb := newBB()

	prices := []int64{
		95, 98, 102, 100, 97, 103, 101, 99, 104, 96,
		105, 98, 100, 102, 97, 99, 103, 101, 98, 100,
	}
	var last map[string]decimal.Decimal
	for _, p := range prices {
		val, _ := bb.Update(makeCandle(p))
		last = val.Data
	}

	if !last["upper"].GreaterThan(last["middle"]) {
		t.Errorf("upper %s should be > middle %s", last["upper"], last["middle"])
	}
	if !last["middle"].GreaterThan(last["lower"]) {
		t.Errorf("middle %s should be > lower %s", last["middle"], last["lower"])
	}
}

func TestBB_IsReady(t *testing.T) {
	bb := newBB()

	for i := range 19 {
		bb.Update(makeCandle(100))
		if bb.IsReady() {
			t.Errorf("IsReady() = true after %d candles, want false", i+1)
		}
	}

	bb.Update(makeCandle(100))
	if !bb.IsReady() {
		t.Error("IsReady() = false after 20 candles, want true")
	}
}
