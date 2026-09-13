package service

import (
	"time"

	"github.com/project/omdbapp/internal/domain"
)

type LibraryStore interface {
	Create(m domain.LibraryEntry) error
	Get(u domain.UserID, m domain.ImdbID) (domain.LibraryEntry, error)
	List(u domain.UserID) ([]domain.LibraryEntry, error)
	Update(u domain.UserID, id domain.ImdbID, mutate func(*domain.LibraryEntry)) (domain.LibraryEntry, error)
	Delete(u domain.UserID, id domain.ImdbID) error
}

type LibraryService struct {
	store LibraryStore
}

type UpdatePath struct {
	Favorite *bool
	Watched  *bool
}

func NewLibraryService(store LibraryStore) *LibraryService {
	return &LibraryService{store: store}
}

func (s *LibraryService) Create(u domain.UserID, m domain.MoviePreview) (domain.LibraryEntry, error) {
	entry := domain.LibraryEntry{
		UserID:       u,
		MoviePreview: m,
		CreatedAt:    time.Now(),
	}

	if err := s.store.Create(entry); err != nil {
		return domain.LibraryEntry{}, err
	}

	return entry, nil
}

func (s *LibraryService) Get(u domain.UserID, id domain.ImdbID) (domain.LibraryEntry, error) {
	return s.store.Get(u, id)
}

func (s *LibraryService) List(u domain.UserID) ([]domain.LibraryEntry, error) {
	return s.store.List(u)
}

func (s *LibraryService) Update(u domain.UserID, id domain.ImdbID, path UpdatePath) (domain.LibraryEntry, error) {
	return s.store.Update(u, id, func(e *domain.LibraryEntry) {
		if path.Favorite != nil {
			e.IsFavorite = *path.Favorite
		}
		if path.Watched != nil {
			e.IsWatched = *path.Watched
		}
	})
}

func (l *LibraryService) Delete(u domain.UserID, id domain.ImdbID) error {
	return l.store.Delete(u, id)
}
