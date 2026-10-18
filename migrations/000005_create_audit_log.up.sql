CREATE TABLE library_audit_log (
    id          BIGSERIAL PRIMARY KEY,
    event_type  TEXT NOT NULL,
    user_id     TEXT NOT NULL,
    imdb_id     TEXT NOT NULL,
    occurred_at TIMESTAMPTZ NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);