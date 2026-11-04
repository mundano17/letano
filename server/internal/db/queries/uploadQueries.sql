-- name: InsertUpload :one
INSERT INTO
    UPLOAD(
        upload_id,
        file_hash,
        file_name,
        file_size,
        uploaded_chunks,
        mime_type,
        upload_status,
        FIELDS,
        user_id
    )
VALUES
    (?, ?, ?, ?, ?, ?, ?, ?, ?) RETURNING upload_id;

-- name: GetUploadedChunksByUploadID :one
SELECT
    uploaded_chunks
FROM
    UPLOAD
WHERE
    upload_id = ?
    AND user_id = ?;

-- name: GetUploadDetailsByUploadID :one
SELECT
    *
FROM
    UPLOAD
WHERE
    upload_id = ?
    AND user_id = ?;

-- name: GetHashByUploadID :one
SELECT
    file_hash
FROM
    UPLOAD
WHERE
    upload_id = ?
    AND user_id = ?;

-- name: UpdateUploadedChunksByID :exec
UPDATE
    UPLOAD
SET
    uploaded_chunks = ?
WHERE
    upload_id = ?
    AND user_id = ?;

-- name: GetStatusByUploadID :one
SELECT
    upload_status
FROM
    upload
WHERE
    upload_id = ?
    AND user_id = ?;

-- name: SetStatusByUploadID :exec
UPDATE
    UPLOAD
SET
    upload_status = ?
WHERE
    upload_id = ?
    AND user_id = ?;