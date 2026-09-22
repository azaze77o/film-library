CREATE TABLE library_entries (
    user_id         TEXT NOT NULL,
    imdb_id         TEXT NOT NULL,
    is_favourite    BOOLEAN NOT NULL DEFAULT FALSE,
    is_watched      BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT library_entries_user_id_fkey
        FOREIGN KEY (user_id)
        REFERENCES users(id)
        ON DELETE CASCADE,

    CONSTRAINT library_entries_imdb_id_fkey
        FOREIGN KEY (imdb_id)
        REFERENCES movies(imdb_id),

    CONSTRAINT library_entries_pkey
        PRIMARY KEY (user_id, imdb_id)
);