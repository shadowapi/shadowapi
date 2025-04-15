package worker

import (
	"context"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/samber/do/v2"
	"github.com/shadowapi/shadowapi/backend/internal/config"
	"github.com/shadowapi/shadowapi/backend/internal/metrics"
	"github.com/shadowapi/shadowapi/backend/internal/queue"
	"github.com/shadowapi/shadowapi/backend/internal/worker/jobs"
	"github.com/shadowapi/shadowapi/backend/internal/worker/monitor"
	"github.com/shadowapi/shadowapi/backend/internal/worker/pipelines"
	"github.com/shadowapi/shadowapi/backend/internal/worker/registry"
	"github.com/shadowapi/shadowapi/backend/internal/worker/scheduler"
	"github.com/shadowapi/shadowapi/backend/internal/worker/types"
)

// Broker routes messages to worker jobs.
type Broker struct {
	ctx     context.Context
	cfg     *config.Config
	log     *slog.Logger
	dbp     *pgxpool.Pool
	queue   *queue.Queue
	monitor *monitor.WorkerMonitor
	cancel  func()
}

// ProvideLazy creates a new Broker without starting it.
// Use this when you want to inject the broker but not have it automatically
// start consuming messages (like in the loader).
func ProvideLazy(i do.Injector) (*Broker, error) {
	ctx := do.MustInvoke[context.Context](i)
	cfg := do.MustInvoke[*config.Config](i)
	dbp := do.MustInvoke[*pgxpool.Pool](i)
	log := do.MustInvoke[*slog.Logger](i).With("service", "broker")
	q := do.MustInvoke[*queue.Queue](i)

	monitoring := monitor.NewWorkerMonitor(log, dbp, false) // phase2 = false for now

	log.Info("Creating broker in lazy mode (worker disabled)")

	b := &Broker{
		ctx:     ctx,
		cfg:     cfg,
		dbp:     dbp,
		log:     log,
		queue:   q,
		monitor: monitoring,
	}

	// Register jobs without starting the broker
	pipelinesMap := pipelines.CreateEmailPipelines(ctx, log, dbp)

	// TODO @reactima different users - different policies
	// pipelinesMap can be constructed with different Contact extractor, different Storages (archived in S3, or in DB), different filters (sync policies)
	// Current implementation is broken
	// Pipeline is attached to Datasource
	// Datasource (email, whatsapp, etc) is attached to the user	pipelinesMap := pipelines.CreateEmailPipelines(ctx, log, dbp)
	registry.RegisterJob(registry.WorkerSubjectEmailScheduledFetch, jobs.ScheduleEmailFetchJobFactory(dbp, log, q, pipelinesMap))
	registry.RegisterJob(registry.WorkerSubjectEmailApplyPipeline, jobs.EmailPipelineMessageJobFactory(dbp, log, q, pipelinesMap))
	registry.RegisterJob(registry.WorkerSubjectTokenRefresh, jobs.TokenRefresherJobFactory(dbp, log, q))

	return b, nil
}

// Provide creates and starts a new Broker.
// This maintains backward compatibility with existing code.
func Provide(i do.Injector) (*Broker, error) {
	// Check command name - never start worker for loader command
	cmdName := os.Args[len(os.Args)-1]
	if cmdName == "loader" {
		b, err := ProvideLazy(i)
		if err != nil {
			return nil, err
		}
		b.log.Info("Detected loader command, worker disabled automatically")
		return b, nil
	}

	// Check environment variable
	if skipWorker := os.Getenv("SA_SKIP_WORKER"); skipWorker == "true" {
		b, err := ProvideLazy(i)
		if err != nil {
			return nil, err
		}
		b.log.Info("Worker disabled via SA_SKIP_WORKER=true")
		return b, nil
	}

	// Standard path - create and start worker
	b, err := ProvideLazy(i)
	if err != nil {
		return nil, err
	}
	b.log.Info("Starting worker broker in normal mode")

	// Start the broker
	if err := b.Start(b.ctx); err != nil {
		b.log.Error("failed to start broker", "error", err)
		return nil, err
	}

	// Start the scheduler
	s := scheduler.NewMultiEmailScheduler(b.log, b.dbp, b.queue)
	s.Start(b.ctx)

	tokenScheduler := scheduler.NewTokenRefresherScheduler(b.log, b.dbp, b.queue)
	tokenScheduler.Start(b.ctx)

	return b, nil
}

