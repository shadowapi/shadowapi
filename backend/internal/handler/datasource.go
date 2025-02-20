package handler

import (
	"context"
	"net/http"

	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/shadowapi/shadowapi/backend/pkg/api"
	"github.com/shadowapi/shadowapi/backend/pkg/query"
)

func (h *Handler) DatasourceSetOAuth2Client(
	ctx context.Context, req *api.DatasourceSetOAuth2ClientReq, params api.DatasourceSetOAuth2ClientParams,
) error {
	log := h.log.With("handler", "DatasourceSetOAuth2Client")
	connectionUUID, err := uuid.FromString(params.UUID)
	if err != nil {
		log.Error("failed to parse connection uuid", "error", err)
		return ErrWithCode(http.StatusBadRequest, E("failed to parse connection uuid"))
	}

	err = query.New(h.dbp).LinkDatasourceWithClient(ctx,
		query.LinkDatasourceWithClientParams{
			UUID:     connectionUUID,
			ClientID: pgtype.Text{String: req.ClientID, Valid: true},
		})
	if err != nil {
		log.Error("failed to link connection with client", "error", err)
		return ErrWithCode(http.StatusInternalServerError, E("failed to link connection with client"))
	}
	return nil
}
