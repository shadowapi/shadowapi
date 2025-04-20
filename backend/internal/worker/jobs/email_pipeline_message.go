package jobs

import (
	"context"
	"encoding/json"

	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shadowapi/shadowapi/backend/internal/queue"
	"github.com/shadowapi/shadowapi/backend/internal/worker/monitor"
	"github.com/shadowapi/shadowapi/backend/internal/worker/registry"
	"github.com/shadowapi/shadowapi/backend/internal/worker/types"
	"github.com/shadowapi/shadowapi/backend/pkg/api"
	"log/slog"
)

type EmailPipelineMessageJobArgs struct {
	PipelineUUID  string          `json:"pipeline_uuid"`
	SchedulerUUID string          `json:"scheduler_uuid"`
	JobUUID       string          `json:"job_uuid"`
	MessageData   json.RawMessage `json:"message_data"`
}

type EmailPipelineMessageJob struct {
	log          *slog.Logger
	dbp          *pgxpool.Pool
	queue        *queue.Queue
	monitor      *monitor.WorkerMonitor
	pipelinesMap *map[string]types.Pipeline

	schedulerUUID string
	jobUUID       string
	pipelineUUID  string
	messageData   json.RawMessage
}

func EmailPipelineMessageJobFactory(
	dbp *pgxpool.Pool,
	log *slog.Logger,
	q *queue.Queue,
	mon *monitor.WorkerMonitor,
	pipelines *map[string]types.Pipeline,
) types.JobFactory {
	return func(data []byte) (types.Job, error) {
		var args EmailPipelineMessageJobArgs
		if err := json.Unmarshal(data, &args); err != nil {
			return nil, err
		}
		recordID := uuid.Must(uuid.NewV7())

		return &EmailPipelineMessageJob{
			log:           log,
			dbp:           dbp,
			queue:         q,
			monitor:       mon,
			pipelinesMap:  pipelines,
			pipelineUUID:  args.PipelineUUID,
			schedulerUUID: args.SchedulerUUID,
			jobUUID:       recordID.String(),
			messageData:   args.MessageData,
		}, nil
	}
}

func (e *EmailPipelineMessageJob) Execute(ctx context.Context) (err error) {
	e.monitor.RecordJobStart(ctx, e.schedulerUUID, e.jobUUID, registry.WorkerSubjectEmailApplyPipeline)
	defer func() {
		status := monitor.StatusDone
		if err != nil {
			status = monitor.StatusFailed
		}
		e.monitor.RecordJobEnd(ctx, "", e.jobUUID, registry.WorkerSubjectEmailApplyPipeline, status, func() string {
			if err != nil {
				return err.Error()
			}
			return ""
		}())
	}()

	var msg api.Message
	if err = json.Unmarshal(e.messageData, &msg); err != nil {
		e.log.Error("unmarshal message failed", "error", err)
		return err
	}

	pl, ok := (*e.pipelinesMap)[e.pipelineUUID]
	if !ok {
		e.log.Error("pipeline not found", "uuid", e.pipelineUUID)
		return nil
	}
	err = pl.Run(ctx, &msg)
	return err
}
