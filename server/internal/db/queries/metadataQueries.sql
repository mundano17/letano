-- name: InsertMetadata :one
INSERT INTO
    METADATA (
        metadata_id,
        user_id,
        file_name,
        file_hash,
        mime_type
    )
VALUES
    (?, ?, ?, ?, ?) RETURNING metadata_id;

-- name: InsertMetadataField :exec
INSERT INTO
    METADATA_FIELDS(metadata_id, field_name, field_value)
VALUES
    (?, ?, ?);