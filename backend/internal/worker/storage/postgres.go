package storage

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"log/slog"

	"github.com/shadowapi/shadowapi/backend/pkg/api"
)

type PostgresStorage struct {
	log *slog.Logger
	dbp *pgxpool.Pool
}

func NewPostgresStorage(log *slog.Logger, dbp *pgxpool.Pool) *PostgresStorage {
	return &PostgresStorage{log: log, dbp: dbp}
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
