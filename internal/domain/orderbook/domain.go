package orderbook

import (
	"fmt"
	"sync"
	"time"

	"github.com/google/btree"
	"github.com/google/uuid"
	domain_order "github.com/nuvotlyuba/trading-engine/internal/domain/order"
	domain_trade "github.com/nuvotlyuba/trading-engine/internal/domain/trade"
	"github.com/shopspring/decimal"
)

type PriceLevel struct {
	Price decimal.Decimal       // Цена
	Total decimal.Decimal       // Суммарный объем
	Queue []*domain_order.Order // Очередь ордеров в порядке FIFO
}

func NewPriceLevel(order *domain_order.Order) *PriceLevel {
	queue := make([]*domain_order.Order, 0)
	return &PriceLevel{
		Price: order.Price,
		Total: order.Remaining(),
		Queue: append(queue, order),
	}
}

type OrderBook struct { // стакан
	Symbol  string
	Bids    map[string]*PriceLevel
	BidTree *btree.BTreeG[decimal.Decimal] //отсортирован по убыванию: [102, 101, 100]

	Asks    map[string]*PriceLevel
	AskTree *btree.BTreeG[decimal.Decimal] // отсортированы по возрастанию: [103, 104, 105]

	mu sync.RWMutex
}

func NewOrderBook(symbol string) *OrderBook {
	return &OrderBook{
		Symbol: symbol,
		Bids:   make(map[string]*PriceLevel, 0),
		BidTree: btree.NewG(32, func(a, b decimal.Decimal) bool {
			return a.GreaterThan(b)
		}),
		Asks: make(map[string]*PriceLevel, 0),
		AskTree: btree.NewG(32, func(a, b decimal.Decimal) bool {
			return a.LessThan(b)
		}),
	}
}

func (ob *OrderBook) BestBid() (decimal.Decimal, bool) {
	ob.mu.RLock()
	defer ob.mu.RUnlock()

	if ob.BidTree.Len() == 0 {
		return decimal.Zero, false
	}
	return ob.BidTree.Min()
}

func (ob *OrderBook) BestAsk() (decimal.Decimal, bool) {
	ob.mu.RLock()
	defer ob.mu.RUnlock()
	if ob.AskTree.Len() == 0 {
		return decimal.Zero, false
	}
	return ob.AskTree.Max()
}

func (ob *OrderBook) AddOrder(order *domain_order.Order) {
	ob.mu.Lock()
	defer ob.mu.Unlock()

	if order.Side == domain_order.SideBuy {
		if level, ok := ob.Bids[order.Price.String()]; !ok {
			ob.Bids[order.Price.String()] = NewPriceLevel(order)
			ob.addKey(order.Price, order.Side)
		} else {
			level.Total = level.Total.Add(order.Remaining())
			level.Queue = append(level.Queue, order)
		}
	}
	if order.Side == domain_order.SideSell {
		if level, ok := ob.Asks[order.Price.String()]; !ok {
			ob.Asks[order.Price.String()] = NewPriceLevel(order)
			ob.addKey(order.Price, order.Side)
		} else {
			level.Total = level.Total.Add(order.Remaining())
			level.Queue = append(level.Queue, order)
		}
	}
}

func (ob *OrderBook) RemoveOrder(order *domain_order.Order) error {
	ob.mu.Lock()
	defer ob.mu.Unlock()

	if order.Side == domain_order.SideBuy {
		level, ok := ob.Bids[order.Price.String()]
		if !ok {
			return fmt.Errorf("not found price level")
		}
		level.Total = level.Total.Sub(order.Remaining())
		level.Queue = removeOrderFromQueue(level.Queue, order.ID)

		if level.Total.IsNegative() || level.Total.IsZero() {
			delete(ob.Bids, order.Price.String())
			ob.removeKey(order.Price, order.Side)
		}

		return nil
	}

	if order.Side == domain_order.SideSell {
		level, ok := ob.Asks[order.Price.String()]
		if !ok {
			return fmt.Errorf("not found price level")
		}
		level.Total = level.Total.Sub(order.Remaining())
		level.Queue = removeOrderFromQueue(level.Queue, order.ID)

		if level.Total.IsNegative() || level.Total.IsZero() {
			delete(ob.Asks, order.Price.String())
			ob.removeKey(order.Price, order.Side)
		}

	}
	return nil
}

func (ob *OrderBook) addKey(price decimal.Decimal, side domain_order.Side) {
	switch side {
	case domain_order.SideBuy:
		ob.BidTree.ReplaceOrInsert(price)
	case domain_order.SideSell:
		ob.AskTree.ReplaceOrInsert(price)
	}
}

func (ob *OrderBook) CancelOrder(order *domain_order.Order) error {
	err := ob.RemoveOrder(order)
	if err != nil {
		return fmt.Errorf("order book -> func cancel order: %w", err)
	}
	order.Cancel()
	/*
			TODO
		    Сделать запись в базу
			Сгенерировать событие order.cancelled в шину
	*/
	return nil
}

