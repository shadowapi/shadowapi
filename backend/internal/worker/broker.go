package worker

import (
	"context"
	"github.com/shadowapi/shadowapi/backend/internal/worker/jobs"
	"github.com/shadowapi/shadowapi/backend/internal/worker/pipelines"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/samber/do/v2"

	"github.com/shadowapi/shadowapi/backend/internal/config"
	"github.com/shadowapi/shadowapi/backend/internal/queue"
	"github.com/shadowapi/shadowapi/backend/internal/worker/registry"
	"github.com/shadowapi/shadowapi/backend/internal/worker/types"
)

type Broker struct {
	ctx    context.Context
	cfg    *config.Config
	log    *slog.Logger
	dbp    *pgxpool.Pool
	queue  *queue.Queue
	cancel func()
}

// Provide creates and starts a new Broker.
func Provide(i do.Injector) (*Broker, error) {
	ctx := do.MustInvoke[context.Context](i)
	cfg := do.MustInvoke[*config.Config](i)
	dbp := do.MustInvoke[*pgxpool.Pool](i)
	log := do.MustInvoke[*slog.Logger](i).With("service", "broker")
	q := do.MustInvoke[*queue.Queue](i)

	b := &Broker{
		ctx:   ctx,
		cfg:   cfg,
		dbp:   dbp,
		log:   log,
		queue: q,
	}

	// Create the email pipeline.
	emailPipeline := pipelines.CreateEmailPipeline(ctx, log, dbp)
	// Register  pipeline job factory.
	registry.RegisterJob(registry.WorkerSubjectEmailFetch, jobs.EmailFetchJobFactory(log, emailPipeline, dbp, q))
	registry.RegisterJob(registry.WorkerSubjectEmailSync, jobs.EmailPipelineJobFactory(emailPipeline, q, log))
	registry.RegisterJob(registry.WorkerSubjectTokenRefresh, jobs.TokenRefresherJobFactory(dbp, log, q))

	if err := b.Start(ctx); err != nil {
		log.Error("failed to start broker", "error", err)
		return nil, err
	}
	b.StartScheduler(ctx)
	return b, nil
}

// Start ensures the stream exists and begins consuming messages.
func (b *Broker) Start(ctx context.Context) error {
	b.log.Debug("Broker starting", "stream", registry.WorkerStream)

	if err := b.queue.Ensure(ctx, registry.WorkerStream, registry.RegistrySubjects); err != nil {
		b.log.Error("failed to ensure stream", "error", err)
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
		b.log.Error("failed to start consumer", "error", err)
		return err
	}
	b.cancel = cancel
	return nil
}

// Shutdown stops the broker.
func (b *Broker) Shutdown(ctx context.Context) error {
	b.cancel()
	return nil
}

// handleMessages routes incoming messages to the appropriate job.
func (b *Broker) handleMessages(ctx context.Context) func(msg queue.Msg) {
	return func(msg queue.Msg) {
		b.log.Debug("Job received", "subject", msg.Subject())
		job, err := registry.CreateJob(msg.Subject(), msg.Data())
		if err != nil {
			if notReadyErr, ok := err.(types.JobNotReadyError); ok {
				b.log.Debug("Job not ready, requeuing", "delay", notReadyErr.Delay)
				_ = msg.NakWithDelay(notReadyErr.Delay)
				return
			}
			b.log.Error("failed to create job", "error", err)
			_ = msg.Term()
			return
		}
		if err := job.Execute(ctx); err != nil {
			if notReadyErr, ok := err.(types.JobNotReadyError); ok {
				b.log.Debug("Job not ready upon execution, requeuing", "delay", notReadyErr.Delay)
				_ = msg.NakWithDelay(notReadyErr.Delay)
				return
			}
			b.log.Error("job execution failed", "error", err)
			_ = msg.Term()
			return
		}
		if err := msg.Ack(); err != nil {
			b.log.Error("failed to acknowledge message", "error", err)
		}
	}
}

