package service

import (
	"context"
	"testing"
	"time"

	"github.com/project/omdbapp/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// Объявлем стуктуру нашего mock-теста
type mockLibraryStore struct {
	mock.Mock
}

// Для каждого метода репозитория управления библиотекрой создаем mock-функции, которые будут вызываться в самом тесте
func (m *mockLibraryStore) Create(ctx context.Context, e domain.LibraryEntry) error {
	args := m.Called(ctx, e)
	return args.Error(0)
}

func (m *mockLibraryStore) Get(ctx context.Context, u domain.UserID, id domain.ImdbID) (domain.LibraryEntry, error) {
	args := m.Called(ctx, u, id)

	var libraryEntry domain.LibraryEntry

	if value := args.Get(0); value != nil {
		libraryEntry = value.(domain.LibraryEntry)
	}

	return libraryEntry, args.Error(1)
}

func (m *mockLibraryStore) List(ctx context.Context, u domain.UserID) ([]domain.LibraryEntry, error) {
	args := m.Called(ctx, u)

	value := args.Get(0)
	if value == nil {
		return nil, args.Error(1)
	}

	entries := value.([]domain.LibraryEntry)
	return entries, args.Error(1)
}

func (m *mockLibraryStore) Update(ctx context.Context, u domain.UserID, id domain.ImdbID, patch domain.LibraryPatch) (domain.LibraryEntry, error) {
	args := m.Called(ctx, u, id, patch)

	var libraryEntry domain.LibraryEntry

	if value := args.Get(0); value != nil {
		libraryEntry = value.(domain.LibraryEntry)
	}

	return libraryEntry, args.Error(1)
}

func (m *mockLibraryStore) Delete(ctx context.Context, u domain.UserID, id domain.ImdbID) error {
	args := m.Called(ctx, u, id)

	return args.Error(0)
}

// Выполняем сами тесты для каждого из сценариев
// Для метода создания фильма

// TestLibraryService_Create_Success - успешное создание записи в библиотеке фильмов
func TestLibraryService_Create_Success(t *testing.T) {
	// Объявляем константы
	fixedTime := time.Date(2026, 9, 28, 12, 0, 0, 0, time.UTC)
	movie := domain.MoviePreview{ID: "tt0108778", Title: "Friends"}
	userID := domain.UserID("default")

	// Объявлем, что мы ждем по результатам теста
	want := domain.LibraryEntry{
		UserID:       userID,
		MoviePreview: movie,
		CreatedAt:    fixedTime,
	}

	// Инициализируем в памяти новый объект mock-функций (возвращзает указатель)
	store := new(mockLibraryStore)
	store.On("Create", mock.Anything, want).
		Return(nil).
		Once()

	// Инициализируем в памяти новый сервис управления библиотекой
	// Передаем нам mock-репозиторий
	svc := NewLibraryService(store)
	svc.now = func() time.Time { return fixedTime }

	got, err := svc.Create(context.Background(), userID, movie)

	require.NoError(t, err)
	assert.Equal(t, want, got)

	store.AssertExpectations(t)
}

// TestLibraryService_Create_InvalidInput - ошибка создания записи в библиотеке фильмов. Не валидный ID.
func TestLibraryService_Create_InvalidInput(t *testing.T) {
	// Допускаем ошибку: ID имеет буквы
	movie := domain.MoviePreview{ID: "tt010877f", Title: "Friends"}
	userID := domain.UserID("default")

	store := new(mockLibraryStore)
	svc := NewLibraryService(store)

	_, err := svc.Create(context.Background(), userID, movie)

	require.ErrorIs(t, err, domain.ErrInvalidInput)
	store.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
}

// ErrAlreadyExists - ошибка создания записи в библиотеке фильмов. Такой фильм уже есть
func TestLibraryService_Create_ErrAlreadyExists(t *testing.T) {
	movie := domain.MoviePreview{ID: "tt0108778", Title: "Friends"}
	userID := domain.UserID("default")

	store := new(mockLibraryStore)
	store.On("Create", mock.Anything, mock.Anything).
		Return(domain.ErrAlreadyExists).
		Once()

	svc := NewLibraryService(store)

	_, err := svc.Create(context.Background(), userID, movie)

	require.ErrorIs(t, err, domain.ErrAlreadyExists)

	store.AssertExpectations(t)
}

// TestLibraryService_Update__Success - успешное обновление записи в библиотеке фильмов
func TestLibraryService_Update_Success(t *testing.T) {
	userID := domain.UserID("default")
	movieID := domain.ImdbID("tt0108778")
	// LibraryPatch ожидает тип указатель, поэтому создаем перменную favourite и передаем ее адрес в структуре
	favourite := true
	patch := domain.LibraryPatch{Favourite: &favourite}

	// Объявлем, что мы ждем по результатам теста
	want := domain.LibraryEntry{
		UserID: userID,
		MoviePreview: domain.MoviePreview{
			ID:    movieID,
			Title: "Friends",
		},
		IsFavourite: true,
	}

	// Инициализируем в памяти новый объект mock-функций (возвращзает указатель)
	store := new(mockLibraryStore)
	store.On("Update", mock.Anything, userID, movieID, patch).
		Return(want, nil).
		Once()

	// Инициализируем в памяти новый сервис управления библиотекой
	// Передаем нам mock-репозиторий
	svc := NewLibraryService(store)

	got, err := svc.Update(context.Background(), userID, movieID, patch)

	require.NoError(t, err)
	assert.Equal(t, want, got)
	store.AssertExpectations(t)
}

// TestLibraryService_Update__InvalidInput - тест с проверкой ошибки
// при запросе без IsFavourite или IsWatched
func TestLibraryService_Update_InvalidInput(t *testing.T) {
	userID := domain.UserID("default")
	movieID := domain.ImdbID("tt0108778")
	patch := domain.LibraryPatch{}

	// Инициализируем в памяти новый объект mock-функций (возвращзает указатель)
	store := new(mockLibraryStore)

	// Инициализируем в памяти новый сервис управления библиотекой
	// Передаем нам mock-репозиторий
	svc := NewLibraryService(store)

	_, err := svc.Update(context.Background(), userID, movieID, patch)

	require.ErrorIs(t, err, domain.ErrInvalidInput)
	store.AssertNotCalled(t, "Update", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

// TestLibraryService_Get_Success - успешный запрос одной записи из библиотеки фильмов
func TestLibraryService_Get_Success(t *testing.T) {
	userID := domain.UserID("default")
	movieID := domain.ImdbID("tt0108778")

	want := domain.LibraryEntry{
		UserID: userID,
		MoviePreview: domain.MoviePreview{
			ID:    movieID,
			Title: "Friends",
		},
		IsFavourite: true,
		IsWatched:   false,
	}

	store := new(mockLibraryStore)
	store.On("Get", mock.Anything, userID, movieID).
		Return(want, nil).
		Once()

	svc := NewLibraryService(store)

	got, err := svc.Get(context.Background(), userID, movieID)

	require.NoError(t, err)
	assert.Equal(t, want, got)
	store.AssertExpectations(t)
}

// TestLibraryService_List_Success - успешный результат получения списка записей из библиотеки фильмов
func TestLibraryService_List_Success(t *testing.T) {
	userID := domain.UserID("default")
	want := []domain.LibraryEntry{
		{
			UserID: userID,
			MoviePreview: domain.MoviePreview{
				ID:    "tt0108778",
				Title: "Friends",
			},
			IsFavourite: true,
		},
	}

	store := new(mockLibraryStore)
	store.On("List", mock.Anything, userID).
		Return(want, nil).
		Once()

	svc := NewLibraryService(store)

	got, err := svc.List(context.Background(), userID)

	require.NoError(t, err)
	assert.Equal(t, want, got)
	store.AssertExpectations(t)
}

// TestLibraryService_Delete_Success - успешно удаляет фильм из библиотеки
func TestLibraryService_Delete_Success(t *testing.T) {
	userID := domain.UserID("default")
	movieID := domain.ImdbID("tt0108778")

	store := new(mockLibraryStore)
	store.On("Delete", mock.Anything, userID, movieID).
		Return(nil).
		Once()

	svc := NewLibraryService(store)

	err := svc.Delete(context.Background(), userID, movieID)

	require.NoError(t, err)
	store.AssertExpectations(t)
}
