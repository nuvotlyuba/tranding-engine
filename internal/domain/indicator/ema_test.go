package indicator

import (
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
