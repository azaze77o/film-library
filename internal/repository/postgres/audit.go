package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/project/omdbapp/internal/domain"
)

type AuditStore struct {
	db *sql.DB
}

func NewAuditStore(db *sql.DB) *AuditStore {
	return &AuditStore{db: db}
}

func (s *AuditStore) Insert(ctx context.Context, e domain.LibraryEvent) error {
	q := psql.Insert("library_audit_log").
		Columns("event_type", "user_id", "imdb_id", "occurred_at").
		Values(
			string(e.Type),
			string(e.UserID),
			string(e.ImdbID),
			e.OccurredAt,
		)

	query, args, err := q.ToSql()
	if err != nil {
		return fmt.Errorf("build insert query: %w", err)
	}

	if _, err := s.db.ExecContext(ctx, query, args...); err != nil {
		return fmt.Errorf("insert library audit log: %w", err)
	}

	return nil
}
