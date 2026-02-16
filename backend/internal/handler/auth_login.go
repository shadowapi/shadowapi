package handler

import (
	"context"
	"fmt"
	"net/http"

	"github.com/shadowapi/shadowapi/backend/pkg/api"
)

// AuthLogin handles the OIDC login redirect.
// GET /api/v1/auth/login?auth_request_id=xxx
// Redirects to the frontend login page with the auth_request_id.
func (h *Handler) AuthLogin(ctx context.Context, params api.AuthLoginParams) (api.AuthLoginRes, error) {
	if h.oauth2Svc == nil {
		return nil, fmt.Errorf("OAuth2 service not configured")
	}

	// Redirect to frontend login page with the auth_request_id
	loginPageURL := fmt.Sprintf("%s/login?auth_request_id=%s", h.oauth2Svc.baseURL, params.AuthRequestID)

	h.log.Debug("redirecting to login page", "auth_request_id", params.AuthRequestID[:min(8, len(params.AuthRequestID))]+"...")
	return &api.AuthLoginFound{
		Location: api.NewOptString(loginPageURL),
	}, nil
}

// AuthLoginSubmit handles login form submission.
// POST /api/v1/auth/login
// Authenticates user credentials and returns an HMAC-signed callback URL to the OIDC server.
func (h *Handler) AuthLoginSubmit(ctx context.Context, req *api.AuthLoginSubmitReq) (api.AuthLoginSubmitRes, error) {
	if h.oauth2Svc == nil {
		return nil, fmt.Errorf("OAuth2 service not configured")
	}

	// Authenticate user
	user, err := h.userManager.AuthenticateUser(ctx, req.Email, req.Password)
	if err != nil {
		h.log.Debug("authentication failed", "email", req.Email, "error", err)
		return nil, err
	}

	subject := user.UUID.Value

	// Build claims for the OIDC callback
	claims := map[string]interface{}{
		"email":          user.Email,
		"email_verified": true,
		"name":           fmt.Sprintf("%s %s", user.FirstName, user.LastName),
		"given_name":     user.FirstName,
		"family_name":    user.LastName,
	}

	// Build HMAC-signed callback URL
	callbackURL, err := h.oauth2Svc.oidcClient.BuildLoginCallbackURL(req.AuthRequestID, subject, claims)
	if err != nil {
		h.log.Error("failed to build login callback URL", "error", err, "subject", subject)
		return nil, ErrWithCode(http.StatusInternalServerError, fmt.Errorf("login callback failed"))
	}

	h.log.Info("user authenticated successfully", "email", req.Email, "subject", subject)

	return &api.AuthLoginSubmitOK{
		RedirectTo: callbackURL,
	}, nil
}
