package handler

import (
	"context"
	"encoding/json"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/shadowapi/shadowapi/backend/internal/converter"
	"net/http"

	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v5"
	"golang.org/x/oauth2"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"

	"github.com/shadowapi/shadowapi/backend/pkg/api"
	"github.com/shadowapi/shadowapi/backend/pkg/query"

	appClients "github.com/shadowapi/shadowapi/backend/internal/oauth2"
)

func (h *Handler) DatasourceEmailRunPipeline(
	ctx context.Context, params api.DatasourceEmailRunPipelineParams,
) (*api.DatasourceEmailRunPipelineOK, error) {
	log := h.log.With("handler", "DatasourceEmailRunPipeline", "connection_uuid", params.UUID)
	connectionUUID, err := uuid.FromString(params.UUID)
	if err != nil {
		log.Error("wrong connectionUUID", "error", err)
		return nil, ErrWithCode(http.StatusBadRequest, E("wrong connectionUUID"))
	}

	ce, err := query.New(h.dbp).GetDatasource(ctx, pgtype.UUID{Bytes: [16]byte(connectionUUID.Bytes()), Valid: true})
	if err != nil && err != pgx.ErrNoRows {
		log.Error("no such connection", "error", err)
		return nil, ErrWithCode(http.StatusBadRequest, E("no such connection"))
	} else if err != nil {
		h.log.Error("failed to get connection", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to get connection"))
	}

	var ds api.DatasourceEmail
	if ce.Datasource.Settings == nil {
		h.log.Error("no settings", "error", err)
		return nil, ErrWithCode(http.StatusBadRequest, E("no settings"))
	}
	if err := json.Unmarshal(ce.Datasource.Settings, &ds); err != nil {
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to unmarshal settings"))
	}

	tokenUUID, err := converter.ConvertOptStringToUUID(ds.OAuth2TokenUUID)
	if err != nil {
		log.Error("invalid OAuth2TokenUUID", "error", err)
		return nil, ErrWithCode(http.StatusBadRequest, E("invalid oauth2 token uuid"))
	}

	token, err := query.New(h.dbp).GetOauth2TokenByUUID(ctx, tokenUUID)
	if err != nil && err != pgx.ErrNoRows {
		h.log.Error("no such oauth2 token", "error", err)
		return nil, ErrWithCode(http.StatusBadRequest, E("no such oauth2 token"))
	} else if err != nil {
		h.log.Error("failed to get oauth2 client token", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to get oauth2 client token"))
	}

	var oaToken oauth2.Token
	if err = json.Unmarshal(token.Token, &oaToken); err != nil {
		h.log.Error("failed to unmarshal oauth2 token", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to unmarshal oauth2 token"))
	}

	oauth2Cfg, err := appClients.GetClientConfig(ctx, h.dbp, token.ClientID)
	if err != nil {
		h.log.Error("failed to get oauth2 client config", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to get oauth2 client config"))
	}

	client := oauth2Cfg.Client(ctx, &oaToken)
	srv, err := gmail.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		h.log.Error("failed to create gmail service", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to create gmail service"))
	}

	user := "me"
	listLablesResponse, err := srv.Users.Labels.List(user).Do()
	if err != nil {
		h.log.Error("failed to list labels", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to list labels"))
	}

	out := api.DatasourceEmailRunPipelineOK{}
	for _, l := range listLablesResponse.Labels {
		ml := api.EmailLabel{
			ID:             api.OptString{Value: l.Id, Set: true},
			HTTPStatusCode: int64(l.HTTPStatusCode),
			Header:         api.EmailLabelHeader(map[string][]string(l.Header)),
			Color: api.OptEmailLabelColor{
				Value: api.EmailLabelColor{
					BackgroundColor: api.OptString{Value: l.Color.BackgroundColor, Set: true},
					TextColor:       api.OptString{Value: l.Color.TextColor, Set: true},
				},
				Set: true,
			},
			LabelListVisibility:   api.OptString{Value: l.LabelListVisibility, Set: true},
			MessageListVisibility: api.OptString{Value: l.MessageListVisibility, Set: true},
			MessagesTotal:         api.OptInt64{Value: l.MessagesTotal, Set: true},
			MessagesUnread:        api.OptInt64{Value: l.MessagesUnread, Set: true},
			Name:                  api.OptString{Value: l.Name, Set: true},
			ThreadsTotal:          api.OptInt64{Value: l.ThreadsTotal, Set: true},
			ThreadsUnread:         api.OptInt64{Value: l.ThreadsUnread, Set: true},
			Type:                  api.OptString{Value: l.Type, Set: true},
		}
		out.Labels = append(out.Labels, ml)
	}
	return &out, nil
}
