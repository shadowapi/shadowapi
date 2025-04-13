// backend/internal/handler/worker_jobs.go
// Provides "read-only" and delete handlers for WorkerJobs,
// similar to how sync_policy is handled, but no create/update endpoints.

package handler

import (
	"context"
	"net/http"

	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/shadowapi/shadowapi/backend/internal/converter"
	"github.com/shadowapi/shadowapi/backend/pkg/api"
	"github.com/shadowapi/shadowapi/backend/pkg/query"
)

// WorkerJobsGet retrieves a specific worker job by uuid.
// GET /workerjobs/{uuid}
func (h *Handler) WorkerJobsGet(ctx context.Context, params api.WorkerJobsGetParams) (*api.WorkerJobs, error) {
	log := h.log.With("handler", "WorkerJobsGet")
	jobUUID, err := uuid.FromString(params.UUID)
	if err != nil {
		log.Error("invalid worker job uuid", "error", err)
		return nil, ErrWithCode(http.StatusBadRequest, E("invalid worker job uuid"))
	}

	q := query.New(h.dbp)
	row, err := q.GetWorkerJob(ctx, pgtype.UUID{Bytes: converter.UToBytes(jobUUID), Valid: true})
	if err != nil {
		log.Error("failed to get worker job", "error", err)
		return nil, ErrWithCode(http.StatusNotFound, E("worker job not found"))
	}

	res, mapErr := qToApiWorkerJobsRow(row.WorkerJob)
	if mapErr != nil {
		log.Error("failed to map worker job row", "error", mapErr)
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to map worker job row"))
	}
	return &res, nil
}

// WorkerJobsList retrieves a list of worker jobs.
// GET /workerjobs
func (h *Handler) WorkerJobsList(ctx context.Context, params api.WorkerJobsListParams) (*api.WorkerJobsListOK, error) {
	log := h.log.With("handler", "WorkerJobsList")

	limit := int32(50)
	offset := int32(0)
	if params.Limit.IsSet() {
		limit = params.Limit.Value
	}
	if params.Offset.IsSet() {
		offset = params.Offset.Value
	}

	q := query.New(h.dbp)
	rows, err := q.ListWorkerJobs(ctx, query.ListWorkerJobsParams{
		Offset: offset,
		Limit:  limit,
	})
	if err != nil {
		log.Error("failed to list worker jobs", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to list worker jobs"))
	}

	out := &api.WorkerJobsListOK{}
	for _, row := range rows {
		mapped, mapErr := qToApiWorkerJobsRow(row.WorkerJob)
		if mapErr != nil {
			log.Error("failed to map worker jobs row", "error", mapErr)
			return nil, ErrWithCode(http.StatusInternalServerError, E("failed to map worker job row"))
		}
		out.Jobs = append(out.Jobs, mapped)
	}
	return out, nil
}

// WorkerJobsDelete deletes a specific worker job by uuid.
// DELETE /workerjobs/{uuid}
func (h *Handler) WorkerJobsDelete(ctx context.Context, params api.WorkerJobsDeleteParams) error {
	log := h.log.With("handler", "WorkerJobsDelete")
	jobUUID, err := uuid.FromString(params.UUID)
	if err != nil {
		log.Error("invalid worker job uuid", "error", err)
		return ErrWithCode(http.StatusBadRequest, E("invalid worker job uuid"))
	}

	err = query.New(h.dbp).DeleteWorkerJob(ctx, pgtype.UUID{Bytes: converter.UToBytes(jobUUID), Valid: true})
	if err != nil {
		log.Error("failed to delete worker job", "error", err)
		return ErrWithCode(http.StatusNotFound, E("worker job not found or could not be deleted"))
	}
	return nil
}

// qToApiWorkerJobsRow converts a query.WorkerJob DB row into api.WorkerJobs object.
func qToApiWorkerJobsRow(dbRow query.WorkerJob) (api.WorkerJobs, error) {
	var res api.WorkerJobs
	// TODO @reactima finish conversion
	//res.UUID.SetTo(dbRow.JobID.Bytes)
	if dbRow.SchedulerUuid != nil {
		res.SchedulerUUID = dbRow.SchedulerUuid.String()
	}
	res.Subject = dbRow.Subject
	res.Status = dbRow.Status

	//if len(dbRow.Data) > 0 {
	//	var dataMap map[string]any
	//	if err := json.Unmarshal(dbRow.Data, &dataMap); err != nil {
	//		return res, err
	//	}
	//	opt := api.NewOptWorkerJobsData()
	//	opt.Value = dataMap
	//	res.Data = opt
	//}

	if dbRow.StartedAt.Valid {
		res.StartedAt.SetTo(dbRow.StartedAt.Time)
	}
	if dbRow.FinishedAt.Valid {
		res.FinishedAt.SetTo(dbRow.FinishedAt.Time)
	}

	return res, nil
}
