package orderbook

import (
	"fmt"
	"testing"

	domain_order "github.com/nuvotlyuba/trading-engine/internal/domain/order"
	"github.com/shopspring/decimal"
)

func TestAddOrder_NewBidLevel(t *testing.T) {
	ob := NewOrderBook("BTCUSDT")
	order := domain_order.NewOrder(
		"BTCUSDT",
		domain_order.SideBuy,
		decimal.NewFromInt(100),
		decimal.NewFromInt(10),
		domain_order.OrderTypeLimit,
	)

	ob.AddOrder(order)

	level, ok := ob.Bids[order.Price.String()]
	if !ok {
		t.Fatal("expexted price level to exist in Bids")
	}
	if len(ob.Bids) != 1 {
		t.Errorf("BidKeys len = %d, want 1", ob.BidTree.Len())
	}
	key, _ := ob.BidTree.Min()
	if !key.Equal(order.Price) {
		t.Errorf("BidKeys[0] = %s, want = %s", level.Total, order.Quantity)
	}
	if !level.Total.Equal(order.Quantity) {
		t.Errorf("Total = %s, want %s", level.Total, order.Quantity)
	}
	if len(level.Queue) != 1 {
		t.Errorf("Queue len = %d, want 1", len(level.Queue))
	}
}

func TestAddOrder_NewAskLevel(t *testing.T) {
	ob := NewOrderBook("BTCUSDT")
	order := domain_order.NewOrder(
		"BTCUSDT",
		domain_order.SideSell,
		decimal.NewFromInt(100),
		decimal.NewFromInt(10),
		domain_order.OrderTypeLimit,
	)

	ob.AddOrder(order)

	level, ok := ob.Asks[order.Price.String()]
	if !ok {
		t.Fatal("expexted price level to exist in Asks")
	}
	if len(ob.Asks) != 1 {
		t.Errorf("AskKeys len = %d, want 1", ob.AskTree.Len())
	}
	key, _ := ob.AskTree.Max()
	if !key.Equal(order.Price) {
		t.Errorf("AskKeys[0] = %s, want = %s", level.Total, order.Quantity)
	}
	if !level.Total.Equal(order.Quantity) {
		t.Errorf("Total = %s, want %s", level.Total, order.Quantity)
	}
	if len(level.Queue) != 1 {
		t.Errorf("Queue len = %d, want 1", len(level.Queue))
	}
}

func TestAddOrder_TwoOrdersInOneLevel(t *testing.T) {
	ob := NewOrderBook("BTCUSDT")
	order1 := domain_order.NewOrder(
		"BTCUSDT",
		domain_order.SideBuy,
		decimal.NewFromInt(100),
		decimal.NewFromInt(10),
		domain_order.OrderTypeLimit,
	)

	order2 := domain_order.NewOrder(
		"BTCUSDT",
		domain_order.SideBuy,
		decimal.NewFromInt(100),
		decimal.NewFromInt(10),
		domain_order.OrderTypeLimit,
	)

	ob.AddOrder(order1)
	t.Logf("after order1: len(Queue)=%d, Total=%s",
		len(ob.Bids[order1.Price.String()].Queue),
		ob.Bids[order1.Price.String()].Total)

	ob.AddOrder(order2)
	t.Logf("after order2: len(Queue)=%d, Total=%s",
		len(ob.Bids[order1.Price.String()].Queue),
		ob.Bids[order1.Price.String()].Total)
	fmt.Println(order1.Price, order2.Price)
	level, ok := ob.Bids[order1.Price.String()]
	if !ok {
		t.Fatal("expected price level to exist in Bids")
	}

	if len(level.Queue) != 2 {
		t.Errorf("Queue len = %d, want 2", len(level.Queue))
	}

	if !level.Total.Equal(decimal.NewFromInt(20)) {
		t.Errorf("Total = %s, want 20", level.Total)
	}
}

func TestAddOrder_TwoOrdersInDifferentLevel(t *testing.T) {
	ob := NewOrderBook("BTCUSDT")
	order1 := domain_order.NewOrder(
		"BTCUSDT",
		domain_order.SideBuy,
		decimal.NewFromInt(100),
		decimal.NewFromInt(10),
		domain_order.OrderTypeLimit,
	)

	order2 := domain_order.NewOrder(
		"BTCUSDT",
		domain_order.SideBuy,
		decimal.NewFromInt(100),
		decimal.NewFromInt(10),
		domain_order.OrderTypeLimit,
	)

	ob.AddOrder(order1)
	t.Logf("after order1: len(Queue)=%d, Total=%s",
		len(ob.Bids[order1.Price.String()].Queue),
		ob.Bids[order1.Price.String()].Total)

	ob.AddOrder(order2)
	t.Logf("after order2: len(Queue)=%d, Total=%s",
		len(ob.Bids[order1.Price.String()].Queue),
		ob.Bids[order1.Price.String()].Total)
	fmt.Println(order1.Price, order2.Price)
	level, ok := ob.Bids[order1.Price.String()]
	if !ok {
		t.Fatal("expected price level to exist in Bids")
	}

	if len(level.Queue) != 2 {
		t.Errorf("Queue len = %d, want 2", len(level.Queue))
	}

	if !level.Total.Equal(decimal.NewFromInt(20)) {
		t.Errorf("Total = %s, want 20", level.Total)
	}
}
