CREATE TABLE movies (
    imdb_id  TEXT PRIMARY KEY,
    title TEXT NOT NULL,
    year TEXT,
    kind TEXT,
    poster TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);