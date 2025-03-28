package storage

import (
	"context"
	"log/slog"

	"github.com/shadowapi/shadowapi/backend/pkg/api"
)

type PostgresStorage struct {
	log *slog.Logger
	// In a full implementation, you might embed a DB connection here.
}

func NewPostgresStorage(log *slog.Logger) *PostgresStorage {
	return &PostgresStorage{log: log}
}

func (p *PostgresStorage) SaveMessage(ctx context.Context, message *api.Message) error {
	p.log.Info("Saving message to Postgres", "message_uuid", message.UUID)
	// Implement Postgres message saving here.
	return nil
}

func (p *PostgresStorage) SaveAttachment(ctx context.Context, file *api.FileObject) error {
	p.log.Info("Saving attachment to Postgres", "file_uuid", file.UUID)
	// Implement Postgres file saving here.
	return nil
}
