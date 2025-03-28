// internal/worker/jobs/email_pipeline.go
package jobs

import (
	"context"
	"encoding/json"
	"time"

	"github.com/gofrs/uuid"
	"github.com/shadowapi/shadowapi/backend/internal/queue"
	"github.com/shadowapi/shadowapi/backend/internal/worker"
	"github.com/shadowapi/shadowapi/backend/pkg/api"
	"log/slog"
)

// EmailPipelineJobArgs contains the incoming job data.
type EmailPipelineJobArgs struct {
	// Raw JSON of the email message (e.g. as produced by a fetcher).
	MessageData json.RawMessage `json:"message_data"`
	// The account’s OAuth token UUID as a string.
	TokenUUID string `json:"token_uuid"`
	// The token’s expiry time.
	TokenExpiry time.Time `json:"token_expiry"`
}

// EmailPipelineJob implements the worker.Job interface.
type EmailPipelineJob struct {
	// Save the raw message data so we can process it.
	MessageData json.RawMessage
	// The pipeline to process the email.
	Pipeline worker.Pipeline
	// Token information for the account.
	TokenUUID   uuid.UUID
	TokenExpiry time.Time
	// A reference to the Queue (used for publishing scheduled token refresh messages).
	Queue *queue.Queue
	// Logger for this job.
	Log *slog.Logger
}

// NewEmailPipelineJob creates a new EmailPipelineJob instance.
func NewEmailPipelineJob(p worker.Pipeline, args EmailPipelineJobArgs, q *queue.Queue, log *slog.Logger) (*EmailPipelineJob, error) {
	tUUID, err := uuid.FromString(args.TokenUUID)
	if err != nil {
		return nil, err
	}
	return &EmailPipelineJob{
		MessageData: args.MessageData,
		Pipeline:    p,
		TokenUUID:   tUUID,
		TokenExpiry: args.TokenExpiry,
		Queue:       q,
		Log:         log,
	}, nil
}

// Execute unmarshals the email message, runs it through the pipeline,
// and if the token is near expiry, schedules a refresh.
func (e *EmailPipelineJob) Execute(ctx context.Context) error {
	var msg api.Message
	if err := json.Unmarshal(e.MessageData, &msg); err != nil {
		e.Log.Error("failed to unmarshal email message", "error", err)
		return err
	}
	e.Log.Info("Processing email message", "message_uuid", msg.UUID)
	// Process the message through the pipeline.
	if err := e.Pipeline.Run(ctx, &msg); err != nil {
		e.Log.Error("pipeline execution failed", "error", err)
		return err
	}
	// Check if the token is near expiry (e.g. less than 30 minutes remaining).
	if time.Until(e.TokenExpiry) < 30*time.Minute {
		e.Log.Info("Token nearing expiry, scheduling refresh", "token_uuid", e.TokenUUID.String())
		// Schedule the token refresh job.
		if err := ScheduleTokenRefresh(ctx, e.Queue, e.TokenUUID, e.TokenExpiry, e.Log); err != nil {
			e.Log.Error("failed to schedule token refresh", "error", err)
			return err
		}
	}
	return nil
}

// EmailPipelineJobFactory returns a worker.JobFactory for email pipeline jobs.
func EmailPipelineJobFactory(p worker.Pipeline, q *queue.Queue, log *slog.Logger) worker.JobFactory {
	return func(data []byte) (worker.Job, error) {
		var args EmailPipelineJobArgs
		if err := json.Unmarshal(data, &args); err != nil {
			return nil, err
		}
		return NewEmailPipelineJob(p, args, q, log)
	}
}
