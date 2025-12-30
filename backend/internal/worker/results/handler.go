// Package results handles job result processing from distributed workers
package results

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/samber/do/v2"

	"github.com/shadowapi/shadowapi/backend/internal/config"
	"github.com/shadowapi/shadowapi/backend/internal/converter"
	"github.com/shadowapi/shadowapi/backend/internal/queue"
	"github.com/shadowapi/shadowapi/backend/internal/worker/subjects"
	"github.com/shadowapi/shadowapi/backend/pkg/query"
)

// JobResult represents a job completion message from grpc2nats
type JobResult struct {
	JobID       string    `json:"job_id"`
	WorkerID    string    `json:"worker_id"`
	Success     bool      `json:"success"`
	Error       string    `json:"error,omitempty"`
	ResultData  []byte    `json:"result_data,omitempty"`
	StartedAt   time.Time `json:"started_at"`
	CompletedAt time.Time `json:"completed_at"`
	DurationMs  int64     `json:"duration_ms"`
}

// Handler subscribes to job results and updates the database
type Handler struct {
	log       *slog.Logger
	cfg       *config.Config
	dbp       *pgxpool.Pool
	queue     *queue.Queue
	cancel    func()
	consumeCC jetstream.ConsumeContext
}

// Provide creates a new result handler for dependency injection
func Provide(i do.Injector) (*Handler, error) {
	log := do.MustInvoke[*slog.Logger](i).With("component", "result-handler")
	cfg := do.MustInvoke[*config.Config](i)
	dbp := do.MustInvoke[*pgxpool.Pool](i)
	q := do.MustInvoke[*queue.Queue](i)

	return &Handler{
		log:   log,
		cfg:   cfg,
		dbp:   dbp,
		queue: q,
	}, nil
}

// Start begins listening for job results
func (h *Handler) Start(ctx context.Context) error {
	h.log.Info("starting result handler", "pattern", subjects.AllResultsPattern())

	// Ensure the results stream exists
	if err := h.queue.Ensure(ctx, subjects.ResultsStreamName, subjects.AllResultSubjects()); err != nil {
		h.log.Error("failed to ensure results stream", "error", err)
		return err
	}

	// Subscribe to all result subjects
	cancel, err := h.queue.Consume(
		ctx,
		subjects.ResultsStreamName,
		subjects.AllResultSubjects(),
		"backend-results",
		h.handleResult,
	)
	if err != nil {
		h.log.Error("failed to subscribe to results", "error", err)
		return err
	}

	h.cancel = cancel
	h.log.Info("result handler started")
	return nil
}

// Stop stops the result handler
func (h *Handler) Stop() {
	if h.cancel != nil {
		h.cancel()
	}
	h.log.Info("result handler stopped")
}

// handleResult processes a job result message
func (h *Handler) handleResult(msg queue.Msg) {
	var result JobResult
	if err := json.Unmarshal(msg.Data(), &result); err != nil {
		h.log.Error("failed to unmarshal job result", "error", err)
		msg.Term()
		return
	}

	h.log.Debug("received job result",
		"job_id", result.JobID,
		"worker_id", result.WorkerID,
		"success", result.Success,
	)

	// Update the worker_jobs table
	ctx := context.Background()
	q := query.New(h.dbp)

	jobUUID, err := uuid.FromString(result.JobID)
	if err != nil {
		h.log.Error("invalid job UUID", "job_id", result.JobID, "error", err)
		msg.Term()
		return
	}

	// Find the worker_jobs record by job_uuid
	workerJob, err := q.GetWorkerJobByJobUUID(ctx, converter.UuidToPgUUID(jobUUID))
	if err != nil {
		h.log.Warn("worker job not found", "job_id", result.JobID, "error", err)
		// Job might have been created but we don't have a record for it yet
		// Just acknowledge and move on
		msg.Ack()
		return
	}

	// Determine status
	status := "done"
	if !result.Success {
		status = "failed"
	}

	// Build error data JSON
	var errorData []byte
	if result.Error != "" {
		errorData = []byte(`{"error":"` + result.Error + `"}`)
	}

	// Update job status
	err = q.UpdateWorkerJobStatus(ctx, query.UpdateWorkerJobStatusParams{
		UUID:       converter.UuidToPgUUID(workerJob.WorkerJob.UUID),
		Status:     status,
		FinishedAt: pgtype.Timestamptz{Time: result.CompletedAt, Valid: true},
		Data:       errorData,
	})
	if err != nil {
		h.log.Error("failed to update job status", "job_id", result.JobID, "error", err)
		// NAK to retry
		msg.Nak()
		return
	}

	h.log.Info("job result processed",
		"job_id", result.JobID,
		"status", status,
		"duration_ms", result.DurationMs,
	)

	msg.Ack()
}
