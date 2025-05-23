package handler

import (
	"context"
	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/shadowapi/shadowapi/backend/internal/converter"
	"net/http"

	"github.com/shadowapi/shadowapi/backend/internal/db"
	"github.com/shadowapi/shadowapi/backend/pkg/api"
	"github.com/shadowapi/shadowapi/backend/pkg/query"
)

// SchedulerCreate implements scheduler-create operation.
//
// POST /scheduler
func (h *Handler) SchedulerCreate(ctx context.Context, req *api.Scheduler) (*api.Scheduler, error) {
	log := h.log.With("handler", "SchedulerCreate")
	return db.InTx(ctx, h.dbp, func(tx pgx.Tx) (*api.Scheduler, error) {
		// Generate a new UUID for the scheduler
		schedulerUUID := uuid.Must(uuid.NewV7())
		pgPipelineUUID, err := converter.ConvertStringToPgUUID(req.PipelineUUID)
		if err != nil {
			log.Error("failed to convert scheduler uuid", "error", err)
			return nil, ErrWithCode(http.StatusBadRequest, E("invalid datasource uuid"))
		}
		// Verify that the pipeline exists
		pipe, err := query.New(tx).GetPipeline(ctx, pgPipelineUUID)
		if err != nil {
			log.Error("failed to get pipeline", "error", err)
			return nil, ErrWithCode(http.StatusInternalServerError, E("failed to get pipeline"))
		}
		_ = pipe // Use pipe if needed

		// Build query parameters for the new scheduler.
		qParams := query.CreateSchedulerParams{
			UUID:           pgtype.UUID{Bytes: converter.UToBytes(schedulerUUID), Valid: true},
			PipelineUuid:   pgPipelineUUID,
			ScheduleType:   req.ScheduleType, // 'cron' or 'one_time'
			CronExpression: converter.ConvertOptNilStringToPgText(req.CronExpression),
			RunAt:          converter.NullTimestamptz(),
			Timezone:       req.Timezone.Or("UTC"),
			NextRun:        converter.NullTimestamptz(),
			LastRun:        converter.NullTimestamptz(),
			IsEnabled:      req.IsEnabled.Or(false),
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
func (h *Handler) SchedulerDelete(ctx context.Context, params api.SchedulerDeleteParams) error {
	log := h.log.With("handler", "SchedulerDelete")
	schUUID, err := converter.ConvertStringToPgUUID(params.UUID.String())
	if err != nil {
		log.Error("invalid scheduler uuid", "error", err)
		return ErrWithCode(http.StatusBadRequest, E("invalid scheduler uuid"))
	}
	err = query.New(h.dbp).DeleteScheduler(ctx, schUUID)
	if err != nil {
		log.Error("failed to delete scheduler", "error", err)
		return ErrWithCode(http.StatusInternalServerError, E("failed to delete scheduler"))
	}
	return nil
}

// SchedulerGet implements scheduler-get operation.
//
// GET /scheduler/{uuid}
func (h *Handler) SchedulerGet(ctx context.Context, params api.SchedulerGetParams) (*api.Scheduler, error) {
	log := h.log.With("handler", "SchedulerGet")
	schUUID, err := converter.ConvertStringToPgUUID(params.UUID.String())
	if err != nil {
		log.Error("invalid scheduler uuid", "error", err)
		return nil, ErrWithCode(http.StatusBadRequest, E("invalid scheduler uuid"))
	}
	sch, err := query.New(h.dbp).GetScheduler(ctx, schUUID)
	if err != nil {
		log.Error("failed to get scheduler", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to get scheduler"))
	}
	out, err := qToApiSchedulerRow(sch)
	if err != nil {
		log.Error("failed to map scheduler", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to map scheduler"))
	}
	return &out, nil
}

// SchedulerList implements scheduler-list operation.
//
// GET /scheduler
func (h *Handler) SchedulerList(ctx context.Context, params api.SchedulerListParams) ([]api.Scheduler, error) {
	log := h.log.With("handler", "SchedulerList")
	limit := int32(50)
	offset := int32(0)
	if params.Limit.IsSet() {
		limit = params.Limit.Value
	}
	if params.Offset.IsSet() {
		offset = params.Offset.Value
	}

	// Build filter parameters (if provided)
	qParams := query.GetSchedulersParams{
		OrderBy:        nil,
		OrderDirection: "desc",
		Offset:         offset,
		Limit:          limit,

		PipelineUuid: "", // set further below if query has ?pipeline_uuid
		IsEnabled:    -1, // -1 means "either"
		IsPaused:     -1, // FIX: use -1 so we don’t implicitly filter out paused rows
	}

	if params.PipelineUUID.IsSet() {
		qParams.PipelineUuid = params.PipelineUUID.Value.String()
	}

	schRows, err := query.New(h.dbp).GetSchedulers(ctx, qParams)
	//schRows, err := query.New(h.dbp).ListSchedulers(ctx, query.ListSchedulersParams{
	//	Offset: int32(0),
	//	Limit:  int32(50),
	//})
	if err != nil {
		log.Error("failed to list schedulers", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to list schedulers"))
	}

	out := make([]api.Scheduler, 0, len(schRows))
	for _, row := range schRows {
		apiItem, err := qToApiSchedulersRow(row)
		//apiItem, err := qToApiListSchedulersRow(row)
		if err != nil {
			log.Error("failed to map scheduler row", "error", err)
			return nil, ErrWithCode(http.StatusInternalServerError, E("failed to map scheduler row"))
		}
		out = append(out, apiItem)
	}
	return out, nil
}

// SchedulerUpdate implements scheduler-update operation.
//
// PUT /scheduler/{uuid}
func (h *Handler) SchedulerUpdate(ctx context.Context, req *api.Scheduler, params api.SchedulerUpdateParams) (*api.Scheduler, error) {
	log := h.log.With("handler", "SchedulerUpdate")
	schUUID, err := converter.ConvertStringToPgUUID(params.UUID.String())
	if err != nil {
		log.Error("invalid scheduler uuid", "error", err)
		return nil, ErrWithCode(http.StatusBadRequest, E("invalid scheduler uuid"))
	}
	return db.InTx(ctx, h.dbp, func(tx pgx.Tx) (*api.Scheduler, error) {
		scheduler, err := query.New(tx).GetScheduler(ctx, schUUID)
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

		uParams := query.UpdateSchedulerParams{
			CronExpression: converter.ConvertOptNilStringToPgText(req.CronExpression),
			RunAt:          converter.NullTimestamptz(),
			Timezone:       req.Timezone.Or("UTC"),
			NextRun:        converter.NullTimestamptz(),
			LastRun:        converter.NullTimestamptz(),
			IsEnabled:      isEnabled,
			UUID:           schUUID,
		}
		err = query.New(tx).UpdateScheduler(ctx, uParams)
		if err != nil {
			log.Error("failed to update scheduler", "error", err)
			return nil, ErrWithCode(http.StatusInternalServerError, E("failed to update scheduler"))
		}
		outRows, err := query.New(tx).GetScheduler(ctx, schUUID)
		if err != nil {
			log.Error("failed to get updated scheduler", "error", err)
			return nil, ErrWithCode(http.StatusInternalServerError, E("failed to get updated scheduler"))
		}
		final, err := qToApiSchedulerRow(outRows)
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
		UUID:           api.NewOptString(s.UUID.String()),
		PipelineUUID:   s.PipelineUuid.String(),
		ScheduleType:   s.ScheduleType,
		CronExpression: api.NewOptNilString(s.CronExpression.String),
		RunAt:          api.NewOptNilDateTime(s.RunAt.Time),
		Timezone:       api.NewOptString(s.Timezone),
		NextRun:        api.NewOptDateTime(s.NextRun.Time),
		LastRun:        api.NewOptDateTime(s.LastRun.Time),
		IsEnabled:      api.NewOptBool(s.IsEnabled),
		CreatedAt:      api.NewOptDateTime(s.CreatedAt.Time),
		UpdatedAt:      api.NewOptDateTime(s.UpdatedAt.Time),
	}
	return out, nil
}

// qToApiSchedulerRow maps a row from GetSchedulers to an api.Scheduler.
func qToApiSchedulerRow(s query.GetSchedulerRow) (api.Scheduler, error) {
	// Map fields from the query type to your API type.
	out := api.Scheduler{
		UUID:           api.NewOptString(s.Scheduler.UUID.String()),
		PipelineUUID:   s.Scheduler.PipelineUuid.String(),
		ScheduleType:   s.Scheduler.ScheduleType,
		CronExpression: api.NewOptNilString(s.Scheduler.CronExpression.String),
		RunAt:          api.NewOptNilDateTime(s.Scheduler.RunAt.Time),
		Timezone:       api.NewOptString(s.Scheduler.Timezone),
		NextRun:        api.NewOptDateTime(s.Scheduler.NextRun.Time),
		LastRun:        api.NewOptDateTime(s.Scheduler.LastRun.Time),
		IsEnabled:      api.NewOptBool(s.Scheduler.IsEnabled),
		CreatedAt:      api.NewOptDateTime(s.Scheduler.CreatedAt.Time),
		UpdatedAt:      api.NewOptDateTime(s.Scheduler.UpdatedAt.Time),
	}
	return out, nil
}

func qToApiSchedulersRow(s query.GetSchedulersRow) (api.Scheduler, error) {
	// Map fields from the query type to your API type.
	out := api.Scheduler{
		UUID:           api.NewOptString(s.UUID.String()),
		PipelineUUID:   s.PipelineUuid.String(),
		ScheduleType:   s.ScheduleType,
		CronExpression: api.NewOptNilString(s.CronExpression.String),
		RunAt:          api.NewOptNilDateTime(s.RunAt.Time),
		Timezone:       api.NewOptString(s.Timezone),
		NextRun:        api.NewOptDateTime(s.NextRun.Time),
		LastRun:        api.NewOptDateTime(s.LastRun.Time),
		IsEnabled:      api.NewOptBool(s.IsEnabled),
		CreatedAt:      api.NewOptDateTime(s.CreatedAt.Time),
		UpdatedAt:      api.NewOptDateTime(s.UpdatedAt.Time),
	}
	return out, nil
}

func qToApiListSchedulersRow(s query.ListSchedulersRow) (api.Scheduler, error) {
	// Map fields from the query type to your API type.
	out := api.Scheduler{
		UUID:           api.NewOptString(s.Scheduler.UUID.String()),
		PipelineUUID:   s.Scheduler.PipelineUuid.String(),
		ScheduleType:   s.Scheduler.ScheduleType,
		CronExpression: api.NewOptNilString(s.Scheduler.CronExpression.String),
		RunAt:          api.NewOptNilDateTime(s.Scheduler.RunAt.Time),
		Timezone:       api.NewOptString(s.Scheduler.Timezone),
		NextRun:        api.NewOptDateTime(s.Scheduler.NextRun.Time),
		LastRun:        api.NewOptDateTime(s.Scheduler.LastRun.Time),
		IsEnabled:      api.NewOptBool(s.Scheduler.IsEnabled),
		CreatedAt:      api.NewOptDateTime(s.Scheduler.CreatedAt.Time),
		UpdatedAt:      api.NewOptDateTime(s.Scheduler.UpdatedAt.Time),
	}
	return out, nil
}
