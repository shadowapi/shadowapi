-- name: GetDatasource :one
SELECT
	sqlc.embed(datasource),
	sqlc.embed(datasource_email)
FROM datasource
LEFT JOIN datasource_email ON datasource_email.datasource_uuid = datasource.uuid
LEFT JOIN oauth2_client ON oauth2_client.id = datasource.oauth2_client_id
WHERE datasource.uuid = @uuid
LIMIT 1;

-- name: ListDatasource :many
SELECT
	sqlc.embed(datasource),
	sqlc.embed(datasource_email)
FROM datasource
LEFT JOIN datasource_email ON datasource_email.datasource_uuid = datasource.uuid
LEFT JOIN oauth2_client ON oauth2_client.id = datasource.oauth2_client_id
ORDER BY datasource.created_at DESC
LIMIT CASE WHEN @limit_records::int = 0 THEN NULL ELSE @limit_records::int END
OFFSET @offset_records::int; 

-- name: CreateDatasource :one
INSERT INTO datasource (
	uuid, name, type, is_enabled, user_uuid
) VALUES (
	@uuid, @name, @type, @is_enabled, @user_uuid
) RETURNING *;

-- name: UpdateDatasource :exec
UPDATE datasource SET
	name              = @name,
  user_uuid         = @user_uuid,
  is_enabled        = @is_enabled,
  oauth2_client_id  = @oauth2_client_id,
  oauth2_token_uuid = @oauth2_token_uuid
WHERE uuid = @uuid
RETURNING *;

-- name: LinkDatasourceWithToken :exec
UPDATE datasource SET oauth2_token_uuid = @oauth2_token_uuid WHERE uuid = @uuid;

-- name: LinkDatasourceWithClient :exec
UPDATE datasource SET
	oauth2_client_id = @client_id
WHERE uuid = @uuid;

-- name: DeleteDatasource :exec
DELETE FROM datasource
WHERE uuid = @uuid;
