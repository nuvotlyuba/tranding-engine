package orderbook

import (
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
		t.Errorf("BidTree.Min() = %s, want %s", key, order.Price)
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
	key, _ := ob.AskTree.Min()
	if !key.Equal(order.Price) {
		t.Errorf("AskKeys.Min() = %s, want = %s", key, order.Price)
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

func TestAddOrder_BidOrdersInDifferentLevel(t *testing.T) {
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
		decimal.NewFromInt(101),
		decimal.NewFromInt(10),
		domain_order.OrderTypeLimit,
	)

	order3 := domain_order.NewOrder(
		"BTCUSDT",
		domain_order.SideBuy,
		decimal.NewFromInt(102),
		decimal.NewFromInt(10),
		domain_order.OrderTypeLimit,
	)

	ob.AddOrder(order1)
	ob.AddOrder(order2)
	ob.AddOrder(order3)

	min, _ := ob.BidTree.Min()
	if !min.Equal(order3.Price) {
		t.Errorf("Min = %s, want %v", min, order3.Price)
	}

	max, _ := ob.BidTree.Max()
	if !max.Equal(order1.Price) {
		t.Errorf("Max= %s, want %v", max, order1.Price)
	}
}

func TestAddOrder_AskOrdersInDifferentLevel(t *testing.T) {
	ob := NewOrderBook("BTCUSDT")
	order1 := domain_order.NewOrder(
		"BTCUSDT",
		domain_order.SideSell,
		decimal.NewFromInt(100),
		decimal.NewFromInt(10),
		domain_order.OrderTypeLimit,
	)

	order2 := domain_order.NewOrder(
		"BTCUSDT",
		domain_order.SideSell,
		decimal.NewFromInt(101),
		decimal.NewFromInt(10),
		domain_order.OrderTypeLimit,
	)

	order3 := domain_order.NewOrder(
		"BTCUSDT",
		domain_order.SideSell,
		decimal.NewFromInt(102),
		decimal.NewFromInt(10),
		domain_order.OrderTypeLimit,
	)

	ob.AddOrder(order1)
	ob.AddOrder(order2)
	ob.AddOrder(order3)

	min, _ := ob.AskTree.Min()
	if !min.Equal(order1.Price) {
		t.Errorf("Min = %s, want %v", min, order1.Price)
	}

	max, _ := ob.AskTree.Max()
	if !max.Equal(order3.Price) {
		t.Errorf("Max= %s, want %v", max, order3.Price)
	}
}

// BestBid / BestAsk
// 1. пустой стакан → возвращает (Zero, false)
// 2. один уровень → возвращает его цену и true
// 3. несколько уровней → возвращает именно лучшую цену (максимум для bid, минимум для ask)
func TestBestBid_EmptyOrderBook(t *testing.T) {
	ob := NewOrderBook("BTCUSDT")

	bestBid, flag := ob.BestBid()
	if !bestBid.IsZero() {
		t.Errorf("BestBid = %s, want = %s", bestBid, decimal.Zero)
	}
	if flag {
		t.Errorf("BestBid flag = %v, want false", flag)
	}
}
func TestBestAsk_EmptyOrderBook(t *testing.T) {
	ob := NewOrderBook("BTCUSDT")

	bestAsk, flag := ob.BestAsk()
	if !bestAsk.IsZero() {
		t.Errorf("BestAsk = %s, want = %s", bestAsk, decimal.Zero)
	}
	if flag {
		t.Errorf("BestAsk flag = %v, want false", flag)
	}
}

func TestBestBid_OneLevelOrderBook(t *testing.T) {
	ob := NewOrderBook("BTCUSDT")

	order := domain_order.NewOrder(
		"BTCUSDT",
		domain_order.SideBuy,
		decimal.NewFromInt(100),
		decimal.NewFromInt(10),
		domain_order.OrderTypeLimit,
	)

	ob.AddOrder(order)

	bestBid, flag := ob.BestBid()
	if !bestBid.Equal(order.Price) {
		t.Errorf("BestBid = %s, want %s", bestBid, order.Price)
	}
	if !flag {
		t.Errorf("Is existed BestBid = %v, want = true", flag)
	}
}

func TestBestAsk_OneLevelOrderBook(t *testing.T) {
	ob := NewOrderBook("BTCUSDT")

	order := domain_order.NewOrder(
		"BTCUSDT",
		domain_order.SideSell,
		decimal.NewFromInt(100),
		decimal.NewFromInt(10),
		domain_order.OrderTypeLimit,
	)

	ob.AddOrder(order)

	bestAsk, flag := ob.BestAsk()
	if !bestAsk.Equal(order.Price) {
		t.Errorf("BestAsk = %s, want %s", bestAsk, order.Price)
	}
	if !flag {
		t.Errorf("BestAsk flag = %v, want true", flag)
	}
}

func TestBestAsk_TwoLevelOrderBook(t *testing.T) {
	ob := NewOrderBook("BTCUSDT")

	order1 := domain_order.NewOrder(
		"BTCUSDT",
		domain_order.SideSell,
		decimal.NewFromInt(100),
		decimal.NewFromInt(10),
		domain_order.OrderTypeLimit,
	)
	order2 := domain_order.NewOrder(
		"BTCUSDT",
		domain_order.SideSell,
		decimal.NewFromInt(101),
		decimal.NewFromInt(10),
		domain_order.OrderTypeLimit,
	)

	ob.AddOrder(order1)
	ob.AddOrder(order2)

	bestAsk, flag := ob.BestAsk()
	if !bestAsk.Equal(order1.Price) {
		t.Errorf("BestAsk = %s, want = %s", bestAsk, order1.Price)
	}
	if !flag {
		t.Errorf("Is existed BestAsk = %v, want = true", flag)
	}
}

func TestBestBid_TwoLevelOrderBook(t *testing.T) {
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
		decimal.NewFromInt(101),
		decimal.NewFromInt(10),
		domain_order.OrderTypeLimit,
	)

	ob.AddOrder(order1)
	ob.AddOrder(order2)

	bestBid, flag := ob.BestBid()
	if !bestBid.Equal(order2.Price) {
		t.Errorf("BestBid = %s, want = %s", bestBid, order2.Price)
	}
	if !flag {
		t.Errorf("Is existed BestBid = %v, want = true", flag)
	}
}

