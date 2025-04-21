package jobs

import (
	"context"
	"encoding/json"
	"github.com/shadowapi/shadowapi/backend/internal/worker/monitor"
	"github.com/shadowapi/shadowapi/backend/internal/worker/registry"
	"github.com/shadowapi/shadowapi/backend/internal/worker/types"
	"log/slog"
	"time"

	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shadowapi/shadowapi/backend/internal/queue"
)

type DummyJobArgs struct {
	SchedulerUUID string    `json:"scheduler_uuid"`
	JobUUID       string    `json:"job_uuid"`
	TokenUUID     uuid.UUID `json:"token_uuid"`
	Expiry        time.Time `json:"expiry"`
}

type DummyJob struct {
	log     *slog.Logger
	dbp     *pgxpool.Pool
	queue   *queue.Queue
	monitor *monitor.WorkerMonitor

	schedulerUUID string
	jobUUID       string
	args          DummyJobArgs
}

func DummyJobFactory(
	dbp *pgxpool.Pool,
	log *slog.Logger,
	q *queue.Queue,
	mon *monitor.WorkerMonitor,
) types.JobFactory {
	return func(data []byte) (types.Job, error) {
		var args DummyJobArgs
		if err := json.Unmarshal(data, &args); err != nil {
			return nil, err
		}

		if time.Now().UTC().Before(args.Expiry) {
			return nil, types.JobNotReadyError{Delay: time.Until(args.Expiry)}
		}

		recordID := uuid.Must(uuid.NewV7())

		return &DummyJob{
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

func (t *DummyJob) Execute(ctx context.Context) (err error) {
	t.monitor.RecordJobStart(ctx, t.schedulerUUID, t.jobUUID, registry.WorkerSubjectDummy)
	defer func() {
		status := monitor.StatusDone
		if err != nil {
			status = monitor.StatusFailed
		}
		t.monitor.RecordJobEnd(ctx, t.schedulerUUID, t.jobUUID, registry.WorkerSubjectDummy, status, func() string {
			if err != nil {
				return err.Error()
			}
			return ""
		}())
	}()

	log := t.log.With("token_uuid", t.args.TokenUUID, "worker", "TokenRefresherJob")

	log.Debug("DummyJob Execute started", "job_uuid", t.jobUUID)

	return nil
}
