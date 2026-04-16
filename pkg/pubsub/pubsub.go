package pubsub

import (
	"errors"
	"log/slog"
	"sync/atomic"
)

type commandType int8

const (
	cmdAdd commandType = iota
	cmdRemove
)

// команда управления подписчиками
type command[T any] struct {
	id int64
	ch chan T
	op commandType
}

// Broadcaster - фан-аут (broadcast) сообщений всем подписчикам
// Один входной канал -> много выходных каналов (по одному на подписчика)
type Broadcaster[T any] struct {
	log   *slog.Logger
	cmdCh chan command[T] // канал для команд

	counter atomic.Int64 // генератор ID подписчиков
	active  atomic.Bool  // открыт ли  subscriber
}

func NewBroadcaster[T any](logger *slog.Logger, topic string, in <-chan T) *Broadcaster[T] {
	if logger == nil {
		logger = slog.Default()
	}

	logger = logger.With("component", "broadcaster", "topic", topic)

	b := &Broadcaster[T]{
		log:   logger,
		cmdCh: make(chan command[T], 64),
	}

	b.active.Store(true)

	go b.run(in)

	return b
}

func (b *Broadcaster[T]) run(in <-chan T) {
	subscribers := make(map[int64]chan T)
	// уникальный ID подписчика
	// канал, в который ты отправляешь сообщения

	for {
		select {
		case cmd, ok := <-b.cmdCh:
			if !ok {
				b.shutdown(subscribers)
				return
			}
			b.handleCommand(subscribers, cmd)
		case msg, ok := <-in:
			if !ok {
				b.shutdown(subscribers)
				return
			}
			b.broadcast(subscribers, msg)
		}
	}
}

// handleCommand — добавление/удаление подписчиков
func (b *Broadcaster[T]) handleCommand(subscribers map[int64]chan T, cmd command[T]) {
	switch cmd.op {
	case cmdAdd:
		subscribers[cmd.id] = cmd.ch
	case cmdRemove:
		if ch, ok := subscribers[cmd.id]; ok {
			delete(subscribers, cmd.id)
			close(ch)
		}
	}
}

// broadcast — отправка сообщения всем подписчикам
func (b *Broadcaster[T]) broadcast(subscribers map[int64]chan T, msg T) {
	for _, ch := range subscribers {
		select {
		case ch <- msg:
		default:
			// если подписчик не успевает, то дропаем
			// важно! не облокируем систему
			b.log.Debug("drop message: slow subscriber")
		}
	}
}

// shutdown — закрытие всех подписчиков
func (b *Broadcaster[T]) shutdown(subscribers map[int64]chan T) {
	b.active.Store(false)

	for _, ch := range subscribers {
		b.safeClose(ch)
	}
}

// safeClose — защита от panic при двойном закрытии
func (b *Broadcaster[T]) safeClose(ch chan T) {
	defer func() {
		if r := recover(); r != nil {
			b.log.Error("channel already closed: %v", r)
		}
	}()
	close(ch)
}

// Add — подписка на поток сообщений
func (b *Broadcaster[T]) Add() (<-chan T, func(), error) {
	if !b.active.Load() {
		return nil, nil, errors.New("closed")
	}

	ch := make(chan T, 1000) // буфер для медленных подписчиков
	id := b.counter.Add(1)

	b.cmdCh <- command[T]{
		id: id,
		ch: ch,
		op: cmdAdd,
	}

	cancel := func() {
		if b.active.Load() {
			b.cmdCh <- command[T]{
				id: id,
				op: cmdRemove,
			}
		}
	}

	return ch, cancel, nil
}

// Close — завершает работу и закрывает всех подписчиков
func (b *Broadcaster[T]) Close() {
	if b.active.CompareAndSwap(true, false) {
		close(b.cmdCh)
	}
}
