package httpapi

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/project/omdbapp/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Объявялем структуру наших  mock_тестов
// Управлениие библиотекой фильмов
type mockLibraryService struct {
	mock.Mock
}

// Сервиса поиска
type mockSearchService struct {
	mock.Mock
}

// Для каждого сервиса управления библиотекрой создаем mock-функции, которые будут вызываться в самом тесте
func (m *mockLibraryService) Create(
	ctx context.Context,
	u domain.UserID,
	p domain.MoviePreview,
) (domain.LibraryEntry, error) {
	args := m.Called(ctx, u, p)
	var libraryEntry domain.LibraryEntry

	if value := args.Get(0); value != nil {
		libraryEntry = value.(domain.LibraryEntry)
	}

	return libraryEntry, args.Error(1)
}

func (m *mockLibraryService) Get(
	ctx context.Context,
	u domain.UserID,
	id domain.ImdbID,
) (domain.LibraryEntry, error) {
	args := m.Called(ctx, u, id)
	var libraryEntry domain.LibraryEntry

	if value := args.Get(0); value != nil {
		libraryEntry = value.(domain.LibraryEntry)
	}

	return libraryEntry, args.Error(1)
}

func (m *mockLibraryService) List(
	ctx context.Context,
	u domain.UserID,
) ([]domain.LibraryEntry, error) {
	args := m.Called(ctx, u)
	value := args.Get(0)

	if value == nil {
		return nil, args.Error(1)
	}

	entries := value.([]domain.LibraryEntry)
	return entries, args.Error(1)
}

func (m *mockLibraryService) Update(
	ctx context.Context,
	u domain.UserID,
	id domain.ImdbID,
	patch domain.LibraryPatch,
) (domain.LibraryEntry, error) {

	args := m.Called(ctx, u, id, patch)

	var libraryEntry domain.LibraryEntry
	if value := args.Get(0); value != nil {
		libraryEntry = value.(domain.LibraryEntry)
	}

	return libraryEntry, args.Error(1)
}

func (m *mockLibraryService) Delete(
	ctx context.Context,
	u domain.UserID,
	id domain.ImdbID,
) error {

	args := m.Called(ctx, u, id)

	return args.Error(0)
}

// Мок сервиса поиск фильмов
func (m *mockSearchService) Search(
	title string,
	page int,
) (domain.SearchResult, error) {
	args := m.Called(title, page)

	var searchResult domain.SearchResult

	if value := args.Get(0); value != nil {
		searchResult = value.(domain.SearchResult)
	}

	return searchResult, args.Error(1)
}

// Далее пишем сами тестовые сценарии: негативные и позитивные
// TestRouter_Create_Success проверяет успешное создание новой записи в библиотеке фильмов
func TestRouter_Create_Success(t *testing.T) {
	movie := domain.MoviePreview{
		ID:     "tt0108778",
		Title:  "Friends",
		Year:   "1994",
		Type:   "series",
		Poster: "http://example",
	}

	created := time.Date(2026, 9, 28, 12, 0, 0, 0, time.UTC)

	entry := domain.LibraryEntry{
		UserID:       defaultUserID,
		MoviePreview: movie,
		CreatedAt:    created,
	}

	// Инициализируем логгер, которые принимает записи и сразу их выбрасывает.
	// т.к. тесту нужен сам объект  *slog.Logger, потому что роутер его вызывает,
	// а текст этих вызовов в сценариях не проверяется.
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	lib := new(mockLibraryService)
	search := new(mockSearchService)

	lib.On("Create", mock.Anything, defaultUserID, movie).
		Return(entry, nil).
		Once()

	// Инициализируем объект запроса
	req := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/library",
		strings.NewReader(`{
					"imdb_id":"tt0108778",
					"title":"Friends",
					"year":"1994",
					"type":"series",
					"poster":"http://example"
					}`,
		),
	)

	// Cоздаёт объект, в который хендлер запишет ответ вместо отправки его по сети.
	rec := httptest.NewRecorder()

	NewRouter(lib, search, logger).ServeHTTP(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)
	assert.JSONEq(t, `{
						"UserID":"default",
						"MoviePreview":{
							"ID":"tt0108778",
							"Title":"Friends",
							"Year":"1994",
							"Type":"series",
							"Poster":"http://example"
							},
						"IsFavourite":false,
						"IsWatched":false,
						"CreatedAt":"2026-09-28T12:00:00Z"
						}`,
		rec.Body.String(),
	)
	lib.AssertExpectations(t)
}