// RemoveOrder
// 1. удалить единственный ордер на уровне → уровень исчез из Bids/Asks и из BidKeys/AskKeys
// 2. удалить один из двух ордеров на уровне → уровень остался, Total уменьшился, Queue длиной 1
// 3. удалить ордер которого нет → возвращает ошибку

func TestRemoveOrder_OnlyOneBidOrder(t *testing.T) {
	order := domain_order.NewOrder(
		"BTCUSDT",
		domain_order.SideBuy,
		decimal.NewFromInt(100),
		decimal.NewFromInt(10),
		domain_order.OrderTypeLimit,
	)

	ob := NewOrderBook("BTCUSDT")
	ob.AddOrder(order)

	err := ob.RemoveOrder(order)
	if err != nil {
		t.Errorf("RemoveOrder() = %v, want nil", err)
	}
	if len(ob.Bids) != 0 {
		t.Errorf("len(ob.Bids) = %v, want 0", len(ob.Bids))
	}
	_, flag := ob.BidTree.Min()
	if flag {
		t.Errorf("Is existed BestBid = %v, want false", flag)
	}
}

func TestRemoveOrder_OnlyOneAskOrder(t *testing.T) {
	order := domain_order.NewOrder(
		"BTCUSDT",
		domain_order.SideSell,
		decimal.NewFromInt(100),
		decimal.NewFromInt(10),
		domain_order.OrderTypeLimit,
	)

	ob := NewOrderBook("BTCUSDT")
	ob.AddOrder(order)

	err := ob.RemoveOrder(order)
	if err != nil {
		t.Errorf("RemoveOrder() = %v, want nil", err)
	}
	if len(ob.Asks) != 0 {
		t.Errorf("len(ob.Asks) = %v, want 0", len(ob.Asks))
	}
	_, flag := ob.AskTree.Max()
	if flag {
		t.Errorf("Is existed BestBid = %v, want false", flag)
	}
}

