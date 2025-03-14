-- name: CreateDatasourceEmail :one
INSERT INTO datasource_email (
	uuid, datasource_uuid, email, provider, password, imap_server, smtp_server, smtp_tls
) VALUES (
	@uuid, @datasource_uuid, @email, @provider, @password, @imap_server, @smtp_server, @smtp_tls
) RETURNING *;

-- name: UpdateDatasourceEmail :exec
UPDATE datasource_email SET
	password = COALESCE(@password, password),
	imap_server = COALESCE(@imap_server, imap_server),
	smtp_server = COALESCE(@smtp_server, smtp_server),
	smtp_tls = COALESCE(@smtp_tls, smtp_tls)
WHERE datasource_uuid = @datasource_uuid
RETURNING *;

-- name: DeleteDatasourceEmail :exec
DELETE FROM datasource_email
WHERE datasource_uuid = @datasource_uuid;
