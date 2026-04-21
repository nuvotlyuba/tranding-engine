package candle

import (
	"testing"
	"time"

	"github.com/shopspring/decimal"
)

func TestUpdate_FirstTrade(t *testing.T) {
	c := NewCandle("BTCUSDT", Period1m, time.Now().Truncate(time.Minute))

	price := decimal.NewFromInt(100)
	qty := decimal.NewFromInt(5)

	c.Update(price, qty)

	if !c.Open.Equal(price) {
		t.Errorf("Open = %s, want = %s", c.Open, price)
	}

	if !c.High.Equal(price) {
		t.Errorf("High = %s, want = %s", c.High, price)
	}

	if !c.Low.Equal(price) {
		t.Errorf("Low = %s, want = %s", c.Low, price)
	}

	if !c.Close.Equal(price) {
		t.Errorf("Low = %s, want = %s", c.Close, price)
	}

	if !c.Volume.Equal(qty) {
		t.Errorf("Volumn = %s, want = %s", c.Volume, qty)
	}

	if c.Trades != 1 {
		t.Errorf("Trades = %d, want = 1", c.Trades)
	}
}

func TestUpdate_HighUpdated(t *testing.T) {
	c := NewCandle("BTCUSDT", Period1m, time.Now().Truncate(time.Minute))

	c.Update(decimal.NewFromInt(100), decimal.NewFromInt(1))
	c.Update(decimal.NewFromInt(110), decimal.NewFromInt(2))

	if !c.High.Equal(decimal.NewFromInt(110)) {
		t.Errorf("High = %s, want 110", c.High)
	}
	if !c.Open.Equal(decimal.NewFromInt(100)) {
		t.Errorf("Open = %s, want 100 (не должен меняться)", c.Open)
	}
	if !c.Close.Equal(decimal.NewFromInt(110)) {
		t.Errorf("Close = %s, want 110", c.Close)
	}
}

func TestUpdate_LowUpdated(t *testing.T) {
	c := NewCandle("BTCUSDT", Period1m, time.Now().Truncate(time.Minute))

	c.Update(decimal.NewFromInt(100), decimal.NewFromInt(1))
	c.Update(decimal.NewFromInt(99), decimal.NewFromInt(2))

	if !c.Low.Equal(decimal.NewFromInt(99)) {
		t.Errorf("Low = %s, want 99", c.Low)
	}
	if !c.Open.Equal(decimal.NewFromInt(100)) {
		t.Errorf("Open = %s, want 100 (не должен меняться)", c.Open)
	}
	if !c.Close.Equal(decimal.NewFromInt(99)) {
		t.Errorf("Close = %s, want 99", c.Close)
	}
}

func TestUpdate_MiddlePrice_OnlyCloseAndVolumeChange(t *testing.T) {
	c := NewCandle("BTCUSDT", Period1m, time.Now().Truncate(time.Minute))

	c.Update(decimal.NewFromInt(100), decimal.NewFromInt(1)) // Open=High=Low=100
	c.Update(decimal.NewFromInt(90), decimal.NewFromInt(1))  // Low=90
	c.Update(decimal.NewFromInt(110), decimal.NewFromInt(1)) // High=110
	c.Update(decimal.NewFromInt(105), decimal.NewFromInt(1)) // между High и Low

	// High и Low не изменились
	if !c.High.Equal(decimal.NewFromInt(110)) {
		t.Errorf("High = %s, want 110", c.High)
	}
	if !c.Low.Equal(decimal.NewFromInt(90)) {
		t.Errorf("Low = %s, want 90", c.Low)
	}
	// Close обновился
	if !c.Close.Equal(decimal.NewFromInt(105)) {
		t.Errorf("Close = %s, want 105", c.Close)
	}
}

func TestUpdate_VolumeUpdated(t *testing.T) {
	c := NewCandle("BTCUSDT", Period1m, time.Now().Truncate(time.Minute))

	c.Update(decimal.NewFromInt(100), decimal.NewFromInt(1))
	c.Update(decimal.NewFromInt(90), decimal.NewFromInt(2))
	c.Update(decimal.NewFromInt(95), decimal.NewFromInt(3))

	if !c.Low.Equal(decimal.NewFromInt(90)) {
		t.Errorf("Low = %s, want 90", c.Low)
	}

	if !c.High.Equal(decimal.NewFromInt(100)) {
		t.Errorf("High = %s, want 100", c.High)
	}
	if !c.Open.Equal(decimal.NewFromInt(100)) {
		t.Errorf("Open = %s, want 100", c.Open)
	}
	if !c.Close.Equal(decimal.NewFromInt(95)) {
		t.Errorf("Close = %s, want 95", c.Close)
	}

	if !c.Volume.Equal(decimal.NewFromInt(6)) {
		t.Errorf("Volume = %s, want 6", c.Volume)
	}

	if c.Trades != 3 {
		t.Errorf("Trades = %d, want 3", c.Trades)
	}
}
