package handler

import (
	"context"
	"github.com/jackc/pgx/v5"
	"github.com/shadowapi/shadowapi/backend/internal/db"
	"net/http"

	"github.com/shadowapi/shadowapi/backend/pkg/api"
	"github.com/shadowapi/shadowapi/backend/pkg/query"
)

func (h *Handler) DatasourceList(
	ctx context.Context, params api.DatasourceListParams,
) ([]api.Datasource, error) {
	log := h.log.With("handler", "StorageList")

	return db.InTx(ctx, h.dbp, func(tx pgx.Tx) ([]api.Datasource, error) {
		arg := query.GetDatasourcesParams{}
		if params.Limit.IsSet() {
			arg.Limit = params.Limit.Value
		}
		if params.Offset.IsSet() {
			arg.Offset = params.Offset.Value
		}

		rows, err := query.New(h.dbp).GetDatasources(ctx, arg)
		if err != nil {
			log.Error("failed to list datasources", "error", err.Error())
			return nil, ErrWithCode(http.StatusInternalServerError, E("failed to list datasources"))
		}

		var datasources []api.Datasource
		for _, row := range rows {
			datasources = append(datasources, QToDatasource(row))
		}
		return datasources, nil
	})
}

//func (h *Handler) DatasourceSetOAuth2Client(
//	ctx context.Context, req *api.DatasourceSetOAuth2ClientReq, params api.DatasourceSetOAuth2ClientParams,
//) error {
//	log := h.log.With("handler", "DatasourceSetOAuth2Client")
//	connectionUUID, err := uuid.FromString(params.UUID)
//	if err != nil {
//		log.Error("failed to parse connection uuid", "error", err)
//		return ErrWithCode(http.StatusBadRequest, E("failed to parse connection uuid"))
//	}
//
//	err = query.New(h.dbp).LinkDatasourceWithClient(ctx,
//		query.LinkDatasourceWithClientParams{
//			UUID:     connectionUUID,
//			ClientID: pgtype.Text{String: req.ClientID, Valid: true},
//		})
//	if err != nil {
//		log.Error("failed to link connection with client", "error", err)
//		return ErrWithCode(http.StatusInternalServerError, E("failed to link connection with client"))
//	}
//	return nil
//}
