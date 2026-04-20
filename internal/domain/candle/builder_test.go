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

func TestBuilder_FirstTrade_CreatesCandle(t *testing.T) {
	b, closed := newTestBuilder("BTCUSDT", Period1m)

	at := time.Date(2024, 1, 1, 10, 0, 30, 0, time.UTC)
	b.ProcessTrade(decimal.NewFromInt(100), decimal.NewFromInt(5), at)

	if closed == nil {
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

/*
ProcessTrade:
1. первая сделка — current не nil, OpenTime = periodStart сделки
2. две сделки в одном периоде — onClose не вызван, current обновлён
3. сделка в новом периоде — onClose вызван ровно один раз с правильной свечой
4. три периода подряд — onClose вызван дважды, каждая закрытая свеча корректна
5. закрытая свеча содержит правильный OpenTime и CloseTime
*/

/*
Current:
6. до первой сделки — возвращает (Candle{}, false)
7. после первой сделки — возвращает (current, true)
*/
