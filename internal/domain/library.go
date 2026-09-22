package domain

import (
	"errors"
	"time"
)

type LibraryEntry struct {
	UserID       UserID
	MoviePreview MoviePreview
	IsFavorite   bool
	IsWatched    bool
	CreatedAt    time.Time
}

// LibraryPatch для обновления признаков фильма: просмотрен, отметка избранное
type LibraryPatch struct {
	Favourite *bool
	Watched   *bool
}

var (
	ErrNotFound      = errors.New("not found")
	ErrAlreadyExists = errors.New("library entry already exists")
	ErrInvalidInput  = errors.New("invalid input")
)
