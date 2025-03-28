package pipelines

import (
	"log/slog"

	"github.com/shadowapi/shadowapi/backend/internal/worker"
	"github.com/shadowapi/shadowapi/backend/internal/worker/extractors"
	"github.com/shadowapi/shadowapi/backend/internal/worker/filters"
	"github.com/shadowapi/shadowapi/backend/internal/worker/storage"
)

// CreateEmailPipeline instantiates a pipeline for processing email messages.
func CreateEmailPipeline(log *slog.Logger) worker.Pipeline {
	// Extractor: extracts contact details from the email message body.
	extractor := extractors.NewContactExtractor()
	// Filter: for example, allow only messages from specific domains.
	filter := filters.NewSyncPolicyFilter([]string{"@example.com", "@mydomain.com"})
	// Storage: persist full messages in Postgres (or S3/hostfiles as needed).
	storageBackend := storage.NewPostgresStorage(log)
	// Create a simple pipeline that uses the three components.
	return worker.NewSimplePipeline(log, extractor, filter, storageBackend)
}
