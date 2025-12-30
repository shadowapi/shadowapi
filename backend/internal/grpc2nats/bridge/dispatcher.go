package bridge

import (
	"context"
	"log/slog"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/shadowapi/shadowapi/backend/internal/grpc2nats/config"
	"github.com/shadowapi/shadowapi/backend/internal/grpc2nats/manager"
	workerv1 "github.com/shadowapi/shadowapi/backend/pkg/proto/worker/v1"
)

// JobDispatcher dispatches jobs to workers
type JobDispatcher struct {
	log       *slog.Logger
	cfg       *config.Config
	router    *manager.Router
	publisher *ResultPublisher
}

// NewJobDispatcher creates a new JobDispatcher
func NewJobDispatcher(
	log *slog.Logger,
	cfg *config.Config,
	router *manager.Router,
	publisher *ResultPublisher,
) *JobDispatcher {
	return &JobDispatcher{
		log:       log.With("component", "dispatcher"),
		cfg:       cfg,
		router:    router,
		publisher: publisher,
	}
}

// Dispatch sends a job to an available worker
func (d *JobDispatcher) Dispatch(ctx context.Context, job *Job) error {
	// Find a suitable worker
	worker, err := d.router.RouteJob(job.WorkspaceSlug, job.IsGlobal)
	if err != nil {
		return err
	}

	// Create job assignment
	deadline := time.Now().Add(5 * time.Minute) // Default 5 minute timeout
	assignment := &workerv1.JobAssignment{
		JobId:         job.ID,
		JobType:       job.Type,
		WorkspaceSlug: job.WorkspaceSlug,
		JobData:       job.Data,
		Deadline:      timestamppb.New(deadline),
	}

	d.log.Info("dispatching job",
		"job_id", job.ID,
		"job_type", job.Type,
		"worker_id", worker.ID,
		"worker_name", worker.Name,
		"workspace", job.WorkspaceSlug,
	)

	// Send job to worker
	if err := worker.SendJob(assignment); err != nil {
		d.log.Error("failed to send job to worker",
			"job_id", job.ID,
			"worker_id", worker.ID,
			"error", err,
		)
		return err
	}

	return nil
}

// DispatchToWorker sends a job to a specific worker
func (d *JobDispatcher) DispatchToWorker(ctx context.Context, job *Job, workerID string) error {
	worker, err := d.router.RouteToWorker(workerID)
	if err != nil {
		return err
	}

	deadline := time.Now().Add(5 * time.Minute)
	assignment := &workerv1.JobAssignment{
		JobId:         job.ID,
		JobType:       job.Type,
		WorkspaceSlug: job.WorkspaceSlug,
		JobData:       job.Data,
		Deadline:      timestamppb.New(deadline),
	}

	d.log.Info("dispatching job to specific worker",
		"job_id", job.ID,
		"job_type", job.Type,
		"worker_id", worker.ID,
	)

	return worker.SendJob(assignment)
}
