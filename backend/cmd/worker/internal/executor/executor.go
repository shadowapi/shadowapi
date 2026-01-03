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

// JobHandler executes a job and returns result data.
type JobHandler func(ctx context.Context, data []byte) (resultData []byte, err error)

// Executor manages job execution for the distributed worker.
type Executor struct {
	log        *slog.Logger
	client     *http.Client
	handlers   map[string]JobHandler
	activeJobs atomic.Int32
}

// Provide creates a new Executor for dependency injection.
func Provide(i do.Injector) (*Executor, error) {
	log := do.MustInvoke[*slog.Logger](i)

	e := &Executor{
		log:      log.With("component", "executor"),
		client:   &http.Client{},
		handlers: make(map[string]JobHandler),
	}
	e.registerHandlers()

	return e, nil
}

// registerHandlers registers all job handlers.
func (e *Executor) registerHandlers() {
	e.handlers["tokenRefresh"] = e.handleTokenRefresh
	e.handlers["testConnectionEmailOAuth"] = e.handleTestConnectionEmailOAuth
	e.handlers["testConnectionPostgres"] = e.handleTestConnectionPostgres
}

// Execute runs a job by its type and returns the result.
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

// ActiveJobs returns the current number of running jobs.
func (e *Executor) ActiveJobs() int32 {
	return e.activeJobs.Load()
}

// IsIdle returns true if no jobs are running.
func (e *Executor) IsIdle() bool {
	return e.activeJobs.Load() == 0
}
