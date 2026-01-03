// Package jobs provides shared job types used by both backend and distributed workers.
package jobs

import "time"

// TableConfig represents a table definition for querying
type TableConfig struct {
	Name   string        `json:"name"`
	Fields []FieldConfig `json:"fields"`
}

// FieldConfig represents a field definition within a table
type FieldConfig struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

// MessageQueryJobArgs holds all data needed for a distributed worker to query messages
// from a PostgreSQL storage and stream them as individual JSON records to NATS.
type MessageQueryJobArgs struct {
	// Job identification
	JobUUID       string `json:"job_uuid"`
	WorkspaceSlug string `json:"workspace_slug"`

	// Query parameters
	Limit  int `json:"limit"`  // Max messages to return per table (default 100)
	Offset int `json:"offset"` // Offset for pagination per table

	// PostgreSQL storage credentials
	StorageHost     string `json:"storage_host"`
	StoragePort     string `json:"storage_port"`
	StorageUser     string `json:"storage_user"`
	StoragePassword string `json:"storage_password"`
	StorageDatabase string `json:"storage_database"`
	StorageOptions  string `json:"storage_options"` // SSL mode, etc.

	// Tables configuration - query all configured tables
	Tables []TableConfig `json:"tables"`

	// NATS subject to publish individual records to
	// Format: shadowapi.data.workspace.{slug}.messages
	NATSSubject string `json:"nats_subject"`
}

// MessageQueryResult contains the outcome of a message query job
type MessageQueryResult struct {
	JobUUID       string `json:"job_uuid"`
	WorkspaceSlug string `json:"workspace_slug"`

	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`

	// Statistics
	TablesQueried     []string `json:"tables_queried"`     // Names of tables that were queried
	MessagesQueried   int      `json:"messages_queried"`   // Total messages found
	MessagesPublished int      `json:"messages_published"` // Messages sent to NATS
	ErrorCount        int      `json:"error_count"`

	// Duration
	DurationMs  int64     `json:"duration_ms"`
	StartedAt   time.Time `json:"started_at"`
	CompletedAt time.Time `json:"completed_at"`
}

// MessageRecord represents a single message to be published to NATS
type MessageRecord struct {
	UUID              string         `json:"uuid"`
	Format            string         `json:"format,omitempty"`
	Type              string         `json:"type,omitempty"`
	Sender            string         `json:"sender"`
	Recipients        []string       `json:"recipients,omitempty"`
	Subject           string         `json:"subject,omitempty"`
	Body              string         `json:"body"`
	BodyParsed        map[string]any `json:"body_parsed,omitempty"`
	Attachments       []any          `json:"attachments,omitempty"`
	Meta              map[string]any `json:"meta,omitempty"`
	ExternalMessageID string         `json:"external_message_id,omitempty"`
	CreatedAt         time.Time      `json:"created_at"`
	UpdatedAt         time.Time      `json:"updated_at,omitempty"`
}
