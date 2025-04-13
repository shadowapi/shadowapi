-- name: CreateFile :one
INSERT INTO "file" (
    uuid,
    storage_type,
    storage_uuid,
    name,
    mime_type,
    size,
    data,
    path,
    is_raw,
    raw_headers,
    has_raw_email,
    is_inline,
    created_at,
    updated_at
) VALUES (
             sqlc.arg('uuid')::uuid,
             sqlc.arg('storage_type'),
             sqlc.arg('storage_uuid')::uuid,
            sqlc.arg('name'),
             sqlc.arg('mime_type'),
             sqlc.arg('size'),
    sqlc.arg('data'),
    sqlc.arg('path'),
    sqlc.arg('is_raw'),
    sqlc.arg('raw_headers'),
    sqlc.arg('has_raw_email'),
    sqlc.arg('is_inline'),
          NOW(),
             NOW()
         ) RETURNING *;

-- name: GetFile :one
SELECT
    sqlc.embed(file)
FROM "file"
WHERE uuid = sqlc.arg('uuid')::uuid;

-- name: ListFiles :many
SELECT
    sqlc.embed(file)
FROM "file"
ORDER BY created_at DESC
LIMIT NULLIF(sqlc.arg('limit')::int, 0)
    OFFSET sqlc.arg('offset');

-- name: GetFiles :many
WITH filtered_files AS (
    SELECT f.*
    FROM "file" f
    WHERE
      -- Filter by storage_type if not empty
        (NULLIF(sqlc.arg('storage_type'), '') IS NULL OR f.storage_type = sqlc.arg('storage_type'))

      -- Filter by storage_uuid if not empty
      AND (NULLIF(sqlc.arg('storage_uuid'), '') IS NULL OR f.storage_uuid = sqlc.arg('storage_uuid')::uuid)

      -- Filter by partial name if not empty (ILIKE for case-insensitive search)
      AND (NULLIF(sqlc.arg('name'), '') IS NULL OR f.name ILIKE '%' || sqlc.arg('name') || '%')

      -- Filter by mime_type if not empty (ILIKE for partial match)
      AND (NULLIF(sqlc.arg('mime_type'), '') IS NULL OR f.mime_type ILIKE '%' || sqlc.arg('mime_type') || '%')

      -- Filter by minimum size if size_min > 0
      AND (COALESCE(sqlc.arg('size_min'), 0) = 0 OR f.size >= sqlc.arg('size_min'))

      -- Filter by maximum size if size_max > 0
      AND (COALESCE(sqlc.arg('size_max'), 0) = 0 OR f.size <= sqlc.arg('size_max'))

      -- Filter by is_raw if not -1 (using integer approach, e.g. -1=ignore, 0=false, 1=true)
      AND (
        NULLIF(sqlc.arg('is_raw')::int, -1) IS NULL
            OR f.is_raw = (sqlc.arg('is_raw')::int)::boolean
        )
)
SELECT
    f.*,
    -- total_count of all matching rows (ignoring limit/offset)
    (SELECT COUNT(*) FROM filtered_files) AS total_count
FROM filtered_files f
ORDER BY
    CASE WHEN sqlc.arg('order_by') = 'created_at' AND sqlc.arg('order_direction') = 'asc'  THEN f.created_at END ASC,
    CASE WHEN sqlc.arg('order_by') = 'created_at' AND sqlc.arg('order_direction') = 'desc' THEN f.created_at END DESC,

    CASE WHEN sqlc.arg('order_by') = 'updated_at' AND sqlc.arg('order_direction') = 'asc'  THEN f.updated_at END ASC,
    CASE WHEN sqlc.arg('order_by') = 'updated_at' AND sqlc.arg('order_direction') = 'desc' THEN f.updated_at END DESC,

    CASE WHEN sqlc.arg('order_by') = 'size' AND sqlc.arg('order_direction') = 'asc'       THEN f.size END ASC,
    CASE WHEN sqlc.arg('order_by') = 'size' AND sqlc.arg('order_direction') = 'desc'      THEN f.size END DESC,

    -- fallback if no valid order_by specified
    f.created_at DESC

LIMIT NULLIF(sqlc.arg('limit')::int, 0)
    OFFSET sqlc.arg('offset')::int;



-- name: UpdateFile :exec
UPDATE "file"
SET
    storage_type  = sqlc.arg('storage_type'),
    storage_uuid  = sqlc.arg('storage_uuid')::uuid,
    name          = sqlc.arg('name'),
    mime_type     = sqlc.arg('mime_type'),
    size          = sqlc.arg('size'),
    data          = sqlc.arg('data'),
    path          = sqlc.arg('path'),
    is_raw        = sqlc.arg('is_raw'),
    raw_headers    = sqlc.arg('raw_headers'),
    has_raw_email  = sqlc.arg('has_raw_email'),
    is_inline      = sqlc.arg('is_inline'),
    updated_at    = NOW()
WHERE uuid = sqlc.arg('uuid')::uuid;

-- name: DeleteFile :exec
DELETE FROM "file"
WHERE uuid = sqlc.arg('uuid')::uuid;
