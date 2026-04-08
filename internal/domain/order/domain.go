package order

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type Side string
type OrderType string
type OrderStatus string

const (
	SideBuy  Side = "buy"
	SideSell Side = "sell"
)

const (
	OrderTypeLimit  OrderType = "limit"
	OrderTypeMarket OrderType = "market"
)

const (
	StatusOpen            OrderStatus = "open"
	StatusFilled          OrderStatus = "filled"
	StatusPartiallyFilled OrderStatus = "partially_filled"
	StatusCancelled       OrderStatus = "cancelled"
)

type Order struct {
	ID        uuid.UUID
	Symbol    string
	Side      Side
	Type      OrderType
	Price     decimal.Decimal // для market ордера = 0
	Quantity  decimal.Decimal // исходный объем
	Filled    decimal.Decimal // сколько уже исполнено
	Status    OrderStatus
	CreatedAt time.Time
}

func NewOrder(symbol string, side Side, price, quantity decimal.Decimal, typeOrder OrderType) *Order {
	return &Order{
		ID:        uuid.New(),
		Symbol:    symbol,
		Side:      side,
		Price:     price,
		Quantity:  quantity,
		Filled:    decimal.Zero,
		Type:      typeOrder,
		Status:    StatusOpen,
		CreatedAt: time.Now().UTC(),
	}
}

func (o *Order) Remaining() decimal.Decimal {
	return o.Quantity.Sub(o.Filled)
}

func (o *Order) IsFilled() bool {
	return o.Remaining().IsZero()
}

func (o *Order) Cancel() {
	o.Status = StatusCancelled
}

func (o *Order) Fill(qty decimal.Decimal) error {
	if qty.IsNegative() || qty.IsZero() {
		return fmt.Errorf("fill qty must be positive, got %s", qty)
	}
	if qty.GreaterThan(o.Remaining()) {
		return fmt.Errorf("fill qty %s exceeds remaining %s", qty, o.Remaining())
	}
	o.Filled = o.Filled.Add(qty)
	if o.IsFilled() {
		o.Status = StatusFilled
	} else {
		o.Status = StatusPartiallyFilled
	}
	return nil
}
