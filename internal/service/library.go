package service

import (
	"context"
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

type LibraryService struct {
	store LibraryStore
}

func NewLibraryService(store LibraryStore) *LibraryService {
	return &LibraryService{store: store}
}

func (s *LibraryService) Create(ctx context.Context, u domain.UserID, m domain.MoviePreview) (domain.LibraryEntry, error) {
	entry := domain.LibraryEntry{
		UserID:       u,
		MoviePreview: m,
		CreatedAt:    time.Now(),
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
