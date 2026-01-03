package handler

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/shadowapi/shadowapi/backend/internal/auth/oauth2"
	"github.com/shadowapi/shadowapi/backend/pkg/api"
)

// AuthLogin handles the Hydra login redirect.
// GET /api/v1/auth/login?login_challenge=xxx
// If the user has a remembered session, accept immediately. Otherwise redirect to frontend login page.
func (h *Handler) AuthLogin(ctx context.Context, params api.AuthLoginParams) (api.AuthLoginRes, error) {
	if h.oauth2Svc == nil {
		return nil, fmt.Errorf("OAuth2 service not configured")
	}

	// Get login request from Hydra
	loginReq, err := h.oauth2Svc.hydraClient.GetLoginRequest(ctx, params.LoginChallenge)
	if err != nil {
		h.log.Error("failed to get login request", "error", err)
		return nil, ErrWithCode(http.StatusBadRequest, fmt.Errorf("invalid login challenge"))
	}

	// If skip is true, the user has a remembered session - accept immediately
	if loginReq.Skip {
		redirectURL, err := h.oauth2Svc.hydraClient.AcceptLoginRequest(
			ctx,
			params.LoginChallenge,
			loginReq.Subject,
			false, // don't extend remember
			0,
		)
		if err != nil {
			h.log.Error("failed to accept skipped login", "error", err)
			return nil, ErrWithCode(http.StatusInternalServerError, fmt.Errorf("login acceptance failed"))
		}

		h.log.Debug("login skipped (remembered session)", "subject", loginReq.Subject)
		return &api.AuthLoginFound{
			Location: api.NewOptString(redirectURL),
		}, nil
	}

	// Redirect to frontend login page with the challenge
	loginPageURL := fmt.Sprintf("%s/login?login_challenge=%s", h.oauth2Svc.baseURL, params.LoginChallenge)

	h.log.Debug("redirecting to login page", "challenge", params.LoginChallenge[:8]+"...")
	return &api.AuthLoginFound{
		Location: api.NewOptString(loginPageURL),
	}, nil
}

// AuthLoginSubmit handles login form submission.
// POST /api/v1/auth/login
// Authenticates user credentials and accepts the Hydra login request.
func (h *Handler) AuthLoginSubmit(ctx context.Context, req *api.AuthLoginSubmitReq) (api.AuthLoginSubmitRes, error) {
	if h.oauth2Svc == nil {
		return nil, fmt.Errorf("OAuth2 service not configured")
	}

	// Authenticate user with dbauth
	user, err := h.userManager.AuthenticateUser(ctx, req.Email, req.Password)
	if err != nil {
		h.log.Debug("authentication failed", "email", req.Email, "error", err)
		return nil, err // Returns 401 for invalid credentials
	}

	// Get subject (user UUID)
	subject := user.UUID.Value

	// Determine remember settings
	remember := req.Remember.Or(false)
	rememberFor := 0
	if remember {
		rememberFor = 3600 // 1 hour
	}

	// Accept login request with Hydra
	redirectURL, err := h.oauth2Svc.hydraClient.AcceptLoginRequest(
		ctx,
		req.LoginChallenge,
		subject,
		remember,
		rememberFor,
	)
	if err != nil {
		h.log.Error("failed to accept login request", "error", err, "subject", subject)
		// Check if the login challenge has expired
		if errors.Is(err, oauth2.ErrLoginChallengeExpired) {
			return nil, ErrWithCode(http.StatusBadRequest, fmt.Errorf("login session expired, please try again"))
		}
		return nil, ErrWithCode(http.StatusInternalServerError, fmt.Errorf("login acceptance failed"))
	}

	h.log.Info("user authenticated successfully", "email", req.Email, "subject", subject)

	return &api.AuthLoginSubmitOK{
		RedirectTo: redirectURL,
	}, nil
}

// AuthConsent handles the Hydra consent redirect.
// GET /api/v1/auth/consent?consent_challenge=xxx
// Auto-approves consent and redirects back to Hydra.
func (h *Handler) AuthConsent(ctx context.Context, params api.AuthConsentParams) (api.AuthConsentRes, error) {
	if h.oauth2Svc == nil {
		return nil, fmt.Errorf("OAuth2 service not configured")
	}

	// Get consent request from Hydra
	consentReq, err := h.oauth2Svc.hydraClient.GetConsentRequest(ctx, params.ConsentChallenge)
	if err != nil {
		h.log.Error("failed to get consent request", "error", err)
		return nil, ErrWithCode(http.StatusBadRequest, fmt.Errorf("invalid consent challenge"))
	}

	// Auto-approve: Accept all requested scopes and audiences
	// No tenant info needed in workspace-based architecture
	redirectURL, err := h.oauth2Svc.hydraClient.AcceptConsentRequest(
		ctx,
		params.ConsentChallenge,
		consentReq.RequestedScope,
		consentReq.RequestedAccessTokenAudience,
		consentReq.Subject,
		true,  // remember consent
		3600,  // remember for 1 hour
		nil,   // no extra access token claims needed
	)
	if err != nil {
		h.log.Error("failed to accept consent request", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, fmt.Errorf("consent acceptance failed"))
	}

	h.log.Debug("consent auto-approved",
		"subject", consentReq.Subject,
		"scopes", consentReq.RequestedScope,
	)

	return &api.AuthConsentFound{
		Location: api.NewOptString(redirectURL),
	}, nil
}
