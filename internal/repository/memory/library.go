package memory

import (
	"time"

	"github.com/project/omdbapp/internal/domain"
)

// LibraryStore - структура хранения библиотеки фильмов
// data - это словарь с под-словарём:
// Пользователь -> Его фильм -> Его состояние в библиотеке
// Записи храним как указатели, т.к. будем обновлять их состояние на месте
type LibraryStore struct {
	data map[domain.UserID]map[domain.ImdbID]*domain.LibraryEntry
}

// Создаём пустую библиотеку (инициализировать map обязательно, иначе обращение к nil-map упадёт при записи)
func NewLibraryStore() *LibraryStore {
	return &LibraryStore{
		data: make(map[domain.UserID]map[domain.ImdbID]*domain.LibraryEntry),
	}

}

// upsert (update or insert) - вспомогательный метод, который проверяет наличие фильма в библиотеке
// создаёт новую запись, если её нет, или возвращает существующую
// возвращает запись библиотеки
func (s *LibraryStore) upsert(u domain.UserID, m domain.MoviePreview) *domain.LibraryEntry {
	// Создаем подсловарь, если у пользователя его еще нет.
	if s.data[u] == nil {
		s.data[u] = make(map[domain.ImdbID]*domain.LibraryEntry)
	}

	entry, ok := s.data[u][m.ID]
	if !ok {
		entry = &domain.LibraryEntry{
			UserID:       u,
			MoviePreview: m,
			CreatedAt:    time.Now(),
		}
		s.data[u][m.ID] = entry
	}
	return entry
}

// AddFavorite добавляет фильм в избранное пользователя
func (s *LibraryStore) AddFavorite(u domain.UserID, m domain.MoviePreview) error {
	s.upsert(u, m).IsFavorite = true
	return nil
}

// MarkWatched делает отметку о том, что фильм просмотрен
func (s *LibraryStore) MarkWatched(u domain.UserID, m domain.MoviePreview) error {
	s.upsert(u, m).IsWatched = true
	return nil
}

// ListMovies предоставляет доступ пользователя к списку его фильмов
func (s *LibraryStore) ListMovies(u domain.UserID) ([]domain.LibraryEntry, error) {
	if s.data[u] == nil {
		return nil, nil
	}
	result := make([]domain.LibraryEntry, 0, len(s.data[u]))
	for _, entry := range s.data[u] {
		result = append(result, *entry)
	}
	return result, nil
}
