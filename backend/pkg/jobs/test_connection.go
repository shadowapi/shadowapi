// Package jobs provides shared job types used by both backend and distributed workers.
package jobs

import "time"

// Error codes for test connection results.
const (
	ErrorCodeAuthFailed          = "auth_failed"
	ErrorCodeInvalidCredentials  = "invalid_credentials"
	ErrorCodeConnectionRefused   = "connection_refused"
	ErrorCodeConnectionTimeout   = "connection_timeout"
	ErrorCodeHostUnreachable     = "host_unreachable"
	ErrorCodeDNSFailure          = "dns_failure"
	ErrorCodeSSLRequired         = "ssl_required"
	ErrorCodeIMAPDisabled        = "imap_disabled"
	ErrorCodeOAuthScopeInsufficient = "oauth_scope_insufficient"
	ErrorCodeUnknown             = "unknown"
)

// TestConnectionEmailOAuthJobArgs holds the arguments for testing OAuth email datasource connectivity.
// This includes all data needed to test IMAP connection without database access.
type TestConnectionEmailOAuthJobArgs struct {
	JobUUID        string `json:"job_uuid"`
	DatasourceUUID string `json:"datasource_uuid"`

	// Email configuration
	Email    string `json:"email"`
	Provider string `json:"provider"` // "gmail" or "google"

	// OAuth2 credentials (passed directly, worker has no DB access)
	AccessToken string `json:"access_token"`

	// IMAP server config (derived from provider or custom)
	IMAPHost string `json:"imap_host"` // e.g., "imap.gmail.com"
	IMAPPort int    `json:"imap_port"` // e.g., 993
}

// TestConnectionPostgresJobArgs holds the arguments for testing PostgreSQL storage connectivity.
// This includes all data needed to test database connection without main app database access.
type TestConnectionPostgresJobArgs struct {
	JobUUID     string `json:"job_uuid"`
	StorageUUID string `json:"storage_uuid"`

	// Connection parameters
	Host     string `json:"host"`
	Port     string `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	Database string `json:"database"` // defaults to "postgres" if empty
	Options  string `json:"options"`  // SSL mode, etc.
}

// TestConnectionResult contains the outcome of a connection test job.
// This is the common result structure for all test connection jobs.
type TestConnectionResult struct {
	JobUUID      string `json:"job_uuid"`
	ResourceType string `json:"resource_type"` // "email_oauth", "postgres"
	ResourceUUID string `json:"resource_uuid"`

	Success      bool   `json:"success"`
	ErrorCode    string `json:"error_code,omitempty"`
	ErrorMessage string `json:"error_message,omitempty"`
	ErrorDetails string `json:"error_details,omitempty"`

	// Timing information
	DurationMs int64     `json:"duration_ms"`
	TestedAt   time.Time `json:"tested_at"`

	// Resource-specific details (e.g., IMAP host/port, DB version)
	Details map[string]any `json:"details,omitempty"`
}
