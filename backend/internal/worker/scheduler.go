// internal/worker/scheduler.go
package worker

import (
	"context"
	"encoding/json"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shadowapi/shadowapi/backend/internal/queue"
	"github.com/shadowapi/shadowapi/backend/internal/worker/jobs"
	"github.com/shadowapi/shadowapi/backend/internal/worker/registry"
	"github.com/shadowapi/shadowapi/backend/pkg/query"
	"log/slog"
)

func ScheduleEmailFetchJobs(ctx context.Context, log *slog.Logger, dbp *pgxpool.Pool, q *queue.Queue) error {
	queries := query.New(dbp)
	// Use a pgtype.UUID with Valid set to false for an empty UUID.
	emptyUUID := pgtype.UUID{Valid: false}
	params := query.GetDatasourcesParams{
		OrderBy:        "created_at",
		OrderDirection: "asc",
		Offset:         0,
		Limit:          100,
		UUID:           emptyUUID,
		UserUUID:       emptyUUID,
		Type:           "email",
		Provider:       "",
		IsEnabled:      true,
		Name:           "",
	}
	ds, err := queries.GetDatasources(ctx, params)
	if err != nil {
		log.Error("Failed to get datasources", "error", err)
		return err
	}
	for _, d := range ds {
		// Use the IsEnabled field directly from the returned row.
		if !d.IsEnabled {
			continue
		}
		accountID := d.UUID.String()
		jobArgs := jobs.EmailFetchJobArgs{
			AccountID:   accountID,
			LastFetched: time.Now().Add(-1 * time.Hour),
		}
		data, err := json.Marshal(jobArgs)
		if err != nil {
			log.Error("Failed to marshal email fetch job args", "error", err, "account", accountID)
			continue
		}
		if err := q.Publish(ctx, registry.WorkerSubjectEmailFetch, data); err != nil {
			log.Error("Failed to publish email fetch job", "error", err, "account", accountID)
			continue
		}
		log.Info("Scheduled email fetch job", "account", accountID)
	}
	return nil
}

func (b *Broker) StartScheduler(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Minute)
	go func() {
		for {
			select {
			case <-ticker.C:
				if err := ScheduleEmailFetchJobs(ctx, b.log, b.dbp, b.queue); err != nil {
					b.log.Error("Scheduler error", "error", err)
				}
			case <-ctx.Done():
				ticker.Stop()
				return
			}
		}
	}()
}
