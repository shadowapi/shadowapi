-- name: CreateFile :one
INSERT INTO "file" (
    uuid,
    storage_type,
    storage_uuid,
    name,
    mime_type,
    size,
    created_at,
    updated_at
) VALUES (
             sqlc.arg('uuid')::uuid,
             sqlc.arg('storage_type'),
             sqlc.arg('storage_uuid')::uuid,
            sqlc.arg('name'),
             sqlc.arg('mime_type'),
             sqlc.arg('size'),
          NOW(),
             NOW()
         ) RETURNING *;

-- name: GetFile :one
SELECT
    *
FROM "file"
WHERE uuid = sqlc.arg('uuid')::uuid;

-- name: ListFiles :many
SELECT
    *
FROM "file"
ORDER BY created_at DESC
LIMIT NULLIF(sqlc.arg('limit')::int, 0)
    OFFSET sqlc.arg('offset');

-- name: UpdateFile :exec
UPDATE "file"
SET
    storage_type = sqlc.arg('storage_type'),
    storage_uuid = sqlc.arg('storage_uuid')::uuid,
    name         = sqlc.arg('name'),
    mime_type    = sqlc.arg('mime_type'),
    size         = sqlc.arg('size'),
    updated_at = NOW()
WHERE uuid = sqlc.arg('uuid')::uuid;

-- name: DeleteFile :exec
DELETE FROM "file"
WHERE uuid = @uuid;