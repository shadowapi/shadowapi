package jobs

import (
	"context"
	"encoding/json"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shadowapi/shadowapi/backend/internal/queue"
	"github.com/shadowapi/shadowapi/backend/internal/worker/types"
	"github.com/shadowapi/shadowapi/backend/pkg/api"
	"log/slog"
)

// EmailPipelineMessageJobArgs contains the incoming job data.
type EmailPipelineMessageJobArgs struct {
	PipelineUUID string          `json:"pipeline_uuid"`
	MessageData  json.RawMessage `json:"message_data"`
}

// EmailPipelineMessageJob implements the worker.Job interface.
type EmailPipelineMessageJob struct {
	PipelineUUID string `json:"pipeline_uuid"`
	// Save the raw message data so we can process it.
	MessageData json.RawMessage

	log          *slog.Logger
	dbp          *pgxpool.Pool
	queue        *queue.Queue
	pipelinesMap *map[string]types.Pipeline
}

func NewEmailPipelineMessageJob(args EmailPipelineMessageJobArgs, dbp *pgxpool.Pool, log *slog.Logger, q *queue.Queue, pipelinesMap *map[string]types.Pipeline) *EmailPipelineMessageJob {
	return &EmailPipelineMessageJob{
		MessageData:  args.MessageData,
		PipelineUUID: args.PipelineUUID,
		log:          log,
		dbp:          dbp,
		queue:        q,
		pipelinesMap: pipelinesMap,
	}
}

func (ep *EmailPipelineMessageJob) Execute(ctx context.Context) error {
	ep.log.Info("Running pipeline", "message", ep.MessageData)

	// Parse the message data into an api.Message struct.
	var message api.Message
	if err := json.Unmarshal(ep.MessageData, &message); err != nil {
		ep.log.Error("failed to unmarshal message data", "error", err)
		return err
	}

	pl, ok := (*ep.pipelinesMap)[ep.PipelineUUID]
	if !ok {
		ep.log.Warn("pipeline not found for datasource", "ep.PipelineUUID", ep.PipelineUUID)
		return nil
	}

	// Run the pipeline.
	if err := pl.Run(ctx, &message); err != nil {
		ep.log.Error("failed to run pipeline", "error", err)
		return err
	}
	ep.log.Info("Pipeline completed successfully", "message_uuid", message.UUID)
	return nil
}

func EmailPipelineMessageJobFactory(dbp *pgxpool.Pool, log *slog.Logger, q *queue.Queue, pipelinesMap *map[string]types.Pipeline) types.JobFactory {
	return func(data []byte) (types.Job, error) {
		var args EmailPipelineMessageJobArgs
		if err := json.Unmarshal(data, &args); err != nil {
			return nil, err
		}
		return NewEmailPipelineMessageJob(args, dbp, log, q, pipelinesMap), nil
	}
}
