package jobs

import (
	"context"
	"encoding/json"
	"github.com/shadowapi/shadowapi/backend/internal/converter"
	"github.com/shadowapi/shadowapi/backend/internal/worker/monitor"
	"github.com/shadowapi/shadowapi/backend/internal/worker/registry"
	"github.com/shadowapi/shadowapi/backend/internal/worker/types"
	"log/slog"
	"time"

	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/oauth2"

	"github.com/shadowapi/shadowapi/backend/internal/db"
	oauthTools "github.com/shadowapi/shadowapi/backend/internal/oauth2"
	"github.com/shadowapi/shadowapi/backend/internal/queue"
	"github.com/shadowapi/shadowapi/backend/pkg/query"
)

// TokenRefresherJobArgs holds the arguments for a token refresh job.
type TokenRefresherJobArgs struct {
	SchedulerUUID string    `json:"scheduler_uuid"`
	JobUUID       string    `json:"job_uuid"`
	TokenUUID     uuid.UUID `json:"token_uuid"`
	Expiry        time.Time `json:"expiry"`
}

// TokenRefresherJob implements the worker.Job interface.
type TokenRefresherJob struct {
	log     *slog.Logger
	dbp     *pgxpool.Pool
	queue   *queue.Queue
	monitor *monitor.WorkerMonitor

	schedulerUUID string
	jobUUID       string
	args          TokenRefresherJobArgs
}

// NewTokenRefresherJob creates a new TokenRefresherJob.
func TokenRefresherJobFactory(
	dbp *pgxpool.Pool,
	log *slog.Logger,
	q *queue.Queue,
	mon *monitor.WorkerMonitor,
) types.JobFactory {
	return func(data []byte) (types.Job, error) {
		var args TokenRefresherJobArgs
		if err := json.Unmarshal(data, &args); err != nil {
			return nil, err
		}

		if time.Now().UTC().Before(args.Expiry) {
			return nil, types.JobNotReadyError{Delay: time.Until(args.Expiry)}
		}

		recordID := uuid.Must(uuid.NewV7())

		return &TokenRefresherJob{
			log:           log,
			dbp:           dbp,
			queue:         q,
			monitor:       mon,
			schedulerUUID: args.SchedulerUUID,
			jobUUID:       recordID.String(),
			args:          args,
		}, nil
	}
}

// Execute performs the token refresh work.
// It uses a transaction to:
//  1. Retrieve the current token data.
//  2. Unmarshal the token.
//  3. Get the OAuth2 client config and refresh the token.
//  4. Reschedule the next refresh by publishing a new token refresh message.
func (t *TokenRefresherJob) Execute(ctx context.Context) (err error) {
	t.monitor.RecordJobStart(ctx, t.schedulerUUID, t.jobUUID, registry.WorkerSubjectTokenRefresh)
	defer func() {
		status := monitor.StatusDone
		if err != nil {
			status = monitor.StatusFailed
		}
		t.monitor.RecordJobEnd(ctx, t.schedulerUUID, t.jobUUID, registry.WorkerSubjectTokenRefresh, status, func() string {
			if err != nil {
				return err.Error()
			}
			return ""
		}())
	}()

	log := t.log.With("token_uuid", t.args.TokenUUID, "worker", "TokenRefresherJob")
	log.Debug("Starting token refresh job")
	refreshedToken, err := db.InTx(ctx, t.dbp, func(tx pgx.Tx) (*oauth2.Token, error) {
		qh := query.New(t.dbp).WithTx(tx)
		tokenRow, err := qh.GetOauth2TokenByUUID(ctx, converter.UuidToPgUUID(t.args.TokenUUID))
		if err != nil {
			return nil, err
		}
		var token oauth2.Token
		if err = json.Unmarshal(tokenRow.Oauth2Token.Token, &token); err != nil {
			log.Error("Failed to unmarshal token data", "error", err)
			return nil, err
		}
		// Use the associated client UUID as the OAuth2 client identifier.
		clientID := tokenRow.Oauth2Token.ClientUuid.String()
		config, err := oauthTools.GetClientConfig(ctx, t.dbp, clientID)
		if err != nil {
			log.Error("Failed to get OAuth2 client config", "error", err)
			return nil, err
		}
		refreshedToken, err := config.TokenSource(ctx, &token).Token()
		if err != nil {
			log.Error("Failed to refresh token, deleting it", "error", err)
			if delErr := qh.DeleteOauth2Token(ctx, converter.UuidToPgUUID(t.args.TokenUUID)); delErr != nil {
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
	return ScheduleTokenRefresh(ctx, t.queue, t.args.TokenUUID, refreshedToken.Expiry, t.log)
}

// ScheduleTokenRefresh schedules a new token refresh job by publishing a message.
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
	return q.Publish(ctx, registry.WorkerSubjectTokenRefresh, msg)
}

/*
// OLD CODE

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
