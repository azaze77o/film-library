package postgres

import (
	"database/sql"
	"testing"
	"time"

	"github.com/project/omdbapp/internal/domain"
	"github.com/stretchr/testify/require"
)

func TestLibraryRow_ToDomain(t *testing.T) {
	// Объявлем перменную created, чтобы во всех случаях
	// фактическая временная отметка сходилась с ожидаемой
	// Иначе тесты всегда будут красными
	created := time.Date(2026, 9, 28, 12, 0, 0, 0, time.UTC)
	tests := []struct {
		name string
		row  libraryRow
		want domain.LibraryEntry
	}{
		{
			name: "null strings become empty",
			row: libraryRow{
				UserID:    "default",
				ImdbID:    "tt0108778",
				Title:     "Friends",
				CreatedAt: created,
			},
			want: domain.LibraryEntry{
				UserID: "default",
				MoviePreview: domain.MoviePreview{
					ID:     "tt0108778",
					Title:  "Friends",
					Year:   "",
					Type:   "",
					Poster: "",
				},
				CreatedAt: created,
			},
		},
		{
			name: "empty strings become empty",
			row: libraryRow{
				UserID:    "default",
				ImdbID:    "tt0108778",
				Title:     "Friends",
				Year:      sql.NullString{String: "", Valid: true},
				Kind:      sql.NullString{String: "", Valid: true},
				Poster:    sql.NullString{String: "", Valid: true},
				CreatedAt: created,
			},
			want: domain.LibraryEntry{
				UserID: "default",
				MoviePreview: domain.MoviePreview{
					ID:     "tt0108778",
					Title:  "Friends",
					Year:   "",
					Type:   "",
					Poster: "",
				},
				CreatedAt: created,
			},
		},
		{
			name: "field correspondence",
			row: libraryRow{
				UserID:      "default",
				ImdbID:      "tt0108778",
				Title:       "Friends",
				Year:        sql.NullString{String: "1994–2004", Valid: true},
				Kind:        sql.NullString{String: "series", Valid: true},
				Poster:      sql.NullString{String: "example", Valid: true},
				IsFavourite: true,
				IsWatched:   false,
				CreatedAt:   created,
			},
			want: domain.LibraryEntry{
				UserID: "default",
				MoviePreview: domain.MoviePreview{
					ID:     "tt0108778",
					Title:  "Friends",
					Year:   "1994–2004",
					Type:   "series",
					Poster: "example",
				},
				IsFavourite: true,
				IsWatched:   false,
				CreatedAt:   created,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			libraryEntry := tt.row.toDomain()

			require.Equal(t, tt.want, libraryEntry)
		})
	}
}
