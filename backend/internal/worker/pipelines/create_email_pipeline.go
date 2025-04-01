package pipelines

import (
	"context"
	"encoding/json"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shadowapi/shadowapi/backend/internal/worker/extractors"
	"github.com/shadowapi/shadowapi/backend/internal/worker/filters"
	"github.com/shadowapi/shadowapi/backend/internal/worker/storage"
	"github.com/shadowapi/shadowapi/backend/internal/worker/types"
	"github.com/shadowapi/shadowapi/backend/pkg/api"
	"github.com/shadowapi/shadowapi/backend/pkg/query"
	"log/slog"
)

// EmailPipeline implements the Pipeline interface.
type EmailPipeline struct {
	log       *slog.Logger
	extractor types.Extractor
	filter    types.Filter
	storage   types.Storage
}

// NewEmailPipeline creates a new EmailPipeline.
func NewEmailPipeline(log *slog.Logger, extractor types.Extractor, filter types.Filter, storage types.Storage) types.Pipeline {
	return &EmailPipeline{
		log:       log,
		extractor: extractor,
		filter:    filter,
		storage:   storage,
	}
}

// Run executes the pipeline steps: filtering, extracting, and saving.
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

// convertSyncPolicy converts a query sync policy row to an API sync policy.
func convertSyncPolicy(row query.GetSyncPoliciesRow) (api.SyncPolicy, error) {
	var policy api.SyncPolicy
	policy.SetUUID(api.NewOptString(row.UUID.String()))

	policy.SetType(row.Type)
	policy.SetBlocklist(row.Blocklist)
	policy.SetExcludeList(row.ExcludeList)
	policy.SetSyncAll(api.NewOptBool(row.SyncAll))

	// Unmarshal the settings JSON bytes into the API SyncPolicySettings type.
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

// CreateEmailPipeline instantiates a pipeline for processing email messages.
// It loads the sync policy for "email" from the database and uses it in the filter.
func CreateEmailPipeline(ctx context.Context, log *slog.Logger, dbp *pgxpool.Pool) types.Pipeline {
	extractor := extractors.NewContactExtractor()

	// Load the sync policy for the "email" service.
	queries := query.New(dbp)
	policies, err := queries.GetSyncPolicies(ctx, query.GetSyncPoliciesParams{
		OrderBy:        "created_at",
		OrderDirection: "desc",
		Offset:         0,
		Limit:          1,
		Type:           "email",
		UUID:           "",
		SyncAll:        -1, // -1 means do not filter by SyncAll in the query.
	})
	var apiPolicy api.SyncPolicy
	if err != nil {
		log.Error("Failed to get sync policy from DB", "error", err)
		// If error occurs, default to a policy that allows all messages.
		apiPolicy = api.SyncPolicy{
			Type:    "email",
			SyncAll: api.NewOptBool(true),
		}
	} else if len(policies) == 0 {
		// No policy exists; default to allowing all messages.
		apiPolicy = api.SyncPolicy{
			Type:    "email",
			SyncAll: api.NewOptBool(true),
		}
	} else {
		// Convert the first returned policy to API format.
		converted, err := convertSyncPolicy(policies[0])
		if err != nil {
			log.Error("Failed to convert sync policy", "error", err)
			// Fallback to allow all messages.
			apiPolicy = api.SyncPolicy{
				Type:    "email",
				SyncAll: api.NewOptBool(true),
			}
		} else {
			apiPolicy = converted
		}
	}

	filter := filters.NewSyncPolicyFilter(apiPolicy, log)
	storageBackend := storage.NewPostgresStorage(log, dbp)
	return NewEmailPipeline(log, extractor, filter, storageBackend)
}