// TestRouter_Create_BadJSON проверяет, что запрос на создание не выполнится из-за битого тела запроса
// В тесте сверяется код ошибки и ожидаемое сообщение с фактитческим
func TestRouter_Create_BadJSON(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	lib := new(mockLibraryService)
	search := new(mockSearchService)

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/library",
		strings.NewReader(`{`),
	)
	rec := httptest.NewRecorder()

	NewRouter(lib, search, logger).ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.JSONEq(t, `{"error":{"code":"INVALID_INPUT","message":"invalid json"}}`, rec.Body.String())

	lib.AssertNotCalled(t, "Create", mock.Anything, mock.Anything, mock.Anything)
}

// TestRouter_Create_AlreadyExists проверяет, что запрос не создает объект, потому что он уже существует
func TestRouter_Create_AlreadyExists(t *testing.T) {
	movie := domain.MoviePreview{
		ID:     "tt0108778",
		Title:  "Friends",
		Year:   "1994",
		Type:   "series",
		Poster: "http://example",
	}

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	lib := new(mockLibraryService)
	search := new(mockSearchService)

	lib.On("Create", mock.Anything, defaultUserID, movie).
		Return(domain.LibraryEntry{}, domain.ErrAlreadyExists).
		Once()

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/library",
		strings.NewReader(`{
					"imdb_id":"tt0108778",
					"title":"Friends",
					"year":"1994",
					"type":"series",
					"poster":"http://example"
					}`,
		),
	)

	rec := httptest.NewRecorder()
	NewRouter(lib, search, logger).ServeHTTP(rec, req)

	assert.Equal(t, http.StatusConflict, rec.Code)
	assert.JSONEq(t, `{"error":{"code":"ALREADY_EXISTS","message":"resource already exists"}}`, rec.Body.String())

	lib.AssertExpectations(t)
}

// TestRouter_Delete_Success проверяет успешное удаление записи из библиотеки
func TestRouter_Delete_Success(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	movieID := domain.ImdbID("tt0108778")

	lib := new(mockLibraryService)
	search := new(mockSearchService)

	lib.On("Delete", mock.Anything, defaultUserID, movieID).
		Return(nil).
		Once()

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/library/tt0108778", nil)
	rec := httptest.NewRecorder()

	NewRouter(lib, search, logger).ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNoContent, rec.Code)
	assert.Empty(t, rec.Body.String())
	assert.Empty(t, rec.Header().Get("Content-Type"))
	lib.AssertExpectations(t)
}

// TestRouter_MethodNotAllowed проверяет, что router не вызыыает сервис по незарегистрированному методу
func TestRouter_MethodNotAllowed(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	lib := new(mockLibraryService)
	search := new(mockSearchService)

	req := httptest.NewRequest(http.MethodPut, "/api/v1/library/tt0108778", nil)
	rec := httptest.NewRecorder()

	NewRouter(lib, search, logger).ServeHTTP(rec, req)

	assert.Equal(t, http.StatusMethodNotAllowed, rec.Code)
	assert.Equal(t, "text/plain; charset=utf-8", rec.Header().Get("Content-Type"))
	assert.Equal(t, "Method Not Allowed\n", rec.Body.String())

	// Дополнительно проверяем, что у указанного пути доступны эти методы.
	// Это наша перестрахивка, т.к. именно у этих методов используется путь из теста.
	assert.Contains(t, rec.Header().Get("Allow"), "GET")
	assert.Contains(t, rec.Header().Get("Allow"), "PATCH")
	assert.Contains(t, rec.Header().Get("Allow"), "DELETE")
}

// TestRouter_Search_MissingQuery проверяет, что router не вызыыает сервис поиска,
// потому что в запросе нет обязательных параметров поиска
func TestRouter_Search_MissingQuery(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	lib := new(mockLibraryService)
	search := new(mockSearchService)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/search", nil)
	rec := httptest.NewRecorder()

	NewRouter(lib, search, logger).ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.JSONEq(t, `{"error":{"code":"INVALID_INPUT","message":"query q is required"}}`, rec.Body.String())
	search.AssertNotCalled(t, "Search", mock.Anything, mock.Anything)
}

// TestRouter_Search_BadPage проверяет, что router вызывает сервис поиска,
// и хендлер при парамтере page с нецелочисленным значением подставляет занчение 1
func TestRouter_Search_BadPage(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	lib := new(mockLibraryService)
	search := new(mockSearchService)

	search.On("Search", "Friends", 1).
		Return(domain.SearchResult{}, nil).
		Once()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/search?q=Friends&page=abc", nil)
	rec := httptest.NewRecorder()

	NewRouter(lib, search, logger).ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	search.AssertExpectations(t)
}
