// Package jobs provides shared job types used by both backend and distributed workers.
package jobs

import (
	"time"
)

// EmailFetchJobArgs holds all data needed for a distributed worker to fetch emails
// and write to external PostgreSQL storage. Workers have no database access.
type EmailFetchJobArgs struct {
	// Job identification
	JobUUID       string `json:"job_uuid"`
	SchedulerUUID string `json:"scheduler_uuid"`
	PipelineUUID  string `json:"pipeline_uuid"`

	// Email OAuth credentials
	Email       string `json:"email"`        // e.g., "user@gmail.com"
	Provider    string `json:"provider"`     // "gmail" or "google"
	AccessToken string `json:"access_token"` // Current OAuth2 access token

	// IMAP configuration (derived from provider or custom)
	IMAPHost string `json:"imap_host"` // e.g., "imap.gmail.com"
	IMAPPort int    `json:"imap_port"` // e.g., 993

	// Fetch parameters
	LastUID     uint32 `json:"last_uid"`      // Last processed IMAP UID (fetch UIDs > this value)
	BatchSize   int    `json:"batch_size"`    // Max emails per job (from scheduler.batch_size)
	MailboxName string `json:"mailbox_name"`  // e.g., "INBOX", "[Gmail]/All Mail"

	// Mapper configuration (extracted from pipeline.flow)
	MapperConfig MapperConfigData `json:"mapper_config"`

	// External PostgreSQL storage credentials
	StorageHost     string `json:"storage_host"`
	StoragePort     string `json:"storage_port"`
	StorageUser     string `json:"storage_user"`
	StoragePassword string `json:"storage_password"`
	StorageDatabase string `json:"storage_database"`
	StorageOptions  string `json:"storage_options"` // SSL mode, etc.
}

// MapperConfigData is a simplified mapper config for job payload
// (mirrors api.MapperConfig structure for JSON serialization)
type MapperConfigData struct {
	Version  string              `json:"version"`
	Mappings []MapperMappingData `json:"mappings"`
}

// MapperMappingData represents a single field mapping configuration
type MapperMappingData struct {
	SourceEntity string               `json:"source_entity"` // "message" or "contact"
	SourceField  string               `json:"source_field"`  // e.g., "subject", "body"
	TargetTable  string               `json:"target_table"`  // e.g., "emails"
	TargetField  string               `json:"target_field"`  // e.g., "email_subject"
	IsEnabled    bool                 `json:"is_enabled"`
	Transform    *MapperTransformData `json:"transform,omitempty"`
}

// MapperTransformData represents a field transformation
type MapperTransformData struct {
	Type   string         `json:"type"`             // "lowercase", "uppercase", etc.
	Params map[string]any `json:"params,omitempty"` // Transform-specific parameters
}

// EmailFetchResult contains the outcome of an email fetch job
type EmailFetchResult struct {
	JobUUID       string `json:"job_uuid"`
	SchedulerUUID string `json:"scheduler_uuid"`
	PipelineUUID  string `json:"pipeline_uuid"`

	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`

	// Statistics
	MessagesFetched int    `json:"messages_fetched"`
	MessagesStored  int    `json:"messages_stored"`
	ErrorCount      int    `json:"error_count"`
	LastUID         uint32 `json:"last_uid"` // Highest processed IMAP UID for incremental fetch

	// Duration
	DurationMs  int64     `json:"duration_ms"`
	StartedAt   time.Time `json:"started_at"`
	CompletedAt time.Time `json:"completed_at"`
}
