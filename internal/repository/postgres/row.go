package postgres

import (
	"database/sql"
	"time"

	"github.com/project/omdbapp/internal/domain"
)

type libraryRow struct {
	UserID      string
	ImdbID      string
	Title       string
	Year        sql.NullString
	Kind        sql.NullString
	Poster      sql.NullString
	IsFavourite bool
	IsWatched   bool
	CreatedAt   time.Time
}

func (r libraryRow) toDomain() domain.LibraryEntry {
	// Обрабатываем null при переводе типов из postgres в domain
	year := ""
	if r.Year.Valid {
		year = r.Year.String
	}

	kind := ""
	if r.Kind.Valid {
		kind = r.Kind.String
	}

	poster := ""
	if r.Poster.Valid {
		poster = r.Poster.String
	}

	return domain.LibraryEntry{
		UserID: domain.UserID(r.UserID),
		MoviePreview: domain.MoviePreview{
			ID:     domain.ImdbID(r.ImdbID),
			Title:  r.Title,
			Year:   year,
			Type:   kind,
			Poster: poster,
		},
		IsFavorite: r.IsFavourite,
		IsWatched:  r.IsWatched,
		CreatedAt:  r.CreatedAt,
	}
}
