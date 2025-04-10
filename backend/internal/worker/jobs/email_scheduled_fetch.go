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
