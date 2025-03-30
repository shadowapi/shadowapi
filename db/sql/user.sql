-- name: CreateUser :one
INSERT INTO "user" (
    uuid,
    email,
    password,
    first_name,
    last_name,
    is_enabled,
    is_admin,
    meta,
    created_at,
    updated_at
) VALUES (
             @uuid,
             @email,
             @password,
             @first_name,
             @last_name,
             @is_enabled,
             @is_admin,
             @meta,
             NOW(),
             NULL
         ) RETURNING *;

-- name: GetUser :one
SELECT *
FROM "user"
WHERE uuid = @uuid
LIMIT 1;

-- name: ListUsers :many
SELECT *
FROM "user"
ORDER BY created_at DESC
LIMIT CASE WHEN @limit_records::int = 0 THEN NULL ELSE @limit_records::int END
    OFFSET @offset_records::int;

-- name: UpdateUser :exec
UPDATE "user"
SET
    email = COALESCE(@email, email),
    password = COALESCE(@password, password),
    first_name = COALESCE(@first_name, first_name),
    last_name = COALESCE(@last_name, last_name),
    is_enabled = COALESCE(@is_enabled, is_enabled),
    is_admin = COALESCE(@is_admin, is_admin),
    meta = COALESCE(@meta, meta),
    updated_at = NOW()
WHERE uuid = @uuid;

-- name: DeleteUser :exec
DELETE FROM "user"
WHERE uuid = @uuid;
