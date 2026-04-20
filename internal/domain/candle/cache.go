package candle

import (
	"fmt"
	"sync"
)

type Cache struct {
	mu      sync.RWMutex
	storage map[string][]Candle // "BTCUSDT_1m" → []Candle
	limit   int                 // максимум свечей на один ключ
}

func NewCache(limit int) *Cache {
	return &Cache{
		storage: make(map[string][]Candle),
		limit:   limit,
	}
}

func (c *Cache) Push(candle Candle) {
	c.mu.Lock()
	defer c.mu.Unlock()

	key := cacheKey(candle.Symbol, candle.Period)
	candles := c.storage[key]

	candles = append(candles, candle)

	if len(candles) > c.limit {
		candles = candles[len(candles)-c.limit:]
	}

	c.storage[key] = candles
}

func (c *Cache) Last(symbol string, period Period, n int) []Candle {
	c.mu.RLock()
	defer c.mu.RUnlock()

	key := cacheKey(symbol, period)
	candles := c.storage[key]

	if len(candles) == 0 {
		return nil
	}

	if n >= len(candles) {
		result := make([]Candle, len(candles))
		copy(result, candles)
		return result
	}

	result := make([]Candle, n)
	copy(result, candles[len(candles)-n:])
	return result
}

func (c *Cache) Latest(symbol string, period Period) (Candle, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	key := cacheKey(symbol, period)
	candles := c.storage[key]

	if len(candles) == 0 {
		return Candle{}, false
	}
	return candles[len(candles)-1], true
}

func (c *Cache) Len(symbol string, period Period) int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return len(c.storage[cacheKey(symbol, period)])
}

func cacheKey(symbol string, period Period) string {
	return fmt.Sprintf("%s_%s", symbol, period)
}
