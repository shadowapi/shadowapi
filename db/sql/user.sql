-- name: GetUser :one
SELECT * FROM "user"
WHERE
	(@uuid::text <> '' AND "uuid"::text = @uuid::text) OR
	(@email::text <> '' AND email = @email::text)
LIMIT 1;

-- name: ListUsers :many
SELECT * FROM "user"
ORDER BY created_at DESC
LIMIT CASE WHEN @limit_records::int = 0 THEN NULL ELSE @limit_records::int END
OFFSET @offset_records::int; 

-- name: CreateUser :one
INSERT INTO "user" (
	"uuid", email, password, first_name, last_name, is_enabled, is_admin
) VALUES(
	@uuid, @email, @password, @first_name, @last_name, @is_enabled, @is_admin
) RETURNING *;

-- name: UpdateUser :exec
UPDATE "user" SET
	email = @email,
	password = COALESCE(@password, password),
	first_name = @first_name,
	last_name = @last_name,
	is_enabled = @is_enabled,
	is_admin = @is_admin
WHERE "uuid" = @uuid;

-- name: DeleteUser :exec
DELETE FROM "user"
WHERE "uuid" = @uuid;
