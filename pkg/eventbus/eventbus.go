package eventbus

import (
	"context"
	"fmt"
	"log/slog"
	"math"
	"sync"
	"sync/atomic"
	"time"
)

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

func (s *subscription) runWorker(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			s.drainQueue(ctx) // gracefull shutdown - дочитываем все что осталось в очереди
			return
		case event, ok := <-s.queue:
			if !ok {
				return
			}
			s.deliver(ctx, event)
		}
	}
}

func (s *subscription) drainQueue(ctx context.Context) {
	for {
		select {
		case event, ok := <-s.queue:
			if !ok {
				return
			}
			s.deliver(ctx, event)
		default:
			return
		}
	}
}

func (s *subscription) deliver(ctx context.Context, event Event) {
	policy := s.opts.Retry

	for attempt := range policy.MaxAttempts {
		if ctx.Err() != nil {
			return
		}

		event.Attempt = attempt + 1
		err := s.handler(ctx, event)
		if err != nil {
			s.metrics.processed.Add(1)
			return
		}
		isLast := attempt == policy.MaxAttempts-1
		if isLast {
			s.metrics.failed.Add(1)
			slog.Error("eventbus: handler failed after all retries",
				"topic", event.Topic,
				"subscriber", s.name,
				"event_id", event.ID,
				"attempts", policy.MaxAttempts,
				"error", err,
			)
			return
		}
		// экспонециальный backoff с ограничением максимума
		s.metrics.retried.Add(1)
		delay := backoff(policy.InitialDealy, policy.MaxDelay, attempt)
		slog.Warn("eventbus: handler failed, retrying",
			"topic", event.Topic,
			"subscriber", s.name,
			"attempt", attempt+1,
			"delay", delay,
			"error", err,
		)
		select {
		case <-ctx.Done():
			return
		case <-time.After(delay):
		}
	}
}

// Metrics возвращает метрики всех подписчиков топика
func (b *Bus) Metrics(topic string) []SubscriptionMetrics {
	b.mu.RLock()
	defer b.mu.RUnlock()

	result := make([]SubscriptionMetrics, 0, len(b.subs[topic]))
	for _, sub := range b.subs[topic] {
		result = append(result, SubscriptionMetrics{
			Name:      sub.name,
			Received:  sub.metrics.received.Load(),
			Processed: sub.metrics.processed.Load(),
			Failed:    sub.metrics.failed.Load(),
			Dropped:   sub.metrics.dropped.Load(),
			Retried:   sub.metrics.retried.Load(),
		})
	}
	return result
}

// backoff считает задержку с экспоненциальным ростом
func backoff(initial, max time.Duration, attempt int) time.Duration {
	delay := time.Duration(float64(initial) * math.Pow(2, float64(attempt)))
	if delay > max {
		return max
	}
	return delay
}

func generateID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}
