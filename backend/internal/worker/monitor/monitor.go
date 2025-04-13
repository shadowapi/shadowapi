package monitor

import (
	"context"
	"encoding/json"
	"time"

	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shadowapi/shadowapi/backend/internal/converter"
	"github.com/shadowapi/shadowapi/backend/pkg/query"
	"log/slog"
)

// WorkerMonitor records job status in Postgres.
type WorkerMonitor struct {
	log    *slog.Logger
	dbp    *pgxpool.Pool
	phase2 bool
}

// WorkerJobStatus defines a simple status representation for UI display.
type WorkerJobStatus struct {
	JobID       string                 `json:"job_id"`
	SchedulerID string                 `json:"scheduler_id,omitempty"`
	Subject     string                 `json:"subject"`
	Status      string                 `json:"status"`
	Data        map[string]interface{} `json:"data,omitempty"`
	StartedAt   time.Time              `json:"started_at"`
	FinishedAt  time.Time              `json:"finished_at,omitempty"`
}

// NewWorkerMonitor creates a new monitor service.
func NewWorkerMonitor(log *slog.Logger, dbp *pgxpool.Pool, phase2 bool) *WorkerMonitor {
	return &WorkerMonitor{
		log:    log,
		dbp:    dbp,
		phase2: phase2,
	}
}

// RecordJobStart creates a new record in worker_jobs with status "running".
func (wm *WorkerMonitor) RecordJobStart(ctx context.Context, jobID, subject string) {
	wm.log.Info("Monitor: job started", "jobID", jobID, "subject", subject)

	q := query.New(wm.dbp)
	parsedUUID, err := uuid.FromString(jobID)
	if err != nil {
		wm.log.Error("RecordJobStart: invalid job UUID", "error", err)
		return
	}

	params := query.CreateWorkerJobParams{
		UUID:          converter.UuidToPgUUID(parsedUUID),
		SchedulerUuid: pgtype.UUID{Valid: false},
		Subject:       subject,
		Status:        "running",
		Data:          []byte("{}"),
		FinishedAt:    NullTimestamptz(), // helper to return a null timestamptz
	}
	_, err = q.CreateWorkerJob(ctx, params)
	if err != nil {
		wm.log.Error("RecordJobStart: failed to create worker job", "error", err)
	}
}

// RecordJobEnd updates the worker_jobs record with the final status.
func (wm *WorkerMonitor) RecordJobEnd(ctx context.Context, jobID, subject, finalStatus, errMsg string) {
	wm.log.Info("Monitor: job ended", "jobID", jobID, "status", finalStatus)

	q := query.New(wm.dbp)
	parsedUUID, err := uuid.FromString(jobID)
	if err != nil {
		wm.log.Error("RecordJobEnd: invalid job UUID", "error", err)
		return
	}

	dataMap := map[string]string{}
	if errMsg != "" {
		dataMap["error"] = errMsg
	}
	dataBytes, _ := json.Marshal(dataMap)

	params := query.UpdateWorkerJobParams{
		Subject:    subject,
		Status:     finalStatus,
		Data:       dataBytes,
		FinishedAt: pgtype.Timestamptz{Time: time.Now().UTC(), Valid: true},
		UUID:       converter.UuidToPgUUID(parsedUUID),
	}
	err = q.UpdateWorkerJob(ctx, params)
	if err != nil {
		wm.log.Error("RecordJobEnd: failed to update worker job", "error", err)
	}
}

// ListRecentJobs returns the most recent jobs for UI display.
func (wm *WorkerMonitor) ListRecentJobs(ctx context.Context, limit int) ([]WorkerJobStatus, error) {
	q := query.New(wm.dbp)
	params := query.ListWorkerJobsParams{
		Offset: 0,
		Limit:  int32(limit),
	}
	rows, err := q.ListWorkerJobs(ctx, params)
	if err != nil {
		return nil, err
	}
	var statuses []WorkerJobStatus
	for _, row := range rows {
		wj := row.WorkerJob
		// Convert pgtype.UUID to a uuid.UUID for string display.
		parsed, err := uuid.FromBytes(wj.JobID.Bytes[:])
		var jobIDStr string
		if err != nil {
			jobIDStr = ""
		} else {
			jobIDStr = parsed.String()
		}
		status := WorkerJobStatus{
			JobID:     jobIDStr,
			Subject:   wj.Subject,
			Status:    wj.Status,
			StartedAt: wj.StartedAt.Time,
		}
		// Check if SchedulerUuid is non-nil.
		if wj.SchedulerUuid != nil {
			status.SchedulerID = wj.SchedulerUuid.String()
		}
		if wj.FinishedAt.Valid {
			status.FinishedAt = wj.FinishedAt.Time
		}
		if len(wj.Data) > 0 {
			var data map[string]interface{}
			if err := json.Unmarshal(wj.Data, &data); err == nil {
				status.Data = data
			}
		}
		statuses = append(statuses, status)
	}
	return statuses, nil
}

// NullTimestamptz returns a null pgtype.Timestamptz.
func NullTimestamptz() pgtype.Timestamptz {
	return pgtype.Timestamptz{Valid: false}
}
