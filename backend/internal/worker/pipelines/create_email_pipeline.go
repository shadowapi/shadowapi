package pipelines

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shadowapi/shadowapi/backend/pkg/api"
	"log/slog"

	"github.com/shadowapi/shadowapi/backend/internal/worker/extractors"
	"github.com/shadowapi/shadowapi/backend/internal/worker/filters"
	"github.com/shadowapi/shadowapi/backend/internal/worker/storage"
	"github.com/shadowapi/shadowapi/backend/internal/worker/types"
)

// SimplePipeline is an example implementation of the Pipeline interface.
type SimplePipeline struct {
	log       *slog.Logger
	extractor types.Extractor
	filter    types.Filter
	storage   types.Storage
}

// NewSimplePipeline creates a new pipeline.
func NewSimplePipeline(log *slog.Logger, extractor types.Extractor, filter types.Filter, storage types.Storage) types.Pipeline {
	return &SimplePipeline{
		log:       log,
		extractor: extractor,
		filter:    filter,
		storage:   storage,
	}
}

// Run executes the pipeline on a message.
func (p *SimplePipeline) Run(ctx context.Context, message *api.Message) error {
	p.log.Info("Running pipeline", "message_uuid", message.UUID)

	// Apply filter.
	if filtered := p.filter.Apply(ctx, message); !filtered {
		p.log.Info("Message blocked by policy", "sender", message.Sender)
		return nil
	}

	// Extract contact.
	contact, err := p.extractor.ExtractContact(message)
	if err != nil {
		p.log.Error("Failed to extract contact", "error", err)
		return err
	}
	p.log.Info("Extracted contact", "contact_uuid", contact.UUID)

	// Save the message.
	if err := p.storage.SaveMessage(ctx, message); err != nil {
		p.log.Error("Failed to save message", "error", err)
		return err
	}
	return nil
}

// CreateEmailPipeline instantiates a pipeline for processing email messages.
func CreateEmailPipeline(ctx context.Context, log *slog.Logger, dbp *pgxpool.Pool) types.Pipeline {
	// Extractor: extracts contact details from the email message body.
	extractor := extractors.NewContactExtractor()
	// Filter: for example, allow only messages from specific domains.
	filter := filters.NewSyncPolicyFilter([]string{"@example.com", "@mydomain.com"})
	// Storage: persist full messages in Postgres (or S3/hostfiles as needed)
	storageBackend := storage.NewPostgresStorage(log, dbp)
	// Create a simple pipeline that uses the three components.
	return NewSimplePipeline(log, extractor, filter, storageBackend)
}
