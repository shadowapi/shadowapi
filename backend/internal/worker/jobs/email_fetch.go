package jobs

import (
	"context"
	"encoding/json"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/shadowapi/shadowapi/backend/internal/worker/registry"
	"github.com/shadowapi/shadowapi/backend/internal/worker/types"
	"github.com/shadowapi/shadowapi/backend/pkg/query"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shadowapi/shadowapi/backend/internal/queue"
	"github.com/shadowapi/shadowapi/backend/pkg/api"
)

// EmailFetchJobArgs holds job arguments (e.g. the datasource account ID).
type EmailFetchJobArgs struct {
	AccountID string `json:"account_id"`
	// Optionally, add a timestamp to fetch only new messages.
	LastFetched time.Time `json:"last_fetched"`
}

// EmailFetchJob implements worker.Job
type EmailFetchJob struct {
	accountID   string
	lastFetched time.Time
	log         *slog.Logger
	// The pipeline combines the extractor, filter, and storage.
	pipeline types.Pipeline
	dbp      *pgxpool.Pool
	queue    *queue.Queue
}

// NewEmailFetchJob constructs an email fetch job.
func NewEmailFetchJob(args EmailFetchJobArgs, log *slog.Logger, pipeline types.Pipeline, dbp *pgxpool.Pool, q *queue.Queue) *EmailFetchJob {
	return &EmailFetchJob{
		accountID:   args.AccountID,
		lastFetched: args.LastFetched,
		log:         log,
		pipeline:    pipeline,
		dbp:         dbp,
		queue:       q,
	}
}

// Execute runs the job:
// 1. Look up the datasource account (and verify that it’s enabled).
// 2. Use an email client (IMAP or Gmail API) to fetch messages since lastFetched.
// 3. For each message, run the pipeline.
func (j *EmailFetchJob) Execute(ctx context.Context) error {
	j.log.Info("Starting email fetch job", "account_id", j.accountID)

	// --- Pseudo-code: Retrieve account credentials from DB ---
	// account, err := db.GetDatasourceEmail(j.accountID)
	// if err != nil { return err }
	// if !account.IsEnabled { j.log.Info("Account disabled"); return nil }
	// If using OAuth2, you may want to check token expiry and schedule a token refresh job.
	// For brevity, that logic is assumed to be handled by the token refresh worker.

	// --- Pseudo-code: Connect to email server and fetch messages ---
	// Here we simulate fetched messages.
	fetched := fetchEmailsForAccount(j.accountID, j.lastFetched)
	if len(fetched) == 0 {
		j.log.Info("No new messages")
		return nil
	}

	// Process each message with the pipeline.
	for _, msg := range fetched {
		// Run the pipeline: extractor, filter and storage.
		if err := j.pipeline.Run(ctx, &msg); err != nil {
			j.log.Error("Pipeline failed", "error", err, "message_uuid", msg.UUID)
		}
	}

	return nil
}

// Some dummy helper that simulates fetching emails
func fetchEmailsForAccount(accountID string, since time.Time) []api.Message {
	// In real code, you’d use IMAP/Gmail API, etc.
	return []api.Message{
		{UUID: "msg-001", Sender: "alice@example.com", Body: `{"first":"Alice","last":"Smith"}`},
	}
}

// EmailFetchJobFactory returns a JobFactory to create EmailFetchJob instances.
func EmailFetchJobFactory(log *slog.Logger, pipeline types.Pipeline, dbp *pgxpool.Pool, q *queue.Queue) types.JobFactory {
	return func(data []byte) (types.Job, error) {
		var args EmailFetchJobArgs
		if err := json.Unmarshal(data, &args); err != nil {
			return nil, err
		}
		return NewEmailFetchJob(args, log, pipeline, dbp, q), nil
	}
}

// NEW
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
		jobArgs := EmailFetchJobArgs{
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
