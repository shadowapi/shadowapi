// internal/worker/scheduler.go
package worker

import (
	"context"
	"time"

	"github.com/shadowapi/shadowapi/backend/internal/worker/jobs"
)

func (b *Broker) StartScheduler(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Minute)
	go func() {
		for {
			select {
			case <-ticker.C:
				if err := jobs.ScheduleEmailFetchJobs(ctx, b.log, b.dbp, b.queue); err != nil {
					b.log.Error("Scheduler error", "error", err)
				}
			case <-ctx.Done():
				ticker.Stop()
				return
			}
		}
	}()
}
