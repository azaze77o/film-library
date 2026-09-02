package domain

import "time"

type LibraryEntry struct {
	UserID       UserID
	MoviePreview MoviePreview
	IsFavorite   bool
	IsWatched    bool
	CreatedAt    time.Time
}
