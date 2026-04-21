package candle

import (
	"testing"
	"time"

	"github.com/shopspring/decimal"
)

func newTestBuilder(symbol string, period Period) (*Builder, *[]Candle) {
	closed := make([]Candle, 0)

	b := NewBuilder(symbol, period, func(c Candle) {
		closed = append(closed, c)
	})

	return b, &closed
}

// 1. первая сделка — current не nil, OpenTime = periodStart сделки
func TestBuilder_FirstTrade_CreatesCandle(t *testing.T) {
	b, closed := newTestBuilder("BTCUSDT", Period1m)

	at := time.Date(2024, 1, 1, 10, 0, 30, 0, time.UTC)
	b.ProcessTrade(decimal.NewFromInt(100), decimal.NewFromInt(5), at)

	if closed == nil && len(*closed) != 0 {
		t.Errorf("onClose calls = %d, want 0", len(*closed))
	}

	current, ok := b.Current()
	if !ok {
		t.Fatalf("Current() = false, want true")
	}

	wantOpen := time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC)
	if !current.OpenTime.Equal(wantOpen) {
		t.Errorf("OpenTime = %v, want %v", current.OpenTime, wantOpen)
	}
}

// 2. две сделки в одном периоде — onClose не вызван, current обновлён
func TestBuilder_TwoTradesInPeriod(t *testing.T) {
	b, closed := newTestBuilder("BTCUSDT", Period1m)

	at := time.Date(2024, 1, 1, 10, 0, 30, 0, time.UTC)
	b.ProcessTrade(decimal.NewFromInt(100), decimal.NewFromInt(5), at)
	b.ProcessTrade(decimal.NewFromInt(101), decimal.NewFromInt(10), at)

	if closed != nil && len(*closed) != 0 {
		t.Errorf("onClose calls = %d, want 0", len(*closed))
	}

	current, ok := b.Current()
	if !ok {
		t.Fatalf("Current() = false, want true")
	}

	if !current.High.Equal(decimal.NewFromInt(101)) {
		t.Errorf("Current High = %s, want = 101", current.High)
	}

	if !current.Low.Equal(decimal.NewFromInt(100)) {
		t.Errorf("Current High = %s, want = 100", current.Low)
	}

}

// 3. сделка в новом периоде — onClose вызван ровно один раз с правильной свечой
func TestBuilder_TradeInNewPeriod(t *testing.T) {
	b, closed := newTestBuilder("BTCUSDT", Period1m)

	at1 := time.Date(2024, 1, 1, 10, 0, 30, 0, time.UTC)
	at2 := at1.Add(2 * time.Minute) // 10:02:30 (новый период для 1m)

	b.ProcessTrade(decimal.NewFromInt(100), decimal.NewFromInt(5), at1)
	b.ProcessTrade(decimal.NewFromInt(101), decimal.NewFromInt(10), at2)

	if len(*closed) != 1 {
		t.Errorf("onClose calls = %d, want 1", len(*closed))
	}

	// Проверяем ПЕРВУЮ (закрытую) свечу
	closedCandle := (*closed)[0]
	if !closedCandle.High.Equal(decimal.NewFromInt(100)) {
		t.Errorf("Closed candle High = %s, want 100", closedCandle.High)
	}
	if !closedCandle.Low.Equal(decimal.NewFromInt(100)) {
		t.Errorf("Closed candle Low = %s, want 100", closedCandle.Low)
	}

	// Проверяем ТЕКУЩУЮ (вторую) свечу
	current, _ := b.Current()
	if !current.High.Equal(decimal.NewFromInt(101)) {
		t.Errorf("Current High = %s, want 101", current.High)
	}
	if !current.Low.Equal(decimal.NewFromInt(101)) {
		t.Errorf("Current Low = %s, want 101", current.Low)
	}
}

