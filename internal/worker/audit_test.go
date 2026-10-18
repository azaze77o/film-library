package worker_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"log/slog"
	"slices"
	"sync"
	"testing"
	"time"

	"github.com/project/omdbapp/internal/domain"
	"github.com/project/omdbapp/internal/worker"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeAuditStore struct {
	mu     sync.Mutex
	events []domain.LibraryEvent
	failOn domain.ImdbID // для этого id insert вернёт ошибку
}

func (f *fakeAuditStore) Insert(ctx context.Context, e domain.LibraryEvent) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if e.ImdbID == f.failOn {
		return errors.New("db is down")
	}

	f.events = append(f.events, e)
	return nil
}

func (f *fakeAuditStore) saved() []domain.LibraryEvent {
	f.mu.Lock()
	defer f.mu.Unlock()

	return slices.Clone(f.events)
}

func TestAuditWorker_ReadOnClose(t *testing.T) {
	events := make(chan domain.LibraryEvent, 2)
	firstEvent := domain.LibraryEvent{Type: domain.EventLibraryEntryCreated, ImdbID: "tt0133093"}
	secondEvent := domain.LibraryEvent{Type: domain.EventLibraryEntryCreated, ImdbID: "tt0108778"}

	// Отправляем оба события в канал
	events <- firstEvent
	events <- secondEvent

	// Закрываем канал
	close(events)

	store := &fakeAuditStore{}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	w := worker.NewAuditWorker(events, store, logger, time.Second)

	w.Run(t.Context())

	assert.Equal(t, []domain.LibraryEvent{firstEvent, secondEvent}, store.saved())
}

func TestAuditWorker_StoreErrorDoesNotStop(t *testing.T) {
	events := make(chan domain.LibraryEvent, 2)
	firstEvent := domain.LibraryEvent{Type: domain.EventLibraryEntryCreated, ImdbID: "tt0133093"}
	secondEvent := domain.LibraryEvent{Type: domain.EventLibraryEntryCreated, ImdbID: "tt0108778"}

	// Отправляем два события в канал
	events <- firstEvent
	events <- secondEvent

	// Закрываем очередь
	close(events)

	store := &fakeAuditStore{failOn: firstEvent.ImdbID}

	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, nil))

	w := worker.NewAuditWorker(events, store, logger, time.Second)

	w.Run(t.Context())

	assert.Equal(t, []domain.LibraryEvent{secondEvent}, store.saved())
	assert.Contains(t, buf.String(), "db is down")
}

func TestAuditWorker_StopOnCancel(t *testing.T) {
	events := make(chan domain.LibraryEvent)
	store := &fakeAuditStore{}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	ctx, cancel := context.WithCancel(t.Context())

	w := worker.NewAuditWorker(events, store, logger, time.Second)
	done := make(chan struct{})
	go func() {
		w.Run(ctx)
		close(done) // сработает только если Run вернулся
	}()

	cancel()

	select {
	case <-done: // ветка успеха, значит тест обработал сигнал отмены контекста
	case <-time.After(time.Second):
		t.Fatal("worker did not stop after cancel")
	}

}

func TestAuditWorker_ProcessesWhileRunning(t *testing.T) {
	events := make(chan domain.LibraryEvent)
	event := domain.LibraryEvent{Type: domain.EventLibraryEntryCreated, ImdbID: "tt0133093"}

	store := &fakeAuditStore{}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	ctx, cancel := context.WithCancel(t.Context())

	w := worker.NewAuditWorker(events, store, logger, time.Second)

	done := make(chan struct{})

	go func() {
		w.Run(ctx)
		close(done)
	}()

	events <- event

	require.Eventually(t, func() bool {
		return len(store.saved()) == 1
	}, time.Second, 10*time.Millisecond)
	assert.Equal(t, []domain.LibraryEvent{event}, store.saved())

	cancel()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("worker did not stop after cancel")
	}
}
