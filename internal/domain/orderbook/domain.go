package orderbook

import (
	"fmt"
	"sort"
	"sync"

	"github.com/google/uuid"
	domain_order "github.com/nuvotlyuba/trading-engine/internal/domain/order"
	"github.com/shopspring/decimal"
	// github.com/google/btree
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
	Bids    map[decimal.Decimal]*PriceLevel
	BidKeys []decimal.Decimal //отсортирован по убыванию: [102, 101, 100]

	Asks    map[decimal.Decimal]*PriceLevel
	AskKeys []decimal.Decimal // отсортированы по возрастанию: [103, 104, 105]

	mu sync.RWMutex
}

func NewOrderBook(symbol string) *OrderBook {
	return &OrderBook{
		Symbol:  symbol,
		Bids:    make(map[decimal.Decimal]*PriceLevel, 0),
		BidKeys: make([]decimal.Decimal, 0),
		Asks:    make(map[decimal.Decimal]*PriceLevel, 0),
		AskKeys: make([]decimal.Decimal, 0),
	}
}

func (ob *OrderBook) BestBid() (decimal.Decimal, bool) {
	ob.mu.RLock()
	defer ob.mu.RUnlock()

	if len(ob.BidKeys) == 0 {
		return decimal.Zero, false
	}
	return ob.BidKeys[0], true
}

func (ob *OrderBook) BestAsk() (decimal.Decimal, bool) {
	ob.mu.RLock()
	defer ob.mu.RUnlock()
	if len(ob.AskKeys) == 0 {
		return decimal.Zero, false
	}
	return ob.AskKeys[0], true
}

func (ob *OrderBook) AddOrder(order *domain_order.Order) {
	ob.mu.Lock()
	defer ob.mu.Unlock()

	if order.Side == domain_order.SideBuy {
		if level, ok := ob.Bids[order.Price]; !ok {
			ob.Bids[order.Price] = NewPriceLevel(order)
			ob.addKey(order.Price, order.Side)
		} else {
			level.Total = level.Total.Add(order.Remaining())
			level.Queue = append(level.Queue, order)
		}
	}
	if order.Side == domain_order.SideSell {
		if level, ok := ob.Asks[order.Price]; !ok {
			ob.Asks[order.Price] = NewPriceLevel(order)
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
		level, ok := ob.Bids[order.Price]
		if !ok {
			return fmt.Errorf("not found price level")
		}
		level.Total = level.Total.Sub(order.Remaining())
		level.Queue = removeOrderFromQueue(level.Queue, order.ID)

		if level.Total.IsZero() {
			delete(ob.Bids, order.Price)
			ob.removeKey(order.Price, order.Side)
		}

		return nil
	}

	if order.Side == domain_order.SideSell {
		level, ok := ob.Asks[order.Price]
		if !ok {
			return fmt.Errorf("not found price level")
		}
		level.Total = level.Total.Sub(order.Remaining())
		level.Queue = removeOrderFromQueue(level.Queue, order.ID)

		if level.Total.IsZero() {
			delete(ob.Asks, order.Price)
			ob.removeKey(order.Price, order.Side)
		}

	}
	return nil
}

func (ob *OrderBook) addKey(key decimal.Decimal, side domain_order.Side) {
	switch side {
	case domain_order.SideBuy:
		ob.BidKeys = append(ob.BidKeys, key)
		sort.Slice(ob.BidKeys, func(i, j int) bool {
			return ob.BidKeys[i].GreaterThan(ob.BidKeys[j])
		})
	case domain_order.SideSell:
		ob.AskKeys = append(ob.AskKeys, key)
		sort.Slice(ob.AskKeys, func(i, j int) bool {
			return ob.AskKeys[i].LessThan(ob.AskKeys[j])
		})
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

func (ob *OrderBook) removeKey(key decimal.Decimal, side domain_order.Side) {
	switch side {
	case domain_order.SideBuy:
		for i, k := range ob.BidKeys {
			if k.Equal(key) {
				ob.BidKeys = append(ob.BidKeys[:i], ob.BidKeys[i+1:]...)
				return
			}
		}
	case domain_order.SideSell:
		for i, k := range ob.AskKeys {
			if k.Equal(key) {
				ob.AskKeys = append(ob.AskKeys[:i], ob.AskKeys[i+1:]...)
				return
			}
		}
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