// 4. три периода подряд — onClose вызван дважды, каждая закрытая свеча корректна
func TestBuilder_TradesInNewPeriods(t *testing.T) {
	b, closed := newTestBuilder("BTCUSDT", Period1m)

	at1 := time.Date(2024, 1, 1, 10, 0, 30, 0, time.UTC)
	at2 := at1.Add(2 * time.Minute) // 10:02:30 (новый период для 1m)
	at3 := at1.Add(3 * time.Minute) // 10:03:30 (новый период для 1m)

	b.ProcessTrade(decimal.NewFromInt(100), decimal.NewFromInt(5), at1)
	b.ProcessTrade(decimal.NewFromInt(101), decimal.NewFromInt(10), at2)
	b.ProcessTrade(decimal.NewFromInt(102), decimal.NewFromInt(20), at3)

	if len(*closed) != 2 {
		t.Errorf("onClose calls = %d, want 2", len(*closed))
	}

	// Проверяем ПЕРВУЮ (закрытую) свечу
	closedCandle1 := (*closed)[0]
	if !closedCandle1.High.Equal(decimal.NewFromInt(100)) {
		t.Errorf("Closed candle High = %s, want 100", closedCandle1.High)
	}
	if !closedCandle1.Low.Equal(decimal.NewFromInt(100)) {
		t.Errorf("Closed candle Low = %s, want 100", closedCandle1.Low)
	}

	// Проверяем ПЕРВУЮ (закрытую) свечу
	closedCandle2 := (*closed)[1]
	if !closedCandle2.High.Equal(decimal.NewFromInt(101)) {
		t.Errorf("Closed candle High = %s, want 101", closedCandle2.High)
	}
	if !closedCandle2.Low.Equal(decimal.NewFromInt(101)) {
		t.Errorf("Closed candle Low = %s, want 101", closedCandle2.Low)
	}

	// Проверяем ТЕКУЩУЮ (третью) свечу
	current, _ := b.Current()
	if !current.High.Equal(decimal.NewFromInt(102)) {
		t.Errorf("Current High = %s, want 102", current.High)
	}
	if !current.Low.Equal(decimal.NewFromInt(102)) {
		t.Errorf("Current Low = %s, want 102", current.Low)
	}
}

// 5. закрытая свеча содержит правильный OpenTime и CloseTime
func TestBuilder_ClosedCandle(t *testing.T) {
	b, closed := newTestBuilder("BTCUSDT", Period1m)

	at1 := time.Date(2024, 1, 1, 10, 0, 30, 0, time.UTC)
	at2 := at1.Add(2 * time.Minute) // 10:02:30 (новый период для 1m)

	b.ProcessTrade(decimal.NewFromInt(100), decimal.NewFromInt(5), at1)
	b.ProcessTrade(decimal.NewFromInt(101), decimal.NewFromInt(10), at2)

	if closed != nil && len(*closed) != 1 {
		t.Errorf("onClose calls = %d, want 1", len(*closed))
	}

	closedCandle := (*closed)[0]
	if !closedCandle.OpenTime.Equal(at1.Truncate(time.Minute)) {
		t.Errorf("Closed candle OpenTime = %v, want %v", closedCandle.OpenTime, at1.Truncate(time.Minute))
	}
	if !closedCandle.CloseTime.Equal(at1.Truncate(time.Minute).Add(periodDuration(Period1m))) {
		t.Errorf("Closed candle CloseTime = %v, want %v", closedCandle.CloseTime, at1)
	}
}

// 6. до первой сделки — возвращает (Candle{}, false)
func TestCurrent_AfterFirstTrade(t *testing.T) {
	b, _ := newTestBuilder("BTCUSDT", Period1m)

	curr, ok := b.Current()

	if ok {
		t.Error("expected ok=false before first trade")
	}

	if curr != (Candle{}) {
		t.Errorf("expected zero Candle, got %+v", curr)
	}
}

// 7. после первой сделки — возвращает (current, true)

func TestCurrent_BeforeFirstTrade(t *testing.T) {
	b, _ := newTestBuilder("BTCUSDT", Period1m)

	at := time.Date(2024, 1, 1, 10, 0, 30, 0, time.UTC)
	b.ProcessTrade(decimal.NewFromInt(100), decimal.NewFromInt(5), at)

	curr, ok := b.Current()

	if !ok {
		t.Error("expected ok=true before first trade")
	}

	if curr == (Candle{}) {
		t.Errorf("expected not zero Candle, got %+v", curr)
	}
}
