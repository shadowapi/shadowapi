package jobs

import (
	"context"
	"encoding/json"
	"time"

	"github.com/shadowapi/shadowapi/backend/internal/worker"
	"github.com/shadowapi/shadowapi/backend/pkg/api"
	//"github.com/shadowapi/shadowapi/backend/internal/worker/extractors"
	//"github.com/shadowapi/shadowapi/backend/internal/worker/filters"
	//"github.com/shadowapi/shadowapi/backend/internal/worker/storage"
	"log/slog"
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
	pipeline worker.Pipeline
}

// NewEmailFetchJob constructs an email fetch job.
func NewEmailFetchJob(args EmailFetchJobArgs, log *slog.Logger, pipeline worker.Pipeline) *EmailFetchJob {
	return &EmailFetchJob{
		accountID:   args.AccountID,
		lastFetched: args.LastFetched,
		log:         log,
		pipeline:    pipeline,
	}
}

// Execute runs the job:
// 1. Look up the datasource account (and verify that it’s enabled).
// 2. Use an email client (e.g. IMAP or Gmail API) to fetch messages.
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

// fetchEmailsForAccount simulates fetching emails for a given account.
// In a real implementation, this would use the account’s credentials.
func fetchEmailsForAccount(accountID string, since time.Time) []api.Message {
	// Example: return a slice of messages (fetched from IMAP or API)
	return []api.Message{
		{
			UUID:   "msg-001",
			Sender: "alice@example.com",
			// Body is assumed to be a JSON payload that can be parsed into a Contact.
			Body: `{"first": "Alice", "last": "Smith", "email": "alice@example.com"}`,
			// You can set additional fields such as attachments here.
		},
		// Additional messages…
	}
}

// EmailFetchJobFactory returns a JobFactory to create EmailFetchJob instances.
func EmailFetchJobFactory(log *slog.Logger, pipeline worker.Pipeline) worker.JobFactory {
	return func(data []byte) (worker.Job, error) {
		var args EmailFetchJobArgs
		if err := json.Unmarshal(data, &args); err != nil {
			return nil, err
		}
		// Optionally: check account is enabled in DB.
		return NewEmailFetchJob(args, log, pipeline), nil
	}
}
