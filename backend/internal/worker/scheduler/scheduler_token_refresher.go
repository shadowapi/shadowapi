package scheduler

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/robfig/cron/v3"
	"golang.org/x/oauth2"

	"github.com/shadowapi/shadowapi/backend/internal/converter"
	"github.com/shadowapi/shadowapi/backend/internal/queue"
	"github.com/shadowapi/shadowapi/backend/internal/worker/monitor"
	"github.com/shadowapi/shadowapi/backend/internal/worker/subjects"
	"github.com/shadowapi/shadowapi/backend/pkg/jobs"
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
	s.log.Debug("token refresh scheduler tick")
	queries := query.New(s.dbp)
	// Query tokens expiring within the next 5 minutes
	tokens, err := queries.GetTokensToRefresh(ctx, "")
	if err != nil {
		s.log.Error("Failed fetching tokens to refresh", "err", err)
		return
	}
	s.log.Debug("found tokens to refresh", "count", len(tokens))
	for _, tokenRow := range tokens {
		// Parse the stored token data
		var storedToken oauth2.Token
		if err := json.Unmarshal(tokenRow.Oauth2Token.Token, &storedToken); err != nil {
			s.log.Error("Failed to unmarshal token data", "token_uuid", tokenRow.Oauth2Token.UUID.String(), "err", err)
			continue
		}

		// Load the OAuth2 client config
		clientRow, err := queries.GetOauth2Client(ctx, converter.UuidPtrToPgUUID(tokenRow.Oauth2Token.ClientUuid))
		if err != nil {
			s.log.Error("Failed to get OAuth2 client", "token_uuid", tokenRow.Oauth2Token.UUID.String(), "client_uuid", tokenRow.Oauth2Token.ClientUuid.String(), "err", err)
			continue
		}

		jobUUID := uuid.Must(uuid.NewV7()).String()
		schedulerUUID := uuid.Must(uuid.NewV7()).String()
		headers := queue.Headers{"X-Job-ID": jobUUID}

		// Build job args with all data needed for token refresh
		jobArgs := jobs.TokenRefresherJobArgs{
			JobUUID:       jobUUID,
			SchedulerUUID: schedulerUUID,
			TokenUUID:     tokenRow.Oauth2Token.UUID.String(),
			Expiry:        time.Now().Add(defaultRefreshInterval),

			// OAuth2 token data
			AccessToken:  storedToken.AccessToken,
			RefreshToken: storedToken.RefreshToken,
			TokenExpiry:  storedToken.Expiry,

			// OAuth2 client config
			ClientID:     clientRow.Oauth2Client.ClientID,
			ClientSecret: clientRow.Oauth2Client.Secret,
			Provider:     clientRow.Oauth2Client.Provider,
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
		s.log.Info("Published token refresher job", "token_uuid", tokenRow.Oauth2Token.UUID.String(), "provider", clientRow.Oauth2Client.Provider)
	}
}
