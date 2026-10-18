package queue

import (
	"context"
	"errors"
	"sync"

	"github.com/project/omdbapp/internal/domain"
)

var (
	ErrClosed = errors.New("queue is closed")
	ErrFull   = errors.New("queue is full")
)

type Queue struct {
	mu     sync.Mutex
	ch     chan domain.LibraryEvent
	closed bool
}

func New(size int) *Queue {
	return &Queue{
		ch: make(chan domain.LibraryEvent, size),
	}
}

// Publish гарантирует одно из состояний: либо событие ушло в очередь, либо возникла ошибка
func (q *Queue) Publish(ctx context.Context, e domain.LibraryEvent) error {
	q.mu.Lock()
	defer q.mu.Unlock() // Снимаем замок при любых обстоятельствах выхода из функции

	if q.closed {
		return ErrClosed
	}

	select {
	case q.ch <- e:
		return nil
	default:
		return ErrFull
	}
}

// Close закрывает канал
func (q *Queue) Close() {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.closed {
		return
	}

	q.closed = true
	close(q.ch)
}

func (q *Queue) Events() <-chan domain.LibraryEvent {
	return q.ch
}
