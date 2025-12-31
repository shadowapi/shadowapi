// Package jobs provides shared job types used by both backend and distributed workers.
package jobs

import "time"

// TokenRefresherJobArgs holds the arguments for a token refresh job.
// This includes all data needed to refresh the token without database access.
type TokenRefresherJobArgs struct {
	SchedulerUUID string    `json:"scheduler_uuid"`
	JobUUID       string    `json:"job_uuid"`
	TokenUUID     string    `json:"token_uuid"`
	Expiry        time.Time `json:"expiry"`

	// OAuth2 token data
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	TokenExpiry  time.Time `json:"token_expiry"`

	// OAuth2 client config (needed to refresh)
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	Provider     string `json:"provider"` // e.g., "google", "gmail" - used to resolve token endpoint
}

// TokenRefreshResult contains the outcome of a token refresh job.
type TokenRefreshResult struct {
	TokenUUID string `json:"token_uuid"`
	Success   bool   `json:"success"`

	// New token data (if successful)
	AccessToken  string    `json:"access_token,omitempty"`
	RefreshToken string    `json:"refresh_token,omitempty"`
	TokenExpiry  time.Time `json:"token_expiry,omitempty"`
	TokenType    string    `json:"token_type,omitempty"`
}