// Start ensures the stream exists and begins consuming messages.
func (b *Broker) Start(ctx context.Context) error {
	b.log.Debug("Broker starting", "stream", registry.WorkerStream)

	if err := b.queue.Ensure(ctx, registry.WorkerStream, registry.RegistrySubjects); err != nil {
		b.log.Error("Failed to ensure stream", "error", err)
		return err
	}
	cancel, err := b.queue.Consume(
		ctx,
		registry.WorkerStream,
		registry.RegistrySubjects,
		"worker-jobs",
		b.handleMessages(ctx),
	)
	if err != nil {
		b.log.Error("Failed to start consumer", "error", err)
		return err
	}
	b.cancel = cancel
	return nil
}

// Shutdown stops the broker.
func (b *Broker) Shutdown(ctx context.Context) error {
	if b.cancel != nil {
		b.cancel()
	}
	return nil
}

// handleMessages routes incoming messages to the appropriate job.
func (b *Broker) handleMessages(ctx context.Context) func(msg queue.Msg) {
	return func(msg queue.Msg) {

		start := time.Now()

		// Extract a jobID from the message headers, fallback if not present
		jobID := msgHeaderToString(msg, "X-Job-ID")
		if jobID == "" {
			// e.g. fallback to a unique ID or do an error
			// For demonstration, let's do a random fallback
			jobID = "job-" + randomShortID()
		}

		b.monitor.RecordJobStart(ctx, jobID, msg.Subject())

		job, err := registry.CreateJob(msg.Subject(), msg.Data())
		if err != nil {
			if notReadyErr, ok := err.(types.JobNotReadyError); ok {
				_ = msg.NakWithDelay(notReadyErr.Delay)
				b.monitor.RecordJobEnd(ctx, jobID, msg.Subject(), "not_ready", notReadyErr.Error())
				return
			}
			b.log.Error("Failed to create job", "error", err)
			_ = msg.Term()
			b.monitor.RecordJobEnd(ctx, jobID, msg.Subject(), "failed", err.Error())
			return
		}
		if err := job.Execute(ctx); err != nil {
			duration := time.Since(start).Seconds()
			metrics.JobExecutedDuration.WithLabelValues(msg.Subject()).Observe(duration)

			if notReadyErr, ok := err.(types.JobNotReadyError); ok {
				_ = msg.NakWithDelay(notReadyErr.Delay)
				b.monitor.RecordJobEnd(ctx, jobID, msg.Subject(), "retry", notReadyErr.Error())
				return
			}
			b.log.Error("Job execution failed", "error", err)
			_ = msg.Term()
			b.monitor.RecordJobEnd(ctx, jobID, msg.Subject(), "failed", err.Error())
			return
		}
		duration := time.Since(start).Seconds()
		metrics.JobExecutedDuration.WithLabelValues(msg.Subject()).Observe(duration)
		_ = msg.Ack()

		b.monitor.RecordJobEnd(ctx, jobID, msg.Subject(), "completed", "")
	}
}

// msgHeaderToString attempts to extract the header value using the HeaderGetter interface.
func msgHeaderToString(m queue.Msg, key string) string {
	if h, ok := m.(queue.HeaderGetter); ok {
		return h.GetHeader(key)
	}
	return ""
}

// randomShortID returns a short identifier based on the current time.
func randomShortID() string {
	return strings.ReplaceAll(time.Now().Format("150405.000"), ".", "")
}
