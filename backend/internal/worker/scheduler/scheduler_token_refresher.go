package scheduler

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/robfig/cron/v3"

	"github.com/shadowapi/shadowapi/backend/internal/queue"
	"github.com/shadowapi/shadowapi/backend/internal/worker/jobs"
	"github.com/shadowapi/shadowapi/backend/internal/worker/monitor"
	"github.com/shadowapi/shadowapi/backend/internal/worker/subjects"
	"github.com/shadowapi/shadowapi/backend/pkg/query"
)

type TokenRefresherScheduler struct {
	log        *slog.Logger
	dbp        *pgxpool.Pool
	queue      *queue.Queue
	cronParser cron.Parser
	interval   time.Duration
	monitor    *monitor.WorkerMonitor
}

var defaultRefreshInterval = 5 * time.Minute

func NewTokenRefresherScheduler(log *slog.Logger, dbp *pgxpool.Pool, q *queue.Queue,
	monitor *monitor.WorkerMonitor) *TokenRefresherScheduler {
	return &TokenRefresherScheduler{
		log:        log,
		dbp:        dbp,
		queue:      q,
		monitor:    monitor,
		cronParser: cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow),
		interval:   defaultRefreshInterval,
	}
}

func (s *TokenRefresherScheduler) Start(ctx context.Context) {
	ticker := time.NewTicker(s.interval)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				s.run(ctx)
			case <-ctx.Done():
				s.log.Info("TokenRefresherScheduler shutting down")
				return
			}
		}
	}()
}

func (s *TokenRefresherScheduler) run(ctx context.Context) {
	queries := query.New(s.dbp)
	// TODO @reactive
	// - review
	// - cut off disabled
	// Query tokens expiring within the next 5 minutes
	tokens, err := queries.GetTokensToRefresh(ctx, nil)
	if err != nil {
		s.log.Error("Failed fetching tokens to refresh", "err", err)
		return
	}
	for _, tokenRow := range tokens {

		jobUUID := uuid.Must(uuid.NewV7()).String()

		schedulerUUID := uuid.Must(uuid.NewV7()).String()
		headers := queue.Headers{"X-Job-ID": jobUUID}

		// TODO @reactima
		// - fix schedulerUUID
		// - review expire
		jobArgs := jobs.TokenRefresherJobArgs{
			JobUUID:       jobUUID,
			SchedulerUUID: schedulerUUID,
			TokenUUID:     tokenRow.Oauth2Token.UUID,
			Expiry:        time.Now().Add(defaultRefreshInterval),
		}
		payload, err := json.Marshal(jobArgs)
		if err != nil {
			s.log.Error("Failed to marshal token refresher job payload", "token_uuid", tokenRow.Oauth2Token.UUID.String(), "err", err)
			continue
		}
		subject := subjects.GlobalJobSubject(subjects.JobTypeTokenRefresh)
		err = s.queue.PublishWithHeaders(ctx, subject, headers, payload)
		if err != nil {
			s.log.Error("Failed to publish token refresher job", "token_uuid", tokenRow.Oauth2Token.UUID.String(), "err", err)
			continue
		}
		s.log.Info("Published token refresher job", "token_uuid", tokenRow.Oauth2Token.UUID.String())
	}
}
