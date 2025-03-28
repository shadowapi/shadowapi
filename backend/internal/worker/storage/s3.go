package storage

import (
	"context"
	"log/slog"

	"github.com/shadowapi/shadowapi/backend/pkg/api"
)

type S3Storage struct {
	log *slog.Logger
	// Additional configuration fields (e.g. bucket name, credentials) can be added.
}

func NewS3Storage(log *slog.Logger) *S3Storage {
	return &S3Storage{log: log}
}

func (s *S3Storage) SaveMessage(ctx context.Context, message *api.Message) error {
	s.log.Info("Saving message to S3", "message_uuid", message.UUID)
	// Implement S3 message saving here.
	return nil
}

func (s *S3Storage) SaveAttachment(ctx context.Context, file *api.FileObject) error {
	s.log.Info("Saving attachment to S3", "file_uuid", file.UUID)
	// Implement S3 file saving here.
	return nil
}
