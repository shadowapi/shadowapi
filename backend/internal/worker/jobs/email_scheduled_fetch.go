package jobs

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shadowapi/shadowapi/backend/internal/converter"
	"github.com/shadowapi/shadowapi/backend/internal/queue"
	"github.com/shadowapi/shadowapi/backend/internal/worker/types"
	"github.com/shadowapi/shadowapi/backend/pkg/api"
	"github.com/shadowapi/shadowapi/backend/pkg/query"
	"log/slog"
	"time"
)

type ScheduledEmailFetchJobArgs struct {
	PipelineUUID string    `json:"pipeline_uuid"`
	LastFetched  time.Time `json:"last_fetched"`
}

type EmailScheduledFetchJob struct {
	log          *slog.Logger
	dbp          *pgxpool.Pool
	queue        *queue.Queue
	pipelinesMap *map[string]types.Pipeline

	pipelineUUID string
	lastFetched  time.Time
}

func NewEmailScheduledFetchJob(args ScheduledEmailFetchJobArgs, dbp *pgxpool.Pool, log *slog.Logger, q *queue.Queue, pipelinesMap *map[string]types.Pipeline) *EmailScheduledFetchJob {
	return &EmailScheduledFetchJob{
		log:          log,
		dbp:          dbp,
		queue:        q,
		pipelinesMap: pipelinesMap,

		pipelineUUID: args.PipelineUUID,
		lastFetched:  args.LastFetched,
	}
}

func (e *EmailScheduledFetchJob) Execute(ctx context.Context) error {
	queries := query.New(e.dbp)
	pipeUUID, err := uuid.FromString(e.pipelineUUID)
	if err != nil {
		e.log.Error("invalid pipeline UUID", "error", err)
		return err
	}
	pipeRow, err := queries.GetPipeline(ctx, pgtype.UUID{Bytes: pipeUUID, Valid: true})
	if err != nil {
		e.log.Error("failed to get pipeline", "error", err)
		return err
	}
	if pipeRow.Pipeline.UUID == uuid.Nil {
		e.log.Error("pipeline not found", "pipeline_uuid", e.pipelineUUID)
		return nil
	}
	if !pipeRow.Pipeline.IsEnabled {
		e.log.Info("pipeline is disabled", "pipeline_uuid", e.pipelineUUID)
		return nil
	}

	dsRow, err := queries.GetDatasource(ctx, converter.UuidPtrToPgUUID(pipeRow.Pipeline.DatasourceUUID))
	if err != nil {
		e.log.Error("failed to get datasource", "error", err)
		return err
	}
	if dsRow.Datasource.UUID == uuid.Nil {
		e.log.Error("datasource not found for pipeline", "pipeline_uuid", e.pipelineUUID)
		return nil
	}
	if !dsRow.Datasource.IsEnabled {
		e.log.Info("datasource is disabled", "datasource_uuid", dsRow.Datasource.UUID.String())
		return nil
	}

	// Token check logic if needed (pseudo-code)
	// If token near expiry -> schedule refresh
	// (In real code you'd fetch and compare OAuth token expiry from DB)
	// e.log.Info("Token checked, scheduling refresh if near expiry")

	e.log.Info("fetching emails", "datasource_uuid", dsRow.Datasource.UUID.String())
	messages := fetchEmails(dsRow.Datasource.UUID.String(), e.lastFetched)
	if len(messages) == 0 {
		e.log.Info("no new messages", "datasource_uuid", dsRow.Datasource.UUID.String())
		return nil
	}

	pl, ok := (*e.pipelinesMap)[dsRow.Datasource.UUID.String()]
	if !ok {
		e.log.Warn("no pipeline found in pipelinesMap for datasource", "datasource_uuid", dsRow.Datasource.UUID.String())
		return nil
	}

	for _, msg := range messages {
		if err := pl.Run(ctx, &msg); err != nil {
			e.log.Error("pipeline run failed", "error", err, "message_uuid", msg.UUID)
		}
	}
	return nil
}

