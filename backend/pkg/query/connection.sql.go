// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.26.0
// source: connection.sql

package query

import (
	"context"

	"github.com/gofrs/uuid"
)

const linkDatasourceWithClient = `-- name: LinkDatasourceWithClient :exec
UPDATE datasource
SET settings = jsonb_set(
        settings,
        '{oauth2_client_id}',
        to_jsonb($1::text),
        true
               ),
    updated_at = NOW()
WHERE uuid = $2
`

type LinkDatasourceWithClientParams struct {
	Oauth2ClientID string    `json:"oauth2_client_id"`
	UUID           uuid.UUID `json:"uuid"`
}

func (q *Queries) LinkDatasourceWithClient(ctx context.Context, arg LinkDatasourceWithClientParams) error {
	_, err := q.db.Exec(ctx, linkDatasourceWithClient, arg.Oauth2ClientID, arg.UUID)
	return err
}

const linkDatasourceWithToken = `-- name: LinkDatasourceWithToken :exec
UPDATE datasource
SET settings = jsonb_set(
        settings,
        '{oauth2_token_uuid}',
        to_jsonb($1::text),
        true
               ),
    updated_at = NOW()
WHERE uuid = $2
`

type LinkDatasourceWithTokenParams struct {
	OAuth2TokenUUID string    `json:"oauth2_token_uuid"`
	UUID            uuid.UUID `json:"uuid"`
}

func (q *Queries) LinkDatasourceWithToken(ctx context.Context, arg LinkDatasourceWithTokenParams) error {
	_, err := q.db.Exec(ctx, linkDatasourceWithToken, arg.OAuth2TokenUUID, arg.UUID)
	return err
}
