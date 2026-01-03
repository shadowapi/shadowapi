// Package executor provides job execution logic for the distributed worker.
// Jobs receive all necessary data in their payload and execute without database access.
package executor

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"sync/atomic"

	"github.com/samber/do/v2"
)

// DataRecord represents a record to be streamed during job execution
type DataRecord struct {
	Subject  string // NATS subject to publish to
	Data     []byte // JSON-encoded record data
	Sequence int64  // Sequence number within the job
}

// RecordSender is a callback function for streaming data records during job execution
type RecordSender func(record DataRecord) error

// JobHandler executes a job and returns result data.
type JobHandler func(ctx context.Context, data []byte) (resultData []byte, err error)

// StreamingJobHandler executes a job that can stream data records.
type StreamingJobHandler func(ctx context.Context, data []byte, sendRecord RecordSender) (resultData []byte, err error)

// Executor manages job execution for the distributed worker.
type Executor struct {
	log               *slog.Logger
	client            *http.Client
	handlers          map[string]JobHandler
	streamingHandlers map[string]StreamingJobHandler
	activeJobs        atomic.Int32
}

// Provide creates a new Executor for dependency injection.
func Provide(i do.Injector) (*Executor, error) {
	log := do.MustInvoke[*slog.Logger](i)

	e := &Executor{
		log:               log.With("component", "executor"),
		client:            &http.Client{},
		handlers:          make(map[string]JobHandler),
		streamingHandlers: make(map[string]StreamingJobHandler),
	}
	e.registerHandlers()

	return e, nil
}

// registerHandlers registers all job handlers.
func (e *Executor) registerHandlers() {
	e.handlers["tokenRefresh"] = e.handleTokenRefresh
	e.handlers["testConnectionEmailOAuth"] = e.handleTestConnectionEmailOAuth
	e.handlers["testConnectionPostgres"] = e.handleTestConnectionPostgres
	e.handlers["emailOAuthFetch"] = e.handleEmailFetch

	// Streaming handlers - these can send data records during execution
	e.streamingHandlers["messageQuery"] = e.handleMessageQuery
}

// IsStreamingJob returns true if the job type requires streaming execution
func (e *Executor) IsStreamingJob(jobType string) bool {
	_, ok := e.streamingHandlers[jobType]
	return ok
}

// Execute runs a job by its type and returns the result.
// For streaming jobs, use ExecuteStreaming instead.
func (e *Executor) Execute(ctx context.Context, jobType string, data []byte) ([]byte, error) {
	handler, ok := e.handlers[jobType]
	if !ok {
		return nil, fmt.Errorf("unknown job type: %s", jobType)
	}

	e.activeJobs.Add(1)
	defer e.activeJobs.Add(-1)

	e.log.Debug("executing job", "job_type", jobType)
	result, err := handler(ctx, data)
	if err != nil {
		e.log.Error("job execution failed", "job_type", jobType, "error", err)
	} else {
		e.log.Debug("job execution completed", "job_type", jobType)
	}
	return result, err
}

// ExecuteStreaming runs a streaming job that can send data records during execution.
func (e *Executor) ExecuteStreaming(ctx context.Context, jobType string, data []byte, sendRecord RecordSender) ([]byte, error) {
	handler, ok := e.streamingHandlers[jobType]
	if !ok {
		return nil, fmt.Errorf("unknown streaming job type: %s", jobType)
	}

	e.activeJobs.Add(1)
	defer e.activeJobs.Add(-1)

	e.log.Debug("executing streaming job", "job_type", jobType)
	result, err := handler(ctx, data, sendRecord)
	if err != nil {
		e.log.Error("streaming job execution failed", "job_type", jobType, "error", err)
	} else {
		e.log.Debug("streaming job execution completed", "job_type", jobType)
	}
	return result, err
}

// ActiveJobs returns the current number of running jobs.
func (e *Executor) ActiveJobs() int32 {
	return e.activeJobs.Load()
}

// IsIdle returns true if no jobs are running.
func (e *Executor) IsIdle() bool {
	return e.activeJobs.Load() == 0
}
