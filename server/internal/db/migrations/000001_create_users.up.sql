CREATE TABLE IF NOT EXISTS USERS (
    user_id TEXT PRIMARY KEY,
    user_name TEXT,
    email TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    created_at TEXT DEFAULT CURRENT_TIMESTAMP,
    refresh_token_hash TEXT
);
