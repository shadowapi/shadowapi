package handler

import (
	"context"
	"encoding/json"
	"github.com/shadowapi/shadowapi/backend/internal/converter"
	"net/http"

	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/shadowapi/shadowapi/backend/internal/db"
	"github.com/shadowapi/shadowapi/backend/pkg/api"
	"github.com/shadowapi/shadowapi/backend/pkg/query"
)

func (h *Handler) PipelineCreate(ctx context.Context, req *api.Pipeline) (*api.Pipeline, error) {
	log := h.log.With("handler", "PipelineCreate")
	return db.InTx(ctx, h.dbp, func(tx pgx.Tx) (*api.Pipeline, error) {
		pipelineUUID := uuid.Must(uuid.NewV7())
		pgDatasourceUUID, err := converter.ConvertStringToPgUUID(req.DatasourceUUID)
		if err != nil {
			log.Error("failed to convert datasource uuid", "error", err)
			return nil, ErrWithCode(http.StatusBadRequest, E("invalid datasource uuid"))
		}
		var flowData []byte
		if req.Flow.IsSet() {
			flowData, err = json.Marshal(req.Flow.Value)
			if err != nil {
				log.Error("failed to marshal pipeline flow", "error", err)
				return nil, ErrWithCode(http.StatusInternalServerError, E("failed to marshal pipeline flow"))
			}
		}
		ds, err := query.New(tx).GetDatasource(ctx, pgDatasourceUUID)
		if err != nil {
			log.Error("failed to get datasource", "error", err)
			return nil, ErrWithCode(http.StatusInternalServerError, E("failed to get datasource"))
		}

		qParams := query.CreatePipelineParams{
			UUID:           pgtype.UUID{Bytes: converter.UToBytes(pipelineUUID), Valid: true},
			DatasourceUUID: pgDatasourceUUID,
			Name:           req.Name,
			Type:           ds.Datasource.Type,
			IsEnabled:      req.IsEnabled.Or(false),
			Flow:           flowData,
		}
		pip, err := query.New(tx).CreatePipeline(ctx, qParams)
		if err != nil {
			log.Error("failed to create pipeline", "error", err)
			return nil, ErrWithCode(http.StatusInternalServerError, E("failed to create pipeline"))
		}
		out, err := qToApiPipeline(pip)
		if err != nil {
			log.Error("failed to map pipeline", "error", err)
			return nil, ErrWithCode(http.StatusInternalServerError, E("failed to map pipeline"))
		}
		return &out, nil
	})
}

