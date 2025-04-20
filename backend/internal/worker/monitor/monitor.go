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

// WorkerMonitor writes worker job lifecycle events to Postgres
// using both scheduler and job UUIDs for traceability
type WorkerMonitor struct {
	log    *slog.Logger
	dbp    *pgxpool.Pool
	phase2 bool
}

const (
	StatusRunning = "running"
	StatusDone    = "done"
	StatusFailed  = "failed"
)

// NewWorkerMonitor creates a new monitor instance
func NewWorkerMonitor(log *slog.Logger, dbp *pgxpool.Pool, phase2 bool) *WorkerMonitor {
	return &WorkerMonitor{log: log, dbp: dbp, phase2: phase2}
}

// RecordJobStart inserts a row with scheduler and job UUIDs and status running
func (wm *WorkerMonitor) RecordJobStart(ctx context.Context, schedulerUUID, jobUUID, subject string) {
	recordID := uuid.Must(uuid.NewV7())
	// convert scheduler UUID
	var schedID pgtype.UUID
	if schedulerUUID != "" {
		var err error
		schedID, err = converter.ConvertStringToPgUUID(schedulerUUID)
		if err != nil {
			wm.log.Error("invalid scheduler uuid", "error", err)
			schedID = pgtype.UUID{Valid: false}
		}
	}
	params := query.CreateWorkerJobParams{
		UUID:          converter.UuidToPgUUID(recordID),
		SchedulerUuid: schedID,
		JobUuid:       pgtype.UUID{Valid: false}, // jobID is created in job factory, nil means it didnt reach a factor
		Subject:       subject,
		Status:        StatusRunning,
		Data:          []byte("{}"),
		FinishedAt:    nullTime(),
	}
	if _, err := query.New(wm.dbp).CreateWorkerJob(ctx, params); err != nil {
		wm.log.Error("create worker job failed", "error", err)
	}
}

// RecordJobEnd updates the row with final status and optional error
func (wm *WorkerMonitor) RecordJobEnd(ctx context.Context, schedulerUUID, jobUUID, subject, finalStatus, errMsg string) {
	// convert scheduler UUID
	var schedID pgtype.UUID
	if schedulerUUID != "" {
		var err error
		schedID, err = converter.ConvertStringToPgUUID(schedulerUUID)
		if err != nil {
			wm.log.Error("invalid scheduler uuid", "error", err)
			schedID = pgtype.UUID{Valid: false}
		}
	}
	// convert job UUID
	jobID, err := converter.ConvertStringToPgUUID(jobUUID)
	if err != nil {
		wm.log.Error("invalid job uuid", "error", err)
		return
	}
	// prepare data payload
	data := map[string]string{}
	if errMsg != "" {
		data["error"] = errMsg
	}
	b, err := json.Marshal(data)
	if err != nil {
		wm.log.Error("failed to marshal error data", "error", err)
	}
	params := query.UpdateWorkerJobParams{
		SchedulerUuid: schedID,
		JobUuid:       pgtype.UUID{Valid: false},
		Subject:       subject,
		Status:        finalStatus,
		Data:          b,
		FinishedAt:    pgtype.Timestamptz{Time: time.Now().UTC(), Valid: true},
		UUID:          jobID,
	}
	if err := query.New(wm.dbp).UpdateWorkerJob(ctx, params); err != nil {
		wm.log.Error("update worker job failed", "error", err)
	}
}

// RecordInstant inserts a completed row immediately and returns its UUID
func (wm *WorkerMonitor) RecordInstant(ctx context.Context, schedulerUUID *uuid.UUID, subject string) string {
	id := uuid.Must(uuid.NewV7())
	var sched pgtype.UUID
	if schedulerUUID != nil {
		sched = converter.UuidToPgUUID(*schedulerUUID)
	}
	params := query.CreateWorkerJobParams{
		UUID:          converter.UuidToPgUUID(id),
		SchedulerUuid: sched,
		JobUuid:       converter.UuidToPgUUID(id),
		Subject:       subject,
		Status:        StatusDone,
		Data:          []byte("{}"),
		FinishedAt:    pgtype.Timestamptz{Time: time.Now().UTC(), Valid: true},
	}
	if _, err := query.New(wm.dbp).CreateWorkerJob(ctx, params); err != nil {
		wm.log.Error("create instant row failed", "error", err)
	}
	return id.String()
}

func nullTime() pgtype.Timestamptz { return pgtype.Timestamptz{Valid: false} }