/*
// OLD CODE
const (
	workerStream              = "worker"
	workerSubject             = "worker.jobs"
	workerSubjectTokenRefresh = workerSubject + ".scheduleTokenRefresh"
)

type Broker struct {
	ctx   context.Context
	cfg   *config.Config
	log   *slog.Logger
	dbp   *pgxpool.Pool
	queue *queue.Queue
	csc   func()
}

// Provide broker service instance for dependency injection
func Provide(i do.Injector) (*Broker, error) {
	ctx := do.MustInvoke[context.Context](i)
	cfg := do.MustInvoke[*config.Config](i)
	dbp := do.MustInvoke[*pgxpool.Pool](i)
	log := do.MustInvoke[*slog.Logger](i).With("service", "broker")
	queue := do.MustInvoke[*queue.Queue](i)

	b := Broker{
		ctx:   ctx,
		cfg:   cfg,
		log:   log,
		dbp:   dbp,
		queue: queue,
	}

	if err := b.Start(ctx); err != nil {
		log.Error("failed to start queue service", "error", err)
		return nil, err
	}
	return &b, nil
}

// Start the broker service
func (b *Broker) Start(ctx context.Context) error {
	log := b.log.With("method", "Start")
	subjects := []string{
		workerSubjectTokenRefresh,
	}
	if err := b.queue.Ensure(ctx, workerStream, subjects); err != nil {
		log.Error("failed to ensure worker stream", "error", err)
		return err
	}

	log.Debug("start waiting jobs", "stream", workerStream)
	cancel, err := b.queue.Consume(
		ctx,
		workerStream,
		subjects,
		"worker-jobs",
		b.handleMessages(ctx, log),
	)
	if err != nil {
		log.Error("failed to get jobs", "error", err)
		return err
	}
	b.csc = cancel
	return nil
}

func (b *Broker) Shutdown(ctx context.Context) error {
	// cancel the consumer
	b.csc()
	return nil
}

// handleMessages handles the incoming messages
func (b *Broker) handleMessages(ctx context.Context, log *slog.Logger) func(msg queue.Msg) {
	return func(msg queue.Msg) {
		log.Debug("job received", "subject", msg.Subject())
		switch msg.Subject() {
		case workerSubjectTokenRefresh:
			b.tokenRefreshHandler(ctx, msg)
			return
		}
		log.Warn("unknown message subject, terminate it", "subject", msg.Subject())
		if err := msg.Term(); err != nil {
			log.Error("failed to terminate message", "error", err)
		}
	}
}

// ScheduleRefresh runs the token refresh worker after 30% of the expiration time
func (b *Broker) ScheduleRefresh(ctx context.Context, tokenUUID uuid.UUID, expiresAt time.Time) error {
	log := b.log.With("token_uuid", tokenUUID, "action", "scheduleTokenRefresh")
	duration := expiresAt.Sub(time.Now().UTC())
	duration = time.Duration(float64(duration) * 0.1)
	log.Info("schedule token refresh", "token_uuid", tokenUUID, "scheduled_at", time.Now().UTC().Add(duration))

	args := &tokenRefresherWorkerArgs{TokenUUID: tokenUUID, Expiry: time.Now().UTC().Add(duration)}
	msg, err := json.Marshal(args)
	if err != nil {
		log.Error("failed to marshal token refresh args", "error", err)
		return err
	}
	return b.queue.Publish(ctx, workerSubjectTokenRefresh, msg)
}

// tokenRefresherWorkerArgs for the token refresh worker
func (b *Broker) tokenRefreshHandler(ctx context.Context, msg queue.Msg) {
	log := b.log.With("method", "tokenRefreshHandler")
	args := &tokenRefresherWorkerArgs{}
	if err := json.Unmarshal(msg.Data(), args); err != nil {
		log.Error("failed to unmarshal token refresh args", "error", err)
		if err := msg.Term(); err != nil {
			log.Error("failed to terminate message", "error", err)
		}
		return
	}
	// This is a scheduled job, postpone it until the scheduled time
	if !time.Now().UTC().After(args.Expiry) {
		log.Debug("job is not ready yet, postpone it", "scheduled_at", args.Expiry)
		// NAK message with delay, we see it again after the scheduled time
		if err := msg.NakWithDelay(time.Until(args.Expiry)); err != nil {
			log.Error("failed to negative acknowledge message", "error", err)
		}
		return
	}
	w := tokenRefresherWorker{dbp: b.dbp, log: b.log}
	if err := w.Work(ctx, b, args); err != nil {
		if err := msg.Term(); err != nil {
			log.Error("failed to terminate message", "error", err)
		}
		return
	}
	if err := msg.Ack(); err != nil {
		log.Error("failed to acknowledge message", "error", err)
	}
}
*/
