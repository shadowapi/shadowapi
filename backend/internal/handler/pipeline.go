package handler

import (
	"context"
	"net/http"

	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/shadowapi/shadowapi/backend/internal/db"
	"github.com/shadowapi/shadowapi/backend/pkg/api"
	"github.com/shadowapi/shadowapi/backend/pkg/query"
)

func (h *Handler) PipelineCreate(ctx context.Context, req *api.PipelineCreateReq) (*api.Pipeline, error) {
	log := h.log.With("handler", "PipelineCreate")
	return db.InTx(ctx, h.dbp, func(tx pgx.Tx) (*api.Pipeline, error) {
		pipelineUUID, err := uuid.NewV7()
		if err != nil {
			log.Error("failed to generate pipeline uuid", "error", err)
			return nil, ErrWithCode(http.StatusInternalServerError, E("failed to generate pipeline uuid"))
		}
		flowData, err := req.Flow.MarshalJSON()
		if err != nil {
			log.Error("failed to marshal pipeline flow data", "error", err)
			return nil, ErrWithCode(http.StatusBadRequest, E("failed to marshal pipeline flow data"))
		}
		pipeline, err := query.New(tx).CreatePipeline(ctx, query.CreatePipelineParams{
			UUID: pipelineUUID,
			Name: req.Name,
			Flow: flowData,
		})
		if err != nil {
			log.Error("failed to create pipeline", "error", err)
			return nil, ErrWithCode(http.StatusInternalServerError, E("failed to create pipeline"))
		}
		ap, err := h.pipelineQueryToAPI(pipeline)
		if err != nil {
			log.Error("failed to convert pipeline query to api", "error", err)
			return nil, ErrWithCode(http.StatusInternalServerError, E("failed to convert pipeline query to api"))
		}
		return &ap, nil
	})
}

func (h *Handler) PipelineUpdate(
	ctx context.Context,
	req *api.PipelineUpdateReq,
	params api.PipelineUpdateParams,
) (*api.Pipeline, error) {
	log := h.log.With("handler", "PipelineUpdate")
	return db.InTx(ctx, h.dbp, func(tx pgx.Tx) (*api.Pipeline, error) {
		pipelineUUID, err := uuid.FromString(params.UUID)
		if err != nil {
			log.Error("failed to parse pipeline uuid", "error", err)
			return nil, ErrWithCode(http.StatusBadRequest, E("failed to parse pipeline uuid"))
		}
		flowData, err := req.Flow.MarshalJSON()
		if err != nil {
			log.Error("failed to marshal pipeline flow data", "error", err)
			return nil, ErrWithCode(http.StatusBadRequest, E("failed to marshal pipeline flow data"))
		}
		err = query.New(tx).UpdatePipeline(ctx, query.UpdatePipelineParams{
			UUID: pipelineUUID,
			Name: req.Name,
			Flow: flowData,
		})
		if err != nil {
			log.Error("failed to update pipeline", "error", err)
			return nil, ErrWithCode(http.StatusInternalServerError, E("failed to update pipeline"))
		}
		pipelines, err := query.New(tx).GetPipelines(ctx, query.GetPipelinesParams{
			UUID:  pgtype.Text{String: pipelineUUID.String(), Valid: true},
			Limit: pgtype.Int4{Int32: 1, Valid: true},
		})
		if err != nil {
			log.Error("failed to get pipeline", "error", err)
			return nil, ErrWithCode(http.StatusInternalServerError, E("failed to get pipeline"))
		}
		if len(pipelines) == 0 {
			log.Error("pipeline not found")
			return nil, ErrWithCode(http.StatusNotFound, E("pipeline not found"))
		}
		result, err := h.pipelineQueryToAPI(pipelines[0])
		if err != nil {
			log.Error("failed to convert pipeline query to api", "error", err)
			return nil, ErrWithCode(http.StatusInternalServerError, E("failed to convert pipeline query to api"))
		}
		return &result, nil
	})
}

