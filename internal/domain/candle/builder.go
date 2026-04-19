package candle

import (
	"sync"
	"time"

	"github.com/shopspring/decimal"
)

type Builder struct {
	symbol  string
	period  Period
	current *Candle      // текущая открытая свеча, nil если еще не было сделок
	mu      sync.Mutex   // защита current от конкурентного доступа
	onClose func(Candle) // колбэк - вызывается когда свеча закрывается
}

func NewBuilder(symbol string, period Period, onClose func(Candle)) *Builder {
	return &Builder{
		symbol:  symbol,
		period:  period,
		onClose: onClose, // функция которая публикует событие candle.Closed в EventBus
	}
}

func (b *Builder) ProccessTrade(price, qty decimal.Decimal, at time.Time) {
	b.mu.Lock()
	defer b.mu.Unlock()

	start := periodStart(at, b.period) // вычисляем начало периода для этой сделки

	// первая сделка вообще - просто открываем свечу
	if b.current == nil {
		b.current = b.newCandle(start)
		b.current.Update(price, qty)
		return
	}

	// сделка отновится к текущему периоду - просто обновляем
	if b.current.OpenTime.Equal(start) {
		b.current.Update(price, qty)
		return
	}

	// период сменился - закрываем текущую свечу и открываем новую
	closed := *b.current // копируем чтобы не пердавать указатель
	b.onClose(closed)    // уведомляем подписчиков

	b.current = b.newCandle(start)
	b.current.Update(price, qty)
}

func (b *Builder) newCandle(openTime time.Time) *Candle {
	return &Candle{
		Symbol:    b.symbol,
		Period:    b.period,
		OpenTime:  openTime,
		CloseTime: openTime.Add(periodDuration(b.period)),
	}
}

// Current возвращает копию текущей незакрытой свечи
// Нужен для отображения живой свечи на графике
func (b *Builder) Current() (Candle, bool) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.current == nil {
		return Candle{}, false
	}

	return *b.current, true
}
