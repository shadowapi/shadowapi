package worker

import (
	"context"
	"log/slog"
	"os"
	"sync"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/samber/do/v2"

	"github.com/shadowapi/shadowapi/backend/internal/config"
	"github.com/shadowapi/shadowapi/backend/internal/queue"
	"github.com/shadowapi/shadowapi/backend/internal/worker/monitor"
	"github.com/shadowapi/shadowapi/backend/internal/worker/scheduler"
	"github.com/shadowapi/shadowapi/backend/internal/worker/subjects"
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

// Scheduler manages job scheduling and publishes to NATS for external workers.
// Jobs are processed by distributed workers via grpc2nats.
type Scheduler struct {
	ctx     context.Context
	cfg     *config.Config
	log     *slog.Logger
	dbp     *pgxpool.Pool
	queue   *queue.Queue
	monitor *monitor.WorkerMonitor
}

// Provide creates and starts the Scheduler.
func Provide(i do.Injector) (*Scheduler, error) {
	ctx := do.MustInvoke[context.Context](i)
	cfg := do.MustInvoke[*config.Config](i)
	dbp := do.MustInvoke[*pgxpool.Pool](i)
	log := do.MustInvoke[*slog.Logger](i).With("service", "scheduler")
	q := do.MustInvoke[*queue.Queue](i)

	monitoring := monitor.NewWorkerMonitor(log, dbp, false)

	s := &Scheduler{
		ctx:     ctx,
		cfg:     cfg,
		dbp:     dbp,
		log:     log,
		queue:   q,
		monitor: monitoring,
	}

	// Check if scheduler should be skipped
	cmdName := os.Args[len(os.Args)-1]
	if cmdName == "loader" {
		log.Info("detected loader command, scheduler disabled")
		return s, nil
	}

	if skipWorker := os.Getenv("BE_SKIP_WORKER"); skipWorker == "true" {
		log.Info("scheduler disabled via BE_SKIP_WORKER=true")
		return s, nil
	}

	// Ensure the jobs stream exists
	if err := q.Ensure(ctx, subjects.StreamName, subjects.AllJobSubjects()); err != nil {
		log.Error("failed to ensure jobs stream", "error", err)
		return nil, err
	}

	log.Info("starting job schedulers")

	// Start email scheduler
	emailScheduler := scheduler.NewMultiEmailScheduler(log, dbp, q, monitoring)
	emailScheduler.Start(ctx)

	// Start token refresh scheduler
	tokenScheduler := scheduler.NewTokenRefresherScheduler(log, dbp, q, monitoring)
	tokenScheduler.Start(ctx)

	log.Info("schedulers started, jobs will be processed by external workers")
	return s, nil
}

// Shutdown stops the scheduler.
func (s *Scheduler) Shutdown(ctx context.Context) error {
	s.log.Info("scheduler shutdown")
	return nil
}