func TestRemoveOrder_TwoBidOrder(t *testing.T) {
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
		decimal.NewFromInt(5),
		domain_order.OrderTypeLimit,
	)

	ob := NewOrderBook("BTCUSDT")
	ob.AddOrder(order1)
	ob.AddOrder(order2)

	err := ob.RemoveOrder(order1)
	if err != nil {
		t.Errorf("RemoveOrder() = %v, want nil", err)
	}
	level, ok := ob.Bids[order1.Price.String()]
	if !ok {
		t.Fatal("price level should still exist")
	}
	if len(level.Queue) != 1 {
		t.Errorf("Queue len = %d, want 1", len(level.Queue))
	}
	if !level.Total.Equal(decimal.NewFromInt(5)) {
		t.Errorf("Total = %s, want 5", level.Total)
	}
}

func TestRemoveOrder_TwoAskOrder(t *testing.T) {
	order1 := domain_order.NewOrder(
		"BTCUSDT",
		domain_order.SideSell,
		decimal.NewFromInt(100),
		decimal.NewFromInt(10),
		domain_order.OrderTypeLimit,
	)

	order2 := domain_order.NewOrder(
		"BTCUSDT",
		domain_order.SideSell,
		decimal.NewFromInt(101),
		decimal.NewFromInt(10),
		domain_order.OrderTypeLimit,
	)

	ob := NewOrderBook("BTCUSDT")
	ob.AddOrder(order1)
	ob.AddOrder(order2)

	err := ob.RemoveOrder(order1)
	if err != nil {
		t.Errorf("RemoveOrder() = %v, want nil", err)
	}
	if len(ob.Asks) != 1 {
		t.Errorf("len(ob.Asks) = %v, want 1", len(ob.Asks))
	}
	_, flag := ob.AskTree.Max()
	if !flag {
		t.Errorf("Is existed BestAsk = %v, want true", flag)
	}
}

func TestRemoveOrder_NotExistedOrder(t *testing.T) {
	order1 := domain_order.NewOrder(
		"BTCUSDT",
		domain_order.SideSell,
		decimal.NewFromInt(100),
		decimal.NewFromInt(10),
		domain_order.OrderTypeLimit,
	)

	order2 := domain_order.NewOrder(
		"BTCUSDT",
		domain_order.SideSell,
		decimal.NewFromInt(101),
		decimal.NewFromInt(10),
		domain_order.OrderTypeLimit,
	)

	ob := NewOrderBook("BTCUSDT")
	ob.AddOrder(order1)

	err := ob.RemoveOrder(order2)
	if err == nil {
		t.Errorf("RemoveOrder() = nil, want err")
	}
}

// CancelOrder
// 1. после вызова → статус ордера StatusCancelled
// 2. после вызова → ордера нет в стакане

func TestCancelOrder_StatusCancelled(t *testing.T) {
	order := domain_order.NewOrder(
		"BTCUSDT",
		domain_order.SideSell,
		decimal.NewFromInt(100),
		decimal.NewFromInt(10),
		domain_order.OrderTypeLimit,
	)

	ob := NewOrderBook("BTCUSDT")
	ob.AddOrder(order)

	err := ob.CancelOrder(order)
	if err != nil {
		t.Errorf("CancelOrder() = %v, want nil", err)
	}

	if order.Status != domain_order.StatusCancelled {
		t.Errorf("Order status = %s, want = %s", order.Status, domain_order.StatusCancelled)
	}
}

