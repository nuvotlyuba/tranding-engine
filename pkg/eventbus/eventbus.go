package eventbus

import (
	"context"
	"fmt"
	"log"
	"sync"
	"sync/atomic"
	"time"
)

// синхронный
// нет защиты от конкурентного доступа

type BusV1 struct {
	subs map[string][]func(payload any)
}

func NewV1() *BusV1 {
	return &BusV1{
		subs: make(map[string][]func(payload any)),
	}
}

func (b *BusV1) Subscribe(topic string, handler func(payload any)) {
	b.subs[topic] = append(b.subs[topic], handler)
}

func (b *BusV1) Publish(topic string, payload any) {
	for _, handler := range b.subs[topic] {
		handler(payload)
	}
}

//  нет защиты от конкурентного доступа к map
// неограниченное число горутин при всплеске событий
// нет graceful shutdown

type BusV2 struct {
	subs map[string][]func(payload any)
}

func NewV2() *BusV2 {
	return &BusV2{
		subs: make(map[string][]func(payload any)),
	}
}

func (b *BusV2) Subscribe(topic string, handler func(payload any)) {
	b.subs[topic] = append(b.subs[topic], handler)
}

func (b *BusV2) Publish(topic string, payload any) {
	for _, handler := range b.subs[topic] {
		go handler(payload)
	}
}

type subscriber struct {
	queue   chan any
	handler func(payload any)
}

type BusV3 struct {
	subs map[string][]*subscriber
	mu   sync.RWMutex
}

func NewV3() *BusV3 {
	return &BusV3{
		subs: make(map[string][]*subscriber),
	}
}
func (b *BusV3) Subscribe(topic string, handler func(payload any), bufferSize int) {
	sub := &subscriber{
		queue: make(chan any, bufferSize),
	}

	go func() {
		for payload := range sub.queue {
			sub.handler(payload)
		}
	}()

	b.mu.Lock()
	b.subs[topic] = append(b.subs[topic], sub)
	b.mu.Unlock()
}

func (b *BusV3) Publish(topic string, payload any) {
	b.mu.RLock()
	subs := b.subs[topic]
	b.mu.RUnlock()

	for _, sub := range subs {
		select {
		case sub.queue <- payload:
		default:
			log.Printf("eventbus: buffer full, topic - %s", topic)
		}
	}
}

type Event struct {
	ID        string
	Topic     string
	Payload   any
	CreatedAt time.Time
	Attempt   int
}

type Handler func(ctx context.Context, event Event) error

type RetryPolicy struct {
	MaxAttempts  int
	InitialDealy time.Duration
	MaxDelay     time.Duration
}

var NoRetry = RetryPolicy{MaxAttempts: 1}
var DefaultRetry = RetryPolicy{
	MaxAttempts:  5,
	InitialDealy: 100 * time.Millisecond,
	MaxDelay:     10 * time.Second,
}

type SubscribeOptions struct {
	Name       string      // имя подписчика
	Workers    int         // количество параллельных воркеров читающих из очереди. использовать 1 для обработчиков которые должны получать события строго порядку
	BufferSize int         // размер буффера очереди событий в штуках
	Retry      RetryPolicy // политика повторных попыток при ошибке обработчика
}

type subscriptionMetrics struct {
	received  atomic.Int64 // сколько событий поступило в очередь подписчика
	processed atomic.Int64 // сколько собтий успешно обработано
	failed    atomic.Int64 // сколько событий не удалось доставить после всех retry
	dropped   atomic.Int64 // сколько событий дропнуто из-за заполненного буфера
	retried   atomic.Int64 // суммарное количество попыток повторной доставки
}

type subscription struct {
	name    string              // имя подписчика из SubscribeOptions.Name
	handler Handler             // функция обработчик которую вызывают воркеры
	queue   chan Event          // буф канал через который Publish передает события воркерам
	opts    SubscribeOptions    // хранятся здесь чтобы воркеры имени доступ к политике retry
	metrics subscriptionMetrics // счетчик для мониторинга этой подписки
}

type Bus struct {
	subs map[string][]*subscription
	mu   sync.RWMutex
}

type SubscriptionMetrics struct {
	// Name — имя подписчика
	Name string
	// Received — получено событий всего
	Received int64
	// Processed — успешно обработано
	Processed int64
	// Failed — не удалось доставить после всех retry
	Failed int64
	// Dropped — дропнуто из-за полного буфера
	Dropped int64
	// Retried — суммарно повторных попыток
	Retried int64
}

func New() *Bus {
	return &Bus{
		subs: make(map[string][]*subscription),
	}
}

func (b *Bus) Subscribe(ctx context.Context, topic string, handler Handler, opts SubscribeOptions) func() {
	if opts.Workers == 0 {
		opts.Workers = 1
	}
	if opts.Name == "" {
		opts.Name = fmt.Sprintf("subscriber-%s", topic)
	}
	if opts.BufferSize == 0 {
		opts.BufferSize = 100
	}
	sub := &subscription{
		name:    opts.Name,
		handler: handler,
		queue:   make(chan Event, opts.BufferSize),
		opts:    opts,
	}

	b.mu.Lock()
	b.subs[topic] = append(b.subs[topic], sub)
	b.mu.Unlock()

	// запускаем пул воркеров
	var wg sync.WaitGroup
	for range opts.Workers {
		wg.Add(1)
		go func() {
			wg.Done()
			sub.runWorker(ctx)
		}()
	}

	return func() {
		b.mu.Lock()
		subs := b.subs[topic]
		for i, s := range subs {
			if s == sub {
				b.subs[topic] = append(subs[:i], subs[i+1:]...)
				break
			}
		}
		b.mu.Unlock()
		wg.Wait()
	}
}
