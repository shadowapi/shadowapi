// backend/internal/handler/worker_jobs.go
// Provides "read-only" and delete handlers for WorkerJobs,
// similar to how sync_policy is handled, but no create/update endpoints.

package handler

import (
	"context"
	"encoding/json"
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

func qToApiWorkerJobsRow(dbRow query.WorkerJob) (api.WorkerJobs, error) {
	var res api.WorkerJobs

	// Convert the UUID to string and set it on WorkerJobs.UUID (which is an OptString).
	res.UUID.SetTo(dbRow.UUID.String())

	// If there's a scheduler UUID, convert it to string.
	if dbRow.SchedulerUuid != nil {
		res.SchedulerUUID = dbRow.SchedulerUuid.String()
	}
	if dbRow.JobUuid != nil {
		res.JobUUID = api.NewOptString(dbRow.JobUuid.String())
	}

	// Basic fields
	res.Subject = dbRow.Subject
	res.Status = dbRow.Status

	// If the Data field is present, unmarshal it into map[string]any (or map[string]jx.Raw).
	if len(dbRow.Data) > 0 {
		var dataMap map[string]interface{}
		if err := json.Unmarshal(dbRow.Data, &dataMap); err != nil {
			return res, err
		}
		// Use the “optional” type for data in WorkerJobs (OptWorkerJobsData).
		optData := api.NewOptWorkerJobsData(nil) // Initialize empty
		// WorkerJobsData is defined as map[string]jx.Raw, but you can store a normal map if needed.
		// Either re-marshal or directly use a cast, e.g. dataMap to map[string]jx.Raw.
		// For simplicity, you can re-marshal then unmarshal into map[string]jx.Raw:
		rawJSON, err := json.Marshal(dataMap)
		if err != nil {
			return res, err
		}
		var workerData api.WorkerJobsData
		if err := json.Unmarshal(rawJSON, &workerData); err != nil {
			return res, err
		}
		optData.Value = workerData
		optData.Set = true
		res.Data = optData
	}

	// Set started_at if valid
	if dbRow.StartedAt.Valid {
		res.StartedAt.SetTo(dbRow.StartedAt.Time)
	}
	// Set finished_at if valid
	if dbRow.FinishedAt.Valid {
		res.FinishedAt.SetTo(dbRow.FinishedAt.Time)
	}

	return res, nil
}
