package jobs

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/shadowapi/shadowapi/backend/internal/worker"
)

type LinkedinJobArgs struct {
	// Add fields as required. For example, you might include a message ID, payload, etc.
	MessageData json.RawMessage `json:"message_data"`
}

type LinkedinJob struct {
	log  *slog.Logger
	args LinkedinJobArgs
}

func NewLinkedinJob(log *slog.Logger, args LinkedinJobArgs) *LinkedinJob {
	return &LinkedinJob{log: log, args: args}
}

func (j *LinkedinJob) Execute(ctx context.Context) error {
	j.log.Info("Executing LinkedIn job")
	// Process the LinkedIn message...
	return nil
}

func LinkedinJobFactory(log *slog.Logger) worker.JobFactory {
	return func(data []byte) (worker.Job, error) {
		var args LinkedinJobArgs
		if err := json.Unmarshal(data, &args); err != nil {
			return nil, err
		}
		return NewLinkedinJob(log, args), nil
	}
}
