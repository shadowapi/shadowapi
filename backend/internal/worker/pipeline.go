// Pipeline execution engine (core logic)
package worker

/* remove
import (
	"context"
	"github.com/shadowapi/shadowapi/backend/internal/worker/types"

	"github.com/shadowapi/shadowapi/backend/pkg/api"
	"log/slog"
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
	if !p.filter.Apply(message) {
		p.log.Info("Message filtered out", "message_uuid", message.UUID)
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
*/
