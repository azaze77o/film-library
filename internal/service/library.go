// Сервис управления библиотекой пользователя с возможными сценариями
// 1. Просмотра списка фильмов библиотеки
// 2. Отметить фильм как избранное
// 3. Отметить фильм как просмотренный

package service

import (
	"github.com/project/omdbapp/internal/domain"
)

type LibraryStore interface {
	AddFavorite(u domain.UserID, m domain.MoviePreview) error
	MarkWatched(u domain.UserID, m domain.MoviePreview) error
	ListMovies(u domain.UserID) ([]domain.LibraryEntry, error)
}

type LibraryService struct {
	store LibraryStore
}

func NewLibraryService(store LibraryStore) *LibraryService {
	return &LibraryService{store: store}
}

func (s *LibraryService) AddFavorite(u domain.UserID, m domain.MoviePreview) error {
	return s.store.AddFavorite(u, m)
}

func (s *LibraryService) MarkWatched(u domain.UserID, m domain.MoviePreview) error {
	return s.store.MarkWatched(u, m)
}

func (s *LibraryService) ListMovies(u domain.UserID) ([]domain.LibraryEntry, error) {
	return s.store.ListMovies(u)
}
