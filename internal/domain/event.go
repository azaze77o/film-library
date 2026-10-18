package domain

import "time"

type EventType string

const EventLibraryEntryCreated EventType = "library_entry.created"

type LibraryEvent struct {
	Type       EventType
	UserID     UserID
	ImdbID     ImdbID
	OccurredAt time.Time
}
