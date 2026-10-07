package service

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/project/omdbapp/internal/domain"
)

type LibraryStore interface {
	Create(ctx context.Context, e domain.LibraryEntry) error
	Get(ctx context.Context, u domain.UserID, id domain.ImdbID) (domain.LibraryEntry, error)
	List(ctx context.Context, u domain.UserID) ([]domain.LibraryEntry, error)
	Update(ctx context.Context, u domain.UserID, id domain.ImdbID, patch domain.LibraryPatch) (domain.LibraryEntry, error)
	Delete(ctx context.Context, u domain.UserID, id domain.ImdbID) error
}

// movieIDPattern - шаблон корректного ID фильма
var movieIDPattern = regexp.MustCompile(`^tt[0-9]+$`)

type LibraryService struct {
	store LibraryStore
	now   func() time.Time // now передается как функция, чтобы можно было протестировать сервис
}

func NewLibraryService(store LibraryStore) *LibraryService {
	return &LibraryService{store: store, now: time.Now}
}

func (s *LibraryService) Create(ctx context.Context, u domain.UserID, m domain.MoviePreview) (domain.LibraryEntry, error) {
	// Выполянем валидацию фильма
	if err := validateMovie(m); err != nil {
		return domain.LibraryEntry{}, err
	}

	// Приводим фильм к доменной модели
	entry := domain.LibraryEntry{
		UserID:       u,
		MoviePreview: m,
		CreatedAt:    s.now(),
	}

	if err := s.store.Create(ctx, entry); err != nil {
		return domain.LibraryEntry{}, err
	}

	return entry, nil
}

func (s *LibraryService) Get(ctx context.Context, u domain.UserID, id domain.ImdbID) (domain.LibraryEntry, error) {
	return s.store.Get(ctx, u, id)
}

func (s *LibraryService) List(ctx context.Context, u domain.UserID) ([]domain.LibraryEntry, error) {
	return s.store.List(ctx, u)
}

func (s *LibraryService) Update(ctx context.Context, u domain.UserID, id domain.ImdbID, patch domain.LibraryPatch) (domain.LibraryEntry, error) {
	if patch.Favourite == nil && patch.Watched == nil {
		return domain.LibraryEntry{}, domain.ErrInvalidInput
	}
	return s.store.Update(ctx, u, id, patch)

}

func (l *LibraryService) Delete(ctx context.Context, u domain.UserID, id domain.ImdbID) error {
	return l.store.Delete(ctx, u, id)
}

// validateMovie  выполняем проверку входного фильма
func validateMovie(m domain.MoviePreview) error {
	if !movieIDPattern.MatchString(string(m.ID)) {
		return fmt.Errorf("%w: invalid object ID", domain.ErrInvalidInput)
	}

	if strings.TrimSpace(m.Title) == "" {
		return fmt.Errorf("%w: title is required", domain.ErrInvalidInput)
	}
	return nil
}
