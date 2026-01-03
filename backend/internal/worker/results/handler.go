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
	"golang.org/x/oauth2"

	"github.com/shadowapi/shadowapi/backend/internal/config"
	"github.com/shadowapi/shadowapi/backend/internal/converter"
	"github.com/shadowapi/shadowapi/backend/internal/jobstore"
	"github.com/shadowapi/shadowapi/backend/internal/queue"
	"github.com/shadowapi/shadowapi/backend/internal/worker/subjects"
	"github.com/shadowapi/shadowapi/backend/pkg/jobs"
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
	jobStore  *jobstore.Store
	cancel    func()
	consumeCC jetstream.ConsumeContext
}

// Provide creates a new result handler for dependency injection
func Provide(i do.Injector) (*Handler, error) {
	log := do.MustInvoke[*slog.Logger](i).With("component", "result-handler")
	cfg := do.MustInvoke[*config.Config](i)
	dbp := do.MustInvoke[*pgxpool.Pool](i)
	q := do.MustInvoke[*queue.Queue](i)
	js := do.MustInvoke[*jobstore.Store](i)

	return &Handler{
		log:      log,
		cfg:      cfg,
		dbp:      dbp,
		queue:    q,
		jobStore: js,
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

	ctx := context.Background()

	// Try to handle test connection results first (they use KV store, not worker_jobs table)
	if len(result.ResultData) > 0 && h.tryHandleTestConnectionResult(ctx, result) {
		h.log.Info("test connection result processed", "job_id", result.JobID)
		msg.Ack()
		return
	}

	// Handle token refresh results
	if refreshResult := isTokenRefreshResult(result.ResultData); refreshResult != nil {
		h.handleTokenRefreshResultDirect(ctx, refreshResult)
		msg.Ack()
		return
	}

	// Update the worker_jobs table for regular jobs
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

// isTokenRefreshResult checks if the result data is a token refresh result by trying to unmarshal it.
// Returns the unmarshalled result if successful, nil otherwise.
func isTokenRefreshResult(data []byte) *jobs.TokenRefreshResult {
	if len(data) == 0 {
		return nil
	}
	var result jobs.TokenRefreshResult
	if err := json.Unmarshal(data, &result); err != nil {
		return nil
	}
	// Token refresh results have a non-empty TokenUUID and no ResourceType field
	if result.TokenUUID != "" {
		return &result
	}
	return nil
}

// tryHandleTestConnectionResult attempts to handle a result as a test connection result.
// Returns true if it was a test connection result (regardless of success), false otherwise.
func (h *Handler) tryHandleTestConnectionResult(ctx context.Context, result JobResult) bool {
	var testResult jobs.TestConnectionResult
	if err := json.Unmarshal(result.ResultData, &testResult); err != nil {
		// Not a test connection result
		return false
	}

	// Check if this looks like a test connection result by verifying required fields
	if testResult.ResourceType == "" || testResult.JobUUID == "" {
		return false
	}

	// Valid resource types for test connections
	if testResult.ResourceType != "email_oauth" && testResult.ResourceType != "postgres" {
		return false
	}

	// Get existing job from KV store
	job, err := h.jobStore.Get(ctx, testResult.JobUUID)
	if err != nil {
		h.log.Error("failed to get test connection job from KV store",
			"job_uuid", testResult.JobUUID,
			"error", err,
		)
		// Still return true - we identified it as a test connection result
		return true
	}

	// Determine status based on success
	status := "completed"
	if !testResult.Success {
		status = "failed"
	}

	// Update job with result
	job.Status = status
	job.Result = result.ResultData
	job.CompletedAt = time.Now().UTC()

	// Store updated job back to KV
	if err := h.jobStore.Put(ctx, job); err != nil {
		h.log.Error("failed to update test connection job in KV store",
			"job_uuid", testResult.JobUUID,
			"error", err,
		)
		return true
	}

	h.log.Info("test connection result stored in KV",
		"job_uuid", testResult.JobUUID,
		"resource_type", testResult.ResourceType,
		"resource_uuid", testResult.ResourceUUID,
		"success", testResult.Success,
	)

	return true
}

// handleTokenRefreshResultDirect processes a pre-parsed token refresh job result
func (h *Handler) handleTokenRefreshResultDirect(ctx context.Context, refreshResult *jobs.TokenRefreshResult) {
	q := query.New(h.dbp)

	if !refreshResult.Success {
		// Token refresh failed - delete the broken token
		h.log.Warn("token refresh failed, deleting token", "token_uuid", refreshResult.TokenUUID)
		tokenUUID, err := uuid.FromString(refreshResult.TokenUUID)
		if err != nil {
			h.log.Error("invalid token UUID", "token_uuid", refreshResult.TokenUUID, "error", err)
			return
		}
		if err := q.DeleteOauth2Token(ctx, converter.UuidToPgUUID(tokenUUID)); err != nil {
			h.log.Error("failed to delete broken token", "token_uuid", refreshResult.TokenUUID, "error", err)
		}
		return
	}

	// Token refresh succeeded - update the token in the database
	tokenUUID, err := uuid.FromString(refreshResult.TokenUUID)
	if err != nil {
		h.log.Error("invalid token UUID", "token_uuid", refreshResult.TokenUUID, "error", err)
		return
	}

	// Build the oauth2.Token to store
	newToken := &oauth2.Token{
		AccessToken:  refreshResult.AccessToken,
		RefreshToken: refreshResult.RefreshToken,
		Expiry:       refreshResult.TokenExpiry,
		TokenType:    refreshResult.TokenType,
	}
	tokenData, err := json.Marshal(newToken)
	if err != nil {
		h.log.Error("failed to marshal new token", "error", err)
		return
	}

	// Update token in database
	if err := q.UpdateOauth2TokenData(ctx, query.UpdateOauth2TokenDataParams{
		UUID:      converter.UuidToPgUUID(tokenUUID),
		Token:     tokenData,
		ExpiresAt: converter.TimeToPgTimestamptz(refreshResult.TokenExpiry),
	}); err != nil {
		h.log.Error("failed to update token", "token_uuid", refreshResult.TokenUUID, "error", err)
		return
	}

	h.log.Info("token updated successfully", "token_uuid", refreshResult.TokenUUID, "new_expiry", refreshResult.TokenExpiry)

	// Schedule next token refresh
	h.scheduleNextTokenRefresh(ctx, tokenUUID, refreshResult.TokenExpiry)
}

// scheduleNextTokenRefresh schedules the next token refresh job
func (h *Handler) scheduleNextTokenRefresh(ctx context.Context, tokenUUID uuid.UUID, expiry time.Time) {
	// Schedule refresh before expiry (e.g., 5 minutes before)
	refreshTime := expiry.Add(-5 * time.Minute)
	if refreshTime.Before(time.Now()) {
		refreshTime = time.Now().Add(5 * time.Minute)
	}

	jobUUID := uuid.Must(uuid.NewV7()).String()
	schedulerUUID := uuid.Must(uuid.NewV7()).String()
	headers := queue.Headers{"X-Job-ID": jobUUID}

	// We need to load the token data again to get the full OAuth2 config
	// This is a simplified version - the scheduler will pick it up on next run
	// For now, just log that rescheduling would happen
	h.log.Info("next token refresh will be scheduled by scheduler",
		"token_uuid", tokenUUID.String(),
		"refresh_time", refreshTime,
	)

	// Note: The TokenRefresherScheduler runs every 5 minutes and will pick up
	// tokens that need refreshing based on their expiry time. We don't need
	// to explicitly schedule here - the scheduler will handle it.
	_ = jobUUID
	_ = schedulerUUID
	_ = headers
}

