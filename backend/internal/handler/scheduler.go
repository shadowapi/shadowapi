package handler

import (
	"context"
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

// SchedulerCreate implements scheduler-create operation.
//
// POST /scheduler
func (h *Handler) SchedulerCreate(ctx context.Context, req *api.Scheduler) (api.SchedulerCreateRes, error) {
	log := h.log.With("handler", "SchedulerCreate")

	// Extract workspace UUID from context - required for workspace-scoped access
	workspaceUUID, err := workspace.RequireWorkspaceUUID(ctx)
	if err != nil {
		log.Error("workspace context required", "error", err)
		return nil, ErrWithCode(http.StatusUnauthorized, E("workspace context required"))
	}

	return db.InTx(ctx, h.dbp, func(tx pgx.Tx) (api.SchedulerCreateRes, error) {
		// Generate a new UUID for the scheduler
		schedulerUUID := uuid.Must(uuid.NewV7())
		pgPipelineUUID, err := converter.ConvertStringToPgUUID(req.PipelineUUID)
		if err != nil {
			log.Error("failed to convert scheduler uuid", "error", err)
			return nil, ErrWithCode(http.StatusBadRequest, E("invalid datasource uuid"))
		}
		// Verify that the pipeline exists and belongs to the workspace
		pipe, err := query.New(tx).GetPipelineByWorkspace(ctx, query.GetPipelineByWorkspaceParams{
			UUID:          pgPipelineUUID,
			WorkspaceUUID: workspaceUUID,
		})
		if err != nil {
			log.Error("failed to get pipeline", "error", err)
			return nil, ErrWithCode(http.StatusInternalServerError, E("failed to get pipeline"))
		}
		_ = pipe // Use pipe if needed

		// Build query parameters for the new scheduler.
		qParams := query.CreateSchedulerParams{
			UUID:                pgtype.UUID{Bytes: converter.UToBytes(schedulerUUID), Valid: true},
			PipelineUuid:        pgPipelineUUID,
			ScheduleType:        req.ScheduleType, // 'cron' or 'one_time'
			CronExpression:      converter.ConvertOptNilStringToPgText(req.CronExpression),
			RunAt:               converter.NullTimestamptz(),
			Timezone:            req.Timezone.Or("UTC"),
			NextRun:             converter.NullTimestamptz(),
			LastRun:             converter.NullTimestamptz(),
			LastUid:             0, // Start from beginning (deprecated)
			IsEnabled:           req.IsEnabled.Or(false),
			IsPaused:            req.IsPaused.Or(false),
			BatchSize:           int32(req.BatchSize.Or(100)),
			SyncState:           pgtype.Text{String: "initial", Valid: true},
			LastSyncTimestamp:   converter.NullTimestamptz(),
			OldestSyncTimestamp: converter.NullTimestamptz(),
			CutoffDate:          converter.ConvertOptNilDateTimeToPgTimestamptz(req.CutoffDate),
		}

		sch, err := query.New(tx).CreateScheduler(ctx, qParams)
		if err != nil {
			log.Error("failed to create scheduler", "error", err)
			return nil, ErrWithCode(http.StatusInternalServerError, E("failed to create scheduler"))
		}

		out, err := qToApiScheduler(sch)
		if err != nil {
			log.Error("failed to map scheduler", "error", err)
			return nil, ErrWithCode(http.StatusInternalServerError, E("failed to map scheduler"))
		}
		return &out, nil
	})
}

// SchedulerDelete implements scheduler-delete operation.
//
// DELETE /scheduler/{uuid}
func (h *Handler) SchedulerDelete(ctx context.Context, params api.SchedulerDeleteParams) (api.SchedulerDeleteRes, error) {
	log := h.log.With("handler", "SchedulerDelete")

	// Extract workspace UUID from context - required for workspace-scoped access
	workspaceUUID, err := workspace.RequireWorkspaceUUID(ctx)
	if err != nil {
		log.Error("workspace context required", "error", err)
		return nil, ErrWithCode(http.StatusUnauthorized, E("workspace context required"))
	}

	schUUID, err := converter.ConvertStringToPgUUID(params.UUID.String())
	if err != nil {
		log.Error("invalid scheduler uuid", "error", err)
		return nil, ErrWithCode(http.StatusBadRequest, E("invalid scheduler uuid"))
	}
	err = query.New(h.dbp).DeleteSchedulerByWorkspace(ctx, query.DeleteSchedulerByWorkspaceParams{
		UUID:          schUUID,
		WorkspaceUUID: workspaceUUID,
	})
	if err != nil {
		log.Error("failed to delete scheduler", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to delete scheduler"))
	}
	return &api.SchedulerDeleteOK{}, nil
}

// SchedulerGet implements scheduler-get operation.
//
// GET /scheduler/{uuid}
func (h *Handler) SchedulerGet(ctx context.Context, params api.SchedulerGetParams) (api.SchedulerGetRes, error) {
	log := h.log.With("handler", "SchedulerGet")

	// Extract workspace UUID from context - required for workspace-scoped access
	workspaceUUID, err := workspace.RequireWorkspaceUUID(ctx)
	if err != nil {
		log.Error("workspace context required", "error", err)
		return nil, ErrWithCode(http.StatusUnauthorized, E("workspace context required"))
	}

	schUUID, err := converter.ConvertStringToPgUUID(params.UUID.String())
	if err != nil {
		log.Error("invalid scheduler uuid", "error", err)
		return nil, ErrWithCode(http.StatusBadRequest, E("invalid scheduler uuid"))
	}
	sch, err := query.New(h.dbp).GetSchedulerByWorkspace(ctx, query.GetSchedulerByWorkspaceParams{
		UUID:          schUUID,
		WorkspaceUUID: workspaceUUID,
	})
	if err != nil {
		log.Error("failed to get scheduler", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to get scheduler"))
	}
	out, err := qToApiSchedulerByWorkspaceRow(sch)
	if err != nil {
		log.Error("failed to map scheduler", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to map scheduler"))
	}
	return &out, nil
}

// SchedulerList implements scheduler-list operation.
//
// GET /scheduler
func (h *Handler) SchedulerList(ctx context.Context, params api.SchedulerListParams) (api.SchedulerListRes, error) {
	log := h.log.With("handler", "SchedulerList")

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

	// Build filter parameters (if provided)
	qParams := query.GetSchedulersWithWorkspaceParams{
		WorkspaceUUID:  workspaceUUID,
		OrderBy:        nil,
		OrderDirection: "desc",
		Offset:         offset,
		Limit:          limit,

		PipelineUuid: "", // set further below if query has ?pipeline_uuid
		IsEnabled:    -1, // -1 means "either"
		IsPaused:     -1, // FIX: use -1 so we don't implicitly filter out paused rows
	}

	if params.PipelineUUID.IsSet() {
		qParams.PipelineUuid = params.PipelineUUID.Value.String()
	}

	schRows, err := query.New(h.dbp).GetSchedulersWithWorkspace(ctx, qParams)
	if err != nil {
		log.Error("failed to list schedulers", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to list schedulers"))
	}

	out := make([]api.Scheduler, 0, len(schRows))
	for _, row := range schRows {
		apiItem, err := qToApiSchedulersWithWorkspaceRow(row)
		if err != nil {
			log.Error("failed to map scheduler row", "error", err)
			return nil, ErrWithCode(http.StatusInternalServerError, E("failed to map scheduler row"))
		}
		out = append(out, apiItem)
	}
	res := api.SchedulerListOKApplicationJSON(out)
	return &res, nil
}

// SchedulerUpdate implements scheduler-update operation.
//
// PUT /scheduler/{uuid}
func (h *Handler) SchedulerUpdate(ctx context.Context, req *api.Scheduler, params api.SchedulerUpdateParams) (api.SchedulerUpdateRes, error) {
	log := h.log.With("handler", "SchedulerUpdate")

	// Extract workspace UUID from context - required for workspace-scoped access
	workspaceUUID, err := workspace.RequireWorkspaceUUID(ctx)
	if err != nil {
		log.Error("workspace context required", "error", err)
		return nil, ErrWithCode(http.StatusUnauthorized, E("workspace context required"))
	}

	schUUID, err := converter.ConvertStringToPgUUID(params.UUID.String())
	if err != nil {
		log.Error("invalid scheduler uuid", "error", err)
		return nil, ErrWithCode(http.StatusBadRequest, E("invalid scheduler uuid"))
	}
	return db.InTx(ctx, h.dbp, func(tx pgx.Tx) (api.SchedulerUpdateRes, error) {
		scheduler, err := query.New(tx).GetSchedulerByWorkspace(ctx, query.GetSchedulerByWorkspaceParams{
			UUID:          schUUID,
			WorkspaceUUID: workspaceUUID,
		})
		if err != nil {
			log.Error("failed to get scheduler", "error", err)
			return nil, ErrWithCode(http.StatusInternalServerError, E("failed to get scheduler"))
		}
		var isEnabled bool
		if req.IsEnabled.IsSet() {
			isEnabled = req.IsEnabled.Value
		} else {
			isEnabled = scheduler.Scheduler.IsEnabled
		}

		var batchSize int32
		if req.BatchSize.IsSet() {
			batchSize = int32(req.BatchSize.Value)
		} else {
			batchSize = scheduler.Scheduler.BatchSize
		}

		var isPaused bool
		if req.IsPaused.IsSet() {
			isPaused = req.IsPaused.Value
		} else {
			isPaused = scheduler.Scheduler.IsPaused
		}

		uParams := query.UpdateSchedulerByWorkspaceParams{
			CronExpression: converter.ConvertOptNilStringToPgText(req.CronExpression),
			RunAt:          converter.NullTimestamptz(),
			Timezone:       req.Timezone.Or("UTC"),
			NextRun:        converter.NullTimestamptz(),
			LastRun:        converter.NullTimestamptz(),
			LastUid:        0, // Preserve existing value via COALESCE
			IsEnabled:      isEnabled,
			IsPaused:       isPaused,
			BatchSize:      batchSize,
			CutoffDate:     converter.ConvertOptNilDateTimeToPgTimestamptz(req.CutoffDate),
			UUID:           schUUID,
			WorkspaceUUID:  workspaceUUID,
		}
		err = query.New(tx).UpdateSchedulerByWorkspace(ctx, uParams)
		if err != nil {
			log.Error("failed to update scheduler", "error", err)
			return nil, ErrWithCode(http.StatusInternalServerError, E("failed to update scheduler"))
		}
		outRows, err := query.New(tx).GetSchedulerByWorkspace(ctx, query.GetSchedulerByWorkspaceParams{
			UUID:          schUUID,
			WorkspaceUUID: workspaceUUID,
		})
		if err != nil {
			log.Error("failed to get updated scheduler", "error", err)
			return nil, ErrWithCode(http.StatusInternalServerError, E("failed to get updated scheduler"))
		}
		final, err := qToApiSchedulerByWorkspaceRow(outRows)
		if err != nil {
			log.Error("failed to map scheduler row after update", "error", err)
			return nil, ErrWithCode(http.StatusInternalServerError, E("failed to map scheduler row"))
		}
		return &final, nil
	})
}

// qToApiScheduler converts a query.Scheduler to an api.Scheduler.
func qToApiScheduler(s query.Scheduler) (api.Scheduler, error) {
	// Map fields from the query type to your API type.
	out := api.Scheduler{
		UUID:                api.NewOptString(s.UUID.String()),
		PipelineUUID:        s.PipelineUuid.String(),
		ScheduleType:        s.ScheduleType,
		CronExpression:      api.NewOptNilString(s.CronExpression.String),
		RunAt:               api.NewOptNilDateTime(s.RunAt.Time),
		Timezone:            api.NewOptString(s.Timezone),
		NextRun:             api.NewOptDateTime(s.NextRun.Time),
		LastRun:             api.NewOptDateTime(s.LastRun.Time),
		IsEnabled:           api.NewOptBool(s.IsEnabled),
		IsPaused:            api.NewOptBool(s.IsPaused),
		BatchSize:           api.NewOptInt(int(s.BatchSize)),
		SyncState:           schedulerSyncStateToAPI(s.SyncState),
		LastSyncTimestamp:   api.NewOptNilDateTime(s.LastSyncTimestamp.Time),
		OldestSyncTimestamp: api.NewOptNilDateTime(s.OldestSyncTimestamp.Time),
		CutoffDate:          api.NewOptNilDateTime(s.CutoffDate.Time),
		CreatedAt:           api.NewOptDateTime(s.CreatedAt.Time),
		UpdatedAt:           api.NewOptDateTime(s.UpdatedAt.Time),
	}
	return out, nil
}

// schedulerSyncStateToAPI converts a pg sync_state to API enum
func schedulerSyncStateToAPI(state pgtype.Text) api.OptSchedulerSyncState {
	if !state.Valid || state.String == "" {
		return api.NewOptSchedulerSyncState(api.SchedulerSyncStateInitial)
	}
	switch state.String {
	case "initial":
		return api.NewOptSchedulerSyncState(api.SchedulerSyncStateInitial)
	case "sync_recent":
		return api.NewOptSchedulerSyncState(api.SchedulerSyncStateSyncRecent)
	case "sync_historical":
		return api.NewOptSchedulerSyncState(api.SchedulerSyncStateSyncHistorical)
	case "sync_complete":
		return api.NewOptSchedulerSyncState(api.SchedulerSyncStateSyncComplete)
	default:
		return api.NewOptSchedulerSyncState(api.SchedulerSyncStateInitial)
	}
}

// qToApiSchedulerRow maps a row from GetSchedulers to an api.Scheduler.
func qToApiSchedulerRow(s query.GetSchedulerRow) (api.Scheduler, error) {
	// Map fields from the query type to your API type.
	out := api.Scheduler{
		UUID:                api.NewOptString(s.Scheduler.UUID.String()),
		PipelineUUID:        s.Scheduler.PipelineUuid.String(),
		ScheduleType:        s.Scheduler.ScheduleType,
		CronExpression:      api.NewOptNilString(s.Scheduler.CronExpression.String),
		RunAt:               api.NewOptNilDateTime(s.Scheduler.RunAt.Time),
		Timezone:            api.NewOptString(s.Scheduler.Timezone),
		NextRun:             api.NewOptDateTime(s.Scheduler.NextRun.Time),
		LastRun:             api.NewOptDateTime(s.Scheduler.LastRun.Time),
		IsEnabled:           api.NewOptBool(s.Scheduler.IsEnabled),
		IsPaused:            api.NewOptBool(s.Scheduler.IsPaused),
		BatchSize:           api.NewOptInt(int(s.Scheduler.BatchSize)),
		SyncState:           schedulerSyncStateToAPI(s.Scheduler.SyncState),
		LastSyncTimestamp:   api.NewOptNilDateTime(s.Scheduler.LastSyncTimestamp.Time),
		OldestSyncTimestamp: api.NewOptNilDateTime(s.Scheduler.OldestSyncTimestamp.Time),
		CutoffDate:          api.NewOptNilDateTime(s.Scheduler.CutoffDate.Time),
		CreatedAt:           api.NewOptDateTime(s.Scheduler.CreatedAt.Time),
		UpdatedAt:           api.NewOptDateTime(s.Scheduler.UpdatedAt.Time),
	}
	return out, nil
}

func qToApiSchedulersRow(s query.GetSchedulersRow) (api.Scheduler, error) {
	// Map fields from the query type to your API type.
	out := api.Scheduler{
		UUID:                api.NewOptString(s.UUID.String()),
		PipelineUUID:        s.PipelineUuid.String(),
		ScheduleType:        s.ScheduleType,
		CronExpression:      api.NewOptNilString(s.CronExpression.String),
		RunAt:               api.NewOptNilDateTime(s.RunAt.Time),
		Timezone:            api.NewOptString(s.Timezone),
		NextRun:             api.NewOptDateTime(s.NextRun.Time),
		LastRun:             api.NewOptDateTime(s.LastRun.Time),
		IsEnabled:           api.NewOptBool(s.IsEnabled),
		IsPaused:            api.NewOptBool(s.IsPaused),
		BatchSize:           api.NewOptInt(int(s.BatchSize)),
		SyncState:           schedulerSyncStateToAPI(s.SyncState),
		LastSyncTimestamp:   api.NewOptNilDateTime(s.LastSyncTimestamp.Time),
		OldestSyncTimestamp: api.NewOptNilDateTime(s.OldestSyncTimestamp.Time),
		CutoffDate:          api.NewOptNilDateTime(s.CutoffDate.Time),
		CreatedAt:           api.NewOptDateTime(s.CreatedAt.Time),
		UpdatedAt:           api.NewOptDateTime(s.UpdatedAt.Time),
	}
	return out, nil
}

func qToApiListSchedulersRow(s query.ListSchedulersRow) (api.Scheduler, error) {
	// Map fields from the query type to your API type.
	out := api.Scheduler{
		UUID:                api.NewOptString(s.Scheduler.UUID.String()),
		PipelineUUID:        s.Scheduler.PipelineUuid.String(),
		ScheduleType:        s.Scheduler.ScheduleType,
		CronExpression:      api.NewOptNilString(s.Scheduler.CronExpression.String),
		RunAt:               api.NewOptNilDateTime(s.Scheduler.RunAt.Time),
		Timezone:            api.NewOptString(s.Scheduler.Timezone),
		NextRun:             api.NewOptDateTime(s.Scheduler.NextRun.Time),
		LastRun:             api.NewOptDateTime(s.Scheduler.LastRun.Time),
		IsEnabled:           api.NewOptBool(s.Scheduler.IsEnabled),
		IsPaused:            api.NewOptBool(s.Scheduler.IsPaused),
		BatchSize:           api.NewOptInt(int(s.Scheduler.BatchSize)),
		SyncState:           schedulerSyncStateToAPI(s.Scheduler.SyncState),
		LastSyncTimestamp:   api.NewOptNilDateTime(s.Scheduler.LastSyncTimestamp.Time),
		OldestSyncTimestamp: api.NewOptNilDateTime(s.Scheduler.OldestSyncTimestamp.Time),
		CutoffDate:          api.NewOptNilDateTime(s.Scheduler.CutoffDate.Time),
		CreatedAt:           api.NewOptDateTime(s.Scheduler.CreatedAt.Time),
		UpdatedAt:           api.NewOptDateTime(s.Scheduler.UpdatedAt.Time),
	}
	return out, nil
}

// qToApiSchedulerByWorkspaceRow maps a workspace-filtered row to an api.Scheduler.
func qToApiSchedulerByWorkspaceRow(s query.GetSchedulerByWorkspaceRow) (api.Scheduler, error) {
	out := api.Scheduler{
		UUID:                api.NewOptString(s.Scheduler.UUID.String()),
		PipelineUUID:        s.Scheduler.PipelineUuid.String(),
		ScheduleType:        s.Scheduler.ScheduleType,
		CronExpression:      api.NewOptNilString(s.Scheduler.CronExpression.String),
		RunAt:               api.NewOptNilDateTime(s.Scheduler.RunAt.Time),
		Timezone:            api.NewOptString(s.Scheduler.Timezone),
		NextRun:             api.NewOptDateTime(s.Scheduler.NextRun.Time),
		LastRun:             api.NewOptDateTime(s.Scheduler.LastRun.Time),
		IsEnabled:           api.NewOptBool(s.Scheduler.IsEnabled),
		IsPaused:            api.NewOptBool(s.Scheduler.IsPaused),
		BatchSize:           api.NewOptInt(int(s.Scheduler.BatchSize)),
		SyncState:           schedulerSyncStateToAPI(s.Scheduler.SyncState),
		LastSyncTimestamp:   api.NewOptNilDateTime(s.Scheduler.LastSyncTimestamp.Time),
		OldestSyncTimestamp: api.NewOptNilDateTime(s.Scheduler.OldestSyncTimestamp.Time),
		CutoffDate:          api.NewOptNilDateTime(s.Scheduler.CutoffDate.Time),
		CreatedAt:           api.NewOptDateTime(s.Scheduler.CreatedAt.Time),
		UpdatedAt:           api.NewOptDateTime(s.Scheduler.UpdatedAt.Time),
	}
	return out, nil
}

// qToApiSchedulersWithWorkspaceRow maps a workspace-filtered schedulers row to an api.Scheduler.
func qToApiSchedulersWithWorkspaceRow(s query.GetSchedulersWithWorkspaceRow) (api.Scheduler, error) {
	out := api.Scheduler{
		UUID:                api.NewOptString(s.UUID.String()),
		PipelineUUID:        s.PipelineUuid.String(),
		ScheduleType:        s.ScheduleType,
		CronExpression:      api.NewOptNilString(s.CronExpression.String),
		RunAt:               api.NewOptNilDateTime(s.RunAt.Time),
		Timezone:            api.NewOptString(s.Timezone),
		NextRun:             api.NewOptDateTime(s.NextRun.Time),
		LastRun:             api.NewOptDateTime(s.LastRun.Time),
		IsEnabled:           api.NewOptBool(s.IsEnabled),
		IsPaused:            api.NewOptBool(s.IsPaused),
		BatchSize:           api.NewOptInt(int(s.BatchSize)),
		SyncState:           schedulerSyncStateToAPI(s.SyncState),
		LastSyncTimestamp:   api.NewOptNilDateTime(s.LastSyncTimestamp.Time),
		OldestSyncTimestamp: api.NewOptNilDateTime(s.OldestSyncTimestamp.Time),
		CutoffDate:          api.NewOptNilDateTime(s.CutoffDate.Time),
		CreatedAt:           api.NewOptDateTime(s.CreatedAt.Time),
		UpdatedAt:           api.NewOptDateTime(s.UpdatedAt.Time),
	}
	return out, nil
}
