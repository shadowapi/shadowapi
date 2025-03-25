-- name: LinkDatasourceWithToken :exec
UPDATE datasource
SET settings = jsonb_set(settings, '{oauth2_token_uuid}', to_jsonb($1::text), true),
    updated_at = NOW()
WHERE uuid = $2;

-- name: LinkDatasourceWithClient :exec
UPDATE datasource
SET settings = jsonb_set(settings, '{oauth2_client_id}', to_jsonb($1::text), true),
    updated_at = NOW()
WHERE uuid = $2;
