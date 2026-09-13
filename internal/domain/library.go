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

var (
	ErrNotFound      = errors.New("not found")
	ErrAlreadyExists = errors.New("library entry alredy exists")
	ErrInvalidInput  = errors.New("invalid input")
)
