package service

import (
	"testing"

	"github.com/project/omdbapp/internal/domain"
	"github.com/stretchr/testify/require"
)

// Тест валидации фильма
func TestValidateMovie(t *testing.T) {
	// Создаем слайс структур, где каждая структура - это тестовый сценарий
	tests := []struct {
		name    string
		movie   domain.MoviePreview
		wantErr error
	}{
		{name: "ok", movie: domain.MoviePreview{ID: "tt0108778", Title: "Friends"}},
		{name: "emty id", movie: domain.MoviePreview{Title: "Friends"}, wantErr: domain.ErrInvalidInput},
		{name: "id without numbers", movie: domain.MoviePreview{ID: "tt", Title: "Friends"}, wantErr: domain.ErrInvalidInput},
		{name: "ID with letters", movie: domain.MoviePreview{ID: "tt3456ab", Title: "Friends"}, wantErr: domain.ErrInvalidInput},
		{name: "empty title", movie: domain.MoviePreview{ID: "tt0108778"}, wantErr: domain.ErrInvalidInput},
		{name: "title empty string", movie: domain.MoviePreview{ID: "tt0108778", Title: "   "}, wantErr: domain.ErrInvalidInput},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateMovie(tt.movie)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
		})
	}
}
