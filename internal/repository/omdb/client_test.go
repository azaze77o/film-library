package omdb

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/project/omdbapp/internal/domain"
	"github.com/stretchr/testify/require"
)

func newTestClient(t *testing.T, h http.HandlerFunc) *Client {
	t.Helper()

	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)

	c := NewClient(srv.URL, "test-key")
	return c
}

func TestClient_Search_Success(t *testing.T) {
	var query url.Values

	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		query = r.URL.Query()

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		_, _ = w.Write([]byte(`{
		"Search":[
		{"Title":
			"Friends",
			"Year":"1994",
			"imdbID":"tt0108778",
			"Type":"series",
			"Poster":"http://example"
		}
			],
		"totalResults":"1",
		"Response":"True"}`))
	})

	got, err := c.Search("Friends", 1)

	require.NoError(t, err)
	require.Equal(t, domain.SearchResult{
		Movies: []domain.MoviePreview{{
			ID:     "tt0108778",
			Title:  "Friends",
			Year:   "1994",
			Type:   "series",
			Poster: "http://example"}},
		TotalResults: 1,
	}, got)
	require.Equal(t, "test-key", query.Get("apiKey"))
	require.Equal(t, "Friends", query.Get("s"))
	require.Equal(t, "movie", query.Get("type"))
	require.Equal(t, "1", query.Get("page"))
}

func TestClient_Search_HTTPStatus(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)

	})

	_, err := c.Search("Friends", 1)

	require.Error(t, err)
	require.Contains(t, err.Error(), "500")
}

func TestClient_Search_PageZero(t *testing.T) {
	var query url.Values

	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		query = r.URL.Query()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		_, _ = w.Write([]byte(`{
		"Search":[
		{"Title":
			"Friends",
			"Year":"1994",
			"imdbID":"tt0108778",
			"Type":"series",
			"Poster":"http://example"
		}
			],
		"totalResults":"1",
		"Response":"True"}`))
	})

	got, err := c.Search("Friends", 0)

	require.NoError(t, err)
	require.Equal(t, domain.SearchResult{
		Movies: []domain.MoviePreview{{
			ID:     "tt0108778",
			Title:  "Friends",
			Year:   "1994",
			Type:   "series",
			Poster: "http://example"}},
		TotalResults: 1,
	}, got)

	require.Equal(t, "1", query.Get("page"))
}

func TestClient_Search_BadJSON(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{`))
	})

	_, err := c.Search("Friends", 1)

	require.Error(t, err)
	require.ErrorIs(t, err, io.ErrUnexpectedEOF)
}

func TestClient_Search_NotFound(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		_, _ = w.Write([]byte(`{"Response":"False","Error":"Movie not found!"}`))
	})

	_, err := c.Search("no-such-movie", 1)

	require.Error(t, err)
	require.Contains(t, err.Error(), "Movie not found!")
	require.NotErrorIs(t, err, domain.ErrNotFound)
}
