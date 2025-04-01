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
	"github.com/shadowapi/shadowapi/backend/internal/worker/scheduler"
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
	s := scheduler.NewScheduler(log, dbp, q)
	s.StartEmailScheduler(ctx)

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