func (h *Handler) PipelineGet(ctx context.Context, params api.PipelineGetParams) (*api.Pipeline, error) {
	log := h.log.With("handler", "PipelineGet")
	pipelineUUID, err := converter.ConvertStringToPgUUID(params.UUID.String())
	if err != nil {
		log.Error("invalid pipeline uuid", "error", err)
		return nil, ErrWithCode(http.StatusBadRequest, E("invalid pipeline uuid"))
	}
	row, err := query.New(h.dbp).GetPipeline(ctx, pipelineUUID)
	if err != nil {
		log.Error("failed to get pipeline", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to get pipeline"))
	}
	out, err := qToApiPipeline(row.Pipeline)
	if err != nil {
		log.Error("failed to map pipeline", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to map pipeline"))
	}
	return &out, nil
}

func (h *Handler) PipelineList(ctx context.Context, params api.PipelineListParams) (*api.PipelineListOK, error) {
	log := h.log.With("handler", "PipelineList")
	limit := int32(50)
	offset := int32(0)
	if params.Limit.IsSet() {
		limit = params.Limit.Value
	}
	if params.Offset.IsSet() {
		offset = params.Offset.Value
	}
	qParams := query.ListPipelinesParams{
		Offset: offset,
		Limit:  limit,
	}
	rows, err := query.New(h.dbp).ListPipelines(ctx, qParams)
	if err != nil {
		log.Error("failed to list pipelines", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to list pipelines"))
	}
	out := &api.PipelineListOK{}
	for _, row := range rows {
		p, err := qToApiPipeline(row.Pipeline)
		if err != nil {
			log.Error("failed to map pipeline row", "error", err)
			return nil, ErrWithCode(http.StatusInternalServerError, E("failed to map pipeline row"))
		}
		out.Pipelines = append(out.Pipelines, p)
	}
	return out, nil
}

func (h *Handler) PipelineUpdate(ctx context.Context, req *api.Pipeline, params api.PipelineUpdateParams) (*api.Pipeline, error) {
	log := h.log.With("handler", "PipelineUpdate")
	pipelineUUID, err := converter.ConvertStringToPgUUID(params.UUID.String())
	if err != nil {
		log.Error("invalid pipeline uuid", "error", err)
		return nil, ErrWithCode(http.StatusBadRequest, E("invalid pipeline uuid"))
	}
	return db.InTx(ctx, h.dbp, func(tx pgx.Tx) (*api.Pipeline, error) {
		existingRow, err := query.New(tx).GetPipeline(ctx, pipelineUUID)
		if err != nil {
			log.Error("failed to get existing pipeline", "error", err)
			return nil, ErrWithCode(http.StatusInternalServerError, E("failed to get existing pipeline"))
		}
		var flowData []byte
		if req.Flow.IsSet() {
			flowData, err = json.Marshal(req.Flow.Value)
			if err != nil {
				log.Error("failed to marshal pipeline flow", "error", err)
				return nil, ErrWithCode(http.StatusInternalServerError, E("failed to marshal pipeline flow"))
			}
		} else {
			flowData = existingRow.Pipeline.Flow
		}
		pgDatasourceUUID, err := converter.ConvertStringToPgUUID(req.DatasourceUUID)
		if err != nil {
			log.Error("failed to convert datasource uuid", "error", err)
			return nil, ErrWithCode(http.StatusBadRequest, E("invalid datasource uuid"))
		}
		uParams := query.UpdatePipelineParams{
			Name:           req.Name,
			Type:           req.Type,
			DatasourceUUID: pgDatasourceUUID,
			IsEnabled:      req.IsEnabled.Or(false),
			Flow:           flowData,
			UUID:           pipelineUUID,
		}
		err = query.New(tx).UpdatePipeline(ctx, uParams)
		if err != nil {
			log.Error("failed to update pipeline", "error", err)
			return nil, ErrWithCode(http.StatusInternalServerError, E("failed to update pipeline"))
		}
		row, err := query.New(tx).GetPipeline(ctx, pipelineUUID)
		if err != nil {
			log.Error("failed to get updated pipeline", "error", err)
			return nil, ErrWithCode(http.StatusInternalServerError, E("failed to get updated pipeline"))
		}
		out, err := qToApiPipeline(row.Pipeline)
		if err != nil {
			log.Error("failed to map updated pipeline", "error", err)
			return nil, ErrWithCode(http.StatusInternalServerError, E("failed to map updated pipeline"))
		}
		return &out, nil
	})
}

func (h *Handler) PipelineDelete(ctx context.Context, params api.PipelineDeleteParams) error {
	log := h.log.With("handler", "PipelineDelete")
	pipelineUUID, err := converter.ConvertStringToPgUUID(params.UUID.String())
	if err != nil {
		log.Error("invalid pipeline uuid", "error", err)
		return ErrWithCode(http.StatusBadRequest, E("invalid pipeline uuid"))
	}
	err = query.New(h.dbp).DeletePipeline(ctx, pipelineUUID)
	if err != nil {
		log.Error("failed to delete pipeline", "error", err)
		return ErrWithCode(http.StatusInternalServerError, E("failed to delete pipeline"))
	}
	return nil
}

// TODO finish convertion
func qToApiPipeline(dbp query.Pipeline) (api.Pipeline, error) {
	out := api.Pipeline{
		UUID:      api.NewOptString(dbp.UUID.String()), // TODO @reactima rethink the whole thing
		Name:      dbp.Name,
		Type:      api.NewOptString(dbp.Type),
		IsEnabled: api.NewOptBool(dbp.IsEnabled),
		CreatedAt: api.NewOptDateTime(dbp.CreatedAt.Time),
		UpdatedAt: api.NewOptDateTime(dbp.UpdatedAt.Time),
	}
	//u, err := uuid.FromString(dbp.UUID.String())
	//if err != nil {
	//	return out, err
	//}
	//out.UUID = api.NewOptUUID(u)
	//if dbp.DatasourceUUID != nil {
	//	out.DatasourceUUID = *dbp.DatasourceUUID
	//} else {
	//	out.DatasourceUUID = uuid.Nil
	//}

	if len(dbp.Flow) > 0 {
		var flowObj api.PipelineFlow
		if err := json.Unmarshal(dbp.Flow, &flowObj); err != nil {
			return out, err
		}
		out.Flow.SetTo(flowObj)
	}
	return out, nil
}
