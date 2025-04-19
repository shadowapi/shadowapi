package jobs

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shadowapi/shadowapi/backend/internal/converter"
	"github.com/shadowapi/shadowapi/backend/internal/oauth2"
	"github.com/shadowapi/shadowapi/backend/internal/queue"
	"github.com/shadowapi/shadowapi/backend/internal/worker/registry"
	"github.com/shadowapi/shadowapi/backend/internal/worker/types"
	"github.com/shadowapi/shadowapi/backend/pkg/api"
	"github.com/shadowapi/shadowapi/backend/pkg/query"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
	"log/slog"
	"strings"
	"time"
)

type EmailOAuthFetchJobArgs struct {
	PipelineUUID string    `json:"pipeline_uuid"`
	LastFetched  time.Time `json:"last_fetched"`
}

type EmailOAuthFetchJob struct {
	log          *slog.Logger
	dbp          *pgxpool.Pool
	queue        *queue.Queue
	pipelinesMap *map[string]types.Pipeline

	pipelineUUID string
	lastFetched  time.Time
}

func NewEmailScheduledFetchJob(args EmailOAuthFetchJobArgs, dbp *pgxpool.Pool, log *slog.Logger, q *queue.Queue, pipelinesMap *map[string]types.Pipeline) *EmailOAuthFetchJob {
	return &EmailOAuthFetchJob{
		log:          log,
		dbp:          dbp,
		queue:        q,
		pipelinesMap: pipelinesMap,

		pipelineUUID: args.PipelineUUID,
		lastFetched:  args.LastFetched,
	}
}

func ScheduleEmailFetchJobFactory(dbp *pgxpool.Pool, log *slog.Logger, q *queue.Queue, pipelinesMap *map[string]types.Pipeline) types.JobFactory {
	return func(data []byte) (types.Job, error) {
		var args EmailOAuthFetchJobArgs
		if err := json.Unmarshal(data, &args); err != nil {
			return nil, err
		}
		if args.PipelineUUID == "" {
			return nil, errors.New("missing pipeline_uuid in job args")
		}
		return NewEmailScheduledFetchJob(args, dbp, log, q, pipelinesMap), nil
	}
}

func (e *EmailOAuthFetchJob) Execute(ctx context.Context) error {
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

	e.log.Info("fetching emails", "datasource_uuid", dsRow.Datasource.UUID.String())

	// ------------------------------------------------------------------
	// Fetch Gmail messages
	// ------------------------------------------------------------------
	msgs, tokUUID, tokExpiry, err := fetchGmailEmails(ctx, dsRow.Datasource, e.lastFetched, e.dbp, e.log)
	if err != nil {
		e.log.Error("failed to fetch Gmail messages", "error", err)
		return err
	}
	if len(msgs) == 0 {
		e.log.Info("no new messages", "datasource_uuid", dsRow.Datasource.UUID.String())
		return nil
	}

	// Queue per‑message pipeline jobs
	for _, m := range msgs {
		raw, err := mustMarshal(m)
		if err != nil {
			e.log.Error("failed to marshal message", "error", err)
			continue
		}

		data, err := json.Marshal(EmailPipelineMessageJobArgs{PipelineUUID: e.pipelineUUID, MessageData: raw})
		if err != nil {
			e.log.Error("failed to marshal pipeline job args", "error", err)
			continue
		}

		if err := e.queue.Publish(ctx, registry.WorkerSubjectEmailApplyPipeline, data); err != nil {
			e.log.Error("failed to publish pipeline job", "error", err)
		}
	}

	// If token expires soon (<2h) schedule a refresh job
	if time.Until(tokExpiry) < 2*time.Hour {
		_ = ScheduleTokenRefresh(ctx, e.queue, tokUUID, tokExpiry, e.log)
	}
	return nil
}

// mustMarshal marshals a value or panics (safe because data is internal).
func mustMarshal(v any) (json.RawMessage, error) {
	b, err := json.Marshal(v)
	return b, err
}

