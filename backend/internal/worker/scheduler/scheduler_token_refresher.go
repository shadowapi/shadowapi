// internal/worker/scheduler/scheduler_token_refresher.go
package scheduler

import (
	"context"
	"encoding/json"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/robfig/cron/v3"
	"github.com/shadowapi/shadowapi/backend/internal/queue"
	"github.com/shadowapi/shadowapi/backend/internal/worker/jobs"
	"github.com/shadowapi/shadowapi/backend/internal/worker/registry"
	"github.com/shadowapi/shadowapi/backend/pkg/query"
	"log/slog"
)

type TokenRefresherScheduler struct {
	log        *slog.Logger
	dbp        *pgxpool.Pool
	queue      *queue.Queue
	cronParser cron.Parser
	interval   time.Duration
}

func NewTokenRefresherScheduler(log *slog.Logger, dbp *pgxpool.Pool, q *queue.Queue) *TokenRefresherScheduler {
	return &TokenRefresherScheduler{
		log:        log,
		dbp:        dbp,
		queue:      q,
		cronParser: cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow),
		interval:   5 * time.Minute,
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
	// Query tokens expiring within the next 5 minutes.
	// TODO @reactive, cut off disabled
	tokens, err := queries.GetTokensToRefresh(ctx, nil)
	if err != nil {
		s.log.Error("Failed fetching tokens to refresh", "err", err)
		return
	}
	for _, tokenRow := range tokens {
		jobArgs := jobs.TokenRefresherJobArgs{
			TokenUUID: tokenRow.Oauth2Token.UUID,
			Expiry:    tokenRow.Oauth2Token.ExpiresAt.Time,
		}
		payload, err := json.Marshal(jobArgs)
		if err != nil {
			s.log.Error("Failed to marshal token refresher job payload", "token_uuid", tokenRow.Oauth2Token.UUID.String(), "err", err)
			continue
		}
		err = s.queue.Publish(ctx, registry.WorkerSubjectTokenRefresh, payload)
		if err != nil {
			s.log.Error("Failed to publish token refresher job", "token_uuid", tokenRow.Oauth2Token.UUID.String(), "err", err)
			continue
		}
		s.log.Info("Published token refresher job", "token_uuid", tokenRow.Oauth2Token.UUID.String())
	}
}
