CREATE INDEX library_entries_user_created_at_idx
    ON library_entries (user_id, created_at DESC);