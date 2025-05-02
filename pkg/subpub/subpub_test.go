package subpub

import (
	"context"
	"log"
	"runtime"
	"sync"
	"testing"
	"time"
)

// тест на несколько слушателей
func TestMultipleSubscribers(t *testing.T) {
	bus := NewSubPub()
	defer func() {
		if err := bus.Close(context.Background()); err != nil {
			log.Println("failed to close bus: %w", err)
		}
	}()

	const subject = "test"
	const subscribersCount = 10

	var wg sync.WaitGroup
	received := make([]string, subscribersCount)

	for i := 0; i < subscribersCount; i++ {
		i := i // чтобы не было замыкания
		wg.Add(1)
		_, err := bus.Subscribe(subject, func(msg interface{}) {
			received[i] = msg.(string)
			wg.Done()
		})
		if err != nil {
			t.Fatalf("subscribe failed: %v", err)
		}
	}

	err := bus.Publish(subject, "test")
	if err != nil {
		t.Fatalf("publish failed: %v", err)
	}

	// ждем пока все получат сообщения
	wg.Wait()

	// проверяем, что все получили сообщение
	for i, val := range received {
		if val != "test" {
			t.Errorf("subscriber %d got wrong value: %s, want test", i, val)
		}
	}
}

// тест на отписку
func TestUnsubscribe(t *testing.T) {
	bus := NewSubPub()
	defer func() {
		if err := bus.Close(context.Background()); err != nil {
			log.Println("failed to close bus: %w", err)
		}
	}()

	const subject = "test"
	var received int

	sub, err := bus.Subscribe(subject, func(msg interface{}) {
		received = msg.(int)
	})
	if err != nil {
		t.Fatalf("subscribe failed: %v", err)
	}

	sub.Unsubscribe()

	err = bus.Publish(subject, "test")
	if err != nil {
		t.Fatalf("publish failed: %v", err)
	}

	time.Sleep(111 * time.Millisecond)

	if received != 0 {
		t.Errorf("received message after unsubscribe: %d, want 0", received)
	}
}

// тест на сохранность порядка
func TestMessageOrder(t *testing.T) {
	bus := NewSubPub()
	defer func() {
		if err := bus.Close(context.Background()); err != nil {
			log.Println("failed to close bus: %w", err)
		}
	}()

	const subject = "test"
	const messagesCount = 100

	var wg sync.WaitGroup
	received := make([]int, 0, messagesCount)
	var mu sync.Mutex

	wg.Add(messagesCount)
	_, err := bus.Subscribe(subject, func(msg interface{}) {
		mu.Lock()
		defer mu.Unlock()
		received = append(received, msg.(int))
		wg.Done()
	})
	if err != nil {
		t.Fatalf("subscribe failed: %v", err)
	}

	for i := 0; i < messagesCount; i++ {
		err := bus.Publish(subject, i)
		if err != nil {
			t.Fatalf("publish failed: %v", err)
		}
	}

	wg.Wait()

	for i := 0; i < messagesCount; i++ {
		if received[i] != i {
			t.Errorf("wrong message order at index %d: got %d, want %d", i, received[i], i)
			break
		}
	}
}

// тест на утечки
func TestGoroutineLeaks(t *testing.T) {
	initialGoroutines := runtime.NumGoroutine()
	bus := NewSubPub()
	const subject = "test"

	sub, err := bus.Subscribe(subject, func(msg interface{}) {})
	if err != nil {
		t.Fatalf("subscribe failed: %v", err)
	}
	sub.Unsubscribe()

	err = bus.Close(context.Background())
	if err != nil {
		t.Fatalf("close failed: %v", err)
	}

	time.Sleep(111 * time.Millisecond)

	finalGoroutines := runtime.NumGoroutine()
	if finalGoroutines > initialGoroutines {
		t.Errorf("goroutine leak: before %d, after %d", initialGoroutines, finalGoroutines)
	}
}
