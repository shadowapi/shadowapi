package jobs

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/oauth2"

	"github.com/shadowapi/shadowapi/backend/internal/db"
	oauthTools "github.com/shadowapi/shadowapi/backend/internal/oauth2"
	"github.com/shadowapi/shadowapi/backend/internal/queue"
	"github.com/shadowapi/shadowapi/backend/internal/worker"
	"github.com/shadowapi/shadowapi/backend/pkg/query"
)

// TokenRefresherJobArgs holds the arguments for a token refresh job.
type TokenRefresherJobArgs struct {
	TokenUUID uuid.UUID `json:"token_uuid"`
	Expiry    time.Time `json:"expiry"`
}

// TokenRefresherJob implements the worker.Job interface.
type TokenRefresherJob struct {
	dbp   *pgxpool.Pool
	log   *slog.Logger
	queue *queue.Queue
	args  TokenRefresherJobArgs
}

// NewTokenRefresherJob creates a new TokenRefresherJob.
func NewTokenRefresherJob(dbp *pgxpool.Pool, log *slog.Logger, q *queue.Queue, args TokenRefresherJobArgs) *TokenRefresherJob {
	return &TokenRefresherJob{
		dbp:   dbp,
		log:   log,
		queue: q,
		args:  args,
	}
}

// Execute performs the token refresh work.
// It uses a transaction to:
//  1. Retrieve the current token data.
//  2. Unmarshal the token.
//  3. Get the OAuth2 client config and refresh the token.
//  4. Reschedule the next refresh by publishing a new token refresh message.
func (t *TokenRefresherJob) Execute(ctx context.Context) error {
	log := t.log.With("token_uuid", t.args.TokenUUID, "worker", "TokenRefresherJob")
	log.Debug("Starting token refresh job")
	token, err := db.InTx(ctx, t.dbp, func(tx pgx.Tx) (*oauth2.Token, error) {
		// Use the query helper with the transaction.
		qh := query.New(t.dbp).WithTx(tx)
		tokenData, err := qh.GetOauth2TokenByUUID(ctx, t.args.TokenUUID)
		if err != nil {
			return nil, err
		}
		var token oauth2.Token
		if err = json.Unmarshal(tokenData.Token, &token); err != nil {
			log.Error("Failed to unmarshal token data", "error", err)
			return nil, err
		}
		config, err := oauthTools.GetClientConfig(ctx, t.dbp, tokenData.ClientID)
		if err != nil {
			log.Error("Failed to get OAuth2 client config", "error", err)
			return nil, err
		}
		refreshedToken, err := config.TokenSource(ctx, &token).Token()
		if err != nil {
			log.Error("Failed to refresh token, deleting it", "error", err)
			if delErr := qh.DeleteOauth2Token(ctx, t.args.TokenUUID); delErr != nil {
				log.Error("Failed to delete broken token", "error", delErr)
			}
			return nil, err
		}
		return refreshedToken, nil
	})
	if err == pgx.ErrNoRows {
		log.Warn("Token not found, cancelling token refresh job")
		return nil
	} else if err != nil {
		return err
	}

	log.Debug("Token refreshed successfully", "token_uuid", t.args.TokenUUID)
	// Reschedule the next token refresh.
	return ScheduleTokenRefresh(ctx, t.queue, t.args.TokenUUID, token.Expiry, t.log)
}

// scheduleTokenRefresh schedules a new token refresh job by publishing a message.
func ScheduleTokenRefresh(ctx context.Context, q *queue.Queue, tokenUUID uuid.UUID, expiry time.Time, log *slog.Logger) error {
	delay := time.Until(expiry)
	log.Debug("Scheduling next token refresh", "delay", delay)
	newArgs := TokenRefresherJobArgs{
		TokenUUID: tokenUUID,
		Expiry:    expiry,
	}
	msg, err := json.Marshal(newArgs)
	if err != nil {
		log.Error("Failed to marshal token refresh args", "error", err)
		return err
	}
	// Publish using the registered worker subject.
	return q.Publish(ctx, worker.WorkerSubjectTokenRefresh, msg)
}

// TokenRefresherJobFactory is the factory for token refresher jobs.
func TokenRefresherJobFactory(dbp *pgxpool.Pool, log *slog.Logger, q *queue.Queue) worker.JobFactory {
	return func(data []byte) (worker.Job, error) {
		var args TokenRefresherJobArgs
		if err := json.Unmarshal(data, &args); err != nil {
			return nil, err
		}
		// If the job is not yet ready, return a JobNotReadyError.
		if time.Now().UTC().Before(args.Expiry) {
			return nil, worker.JobNotReadyError{Delay: time.Until(args.Expiry)}
		}
		return NewTokenRefresherJob(dbp, log, q, args), nil
	}
}
