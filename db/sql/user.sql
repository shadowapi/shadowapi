-- name: CreateUser :one
INSERT INTO "user" (
    uuid,
    email,
    password,
    first_name,
    last_name,
    is_enabled,
    is_admin,
    zitadel_subject,
    meta,
    created_at,
    updated_at
) VALUES (
             sqlc.arg('uuid')::uuid,
             sqlc.arg('email'),
             sqlc.arg('password'),
             sqlc.arg('first_name'),
             sqlc.arg('last_name'),
             sqlc.arg('is_enabled')::boolean,
            sqlc.arg('is_admin')::boolean,
            sqlc.arg('zitadel_subject'),
            sqlc.arg('meta'),
             NOW(),
             NULL
         ) RETURNING *;

-- name: GetUser :one
SELECT
    *
FROM "user"
WHERE uuid = sqlc.arg('uuid')::uuid;

-- name: UpdateUserNames :exec
UPDATE "user"
SET
    first_name = sqlc.arg('first_name'),
    last_name = sqlc.arg('last_name'),
    updated_at = NOW()
WHERE uuid = sqlc.arg('uuid')::uuid;

-- name: GetUserByZitadelSubject :one
SELECT
    *
FROM "user"
WHERE zitadel_subject = sqlc.arg('zitadel_subject');

-- name: ListUsers :many
SELECT
   *
FROM "user"
ORDER BY created_at DESC
LIMIT NULLIF(sqlc.arg('limit')::int, 0)
    OFFSET sqlc.arg('offset');

-- name: UpdateUser :exec
UPDATE "user"
SET
    email = sqlc.arg('email'),
    password = sqlc.arg('password'),
    first_name = sqlc.arg('first_name'),
    last_name = sqlc.arg('last_name'),
    is_enabled =  sqlc.arg('is_enabled')::boolean,
    is_admin = sqlc.arg('is_admin')::boolean,
    zitadel_subject = sqlc.arg('zitadel_subject'),
    meta = sqlc.arg('meta'),
    updated_at = NOW()
WHERE uuid = sqlc.arg('uuid')::uuid;

-- name: DeleteUser :exec
DELETE FROM "user"
WHERE uuid = sqlc.arg('uuid')::uuid;
