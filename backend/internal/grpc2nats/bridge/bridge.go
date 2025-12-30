package bridge

import (
	"context"
	"log/slog"
	"strings"
	"sync"

	"github.com/nats-io/nats.go/jetstream"
	"github.com/samber/do/v2"

	"github.com/shadowapi/shadowapi/backend/internal/grpc2nats/config"
	"github.com/shadowapi/shadowapi/backend/internal/grpc2nats/manager"
	"github.com/shadowapi/shadowapi/backend/internal/grpc2nats/natsconn"
)

// Bridge orchestrates the flow between NATS and gRPC workers
type Bridge struct {
	log        *slog.Logger
	cfg        *config.Config
	conn       *natsconn.Connection
	manager    *manager.WorkerManager
	router     *manager.Router
	dispatcher *JobDispatcher
	publisher  *ResultPublisher

	consumeCtx jetstream.ConsumeContext
	mu         sync.Mutex
	running    bool
}

// Provide creates a Bridge for dependency injection
func Provide(i do.Injector) (*Bridge, error) {
	log := do.MustInvoke[*slog.Logger](i).With("component", "bridge")
	cfg := do.MustInvoke[*config.Config](i)
	conn := do.MustInvoke[*natsconn.Connection](i)
	mgr := do.MustInvoke[*manager.WorkerManager](i)

	router := manager.NewRouter(mgr)
	publisher := NewResultPublisher(log, cfg, conn)
	dispatcher := NewJobDispatcher(log, cfg, router, publisher)

	return &Bridge{
		log:        log,
		cfg:        cfg,
		conn:       conn,
		manager:    mgr,
		router:     router,
		dispatcher: dispatcher,
		publisher:  publisher,
	}, nil
}

// Start begins subscribing to NATS for jobs
func (b *Bridge) Start(ctx context.Context) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.running {
		return nil
	}

	subjects := b.cfg.Subjects()

	// Ensure the jobs stream exists
	streamName := "jobs"
	streamSubjects := []string{subjects.JobsAll()}

	_, err := b.conn.EnsureStream(ctx, streamName, streamSubjects)
	if err != nil {
		return err
	}

	b.log.Info("subscribing to job subjects", "pattern", subjects.JobsAll())

	// Subscribe to all job subjects
	cc, err := b.conn.Subscribe(
		ctx,
		streamName,
		[]string{subjects.JobsAll()},
		"grpc2nats-jobs",
		b.handleJob,
	)
	if err != nil {
		return err
	}

	b.consumeCtx = cc
	b.running = true

	b.log.Info("bridge started", "instance_id", b.cfg.InstanceID)
	return nil
}

// Stop stops the bridge
func (b *Bridge) Stop() {
	b.mu.Lock()
	defer b.mu.Unlock()

	if !b.running {
		return
	}

	if b.consumeCtx != nil {
		b.consumeCtx.Stop()
	}

	b.running = false
	b.log.Info("bridge stopped")
}

// handleJob processes incoming job messages from NATS
func (b *Bridge) handleJob(msg jetstream.Msg) {
	ctx := context.Background()

	// Parse subject to extract job info
	// Format: {prefix}.jobs.{global|workspace.{slug}}.{job_type}
	subject := msg.Subject()
	parts := strings.Split(subject, ".")

	if len(parts) < 4 {
		b.log.Warn("invalid job subject format", "subject", subject)
		msg.Term()
		return
	}

	// Extract job metadata from headers
	jobID := msg.Headers().Get("X-Job-ID")
	if jobID == "" {
		b.log.Warn("job missing X-Job-ID header", "subject", subject)
		msg.Term()
		return
	}

	// Parse the subject structure
	// {prefix}.jobs.global.{job_type} or {prefix}.jobs.workspace.{slug}.{job_type}
	var workspaceSlug string
	var jobType string
	var isGlobal bool

	// Skip prefix and "jobs"
	jobPath := parts[2:]
	if len(jobPath) >= 2 && jobPath[0] == "global" {
		isGlobal = true
		jobType = strings.Join(jobPath[1:], ".")
	} else if len(jobPath) >= 3 && jobPath[0] == "workspace" {
		workspaceSlug = jobPath[1]
		jobType = strings.Join(jobPath[2:], ".")
	} else {
		b.log.Warn("cannot parse job subject", "subject", subject, "path", jobPath)
		msg.Term()
		return
	}

	b.log.Debug("received job",
		"job_id", jobID,
		"job_type", jobType,
		"workspace", workspaceSlug,
		"is_global", isGlobal,
	)

	// Dispatch the job
	err := b.dispatcher.Dispatch(ctx, &Job{
		ID:            jobID,
		Type:          jobType,
		WorkspaceSlug: workspaceSlug,
		IsGlobal:      isGlobal,
		Data:          msg.Data(),
	})

	if err != nil {
		if err == manager.ErrNoWorkersAvailable {
			// NAK with delay to retry later
			b.log.Debug("no workers available, will retry",
				"job_id", jobID,
				"job_type", jobType,
			)
			msg.NakWithDelay(5_000_000_000) // 5 seconds
			return
		}

		b.log.Error("failed to dispatch job",
			"job_id", jobID,
			"error", err,
		)
		msg.Term()
		return
	}

	// Job successfully dispatched
	msg.Ack()
}

// Job represents a job to be dispatched to a worker
type Job struct {
	ID            string
	Type          string
	WorkspaceSlug string
	IsGlobal      bool
	Data          []byte
}
