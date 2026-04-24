package indicator

import (
	"time"

	domain_candle "github.com/nuvotlyuba/trading-engine/internal/domain/candle"
	"github.com/shopspring/decimal"
)

type Indicator interface {
	Update(candle domain_candle.Candle) (Value, error)
	Name() string
	Symbol() string
	Period() domain_candle.Period
	WarmUp() int   // возвращет сколько свечей нужно до первого валидного значения
	ISReady() bool // возвращает true когда индикатор прогрелся
}

type Value struct { // результат вычисления индикатора на одной свече
	Inducator string
	Symbol    string
	Period    domain_candle.Period
	Timestamp time.Time
	Data      map[string]decimal.Decimal
}
