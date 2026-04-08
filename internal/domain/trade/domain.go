package trade

import (
	"time"

	domain_order "github.com/nuvotlyuba/trading-engine/internal/domain/order"
	"github.com/shopspring/decimal"
)

type Trade struct {
	ID          string
	Symbol      string
	BuyOrderID  string
	SellOrderID string
	Price       decimal.Decimal
	Quantity    decimal.Decimal
	ExecutedAt  time.Time
}

type MatchResult struct {
	Trades        []Trade
	FilledOrder   *domain_order.Order // входящий ордер после обработки
	UpdatedLevels []decimal.Decimal   // какие ценовые уровни изменились
}