func TestCancelOrder_OrderNotExistedInOrderBook(t *testing.T) {
	order1 := domain_order.NewOrder(
		"BTCUSDT",
		domain_order.SideSell,
		decimal.NewFromInt(100),
		decimal.NewFromInt(10),
		domain_order.OrderTypeLimit,
	)

	order2 := domain_order.NewOrder(
		"BTCUSDT",
		domain_order.SideSell,
		decimal.NewFromInt(100),
		decimal.NewFromInt(10),
		domain_order.OrderTypeLimit,
	)

	ob := NewOrderBook("BTCUSDT")
	ob.AddOrder(order1)

	err := ob.CancelOrder(order2)
	if err == nil {
		t.Errorf("CancelOrder() = nil, want err")
	}
}

func TestMatching_BuyLimit_NoAsks(t *testing.T) {
	ob := NewOrderBook("BTCUSDT")

	incoming := domain_order.NewOrder(
		"BTCUSDT",
		domain_order.SideBuy,
		decimal.NewFromInt(100),
		decimal.NewFromInt(10),
		domain_order.OrderTypeLimit,
	)

	result, err := ob.Matching(incoming)

	if err != nil {
		t.Fatalf("Matching() error = %v, want nil", err)
	}
	// трейдов нет
	if len(result.Trades) != 0 {
		t.Errorf("trades len = %d, want 0", len(result.Trades))
	}
	// ордер добавлен в стакан
	level, ok := ob.Bids[incoming.Price.String()]
	if !ok {
		t.Fatal("expected order to be added to Bids")
	}
	if len(level.Queue) != 1 {
		t.Errorf("Queue len = %d, want 1", len(level.Queue))
	}
	// статус остался open
	if incoming.Status != domain_order.StatusOpen {
		t.Errorf("Status = %s, want open", incoming.Status)
	}
}

func TestMatching_SellLimit_NoBids(t *testing.T) {
	ob := NewOrderBook("BTCUSDT")

	incoming := domain_order.NewOrder(
		"BTCUSDT",
		domain_order.SideSell,
		decimal.NewFromInt(100),
		decimal.NewFromInt(10),
		domain_order.OrderTypeLimit,
	)

	result, err := ob.Matching(incoming)

	if err != nil {
		t.Fatalf("Matching() error = %v, want nil", err)
	}
	// трейдов нет
	if len(result.Trades) != 0 {
		t.Errorf("trades len = %d, want 0", len(result.Trades))
	}
	// ордер добавлен в стакан
	level, ok := ob.Asks[incoming.Price.String()]
	if !ok {
		t.Fatal("expected order to be added to Asks")
	}
	if len(level.Queue) != 1 {
		t.Errorf("Queue len = %d, want 1", len(level.Queue))
	}
	// статус остался open
	if incoming.Status != domain_order.StatusOpen {
		t.Errorf("Status = %s, want open", incoming.Status)
	}
}

