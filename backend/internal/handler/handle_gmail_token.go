package handler

import (
	"context"
	"log/slog"
	"net/http"
	"net/url"

	"golang.org/x/oauth2"

	"github.com/shadowapi/shadowapi/backend/pkg/api"
)

func (h *Handler) handleGmailToken(
	ctx context.Context,
	log *slog.Logger,
	token *oauth2.Token,
	stateQuery url.Values,
	clientID string,
) (*api.OAuth2ClientCallbackFound, error) {
	return nil, ErrWithCode(http.StatusInternalServerError, E("not implemented"))

	/*
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
	*/

}
