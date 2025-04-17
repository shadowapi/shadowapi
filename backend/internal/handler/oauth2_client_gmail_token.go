package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"golang.org/x/oauth2"
	"log/slog"

	"github.com/shadowapi/shadowapi/backend/internal/converter"
	"github.com/shadowapi/shadowapi/backend/internal/db"
	"github.com/shadowapi/shadowapi/backend/pkg/api"
	"github.com/shadowapi/shadowapi/backend/pkg/query"
)

func (h *Handler) handleGmailToken(
	ctx context.Context,
	log *slog.Logger,
	token *oauth2.Token,
	stateQuery url.Values,
	clientID string,
) (*api.OAuth2ClientCallbackFound, error) {
	// Marshal token to JSON
	tokenData, err := json.Marshal(token)
	if err != nil {
		log.Error("failed to marshal token", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed marshal token"))
	}

	// Extract datasource UUID from state
	dsID := stateQuery.Get("datasource_uuid")
	if dsID == "" {
		log.Error("missing datasource_uuid in state query")
		return nil, ErrWithCode(http.StatusBadRequest, E("missing datasource_uuid parameter"))
	}
	goDSUUID, err := uuid.FromString(dsID)
	if err != nil {
		log.Error("invalid datasource_uuid", "error", err)
		return nil, ErrWithCode(http.StatusBadRequest, E("invalid datasource_uuid parameter"))
	}

	// Convert clientID to pgtype.UUID
	pgClientUUID, err := converter.ConvertStringToPgUUID(clientID)
	if err != nil {
		log.Error("failed to convert client UUID", "error", err)
		return nil, ErrWithCode(http.StatusBadRequest, E("invalid client UUID"))
	}

	// Optionally extract user_uuid from state
	pgUserUUID := pgtype.UUID{Valid: false}
	if u := stateQuery.Get("user_uuid"); u != "" {
		goUserUUID, err := uuid.FromString(u)
		if err != nil {
			log.Error("invalid user_uuid", "error", err)
			return nil, ErrWithCode(http.StatusBadRequest, E("invalid user_uuid parameter"))
		}
		pgUserUUID = converter.UuidToPgUUID(goUserUUID)
	}

	// New token UUIDs
	goTokenUUID := uuid.Must(uuid.NewV7())
	pgTokenUUID := converter.UuidToPgUUID(goTokenUUID)

	return db.InTx(ctx, h.dbp, func(tx pgx.Tx) (*api.OAuth2ClientCallbackFound, error) {
		q := query.New(tx)

		// Delete existing token for this client
		_ = q.DeleteOauth2TokenByClientUUID(ctx, pgClientUUID)

		// Create new oauth2_token
		if _, err := q.CreateOauth2Token(ctx, query.CreateOauth2TokenParams{
			UUID:       pgTokenUUID,
			ClientUuid: pgClientUUID,
			UserUUID:   pgUserUUID,
			Token:      tokenData,
		}); err != nil {
			log.Error("failed to add oauth2 token", "error", err)
			return nil, ErrWithCode(http.StatusInternalServerError, E("can't add oauth2 token"))
		}

		// Link datasource with token
		if err := q.LinkDatasourceWithToken(ctx, query.LinkDatasourceWithTokenParams{
			UUID:            goDSUUID,
			OAuth2TokenUUID: goTokenUUID.String(),
		}); err != nil {
			log.Error("failed to link datasource with token", "error", err)
			return nil, ErrWithCode(http.StatusInternalServerError, E("failed to link datasource with token"))
		}

		// Create oauth2_subject
		goSubjectUUID := uuid.Must(uuid.NewV7())
		pgSubjectUUID := converter.UuidToPgUUID(goSubjectUUID)
		if _, err := q.CreateOauth2Subject(ctx, query.CreateOauth2SubjectParams{
			UUID:       pgSubjectUUID,
			UserUUID:   pgUserUUID,
			ClientUuid: pgClientUUID,
		}); err != nil {
			log.Error("failed to add oauth2 subject", "error", err)
			return nil, ErrWithCode(http.StatusInternalServerError, E("can't add oauth2 subject"))
		}

		// Redirect location
		location, err := url.Parse("/datasources/" + dsID)
		if err != nil {
			log.Error("can't parse location", "error", err)
			return nil, ErrWithCode(http.StatusInternalServerError, E("can't parse location"))
		}

		return &api.OAuth2ClientCallbackFound{
			Location: api.OptURI{Value: *location, Set: true},
		}, nil
	})
}

/*
func (h *Handler) handleGmailTokenOld(
	ctx context.Context,
	log *slog.Logger,
	token *oauth2.Token,
	stateQuery url.Values,
	clientID string,
) (*api.OAuth2ClientCallbackFound, error) {
	return nil, ErrWithCode(http.StatusInternalServerError, E("not implemented"))

		var tokenData []byte

		tokenData, err := json.Marshal(token)
		if err != nil {
			log.Error("marshal api token", "error", err)
			return nil, ErrWithCode(http.StatusInternalServerError, E("failed marshal api token"))
		}

		datasourceUUID := uuid.FromStringOrNil(stateQuery.Get("datasource_uuid"))
		if datasourceUUID.IsNil() {
			log.Error("missing datasource_uuid parameter")
			return nil, ErrWithCode(http.StatusBadRequest, E("missing datasource_uuid parameter"))
		}

		return db.InTx(ctx, h.dbp, func(tx pgx.Tx) (*api.OAuth2ClientCallbackFound, error) {
			db := query.New(tx)


			err := db.DeleteOauth2TokenByClientID(ctx, clientID)
			if err != nil && err != pgx.ErrNoRows {
				log.Error("can't delete oauth2 token by client id", "error", err)
				return nil, ErrWithCode(http.StatusInternalServerError, E("can't delete oauth2 token by client id"))
			}

			tokenUUID := uuid.Must(uuid.NewV7())
			err = db.AddOauth2Token(ctx, query.AddOauth2TokenParams{
				UUID:     tokenUUID,
				ClientID: clientID,
				Token:    tokenData,
			})
			if err != nil {
				log.Error("can't add oauth2 token", "error", err)
				return nil, ErrWithCode(http.StatusInternalServerError, E("can't add oauth2 token"))
			}

			err = db.LinkDatasourceWithToken(ctx, query.LinkDatasourceWithTokenParams{
				UUID:            datasourceUUID,
				OAuth2TokenUUID: tokenUUID.String(),
			})
			if err != nil {
				log.Error("link datasource with token", "error", err)
				return nil, ErrWithCode(http.StatusInternalServerError, E("link datasource with token"))
			}

			// TODO @reactima token needs to be sent and processed as a message
			//log.Info("add token refresh job to worker pool", "token_uuid", tokenUUID, "expiry", token.Expiry)
			//if err = h.wbr.ScheduleRefresh(ctx, tokenUUID, token.Expiry); err != nil {
			//	log.Error("fail add job to worker pool", "error", err)
			//	return nil, ErrWithCode(http.StatusInternalServerError, E("fail add job to worker pool"))
			//}

			location, err := url.Parse("/datasources/" + datasourceUUID.String())
			if err != nil {
				log.Error("can't parse location", "error", err)
				return nil, ErrWithCode(http.StatusInternalServerError, E("can't parse location"))
			}

			return &api.OAuth2ClientCallbackFound{
				Location: api.OptURI{Value: *location, Set: true},
			}, nil
		})


}
*/