func TestMatching_BuyLimit_OneAsk_RemoveLevel(t *testing.T) {
	ob := NewOrderBook("BTCUSDT")

	incoming := domain_order.NewOrder(
		"BTCUSDT",
		domain_order.SideBuy,
		decimal.NewFromInt(100),
		decimal.NewFromInt(10),
		domain_order.OrderTypeLimit,
	)

	orderAsk := domain_order.NewOrder(
		"BTCUSDT",
		domain_order.SideSell,
		decimal.NewFromInt(99),
		decimal.NewFromInt(10),
		domain_order.OrderTypeLimit,
	)
	ob.addOrder(orderAsk)

	result, err := ob.Matching(incoming)

	if err != nil {
		t.Fatalf("Matching() error = %v, want nil", err)
	}

	if len(result.Trades) != 1 {
		t.Errorf("trades len = %d, want 1", len(result.Trades))
	}

	if len(result.UpdatedLevels) != 1 {
		t.Errorf("len(UpdatedLevels) = %d, want 1", len(result.UpdatedLevels))
	}
	if result.UpdatedLevels[0] != orderAsk.Price {
		t.Errorf("UpdatedLevel price = %v , want = %v", result.UpdatedLevels[0], orderAsk.Price)
	}

	if result.Trades[0].BuyOrderID != incoming.ID {
		t.Errorf("trade butOrderID = %s, want %s", result.Trades[0].BuyOrderID, incoming.ID)
	}

	if result.Trades[0].SellOrderID != orderAsk.ID {
		t.Errorf("trade sellOrderID = %s, want %s", result.Trades[0].SellOrderID, orderAsk.ID)
	}

	if len(ob.Asks) != 0 || len(ob.Bids) != 0 {
		t.Fatal("there are no order in Asks and in Bids")
	}
	if incoming.Status != domain_order.StatusFilled {
		t.Errorf("Status = %s, want filled", incoming.Status)
	}

	if orderAsk.Status != domain_order.StatusFilled {
		t.Errorf("Status = %s, want filled", incoming.Status)
	}
}

func TestMatching_SellLimit_OneBid_RemoveLevel(t *testing.T) {
	ob := NewOrderBook("BTCUSDT")

	incoming := domain_order.NewOrder(
		"BTCUSDT",
		domain_order.SideSell,
		decimal.NewFromInt(99),
		decimal.NewFromInt(10),
		domain_order.OrderTypeLimit,
	)

	orderBid := domain_order.NewOrder(
		"BTCUSDT",
		domain_order.SideBuy,
		decimal.NewFromInt(100),
		decimal.NewFromInt(10),
		domain_order.OrderTypeLimit,
	)
	ob.addOrder(orderBid)

	result, err := ob.Matching(incoming)

	if err != nil {
		t.Fatalf("Matching() error = %v, want nil", err)
	}

	if len(result.Trades) != 1 {
		t.Errorf("trades len = %d, want 1", len(result.Trades))
	}

	if len(result.UpdatedLevels) != 1 {
		t.Errorf("len(UpdatedLevels) = %d, want 1", len(result.UpdatedLevels))
	}
	if result.UpdatedLevels[0] != orderBid.Price {
		t.Errorf("UpdatedLevel price = %v , want = %v", result.UpdatedLevels[0], orderBid.Price)
	}

	if result.Trades[0].SellOrderID != incoming.ID {
		t.Errorf("trade sellOrderID = %s, want %s", result.Trades[0].SellOrderID, incoming.ID)
	}

	if result.Trades[0].BuyOrderID != orderBid.ID {
		t.Errorf("trade buyOrderID = %s, want %s", result.Trades[0].BuyOrderID, orderBid.ID)
	}

	if len(ob.Asks) != 0 || len(ob.Bids) != 0 {
		t.Fatal("there are no order in Asks and in Bids")
	}
	if incoming.Status != domain_order.StatusFilled {
		t.Errorf("Status = %s, want filled", incoming.Status)
	}

	if orderBid.Status != domain_order.StatusFilled {
		t.Errorf("Status = %s, want filled", incoming.Status)
	}
}

// 5. buy limit — частичное исполнение → 1 trade, входящий ордер частично filled и добавлен в стакан
// 6. buy limit — цена ниже лучшего ask → матчинга нет, ордер добавляется в стакан
// 7. sell limit — цена выше лучшего bid → матчинга нет, ордер добавляется в стакан

// Сложные сценарии:
// 8. buy limit съедает несколько уровней asks → несколько trades, все уровни удалены
// 9. buy limit — встречный ордер больше входящего → входящий filled полностью, встречный частично остаётся в стакане
// 10. два ордера на одном уровне — входящий съедает первый полностью и частично второй → FIFO соблюдён
// 11. market buy — исполняется по любой цене независимо от bestAsk → trade создан
// 12. market sell — исполняется по любой цене независимо от bestBid → trade создан
