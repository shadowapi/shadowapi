package jobs

import (
	"context"
	"github.com/shadowapi/shadowapi/backend/internal/worker/types"
	"log/slog"
)

type TelegramFetchMockArgs struct {
	AccountID string `json:"account_id"`
}

type TelegramFetchMockJob struct {
	log       *slog.Logger
	accountID string
}

func NewTelegramFetchMockJob(log *slog.Logger, acc string) *TelegramFetchMockJob {
	return &TelegramFetchMockJob{log: log, accountID: acc}
}

func (j *TelegramFetchMockJob) Execute(ctx context.Context) error {
	j.log.Info("Mock Telegram fetch job", "account_id", j.accountID)
	// Do nothing or return a mock message
	return nil
}

func TelegramFetchMockJobFactory(log *slog.Logger) types.JobFactory {
	return func(data []byte) (types.Job, error) {
		// Unmarshal and return
		return NewTelegramFetchMockJob(log, "mock-telegram"), nil
	}
}

// Similarly for WhatsApp & LinkedIn mocks...
