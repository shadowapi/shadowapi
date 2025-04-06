package storage

import (
	"context"
	"fmt"
	"github.com/shadowapi/shadowapi/backend/internal/converter"
	"log/slog"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/shadowapi/shadowapi/backend/pkg/api"
	"github.com/shadowapi/shadowapi/backend/pkg/query"
)

type S3Storage struct {
	log      *slog.Logger
	s3Client *s3.S3
	bucket   string
	pgdb     *query.Queries
}

func NewS3Storage(log *slog.Logger, s3Client *s3.S3, bucketName string, pgdb *query.Queries) *S3Storage {
	return &S3Storage{
		log:      log,
		s3Client: s3Client,
		bucket:   bucketName,
		pgdb:     pgdb,
	}
}

func (s *S3Storage) SaveMessage(ctx context.Context, message *api.Message) error {

	if s.pgdb == nil {
		s.log.Warn("pgdb is nil, skipping DB insert for SaveMessage")
		return nil
	}
	s.log.Info("Saving message to S3 (metadata in Postgres)", "message_uuid", message.GetUUID())

	u, err := uuid.FromString(message.GetUUID())
	if err != nil {
		s.log.Error("invalid message UUID", "error", err)
		return err
	}
	var arr [16]byte
	copy(arr[:], u.Bytes())
	uid := pgtype.UUID{Bytes: arr, Valid: true}

	_, err = s.pgdb.CreateMessage(ctx, query.CreateMessageParams{
		UUID:       uid,
		Sender:     message.GetSender(),
		Recipients: message.GetRecipients(),
		Subject:    converter.OptionalText(message.GetSubject()),
		Body:       message.GetBody(),
		Source:     nil,
		Type:       nil,
		ChatUuid:   nil,
		ThreadUuid: nil,
		BodyParsed: nil,
		Reactions:  nil,
	})
	if err != nil {
		s.log.Error("failed to insert message record (S3 mode)", "error", err)
		return err
	}

	for _, att := range message.GetAttachments() {
		if err := s.SaveAttachment(ctx, &att); err != nil {
			return err
		}
	}
	return nil
}

func (s *S3Storage) SaveAttachment(ctx context.Context, file *api.FileObject) error {
	if s.pgdb == nil {
		s.log.Warn("pgdb is nil, skipping DB insert for SaveAttachment")
	}
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

	key := s.generateObjectKey(fileUUID, name)

	if file.GetData().IsSet() && s.s3Client != nil {
		dataBytes := []byte(file.GetData().Or(""))
		_, s3Err := s.s3Client.PutObjectWithContext(ctx, &s3.PutObjectInput{
			Bucket:      aws.String(s.bucket),
			Key:         aws.String(key),
			Body:        converter.BytesToReader(dataBytes),
			ContentType: aws.String(mime),
		})
		if s3Err != nil {
			s.log.Error("S3 upload failed", "error", s3Err)
			return s3Err
		}
	} else {
		s.log.Warn("No data in file object or s3Client is nil, skipping S3 upload", "file_uuid", fileUUID)
	}

	if s.pgdb != nil {
		_, err = s.pgdb.CreateFile(ctx, query.CreateFileParams{
			UUID:        uid,
			StorageType: "s3",
			StorageUuid: pgtype.UUID{Valid: false},
			Name:        name,
			MimeType:    converter.PgText(mime),
			Size:        converter.PgInt8(size),
			Data:        nil,
			Path:        converter.PgText(key),
			IsRaw:       converter.PgBool(false),
		})
		if err != nil {
			s.log.Error("failed to insert file record in DB (S3)", "error", err)
			return err
		}
	}

	s.log.Info("Attachment saved to S3 & metadata to Postgres", "file_uuid", fileUUID, "key", key)
	return nil
}

func (s *S3Storage) generateObjectKey(fileUUID, originalName string) string {
	ext := converter.FileExt(originalName)
	sub := fileUUID[:2]
	return fmt.Sprintf("files/%s/%s%s", sub, fileUUID, ext)
}
