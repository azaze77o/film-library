package memory

import (
	"sync"

	"github.com/project/omdbapp/internal/domain"
)

// LibraryStore - структура хранения библиотеки фильмов
// data - это словарь с под-словарём:
// Пользователь -> Его фильм -> Его состояние в библиотеке
// Записи храним как указатели, т.к. будем обновлять их состояние на месте
type LibraryStore struct {
	mu   sync.RWMutex
	data map[domain.UserID]map[domain.ImdbID]*domain.LibraryEntry
}

// Создаём пустую библиотеку (инициализировать map обязательно, иначе обращение к nil-map упадёт при записи)
func NewLibraryStore() *LibraryStore {
	return &LibraryStore{
		data: make(map[domain.UserID]map[domain.ImdbID]*domain.LibraryEntry),
	}

}

// Create - создает для пользователя новую библиотеку, если ее еще нет и добавляет в нее фильм
func (s *LibraryStore) Create(e domain.LibraryEntry) error {
	// Блокируем создание библиотеки для предотвращения конкурентных записей
	s.mu.Lock()
	defer s.mu.Unlock()

	// Создаем подсловарь, если у пользователя его еще нет.
	if s.data[e.UserID] == nil {
		s.data[e.UserID] = make(map[domain.ImdbID]*domain.LibraryEntry)
	}

	id := e.MoviePreview.ID

	// Возвращаем ошибку, если такой фильм уже есть в библиотеке
	if _, ok := s.data[e.UserID][id]; ok {
		return domain.ErrAlreadyExists
	}

	stored := e
	s.data[e.UserID][id] = &stored

	return nil
}

// Get - получает конкретный запись с фильмом из коллекции пользователя
func (s *LibraryStore) Get(u domain.UserID, id domain.ImdbID) (domain.LibraryEntry, error) {
	// Блокируем чтение фильма пользователя, но не ограничиваем паралелльное чтение
	s.mu.RLock()
	defer s.mu.RUnlock()

	entries := s.data[u]
	if entries == nil {
		return domain.LibraryEntry{}, domain.ErrNotFound
	}

	entry, ok := entries[id]
	if !ok {
		return domain.LibraryEntry{}, domain.ErrNotFound
	}

	return *entry, nil
}

// List - получает список всех запсией из коллекции пользователя
func (s *LibraryStore) List(u domain.UserID) ([]domain.LibraryEntry, error) {
	// Блокируем чтение фильмов пользователя, но не ограничиваем паралелльное чтение
	s.mu.RLock()
	defer s.mu.RUnlock()

	entries := s.data[u]

	if entries == nil {
		return []domain.LibraryEntry{}, nil
	}

	result := make([]domain.LibraryEntry, 0, len(entries))
	for _, entry := range entries {
		result = append(result, *entry)
	}
	return result, nil
}

func (s *LibraryStore) Update(u domain.UserID, id domain.ImdbID, mutate func(*domain.LibraryEntry)) (domain.LibraryEntry, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Присваиваем коллекцию пользователя в переменную,
	// Если ее нет, то возвращаем ошибку
	entries := s.data[u]
	if entries == nil {
		return domain.LibraryEntry{}, domain.ErrNotFound
	}

	// Выбираем из коллекции фильма и присваиваем его в переменную
	// Если его нет, то возвращаем ошибку.
	entry, ok := entries[id]
	if !ok {
		return domain.LibraryEntry{}, domain.ErrNotFound
	}

	mutate(entry)

	return *entry, nil
}

func (s *LibraryStore) Delete(u domain.UserID, id domain.ImdbID) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	entries := s.data[u]

	if entries == nil {
		return domain.ErrNotFound
	}

	_, ok := entries[id]
	if !ok {
		return domain.ErrNotFound
	}

	delete(entries, id)

	return nil

}
