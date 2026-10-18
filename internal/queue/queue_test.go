package queue_test

import (
	"errors"
	"sync"
	"testing"

	"github.com/project/omdbapp/internal/domain"
	"github.com/project/omdbapp/internal/queue"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newEvent(id domain.ImdbID) domain.LibraryEvent {
	return domain.LibraryEvent{
		Type:   domain.EventLibraryEntryCreated,
		UserID: "default",
		ImdbID: id,
	}
}

func TestQueue_PublishAndReceive(t *testing.T) {
	q := queue.New(1)
	want := newEvent("tt33735025")

	err := q.Publish(t.Context(), want)
	require.NoError(t, err)

	got := <-q.Events()
	assert.Equal(t, want, got)
}

func TestQueue_Full(t *testing.T) {
	q := queue.New(1)

	firstEvent := newEvent("tt33735025")
	secondEvent := newEvent("tt0108778")

	err := q.Publish(t.Context(), firstEvent)
	require.NoError(t, err)

	err = q.Publish(t.Context(), secondEvent)
	require.ErrorIs(t, err, queue.ErrFull)
}

func TestQueue_PublishAfterClose(t *testing.T) {
	q := queue.New(1)
	q.Close()

	event := newEvent("tt33735025")

	err := q.Publish(t.Context(), event)
	require.ErrorIs(t, err, queue.ErrClosed)
}

func TestQueue_CloseTwice(t *testing.T) {
	q := queue.New(1)
	q.Close()

	assert.NotPanics(t, func() { q.Close() })
}

func TestQueue_ReadAfterClose(t *testing.T) {
	q := queue.New(2)

	firstEvent := newEvent("tt33735025")
	secondEvent := newEvent("tt0108778")

	// Отправляем первое событие
	err := q.Publish(t.Context(), firstEvent)
	require.NoError(t, err)

	// Отправляем второе событие
	err = q.Publish(t.Context(), secondEvent)
	require.NoError(t, err)

	// Закрываем очередь
	q.Close()

	// Объявляем переменную, в которую сложим все полученные события
	var got []domain.LibraryEvent

	// Читаем очередь и складываем результат в слайс
	for e := range q.Events() {
		got = append(got, e)
	}

	assert.Equal(t, []domain.LibraryEvent{firstEvent, secondEvent}, got)
}

func TestQueue_ConcurrentPublishAndClose(t *testing.T) {
	q := queue.New(10)
	event := newEvent("tt33735025")

	var wg sync.WaitGroup

	for range 100 {
		wg.Go(func() {
			err := q.Publish(t.Context(), event)
			if err != nil && errors.Is(err, queue.ErrClosed) && errors.Is(err, queue.ErrFull) {
				t.Errorf("unexpected code %d", err)
			}
		})

	}

	q.Close()

	wg.Wait()
}
