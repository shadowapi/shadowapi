-- name: GetOauth2State :one
SELECT
  *
FROM oauth2_state
WHERE 
  uuid = @uuid AND expired_at > NOW()
LIMIT 1;

-- name: CreateOauth2State :one
INSERT INTO oauth2_state (
  uuid, client_name, client_id, state, expired_at
) VALUES (
  @uuid, @client_name, @client_id, @state, NOW() + interval '1 hour'
) RETURNING *;

-- name: CreateOauth2Subject :one
INSERT INTO oauth2_subject (
  uuid, user_uuid, client_id, client_name, expired_at
) VALUES (
  @uuid, @user_uuid, @client_id, @client_name, @expired_at
) RETURNING *;

-- name: CreateOauth2Client :one
INSERT INTO oauth2_client (
  id, provider, name, secret
) VALUES (
  @id, @provider, @name, @secret
) RETURNING *;

-- name: GetOauth2Client :one
SELECT
  *
FROM oauth2_client
WHERE 
  id = @id
LIMIT 1;

-- name: GetOauth2TokensByClientID :many
SELECT
  *
FROM oauth2_token
WHERE "client_id" = @id;

-- name: GetOauth2TokenByUUID :one
SELECT
  *
FROM oauth2_token
WHERE "uuid" = @uuid
LIMIT 1;

-- name: GetOauth2ClientTokens :many
SELECT
  ot.*,
  c.name
FROM oauth2_token AS ot
LEFT JOIN datasource AS c ON c.oauth2_token_uuid = ot.uuid
WHERE 
  client_id = @client_id;

-- name: AddOauth2Token :exec
INSERT INTO oauth2_token (
  uuid, client_id, token
) VALUES (
  @uuid, @client_id, @token
);

-- name: UpdateOauth2Token :exec
UPDATE oauth2_token SET
  token = @token
WHERE uuid = @uuid;

-- name: DeleteOauth2Token :exec
DELETE FROM oauth2_token
WHERE uuid = @uuid;

-- name: DeleteOauth2TokenByClientID :exec
DELETE FROM oauth2_token
WHERE client_id = @client_id;

-- name: UpdateOauth2Client :exec
UPDATE oauth2_client SET
  name = @name,
  provider = @provider,
  secret = @secret
WHERE id = @id;

-- name: DeleteOauth2Client :exec
DELETE FROM oauth2_client
WHERE id = @id;

-- name: ListOauth2Clients :many
SELECT
  *
FROM oauth2_client;