func removeOrderFromQueue(orders []*domain_order.Order, orderID uuid.UUID) []*domain_order.Order {
	for i, order := range orders {
		if order.ID == orderID {
			return append(orders[:i], orders[i+1:]...)
		}
	}
	return orders
}

func (ob *OrderBook) removeKey(price decimal.Decimal, side domain_order.Side) {
	switch side {
	case domain_order.SideBuy:
		ob.BidTree.Delete(price)
	case domain_order.SideSell:
		ob.AskTree.Delete(price)
	}
}

/*

Пришёл новый ордер на покупку (bid) по цене 101:

1. Смотрим на лучший ask (минимальная цена продажи)
2. Если лучший ask <= 101 — матчим:
   а. Берём первый ордер из очереди этого уровня (FIFO)
   б. Определяем объём сделки = min(входящий.Remaining, встречный.Remaining)
   в. Создаём Trade
   г. Обновляем Filled у обоих ордеров
   д. Если встречный ордер полностью исполнен — убираем его из очереди
   е. Если уровень опустел — убираем уровень из стакана и из AskKeys
3. Повторяем пока входящий ордер не исполнен или asks не кончились
4. Если у входящего остался остаток — добавляем его в Bids
*/

func (ob *OrderBook) Matching(order *domain_order.Order) (*domain_trade.MatchResult, error) {
	ob.mu.Lock()
	defer ob.mu.Unlock()

	var trades []domain_trade.Trade
	var updatedLevels []decimal.Decimal
	var filledOrder *domain_order.Order

	switch order.Side {
	case domain_order.SideBuy:
		for !order.Remaining().IsZero() && ob.AskTree.Len() > 0 {
			bestAskPrice, ok := ob.BestAsk()
			if !ok {
				return nil, nil
			}

			if order.Type == domain_order.OrderTypeLimit && order.Price.LessThan(bestAskPrice) {
				break
			}

			level, _ := ob.Asks[bestAskPrice.String()]

			for len(level.Queue) > 0 && order.Remaining().IsPositive() {
				resting := level.Queue[0]

				tradeQty := decimal.Min(order.Remaining(), resting.Remaining())
				order.Fill(tradeQty)
				resting.Fill(tradeQty)

				trade := domain_trade.Trade{
					ID:          uuid.New(),
					Symbol:      order.Symbol,
					BuyOrderID:  order.ID,
					SellOrderID: resting.ID,
					Price:       bestAskPrice,
					Quantity:    tradeQty,
					ExecutedAt:  time.Now().UTC(),
				}

				level.Total = level.Total.Sub(tradeQty)

				if resting.IsFilled() {
					level.Queue = level.Queue[1:]
				}

				if len(level.Queue) == 0 {
					delete(ob.Asks, bestAskPrice.String())
					ob.removeKey(order.Price, domain_order.SideSell)
				}

				updatedLevels = append(updatedLevels, bestAskPrice)
				trades = append(trades, trade)
			}

		}

	case domain_order.SideSell:
		for !order.Remaining().IsZero() && ob.BidTree.Len() > 0 {
			bestBidPrice, ok := ob.BestBid()
			if !ok {
				return nil, nil
			}

			if order.Price.IsZero() && order.Price.LessThan(bestBidPrice) {
				break // Не можем исполнить, цена слишком низкая
			}

			level, _ := ob.Bids[bestBidPrice.String()]

			for len(level.Queue) > 0 && order.Remaining().IsPositive() {
				resting := level.Queue[0]

				tradeQty := decimal.Min(order.Remaining(), resting.Remaining())
				err := order.Fill(tradeQty)
				if err != nil {
					return nil, err
				}
				err = resting.Fill(tradeQty)
				if err != nil {
					return nil, err
				}

				trade := domain_trade.Trade{
					ID:          uuid.New(),
					Symbol:      order.Symbol,
					BuyOrderID:  resting.ID,
					SellOrderID: order.ID,
					Price:       bestBidPrice,
					Quantity:    tradeQty,
					ExecutedAt:  time.Now().UTC(),
				}

				level.Total = level.Total.Sub(tradeQty)

				if resting.IsFilled() {
					level.Queue = level.Queue[1:]
				}

				if len(level.Queue) == 0 {
					delete(ob.Bids, order.Price.String())
					ob.removeKey(order.Price, domain_order.SideBuy)
				}

				updatedLevels = append(updatedLevels, bestBidPrice)
				trades = append(trades, trade)
			}
		}
		if !order.Remaining().IsZero() && order.Type == domain_order.OrderTypeLimit {
			ob.AddOrder(order)
		}
	}

	return &domain_trade.MatchResult{
		Trades:        trades,
		UpdatedLevels: updatedLevels,
		FilledOrder:   filledOrder,
	}, nil
}
