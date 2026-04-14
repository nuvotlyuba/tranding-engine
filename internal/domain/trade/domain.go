package trade

import (
	"time"

	"github.com/google/uuid"
	domain_order "github.com/nuvotlyuba/trading-engine/internal/domain/order"
	"github.com/shopspring/decimal"
)

type Trade struct {
	ID          uuid.UUID
	Symbol      string
	BuyOrderID  uuid.UUID
	SellOrderID uuid.UUID
	Price       decimal.Decimal
	Quantity    decimal.Decimal
	ExecutedAt  time.Time
}

type MatchResult struct {
	Trades        []Trade
	FilledOrder   *domain_order.Order // входящий ордер после обработки
	UpdatedLevels []decimal.Decimal   // какие ценовые уровни изменились
}
