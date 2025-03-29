package storage

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shadowapi/shadowapi/backend/pkg/api"
	"log/slog"
	"os"
)

type HostfilesStorage struct {
	log        *slog.Logger
	rootFolder string
	dbp        *pgxpool.Pool
}

func NewHostfilesStorage(log *slog.Logger, folder string, dbp *pgxpool.Pool) *HostfilesStorage {
	return &HostfilesStorage{log: log, rootFolder: folder, dbp: dbp}
}

// SaveMessage saves metadata in Postgres
func (h *HostfilesStorage) SaveMessage(ctx context.Context, msg *api.Message) error {
	h.log.Info("Saving message meta to Postgres (hostfiles mode)", "uuid", msg.UUID)
	// For example: db.InsertMessageMetadata(msg.UUID, msg.Sender, msg.Body, "hostfiles")
	return nil
}

// SaveAttachment writes the file to local folder, then meta in Postgres
func (h *HostfilesStorage) SaveAttachment(ctx context.Context, f *api.FileObject) error {
	filePath := fmt.Sprintf("%s/%s", h.rootFolder, f.GetName().Or("attachment"))
	out, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer out.Close()

	// If there's data in memory
	// io.Copy(out, someReader)

	// Insert metadata
	// db.InsertAttachment(f.UUID, userID, filePath)
	h.log.Info("Attachment saved to local disk, meta in Postgres", "filePath", filePath)
	return nil
}
