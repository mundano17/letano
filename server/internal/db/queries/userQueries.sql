-- name: CreateUsersTable :exec
CREATE TABLE IF NOT EXISTS USERS (
    user_id TEXT PRIMARY KEY,
    user_name TEXT,
    email TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    created_at TEXT DEFAULT CURRENT_TIMESTAMP,
    refresh_token_hash TEXT
);

-- name: CreateUser :one
INSERT INTO
    users (user_id, email, password_hash)
VALUES
    (?, ?, ?) RETURNING *;

-- name: UpdateRefreshTokenHash :exec
UPDATE
    users
SET
    refresh_token_hash = ?
WHERE
    user_id = ?;

-- name: GetRefreshTokenHashByUserID :one
SELECT
    refresh_token_hash
FROM
    users
WHERE
    user_id = ?;

-- name: SetRefreshTokenHashToNULL :exec
UPDATE
    users
SET
    refresh_token_hash = NULL
WHERE
    user_id = ?;

-- name: GetUserByEmail :one
SELECT
    *
FROM
    users
WHERE
    email = ?;

-- name: GetUserByID :one
SELECT
    *
FROM
    users
WHERE
    user_id = ?;

-- name: DeleteUserByID :execrows
DELETE FROM
    users
WHERE
    user_id = ?;

-- name: DeleteUserByEmail :exec
DELETE FROM
    users
WHERE
    email = ?;

-- name: ListUsers :many
SELECT
    *
FROM
    users
ORDER BY
    user_id;