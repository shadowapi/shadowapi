// internal/worker/scheduler.go
package scheduler

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shadowapi/shadowapi/backend/internal/queue"
	"github.com/shadowapi/shadowapi/backend/internal/worker/jobs"
	"time"

	"log/slog"
)

type Scheduler struct {
	log   *slog.Logger
	dbp   *pgxpool.Pool
	queue *queue.Queue
}

func NewScheduler(log *slog.Logger, dbp *pgxpool.Pool, q *queue.Queue) *Scheduler {
	return &Scheduler{
		log:   log,
		dbp:   dbp,
		queue: q,
	}
}

func (s *Scheduler) StartEmailScheduler(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Minute)
	go func() {
		for {
			select {
			case <-ticker.C:
				if err := jobs.ScheduleEmailFetchJobs(ctx, s.log, s.dbp, s.queue); err != nil {
					s.log.Error("Scheduler error", "error", err)
				}
			case <-ctx.Done():
				ticker.Stop()
				return
			}
		}
	}()
}