// fetchGmailEmails fetches new Gmail messages created after `since`.
// It returns the messages plus the token UUID and expiry so the caller can queue a refresh job.
func fetchGmailEmails(
	ctx context.Context,
	ds query.Datasource,
	since time.Time,
	dbp *pgxpool.Pool,
	log *slog.Logger,
) ([]api.Message, uuid.UUID, time.Time, error) {

	// 1. Parse datasource.settings to obtain OAuth2 token/client UUIDs
	if ds.Type != "email_oauth" {
		log.Error("invalid datasource type", "type", ds.Type)
		return nil, uuid.Nil, time.Time{}, errors.New("invalid datasource type")
	}

	var cfg api.DatasourceEmailOAuth
	if err := json.Unmarshal(ds.Settings, &cfg); err != nil {
		log.Error("failed unmarshal DatasourceEmailOAuth settings", "error", err)
		return nil, uuid.Nil, time.Time{}, err
	}

	if cfg.OAuth2ClientUUID == "" {
		log.Error("invalid OAuth2ClientUUID  in settings")
		return nil, uuid.Nil, time.Time{}, errors.New("invalid OAuth2ClientUUID in settings")
	}

	// from oauth2_client table
	clientConfig, err := oauth2.GetClientConfig(ctx, dbp, cfg.OAuth2ClientUUID)
	if err != nil {
		log.Error("failed to build client config", "error", err)
		return nil, uuid.Nil, time.Time{}, err
	}

	// Get tokenUUID from oauth2_token table
	queries := query.New(dbp)
	pgClientUUID, _ := converter.ConvertStringToPgUUID(cfg.OAuth2ClientUUID)
	rows, err := queries.GetOauth2TokensByClientUUID(ctx, pgClientUUID)
	if err != nil {
		log.Error("failed to query tokens by client", "error", err)
		return nil, uuid.Nil, time.Time{}, err
	}
	if len(rows) == 0 {
		log.Error("no tokens found for client", "client_uuid", cfg.OAuth2ClientUUID)
		return nil, uuid.Nil, time.Time{}, errors.New("no oauth2 token found for client")
	}
	latest := rows[0]
	for _, r := range rows {
		lu := r.Oauth2Token.UpdatedAt
		if !latest.Oauth2Token.UpdatedAt.Valid || (lu.Valid && lu.Time.After(latest.Oauth2Token.UpdatedAt.Time)) {
			latest = r
		}
	}
	tokenUUID := latest.Oauth2Token.UUID
	if tokenUUID == uuid.Nil {
		log.Error("invalid token UUID in settings")
		return nil, uuid.Nil, time.Time{}, errors.New("invalid token UUID in settings")
	}

	// Build persistent token store from oauth2_token table – will auto‑persist refreshes back to DB.
	tokenStore, err := oauth2.NewTokenStore(ctx, &clientConfig.Config, dbp, tokenUUID, 5*time.Minute)
	if err != nil {
		log.Error("failed to create token store", "error", err)
		return nil, uuid.Nil, time.Time{}, err
	}

	token, err := tokenStore.Token()
	if err != nil {
		return nil, uuid.Nil, time.Time{}, err
	}

	httpClient := clientConfig.Config.Client(ctx, token)
	gmailSvc, err := gmail.NewService(ctx, option.WithHTTPClient(httpClient))
	if err != nil {
		return nil, uuid.Nil, time.Time{}, err
	}

	// TODO @reactima use query policy??
	// Use Gmail search query `after:YYYY/MM/DD`.
	queryStr := fmt.Sprintf("after:%s", since.Format("2024/01/01"))
	listRes, err := gmailSvc.Users.Messages.List("me").Q(queryStr).Do()
	if err != nil {
		return nil, uuid.Nil, time.Time{}, err
	}

	if len(listRes.Messages) == 0 {
		return []api.Message{}, tokenUUID, token.Expiry, nil
	}

	header := func(headers []*gmail.MessagePartHeader, name string) string {
		for _, h := range headers {
			if strings.EqualFold(h.Name, name) {
				return h.Value
			}
		}
		return ""
	}

	var messages []api.Message
	for _, meta := range listRes.Messages {
		fullMsg, err := gmailSvc.Users.Messages.Get("me", meta.Id).Format("full").Do()
		if err != nil {
			log.Warn("failed to fetch full Gmail message", "id", meta.Id, "error", err)
			continue
		}
		sub := header(fullMsg.Payload.Headers, "Subject")
		var subjectOpt api.OptString
		if sub != "" {
			subjectOpt = api.NewOptString(sub)
		}
		msg := api.Message{
			UUID:       api.NewOptString("gmail-" + meta.Id),
			Type:       "email",
			Format:     "email",
			Sender:     header(fullMsg.Payload.Headers, "From"),
			Recipients: []string{header(fullMsg.Payload.Headers, "To")},
			Subject:    subjectOpt,
			Body:       fullMsg.Snippet,
		}
		messages = append(messages, msg)
	}
	return messages, tokenUUID, token.Expiry, nil
}
