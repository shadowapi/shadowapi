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
             @uuid,
             @storage_type,
             @storage_uuid,
             @name,
             @mime_type,
             @size,
             NOW(),
             NULL
         ) RETURNING *;

-- name: GetFile :one
SELECT
    *
FROM "file"
WHERE
    uuid = @uuid
LIMIT 1;

-- name: ListFiles :many
SELECT
    *
FROM "file"
ORDER BY created_at DESC
LIMIT CASE WHEN @limit_records::int = 0 THEN NULL ELSE @limit_records::int END
    OFFSET @offset_records::int;

-- name: UpdateFile :exec
UPDATE "file"
SET
    storage_type = COALESCE(@storage_type, storage_type),
    storage_uuid = COALESCE(@storage_uuid, storage_uuid),
    name         = COALESCE(@name, name),
    mime_type    = COALESCE(@mime_type, mime_type),
    size         = COALESCE(@size, size),
    updated_at   = NOW()
WHERE uuid = @uuid;

-- name: DeleteFile :exec
DELETE FROM "file"
WHERE uuid = @uuid;