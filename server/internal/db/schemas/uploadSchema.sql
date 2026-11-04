CREATE TABLE IF NOT EXISTS UPLOAD (
    user_id TEXT NOT NULL,
    upload_id TEXT PRIMARY KEY,
    file_hash TEXT NOT NULL,
    file_name TEXT NOT NULL,
    file_size INTEGER NOT NULL,
    uploaded_chunks BLOB NOT NULL,
    mime_type TEXT NOT NULL,
    upload_status TEXT NOT NULL,
    FIELDS TEXT,
    created_at TEXT DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES USERS(user_id)
);

CREATE INDEX IF NOT EXISTS idx_upload_user_id ON upload(user_id);