package subpub

import (
	"context"
	"errors"
	"sync"
)

var (
	ErrSubPubClosed = errors.New("subpub is closed")
)

type MessageHandler func(msg interface{})

type Subscription interface {
	Unsubscribe()
}

type SubPub interface {
	Subscribe(subject string, cb MessageHandler) (Subscription, error)
	Publish(subject string, msg interface{}) error
	Close(ctx context.Context) error
}

// реализуем интерфейс Subscription
type subscription struct {
	subject  string // событие, на которое подписаны
	handler  MessageHandler
	bus      *subPub // ссылка на шину
	mu       sync.Mutex
	isActive bool
	msgChan  chan interface{} // канал для сохранения порядка сообщений
}

// Реализация шины (интерфейс SubPub)
type subPub struct {
	subjects map[string][]*subscription // хранилище подписок
	mu       sync.Mutex
	closed   bool
	pubWg    sync.WaitGroup // отслеживаем публикации
}

func NewSubPub() SubPub {
	return &subPub{
		subjects: make(map[string][]*subscription),
	}
}

func (s *subscription) Unsubscribe() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.isActive {
		return
	}

	s.isActive = false
	close(s.msgChan)
	// удаляем подписку в шине
	s.bus.removeSubscription(s)
}

// создание подписки
func (sb *subPub) Subscribe(subject string, cb MessageHandler) (Subscription, error) {
	sb.mu.Lock()
	defer sb.mu.Unlock()

	if sb.closed {
		return nil, ErrSubPubClosed
	}

	sub := &subscription{
		subject:  subject,
		handler:  cb,
		bus:      sb,
		isActive: true,
		msgChan:  make(chan interface{}, 100),
	}

	sb.subjects[subject] = append(sb.subjects[subject], sub)

	// асинхронно "слушаем" сообщения
	go func() {
		// чтение из канала обеспечивает FIFO
		for msg := range sub.msgChan {
			sub.mu.Lock()
			if sub.isActive {
				sub.handler(msg)
			}
			sub.mu.Unlock()
		}
	}()

	return sub, nil
}

// публикация события
func (sb *subPub) Publish(subject string, msg interface{}) error {
	sb.mu.Lock()
	defer sb.mu.Unlock()

	if sb.closed {
		return ErrSubPubClosed
	}

	subs, ok := sb.subjects[subject]
	// нет слушателей - не публикуем
	if !ok {
		return nil
	}

	// +1 публикация
	sb.pubWg.Add(1)
	defer sb.pubWg.Done()

	for _, sub := range subs {
		sub.mu.Lock()
		if sub.isActive {
			sub.msgChan <- msg // отправляем сообщение в канал слушателя
		}
		sub.mu.Unlock()
	}

	return nil
}

// завершаем работу шины
func (sb *subPub) Close(ctx context.Context) error {
	sb.mu.Lock()
	if sb.closed {
		sb.mu.Unlock()
		return nil
	}
	sb.closed = true
	sb.mu.Unlock()

	// ждем, пока завершаются публикации
	pubDone := make(chan struct{})
	go func() {
		sb.pubWg.Wait()
		close(pubDone)
	}()

	select {
	case <-pubDone:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}

}

// метод, удаляющий подписку
func (sb *subPub) removeSubscription(s *subscription) {
	sb.mu.Lock()
	defer sb.mu.Unlock()

	// получаем слайс подписок на событие
	subs, ok := sb.subjects[s.subject]
	if !ok {
		return
	}

	for i, curr := range subs {
		if curr == s {
			// удаляем элемент и обнуляем последний элемент слайса
			subs[i] = subs[len(subs)-1]
			subs[len(subs)-1] = nil
			sb.subjects[s.subject] = subs[:len(subs)-1]
			break
		}
	}

	// если эта подписка была последней, удалим событие
	if len(sb.subjects[s.subject]) == 0 {
		delete(sb.subjects, s.subject)
	}
}
