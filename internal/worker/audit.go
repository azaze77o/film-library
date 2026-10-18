package worker

import (
	"context"
	"log/slog"
	"time"

	"github.com/project/omdbapp/internal/domain"
)

type AuditStore interface {
	Insert(ctx context.Context, e domain.LibraryEvent) error
}

type AuditWorker struct {
	events  <-chan domain.LibraryEvent
	store   AuditStore
	logger  *slog.Logger
	timeout time.Duration
}

func NewAuditWorker(
	events <-chan domain.LibraryEvent,
	store AuditStore,
	logger *slog.Logger,
	timeout time.Duration,
) *AuditWorker {
	return &AuditWorker{
		events:  events,
		store:   store,
		logger:  logger,
		timeout: timeout,
	}
}

func (w *AuditWorker) Run(ctx context.Context) {
	w.logger.Info("audit worker started")
	defer w.logger.Info("audit worker stopped")

	for {
		select {
		case <-ctx.Done():
			return
		case e, ok := <-w.events:
			if !ok {
				return
			}
			w.handle(ctx, e)
		}
	}
}

// handle обрабатывает одно событие: записывает его в базу и если не уадлось, логирует ошибку
// Причины вынос запись события связано с defer cancel(), которое работает только при выходе из функции.
// За каждую итерацию в цикле Run накопилось множество timeout, которые мы долго ни убивалось, что приводило к утечке ресурсов.
// Еще одна причина заключается в том, что ошибка не останавливает воркер.
// В handle после ошибки стоит return. Он выходит только из handle, а не из Run. Цикл в Run продолжается.
func (w *AuditWorker) handle(ctx context.Context, e domain.LibraryEvent) {
	ctx, cancel := context.WithTimeout(ctx, w.timeout)
	defer cancel()

	if err := w.store.Insert(ctx, e); err != nil {
		w.logger.Error(
			"insert library event",
			"error", err,
			"event_type", e.Type,
			"imdb_id", e.ImdbID,
		)
		return
	}
}
