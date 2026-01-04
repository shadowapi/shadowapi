package handler

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/shadowapi/shadowapi/backend/internal/converter"
	"github.com/shadowapi/shadowapi/backend/internal/db"
	"github.com/shadowapi/shadowapi/backend/internal/workspace"
	"github.com/shadowapi/shadowapi/backend/pkg/api"
	"github.com/shadowapi/shadowapi/backend/pkg/query"
)

func (h *Handler) PipelineCreate(ctx context.Context, req *api.Pipeline) (api.PipelineCreateRes, error) {
	log := h.log.With("handler", "PipelineCreate")

	// Extract workspace UUID from context - required for workspace-scoped access
	workspaceUUID, err := workspace.RequireWorkspaceUUID(ctx)
	if err != nil {
		log.Error("workspace context required", "error", err)
		return nil, ErrWithCode(http.StatusUnauthorized, E("workspace context required"))
	}

	return db.InTx(ctx, h.dbp, func(tx pgx.Tx) (api.PipelineCreateRes, error) {
		pipelineUUID := uuid.Must(uuid.NewV7())
		pgDatasourceUUID, err := converter.ConvertStringToPgUUID(req.DatasourceUUID)
		if err != nil {
			log.Error("failed to convert datasource uuid", "error", err)
			return nil, ErrWithCode(http.StatusBadRequest, E("invalid datasource uuid"))
		}
		pgStorageUUID, err := converter.ConvertStringToPgUUID(req.StorageUUID)
		if err != nil {
			log.Error("failed to convert storage uuid", "error", err)
			return nil, ErrWithCode(http.StatusBadRequest, E("invalid storage uuid"))
		}
		var flowData []byte
		if req.Flow.IsSet() {
			flowData, err = json.Marshal(req.Flow.Value)
			if err != nil {
				log.Error("failed to marshal pipeline flow", "error", err)
				return nil, ErrWithCode(http.StatusInternalServerError, E("failed to marshal pipeline flow"))
			}
		} else {
			// Default to empty JSON object for JSONB NOT NULL constraint
			flowData = []byte("{}")
		}

		// Validate datasource belongs to workspace
		ds, err := query.New(tx).GetDatasourceByWorkspace(ctx, query.GetDatasourceByWorkspaceParams{
			UUID:          pgDatasourceUUID,
			WorkspaceUUID: workspaceUUID,
		})
		if err != nil {
			log.Error("failed to get datasource", "error", err)
			return nil, ErrWithCode(http.StatusInternalServerError, E("failed to get datasource"))
		}

		// Handle optional worker_uuid
		var pgWorkerUUID pgtype.UUID
		if req.WorkerUUID.IsSet() && !req.WorkerUUID.IsNull() && req.WorkerUUID.Value != "" {
			pgWorkerUUID, err = converter.ConvertStringToPgUUID(req.WorkerUUID.Value)
			if err != nil {
				log.Error("failed to convert worker uuid", "error", err)
				return nil, ErrWithCode(http.StatusBadRequest, E("invalid worker uuid"))
			}
		}

		qParams := query.CreatePipelineParams{
			UUID:           pgtype.UUID{Bytes: converter.UToBytes(pipelineUUID), Valid: true},
			WorkspaceUUID:  workspaceUUID,
			DatasourceUUID: pgDatasourceUUID,
			StorageUuid:    pgStorageUUID,
			WorkerUUID:     pgWorkerUUID,
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

func (h *Handler) PipelineGet(ctx context.Context, params api.PipelineGetParams) (api.PipelineGetRes, error) {
	log := h.log.With("handler", "PipelineGet")

	// Extract workspace UUID from context - required for workspace-scoped access
	workspaceUUID, err := workspace.RequireWorkspaceUUID(ctx)
	if err != nil {
		log.Error("workspace context required", "error", err)
		return nil, ErrWithCode(http.StatusUnauthorized, E("workspace context required"))
	}

	pipelineUUID, err := converter.ConvertStringToPgUUID(params.UUID.String())
	if err != nil {
		log.Error("invalid pipeline uuid", "error", err)
		return nil, ErrWithCode(http.StatusBadRequest, E("invalid pipeline uuid"))
	}
	row, err := query.New(h.dbp).GetPipelineByWorkspace(ctx, query.GetPipelineByWorkspaceParams{
		UUID:          pipelineUUID,
		WorkspaceUUID: workspaceUUID,
	})
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

func (h *Handler) PipelineList(ctx context.Context, params api.PipelineListParams) (api.PipelineListRes, error) {
	log := h.log.With("handler", "PipelineList")

	// Extract workspace UUID from context - required for workspace-scoped access
	workspaceUUID, err := workspace.RequireWorkspaceUUID(ctx)
	if err != nil {
		log.Error("workspace context required", "error", err)
		return nil, ErrWithCode(http.StatusUnauthorized, E("workspace context required"))
	}

	limit := int32(50)
	offset := int32(0)
	if params.Limit.IsSet() {
		limit = params.Limit.Value
	}
	if params.Offset.IsSet() {
		offset = params.Offset.Value
	}
	qParams := query.ListPipelinesByWorkspaceParams{
		WorkspaceUUID: workspaceUUID,
		Offset:        offset,
		Limit:         limit,
	}
	rows, err := query.New(h.dbp).ListPipelinesByWorkspace(ctx, qParams)
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

func (h *Handler) PipelineUpdate(ctx context.Context, req *api.Pipeline, params api.PipelineUpdateParams) (api.PipelineUpdateRes, error) {
	log := h.log.With("handler", "PipelineUpdate")

	// Extract workspace UUID from context - required for workspace-scoped access
	workspaceUUID, err := workspace.RequireWorkspaceUUID(ctx)
	if err != nil {
		log.Error("workspace context required", "error", err)
		return nil, ErrWithCode(http.StatusUnauthorized, E("workspace context required"))
	}

	pipelineUUID, err := converter.ConvertStringToPgUUID(params.UUID.String())
	if err != nil {
		log.Error("invalid pipeline uuid", "error", err)
		return nil, ErrWithCode(http.StatusBadRequest, E("invalid pipeline uuid"))
	}
	return db.InTx(ctx, h.dbp, func(tx pgx.Tx) (api.PipelineUpdateRes, error) {
		existingRow, err := query.New(tx).GetPipelineByWorkspace(ctx, query.GetPipelineByWorkspaceParams{
			UUID:          pipelineUUID,
			WorkspaceUUID: workspaceUUID,
		})
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
		pgStorageUUID, err := converter.ConvertStringToPgUUID(req.StorageUUID)
		if err != nil {
			log.Error("failed to convert storage uuid", "error", err)
			return nil, ErrWithCode(http.StatusBadRequest, E("invalid datasource uuid"))
		}

		newType := existingRow.Pipeline.Type // default
		if req.Type.IsSet() {                // only overwrite if client sent it
			newType = req.Type.Value
		}

		// Handle optional worker_uuid
		var pgWorkerUUID pgtype.UUID
		if req.WorkerUUID.IsSet() && !req.WorkerUUID.IsNull() && req.WorkerUUID.Value != "" {
			pgWorkerUUID, err = converter.ConvertStringToPgUUID(req.WorkerUUID.Value)
			if err != nil {
				log.Error("failed to convert worker uuid", "error", err)
				return nil, ErrWithCode(http.StatusBadRequest, E("invalid worker uuid"))
			}
		}

		uParams := query.UpdatePipelineByWorkspaceParams{
			Name:           req.Name,
			Type:           newType,
			DatasourceUUID: pgDatasourceUUID,
			StorageUuid:    pgStorageUUID,
			WorkerUUID:     pgWorkerUUID,
			IsEnabled:      req.IsEnabled.Or(false),
			Flow:           flowData,
			UUID:           pipelineUUID,
			WorkspaceUUID:  workspaceUUID,
		}
		err = query.New(tx).UpdatePipelineByWorkspace(ctx, uParams)
		if err != nil {
			log.Error("failed to update pipeline", "error", err)
			return nil, ErrWithCode(http.StatusInternalServerError, E("failed to update pipeline"))
		}
		row, err := query.New(tx).GetPipelineByWorkspace(ctx, query.GetPipelineByWorkspaceParams{
			UUID:          pipelineUUID,
			WorkspaceUUID: workspaceUUID,
		})
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

func (h *Handler) PipelineDelete(ctx context.Context, params api.PipelineDeleteParams) (api.PipelineDeleteRes, error) {
	log := h.log.With("handler", "PipelineDelete")

	// Extract workspace UUID from context - required for workspace-scoped access
	workspaceUUID, err := workspace.RequireWorkspaceUUID(ctx)
	if err != nil {
		log.Error("workspace context required", "error", err)
		return nil, ErrWithCode(http.StatusUnauthorized, E("workspace context required"))
	}

	pipelineUUID, err := converter.ConvertStringToPgUUID(params.UUID.String())
	if err != nil {
		log.Error("invalid pipeline uuid", "error", err)
		return nil, ErrWithCode(http.StatusBadRequest, E("invalid pipeline uuid"))
	}
	err = query.New(h.dbp).DeletePipelineByWorkspace(ctx, query.DeletePipelineByWorkspaceParams{
		UUID:          pipelineUUID,
		WorkspaceUUID: workspaceUUID,
	})
	if err != nil {
		log.Error("failed to delete pipeline", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to delete pipeline"))
	}
	return &api.PipelineDeleteOK{}, nil
}

// qToApiPipeline converts a db pipeline row into an API pipeline, handling nullable fields safely.
func qToApiPipeline(dbp query.Pipeline) (api.Pipeline, error) {
	// Safely handle potential nil UUID pointers for datasource and storage
	dsUUID := ""
	if dbp.DatasourceUUID != nil {
		dsUUID = dbp.DatasourceUUID.String()
	}
	storageUUID := ""
	if dbp.StorageUuid != nil {
		storageUUID = dbp.StorageUuid.String()
	}

	out := api.Pipeline{
		UUID:           api.NewOptString(dbp.UUID.String()),
		DatasourceUUID: dsUUID,
		StorageUUID:    storageUUID,
		Name:           dbp.Name,
		Type:           api.NewOptString(dbp.Type),
		IsEnabled:      api.NewOptBool(dbp.IsEnabled),
		CreatedAt:      api.NewOptDateTime(dbp.CreatedAt.Time),
		UpdatedAt:      api.NewOptDateTime(dbp.UpdatedAt.Time),
	}

	// Handle nullable worker_uuid
	if dbp.WorkerUUID != nil {
		out.WorkerUUID = api.NewOptNilString(dbp.WorkerUUID.String())
	}

	// Unmarshal Flow JSON if present
	if len(dbp.Flow) > 0 {
		var flowObj api.PipelineFlow
		if err := json.Unmarshal(dbp.Flow, &flowObj); err != nil {
			return out, err
		}
		out.Flow.SetTo(flowObj)
	}

	return out, nil
}
