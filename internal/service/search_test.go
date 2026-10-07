package service

import (
	"testing"

	"github.com/project/omdbapp/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockSearch struct {
	mock.Mock
}

func (m *mockSearch) Search(title string, page int) (domain.SearchResult, error) {
	args := m.Called(title, page)

	var searchResult domain.SearchResult

	if value := args.Get(0); value != nil {
		searchResult = value.(domain.SearchResult)
	}

	return searchResult, args.Error(1)
}

// TestSearchService_Search_Success - возвращает успешный результат поиска фильмов по заголовку
func TestSearchService_Search_Success(t *testing.T) {
	title := "Friends"
	page := 1
	movies := []domain.MoviePreview{{
		ID:     "tt0108778",
		Title:  "Friends",
		Year:   "1994–2004",
		Type:   "series",
		Poster: "example",
	}, {
		ID:     "tt33735025",
		Title:  "The Friends",
		Year:   "1919967",
		Type:   "movie",
		Poster: "example",
	},
	}

	want := domain.SearchResult{
		Movies:       movies,
		TotalResults: 2,
	}

	omdb := new(mockSearch)
	omdb.On("Search", title, page).
		Return(want, nil).
		Once()

	svc := NewSearchService(omdb)
	got, err := svc.Search(title, page)

	require.NoError(t, err)
	assert.Equal(t, want, got)
	omdb.AssertExpectations(t)
}

// TestSearchService_Search_InvalidInput - тест на негативный сценарий запроса, в котором пустой заголовок
func TestSearchService_Search_InvalidInput(t *testing.T) {
	title := "     "
	page := 1

	omdb := new(mockSearch)
	svc := NewSearchService(omdb)
	_, err := svc.Search(title, page)

	require.ErrorIs(t, err, domain.ErrInvalidInput)

	omdb.AssertNotCalled(t, "Search", mock.Anything, mock.Anything)
}
