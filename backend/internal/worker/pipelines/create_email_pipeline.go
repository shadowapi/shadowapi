package pipelines

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shadowapi/shadowapi/backend/internal/worker/extractors"
	"github.com/shadowapi/shadowapi/backend/internal/worker/filters"
	stor "github.com/shadowapi/shadowapi/backend/internal/worker/storage"
	"github.com/shadowapi/shadowapi/backend/internal/worker/types"
	"github.com/shadowapi/shadowapi/backend/pkg/api"
	"github.com/shadowapi/shadowapi/backend/pkg/query"
	"log/slog"
)

type EmailPipeline struct {
	log       *slog.Logger
	extractor types.Extractor
	filter    types.Filter
	storage   types.Storage
}

func NewEmailPipeline(log *slog.Logger, extractor types.Extractor, filter types.Filter, storage types.Storage) types.Pipeline {
	return &EmailPipeline{
		log:       log,
		extractor: extractor,
		filter:    filter,
		storage:   storage,
	}
}

func (p *EmailPipeline) Run(ctx context.Context, message *api.Message) error {
	p.log.Info("Running pipeline", "message_uuid", message.UUID)
	if !p.filter.Apply(ctx, message) {
		p.log.Info("Message blocked by sync policy", "sender", message.Sender)
		return nil
	}
	contact, err := p.extractor.ExtractContact(message)
	if err != nil {
		p.log.Error("Failed to extract contact", "error", err)
		return err
	}
	p.log.Info("Extracted contact", "contact_uuid", contact.UUID)
	if err := p.storage.SaveMessage(ctx, message); err != nil {
		p.log.Error("Failed to save message", "error", err)
		return err
	}
	return nil
}

func CreateEmailPipelines(ctx context.Context, log *slog.Logger, dbp *pgxpool.Pool) *map[string]types.Pipeline {
	pipelinesMap := make(map[string]types.Pipeline)
	q := query.New(dbp)
	params := query.GetPipelinesParams{
		OrderBy:        "created_at",
		OrderDirection: "asc",
		Offset:         0,
		Limit:          100,
		Type:           "email",
		UUID:           pgtype.UUID{Valid: false},
		IsEnabled:      1,
		Name:           "",
	}
	pipes, err := q.GetPipelines(ctx, params)
	if err != nil {
		log.Error("Failed to fetch email pipelines", "error", err)
		return &pipelinesMap
	}
	for _, pipe := range pipes {
		syncParams := query.GetSyncPoliciesParams{
			OrderBy:        "created_at",
			OrderDirection: "desc",
			Offset:         0,
			Limit:          1,
			Type:           "email",
			UUID:           "",
			SyncAll:        -1,
		}
		policies, err := q.GetSyncPolicies(ctx, syncParams)
		var apiPolicy api.SyncPolicy
		if err != nil || len(policies) == 0 {
			apiPolicy = api.SyncPolicy{
				Type:    api.NewOptString("email"),
				SyncAll: api.NewOptBool(true),
			}
		} else {
			converted, err := convertSyncPolicy(policies[0])
			if err != nil {
				log.Error("Failed to convert sync policy", "error", err)
				apiPolicy = api.SyncPolicy{
					Type:    api.NewOptString("email"),
					SyncAll: api.NewOptBool(true),
				}
			} else {
				apiPolicy = converted
			}
		}
		filter := filters.NewSyncPolicyFilter(apiPolicy, log)
		extractor := extractors.NewContactExtractor()
		if pipe.DatasourceUUID == nil {
			log.Error("Pipeline has nil DatasourceUUID")
			continue
		}
		dsRow, err := q.GetDatasource(ctx, uuidToPgUUID(*pipe.DatasourceUUID))
		if err != nil {
			log.Error("Failed to get datasource", "error", err)
			continue
		}
		storageUUID, err := extractStorageUUID(dsRow.Datasource.Settings)
		if err != nil {
			log.Error("Failed to extract storage UUID from datasource settings", "error", err)
			continue
		}
		uuidArr, err := bytesToUUID(storageUUID)
		if err != nil {
			log.Error("Invalid storage UUID", "error", err)
			continue
		}
		storageRow, err := q.GetStorage(ctx, pgtype.UUID{Bytes: uuidArr, Valid: true})
		if err != nil {
			log.Error("Failed to get storage", "error", err)
			continue
		}
		var storageBackend types.Storage
		switch storageRow.Storage.Type {
		case "s3":
			// Provide all required args to match the function signature:
			// func NewS3Storage(log *slog.Logger, s3Client *s3.S3, bucketName string, pgdb *query.Queries) *S3Storage
			storageBackend = stor.NewS3Storage(
				log,
				nil,                   // placeholder s3Client
				"example-bucket-name", // or read from config
				query.New(dbp),        // pass your Queries
			)

		case "hostfiles":
			storageBackend = stor.NewHostfilesStorage(log, "./data", dbp)
		case "postgres":
			storageBackend = stor.NewPostgresStorage(log, dbp)
		default:
			log.Error("Unknown storage type", "type", storageRow.Storage.Type)
			continue
		}
		pipeline := NewEmailPipeline(log, extractor, filter, storageBackend)
		pipelinesMap[dsRow.Datasource.UUID.String()] = pipeline
	}
	return &pipelinesMap
}

func extractStorageUUID(settings []byte) ([]byte, error) {
	if len(settings) == 0 {
		return nil, errors.New("empty storage settings")
	}
	// In a real implementation, parse JSON to extract a 16-byte UUID.
	return settings, nil
}

func bytesToUUID(b []byte) ([16]byte, error) {
	var arr [16]byte
	if len(b) != 16 {
		return arr, errors.New("invalid UUID length")
	}
	copy(arr[:], b)
	return arr, nil
}

func uuidToPgUUID(u uuid.UUID) pgtype.UUID {
	var pg pgtype.UUID
	copy(pg.Bytes[:], u.Bytes())
	pg.Valid = true
	return pg
}

func convertSyncPolicy(row query.GetSyncPoliciesRow) (api.SyncPolicy, error) {
	var policy api.SyncPolicy
	policy.SetUUID(api.NewOptString(row.UUID.String()))
	policy.SetType(api.NewOptString(row.Type))
	policy.SetBlocklist(row.Blocklist)
	policy.SetExcludeList(row.ExcludeList)
	policy.SetSyncAll(api.NewOptBool(row.SyncAll))
	var settings api.SyncPolicySettings
	if len(row.Settings) > 0 {
		if err := json.Unmarshal(row.Settings, &settings); err != nil {
			return policy, err
		}
	} else {
		settings = make(api.SyncPolicySettings)
	}
	policy.SetSettings(api.NewOptSyncPolicySettings(settings))
	policy.SetCreatedAt(api.NewOptDateTime(row.CreatedAt.Time))
	policy.SetUpdatedAt(api.NewOptDateTime(row.UpdatedAt.Time))
	return policy, nil
}
