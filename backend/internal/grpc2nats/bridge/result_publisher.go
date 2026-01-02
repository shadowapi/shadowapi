package bridge

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/samber/do/v2"

	"github.com/shadowapi/shadowapi/backend/internal/grpc2nats/config"
	"github.com/shadowapi/shadowapi/backend/internal/grpc2nats/natsconn"
)

// JobResult represents the result of a completed job
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

// ResultPublisher publishes job results to NATS
type ResultPublisher struct {
	log  *slog.Logger
	cfg  *config.Config
	conn *natsconn.Connection
}

// NewResultPublisher creates a new ResultPublisher
func NewResultPublisher(
	log *slog.Logger,
	cfg *config.Config,
	conn *natsconn.Connection,
) *ResultPublisher {
	return &ResultPublisher{
		log:  log.With("component", "result-publisher"),
		cfg:  cfg,
		conn: conn,
	}
}

// Provide creates a ResultPublisher for dependency injection
func ProvideResultPublisher(i do.Injector) (*ResultPublisher, error) {
	log := do.MustInvoke[*slog.Logger](i)
	cfg := do.MustInvoke[*config.Config](i)
	conn := do.MustInvoke[*natsconn.Connection](i)
	return NewResultPublisher(log, cfg, conn), nil
}

// Publish sends a job result to NATS
func (p *ResultPublisher) Publish(ctx context.Context, result *JobResult) error {
	subjects := p.cfg.Subjects()
	subject := subjects.Results(result.JobID)

	data, err := json.Marshal(result)
	if err != nil {
		p.log.Error("failed to marshal result", "job_id", result.JobID, "error", err)
		return err
	}

	headers := map[string]string{
		"X-Job-ID":    result.JobID,
		"X-Worker-ID": result.WorkerID,
		"X-Success":   boolToString(result.Success),
	}

	if err := p.conn.PublishWithHeaders(ctx, subject, headers, data); err != nil {
		p.log.Error("failed to publish result", "job_id", result.JobID, "error", err)
		return err
	}

	p.log.Debug("result published",
		"job_id", result.JobID,
		"worker_id", result.WorkerID,
		"success", result.Success,
		"subject", subject,
	)

	return nil
}

// PublishError publishes an error result for a job
func (p *ResultPublisher) PublishError(ctx context.Context, jobID, workerID, errMsg string) error {
	result := &JobResult{
		JobID:       jobID,
		WorkerID:    workerID,
		Success:     false,
		Error:       errMsg,
		CompletedAt: time.Now().UTC(),
	}
	return p.Publish(ctx, result)
}

func boolToString(b bool) string {
	if b {
		return "true"
	}
	return "false"
}
