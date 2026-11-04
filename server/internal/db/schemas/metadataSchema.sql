CREATE TABLE IF NOT EXISTS METADATA (
    metadata_id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL REFERENCES USERS(user_id),
    file_name TEXT NOT NULL,
    file_hash TEXT NOT NULL,
    uploaded_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    mime_type TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_metadata_user_id ON METADATA(user_id);

CREATE TABLE IF NOT EXISTS METADATA_FIELDS (
    metadata_id TEXT NOT NULL REFERENCES metadata(metadata_id) ON DELETE CASCADE,
    field_name TEXT NOT NULL,
    field_value TEXT NOT NULL,
    PRIMARY KEY (metadata_id, field_name)
);