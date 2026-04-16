package pubsub

import (
	"context"
	"encoding"
	"sync"
)

// TopicPublisher — удобный publisher для конкретного топика
type TopicPublisher struct {
	topic string
	bus   *InMemoryBus
}

// topic — внутреннее представление топика
type topic struct {
	broadcaster *Broadcaster[[]byte]
	in          chan []byte
}

// InMemoryBus — простой pub/sub внутри одного процесса
type InMemoryBus struct {
	mu     sync.Mutex
	topics map[string]*topic
}

func NewInMemoryBus() *InMemoryBus {
	return &InMemoryBus{
		topics: make(map[string]*topic),
	}
}

func (bus *InMemoryBus) Init(ctx context.Context) error {
	return nil
}

func (bus *InMemoryBus) Shutdown() {
	bus.mu.Lock()
	defer bus.mu.Unlock()

	for _, t := range bus.topics {
		t.broadcaster.Close()
		close(t.in)
	}
}

// getOrCreateTopic — скрывает логику создания топика
func (bus *InMemoryBus) getOrCreateTopic(name string) *topic {
	t, ok := bus.topics[name]
	if ok {
		return t
	}

	in := make(chan []byte, 1000)

	t = &topic{
		in:          in,
		broadcaster: NewBroadcaster(nil, name, in),
	}

	bus.topics[name] = t

	return t
}

// Subscribe — подписка на топик
func (bus *InMemoryBus) Subscribe(topicName string) (<-chan []byte, func(), error) {
	bus.mu.Lock()
	defer bus.mu.Unlock()

	t := bus.getOrCreateTopic(topicName)
	return t.broadcaster.Add()
}

// Publish — отправка сообщений в топик
func (bus *InMemoryBus) Publish(
	ctx context.Context,
	topicName string,
	msgs ...encoding.BinaryMarshaler,
) error {
	bus.mu.Lock()
	t, ok := bus.topics[topicName]
	bus.mu.Unlock()

	if !ok {
		return nil // или ошибка — зависит от твоей логики
	}

	for _, msg := range msgs {
		buf, err := msg.MarshalBinary()
		if err != nil {
			return err
		}

		select {
		case t.in <- buf:
		case <-ctx.Done():
			return ctx.Err()
		default:
			// канал переполнен — дропаем
		}
	}

	return nil
}

// UseTopic — удобный способ получить publisher для одного топика
func (bus *InMemoryBus) UseTopic(topic string) *TopicPublisher {
	return &TopicPublisher{
		topic: topic,
		bus:   bus,
	}
}

func (tp *TopicPublisher) Publish(ctx context.Context, msg encoding.BinaryMarshaler) error {
	return tp.bus.Publish(ctx, tp.topic, msg)
}
