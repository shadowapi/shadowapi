package storage

import (
	"context"
	"fmt"
	"github.com/shadowapi/shadowapi/backend/internal/converter"
	"os"
	"path/filepath"

	"log/slog"

	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/shadowapi/shadowapi/backend/pkg/api"
	"github.com/shadowapi/shadowapi/backend/pkg/query"
)

type HostfilesStorage struct {
	log        *slog.Logger
	rootFolder string
	dbp        *pgxpool.Pool
}

func NewHostfilesStorage(log *slog.Logger, folder string, dbp *pgxpool.Pool) *HostfilesStorage {
	return &HostfilesStorage{log: log, rootFolder: folder, dbp: dbp}
}

func (h *HostfilesStorage) SaveMessage(ctx context.Context, msg *api.Message) error {
	h.log.Info("Saving message meta (hostfiles mode)", "message_uuid", msg.GetUUID())

	// Insert into the message table (same as before).
	u, err := uuid.FromString(msg.GetUUID())
	if err != nil {
		h.log.Error("invalid message UUID", "uuid", msg.GetUUID(), "error", err)
		return err
	}
	var arr [16]byte
	copy(arr[:], u.Bytes())
	uid := pgtype.UUID{Bytes: arr, Valid: true}

	q := query.New(h.dbp)
	_, err = q.CreateMessage(ctx, query.CreateMessageParams{
		UUID:       uid,
		Sender:     msg.GetSender(),
		Recipients: msg.GetRecipients(),
		Subject:    converter.OptionalText(msg.GetSubject()),
		Body:       msg.GetBody(),
		// ... etc ...
	})
	if err != nil {
		h.log.Error("failed to insert message record (hostfiles)", "error", err)
		return err
	}

	for _, att := range msg.GetAttachments() {
		if err := h.SaveAttachment(ctx, &att); err != nil {
			h.log.Error("failed to save attachment (hostfiles)", "error", err)
			return err
		}
	}
	return nil
}

// SaveAttachment writes the file to disk and then inserts the metadata into "file".
func (h *HostfilesStorage) SaveAttachment(ctx context.Context, file *api.FileObject) error {
	fileUUID := file.GetUUID().Or("")
	u, err := uuid.FromString(fileUUID)
	if err != nil {
		h.log.Error("invalid file UUID", "uuid", fileUUID, "error", err)
		return err
	}

	name := file.GetName() // .Or("attachment")
	mime := file.GetMimeType().Or("application/octet-stream")
	size := file.GetSize().Or(0)

	// Subfolder approach: e.g. "ab" from the first 2 chars of the UUID, etc.
	sub := fileUUID[:2]
	dir := filepath.Join(h.rootFolder, sub)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("failed to create subdir: %w", err)
	}
	ext := filepath.Ext(name)
	// We'll just store everything as "regular" attachments for now:
	finalName := fmt.Sprintf("%s%s", fileUUID, ext)
	if ext == "" {
		finalName = fileUUID // no extension
	}

	filePath := filepath.Join(dir, finalName)
	h.log.Info("Saving file locally", "filePath", filePath)

	out, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer out.Close()

	if file.GetData().IsSet() {
		// If your .Data is raw or base64, decode if needed. Here we treat it as plain bytes.
		fileBytes := []byte(file.GetData().Or(""))
		if _, copyErr := out.Write(fileBytes); copyErr != nil {
			return copyErr
		}
	}

	q := query.New(h.dbp)
	var arr [16]byte
	copy(arr[:], u.Bytes())
	uid := pgtype.UUID{Bytes: arr, Valid: true}

	_, err = q.CreateFile(ctx, query.CreateFileParams{
		UUID:        uid,
		StorageType: "hostfiles",
		StorageUuid: pgtype.UUID{Valid: false}, // or set if you have a 'storage' record
		Name:        name,
		MimeType:    converter.PgText(mime),
		Size:        converter.PgInt8(size),
		Data:        nil, // no data in Postgres
		Path:        converter.PgText(filePath),
		IsRaw:       converter.PgBool(false), // or detect if raw
	})
	if err != nil {
		return err
	}

	h.log.Info("Attachment saved locally, meta in Postgres", "filePath", filePath)
	return nil
}
