-- name: LinkDatasourceWithToken :exec
UPDATE datasource
SET settings = jsonb_set(
        settings,
        '{oauth2_token_uuid}',
        to_jsonb(sqlc.arg(oauth2_token_uuid)::text),
        true
               ),
    updated_at = NOW()
WHERE uuid = sqlc.arg(uuid);

-- name: LinkDatasourceWithClient :exec
UPDATE datasource
SET settings = jsonb_set(
        settings,
        '{oauth2_client_id}',
        to_jsonb(sqlc.arg(oauth2_client_id)::text),
        true
               ),
    updated_at = NOW()
WHERE uuid = sqlc.arg(uuid);