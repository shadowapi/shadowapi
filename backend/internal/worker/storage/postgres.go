package storage

import (
	"context"
	"github.com/shadowapi/shadowapi/backend/internal/converter"
	"log/slog"

	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/shadowapi/shadowapi/backend/pkg/api"
	"github.com/shadowapi/shadowapi/backend/pkg/query"
)

type PostgresStorage struct {
	log  *slog.Logger
	dbp  *pgxpool.Pool
	pgdb *query.Queries
}

func NewPostgresStorage(log *slog.Logger, dbp *pgxpool.Pool) *PostgresStorage {
	return &PostgresStorage{log: log, dbp: dbp}
}

func (s *PostgresStorage) SaveMessage(ctx context.Context, message *api.Message) error {
	s.log.Info("Saving message to Postgres", "message_uuid", message.GetUUID())

	// Insert into "message" table (same as before).
	// Here for brevity, only the essential fields are shown:
	u, err := uuid.FromString(message.UUID.Value)
	if err != nil {
		s.log.Error("invalid message UUID", "uuid", message.GetUUID(), "error", err)
		return err
	}
	var arr [16]byte
	copy(arr[:], u.Bytes())
	uid := pgtype.UUID{Bytes: arr, Valid: true}

	chatUUID, err := converter.ConvertOptStringToPgUUID(message.ChatUUID)
	if err != nil {
		s.log.Error("invalid chat UUID", "error", err)
		return err
	}
	threadUuid, err := converter.ConvertOptStringToPgUUID(message.ThreadUUID)
	if err != nil {
		s.log.Error("invalid thread UUID", "error", err)
		return err
	}

	_, err = s.pgdb.CreateMessage(ctx, query.CreateMessageParams{
		UUID:       uid,
		Sender:     message.GetSender(),
		Recipients: message.GetRecipients(),
		Subject:    converter.OptionalText(message.GetSubject()),
		Body:       message.GetBody(),
		Format:     message.Format,
		Type:       message.Type,
		ChatUuid:   chatUUID,
		ThreadUuid: threadUuid,
		BodyParsed: nil,
		Reactions:  nil,
	})

	if err != nil {
		s.log.Error("failed to insert message record", "error", err)
		return err
	}

	// If you wish to separately store attachments:
	for _, att := range message.GetAttachments() {
		if err := s.SaveAttachment(ctx, &att); err != nil {
			s.log.Error("failed to save attachment in Postgres", "error", err)
			return err
		}
	}

	return nil
}

// SaveAttachment inserts an attachment into the file table. No "isRaw" param here.
func (s *PostgresStorage) SaveAttachment(ctx context.Context, file *api.FileObject) error {
	s.log.Info("Saving file to Postgres", "file_uuid", file.GetUUID().Or(""))

	fileUUID := file.GetUUID().Or("")
	u, err := uuid.FromString(fileUUID)
	if err != nil {
		s.log.Error("invalid file UUID", "uuid", fileUUID, "error", err)
		return err
	}
	var arr [16]byte
	copy(arr[:], u.Bytes())
	uid := pgtype.UUID{Bytes: arr, Valid: true}

	name := file.GetName() // .Or("attachment")
	mime := file.GetMimeType().Or("application/octet-stream")
	size := file.GetSize().Or(0)

	// data field should contain the raw bytes if the user provided them.
	// For example, if `file.Data` is a base64 string or raw bytes in memory, adapt as needed.
	var fileBytes []byte
	if file.GetData().IsSet() {
		// If your FileObject.Data is a base64 string, decode it
		fileBytes = []byte(file.GetData().Or(""))
		// or do actual base64 decode if that is how it's stored
		// fileBytes, _ = base64.StdEncoding.DecodeString(...)
	}

	q := query.New(s.dbp)
	_, err = q.CreateFile(ctx, query.CreateFileParams{
		UUID: uid,

		StorageType: "postgres",
		StorageUuid: pgtype.UUID{Valid: false}, // or set if you have a storage record
		Name:        name,
		MimeType:    converter.PgText(mime),
		Size:        converter.PgInt8(size),
		Data:        fileBytes,
		Path:        pgtype.Text{Valid: false},
		IsRaw:       converter.PgBool(false), // or detect if raw
	})
	if err != nil {
		s.log.Error("failed to insert file record in DB (postgres)", "error", err)
		return err
	}

	s.log.Info("File saved in Postgres", "name", name)
	return nil
}
