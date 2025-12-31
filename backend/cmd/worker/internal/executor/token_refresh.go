package executor

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"golang.org/x/oauth2"
	googleOAuth2 "golang.org/x/oauth2/google"

	"github.com/shadowapi/shadowapi/backend/pkg/jobs"
)

// handleTokenRefresh executes a token refresh job.
// It makes an HTTP call to the OAuth provider to refresh the token.
func (e *Executor) handleTokenRefresh(ctx context.Context, data []byte) ([]byte, error) {
	var args jobs.TokenRefresherJobArgs
	if err := json.Unmarshal(data, &args); err != nil {
		return nil, fmt.Errorf("failed to unmarshal job args: %w", err)
	}

	e.log.Debug("refreshing token",
		"token_uuid", args.TokenUUID,
		"provider", args.Provider,
	)

	// Build OAuth2 config from job data
	endpoint, err := resolveEndpoint(args.Provider)
	if err != nil {
		return e.buildErrorResult(args.TokenUUID, err)
	}

	config := &oauth2.Config{
		ClientID:     args.ClientID,
		ClientSecret: args.ClientSecret,
		Endpoint:     endpoint,
	}

	// Create token from job data
	token := &oauth2.Token{
		AccessToken:  args.AccessToken,
		RefreshToken: args.RefreshToken,
		Expiry:       args.TokenExpiry,
	}

	// Refresh token (makes HTTP call to OAuth provider)
	newToken, err := config.TokenSource(ctx, token).Token()
	if err != nil {
		e.log.Error("token refresh failed",
			"token_uuid", args.TokenUUID,
			"error", err,
		)
		return e.buildErrorResult(args.TokenUUID, err)
	}

	e.log.Info("token refreshed successfully",
		"token_uuid", args.TokenUUID,
		"new_expiry", newToken.Expiry,
	)

	// Return successful result
	result := jobs.TokenRefreshResult{
		TokenUUID:    args.TokenUUID,
		Success:      true,
		AccessToken:  newToken.AccessToken,
		RefreshToken: newToken.RefreshToken,
		TokenExpiry:  newToken.Expiry,
		TokenType:    newToken.TokenType,
	}
	return json.Marshal(result)
}

// resolveEndpoint returns the OAuth2 endpoint for the given provider.
func resolveEndpoint(provider string) (oauth2.Endpoint, error) {
	switch strings.ToLower(provider) {
	case "google", "gmail":
		return googleOAuth2.Endpoint, nil
	default:
		return oauth2.Endpoint{}, fmt.Errorf("unknown provider: %s", provider)
	}
}

// buildErrorResult creates a failure result JSON.
func (e *Executor) buildErrorResult(tokenUUID string, err error) ([]byte, error) {
	result := jobs.TokenRefreshResult{
		TokenUUID: tokenUUID,
		Success:   false,
	}
	return json.Marshal(result)
}
