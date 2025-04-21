package worker

import (
	"context"
	"log/slog"
	"os"
	"strings"
	"sync"
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

var (
	jobCancelsMu sync.RWMutex
	jobCancels   = make(map[string]context.CancelFunc)
)

// registerCancel stores the cancel func under jobUUID,
// and removes it when the jobCtx is done.
func registerCancel(jobUUID string, cancel context.CancelFunc, jobCtx context.Context) {
	jobCancelsMu.Lock()
	jobCancels[jobUUID] = cancel
	jobCancelsMu.Unlock()

	go func() {
		<-jobCtx.Done()
		jobCancelsMu.Lock()
		delete(jobCancels, jobUUID)
		jobCancelsMu.Unlock()
	}()
}

// CancelJob cancels a running job by its UUID.
// Returns true if the job was found and cancelled.
func CancelJob(jobUUID string) bool {
	jobCancelsMu.RLock()
	cancel, ok := jobCancels[jobUUID]
	jobCancelsMu.RUnlock()
	if !ok {
		return false
	}
	cancel()
	return true
}

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

	// pipelinesMap is map of Datasource UUID to Pipeline
	// - can be constructed with different Contact extractor, different Storages (archived in S3, or in DB)
	//.- can have different filters (sync policies)
	// Pipeline is attached to Datasource
	// Datasource (email, whatsapp, etc) is attached to User
	pipelinesMap := pipelines.CreateEmailPipelines(ctx, log, dbp)

	// Register jobs without starting the broker
	registry.RegisterJob(registry.WorkerSubjectEmailOAuthFetch, jobs.EmailOAuthFetchJobFactory(dbp, log, q, monitoring, pipelinesMap))
	registry.RegisterJob(registry.WorkerSubjectEmailApplyPipeline, jobs.EmailPipelineMessageJobFactory(dbp, log, q, monitoring, pipelinesMap))
	registry.RegisterJob(registry.WorkerSubjectTokenRefresh, jobs.TokenRefresherJobFactory(dbp, log, q, monitoring))
	registry.RegisterJob(registry.WorkerSubjectDummy, jobs.DummyJobFactory(dbp, log, q, monitoring))

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

	// Start Schedulers one by one
	// MultiEmailScheduler will activate either email or email_oauth type of pipelines
	s := scheduler.NewMultiEmailScheduler(b.log, b.dbp, b.queue, b.monitor)
	s.Start(b.ctx)

	// TODO @reactima rethink
	// this will schedule to token refresh jobs even if no active pipelines
	tokenScheduler := scheduler.NewTokenRefresherScheduler(b.log, b.dbp, b.queue, b.monitor)
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
		jobID := msgHeaderToString(msg, "X-Job-ID")
		if jobID == "" {
			b.log.Error("Broker handleMessages missing X-Job-ID header", "subject", msg.Subject())
			_ = msg.Term()
			return
		}

		jobCtx, cancel := context.WithCancel(ctx)
		registerCancel(jobID, cancel, jobCtx)

		job, err := registry.CreateJob(msg.Subject(), jobID, msg.Data())
		if err != nil {
			if notReady, ok := err.(types.JobNotReadyError); ok {
				_ = msg.NakWithDelay(notReady.Delay)
				return
			}
			b.log.Error("Broker handleMessages failed to create job", "error", err)
			_ = msg.Term()
			return
		}

		if err := job.Execute(jobCtx); err != nil {
			duration := time.Since(start).Seconds()
			metrics.JobExecutedDuration.WithLabelValues(msg.Subject(), "failure").Observe(duration)
			if notReady, ok := err.(types.JobNotReadyError); ok {
				_ = msg.NakWithDelay(notReady.Delay)
				return
			}
			b.log.Error("Broker handleMessages job execution failed", "error", err)
			_ = msg.Term()
			return
		}

		duration := time.Since(start).Seconds()
		metrics.JobExecutedDuration.WithLabelValues(msg.Subject(), "success").Observe(duration)
		_ = msg.Ack()
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