func (h *Handler) PipelineList(ctx context.Context, params api.PipelineListParams) (*api.PipelineListOK, error) {
	log := h.log.With("handler", "PipelineList")
	return db.InTx(ctx, h.dbp, func(tx pgx.Tx) (*api.PipelineListOK, error) {
		qrParams := query.GetPipelinesParams{}
		if params.Limit.IsSet() {
			qrParams.Limit = pgtype.Int4{Int32: params.Limit.Value, Valid: true}
		}
		if params.Offset.IsSet() {
			qrParams.Offset = pgtype.Int4{Int32: params.Offset.Value, Valid: true}
		}
		out, err := query.New(tx).GetPipelines(ctx, qrParams)
		if err != nil {
			log.Error("failed to get pipeline", "error", err)
			return nil, ErrWithCode(http.StatusInternalServerError, E("failed to get pipeline"))
		}
		result := &api.PipelineListOK{
			Pipelines: []api.Pipeline{},
		}
		for _, o := range out {
			resultItem, err := h.pipelineQueryToAPI(o)
			if err != nil {
				log.Error("failed to convert pipeline query to api", "error", err)
				return nil, ErrWithCode(http.StatusInternalServerError, E("failed to convert pipeline query to api"))
			}
			result.Pipelines = append(result.Pipelines, resultItem)
		}
		return result, nil
	})
}

func (h *Handler) PipelineGet(ctx context.Context, params api.PipelineGetParams) (*api.Pipeline, error) {
	log := h.log.With("handler", "PipelineGet")
	return db.InTx(ctx, h.dbp, func(tx pgx.Tx) (*api.Pipeline, error) {
		pipelineUUID, err := uuid.FromString(params.UUID)
		if err != nil {
			log.Error("failed to parse pipeline uuid", "error", err)
			return nil, ErrWithCode(http.StatusBadRequest, E("failed to parse pipeline uuid"))
		}
		pipelines, err := query.New(tx).GetPipelines(ctx, query.GetPipelinesParams{
			UUID:  pgtype.Text{String: pipelineUUID.String(), Valid: true},
			Limit: pgtype.Int4{Int32: 1, Valid: true},
		})
		if err != nil {
			log.Error("failed to get pipeline", "error", err)
			return nil, ErrWithCode(http.StatusInternalServerError, E("failed to get pipeline"))
		}
		if len(pipelines) == 0 {
			log.Error("pipeline not found")
			return nil, ErrWithCode(http.StatusNotFound, E("pipeline not found"))
		}
		result, err := h.pipelineQueryToAPI(pipelines[0])
		if err != nil {
			log.Error("failed to convert pipeline query to api", "error", err)
			return nil, ErrWithCode(http.StatusInternalServerError, E("failed to convert pipeline query to api"))
		}
		return &result, nil
	})
}

func (h *Handler) PipelineDelete(ctx context.Context, params api.PipelineDeleteParams) error {
	log := h.log.With("handler", "PipelineGet")
	pipelineUUID, err := uuid.FromString(params.UUID)
	if err != nil {
		log.Error("failed to parse pipeline uuid", "error", err)
		return ErrWithCode(http.StatusBadRequest, E("failed to parse pipeline uuid"))
	}
	err = query.New(h.dbp).DeletePipeline(ctx, pipelineUUID)
	if err != nil {
		log.Error("failed to delete pipeline", "error", err)
		return ErrWithCode(http.StatusInternalServerError, E("failed to delete pipeline"))
	}
	return nil
}

func (h *Handler) pipelineQueryToAPI(from query.Pipeline) (to api.Pipeline, err error) {
	flow := api.PipelineFlow{}
	if err = flow.UnmarshalJSON(from.Flow); err != nil {
		return
	}
	return api.Pipeline{
		UUID:      from.UUID.String(),
		Name:      from.Name,
		Flow:      flow,
		CreatedAt: api.NewOptDateTime(from.CreatedAt),
		UpdatedAt: api.NewOptDateTime(from.UpdatedAt.Time),
	}, nil
}