func fetchEmails(datasourceUUID string, since time.Time) []api.Message {
	// Replace with real IMAP / Gmail API / etc.

	/*
		// Parse the datasource settings for OAuth2 credentials.
			if len(ds.Settings) == 0 {
				return nil, errors.New("empty datasource settings")
			}
			var settings emailOAuthSettings
			if err := json.Unmarshal(ds.Settings, &settings); err != nil {
				log.Error("failed to unmarshal datasource settings", "error", err)
				return nil, err
			}
			if settings.OAuth2ClientUUID == "" || settings.OAuth2TokenUUID == "" {
				return nil, errors.New("missing OAuth2 configuration in datasource settings")
			}
			// Get the OAuth2 client configuration.
			clientConfig, err := oauth2.GetClientConfig(ctx, dbp, settings.OAuth2ClientUUID)
			if err != nil {
				log.Error("failed to get client config", "error", err)
				return nil, err
			}
			// Convert token UUID string to uuid.UUID.
			tokenUUID, err := uuid.FromString(settings.OAuth2TokenUUID)
			if err != nil {
				log.Error("invalid OAuth2 token UUID", "token_uuid", settings.OAuth2TokenUUID, "error", err)
				return nil, err
			}
			// Create a token store and get a valid token.
			store, err := newTokenStore(ctx, clientConfig, dbp, tokenUUID, time.Minute)
			if err != nil {
				log.Error("failed to create token store", "error", err)
				return nil, err
			}
			token, err := store.Token()
			if err != nil {
				log.Error("failed to retrieve OAuth2 token", "error", err)
				return nil, err
			}
			// Create an HTTP client using the token.
			httpClient := clientConfig.Client(ctx, token)
			// Create the Gmail API service.
			gmailService, err := gmail.NewService(ctx, option.WithHTTPClient(httpClient))
			if err != nil {
				log.Error("failed to create Gmail service", "error", err)
				return nil, err
			}
			// Build a query string for messages after the given date.
			queryStr := fmt.Sprintf("after:%s", since.Format("2006/01/02"))
			req := gmailService.Users.Messages.List("me").Q(queryStr)
			res, err := req.Do()
			if err != nil {
				log.Error("failed to list messages", "error", err, "query", queryStr)
				return nil, err
			}
			if res.ResultSizeEstimate == 0 || len(res.Messages) == 0 {
				return []api.Message{}, nil
			}
			var messages []api.Message
			// Helper function to extract header values.
			getHeader := func(headers []*gmail.MessagePartHeader, name string) string {
				for _, h := range headers {
					if strings.EqualFold(h.Name, name) {
						return h.Value
					}
				}
				return ""
			}
			// Loop over each message and fetch its full content.
			for _, m := range res.Messages {
				fullMsg, err := gmailService.Users.Messages.Get("me", m.Id).Format("full").Do()
				if err != nil {
					log.Error("failed to get full message", "message_id", m.Id, "error", err)
					continue
				}
				var subjectOpt api.OptString
				subject := getHeader(fullMsg.Payload.Headers, "Subject")
				if subject != "" {
					subjectOpt = api.NewOptString(subject)
				} else {
					subjectOpt = api.NewOptString("")
				}
				// Prepare the api.Message. Using snippet as the body.
				emailMsg := api.Message{
					UUID:       api.NewOptString("gmail-" + m.Id),
					Format:     "email",
					Type:       "email",
					Sender:     getHeader(fullMsg.Payload.Headers, "From"),
					Recipients: []string{getHeader(fullMsg.Payload.Headers, "To")},
					Subject:    subjectOpt,
					Body:       fullMsg.Snippet,
				}
				messages = append(messages, emailMsg)
			}
			log.Info("Gmail fetch completed", "count", len(messages))
			return messages, nil
	*/
	return []api.Message{
		{
			UUID:   api.NewOptString("email-" + datasourceUUID + "-001"),
			Sender: "test-sender@example.com",
			Body:   `{"first":"Tester","last":"Email"}`,
		},
	}
}

func ScheduleEmailFetchJobFactory(dbp *pgxpool.Pool, log *slog.Logger, q *queue.Queue, pipelinesMap *map[string]types.Pipeline) types.JobFactory {
	return func(data []byte) (types.Job, error) {
		var args ScheduledEmailFetchJobArgs
		if err := json.Unmarshal(data, &args); err != nil {
			return nil, err
		}
		if args.PipelineUUID == "" {
			return nil, errors.New("missing pipeline_uuid in job args")
		}
		return NewEmailScheduledFetchJob(args, dbp, log, q, pipelinesMap), nil
	}
}
