package handler

import (
	"context"

	"github.com/shadowapi/shadowapi/backend/pkg/api"
)

// AuthLogin implements auth-login operation.
//
// Authenticate user with email and password.
//
// POST /auth/login
func (h *Handler) AuthLogin(ctx context.Context, req *api.AuthLoginRequest) (api.AuthLoginRes, error) {
	h.log.Info("auth login attempt", "email", req.Email)

	// For now, return a simple success response
	// In a real implementation, you would validate credentials against your auth system
	if req.Email == "" || req.Password == "" {
		return &api.AuthLoginUnauthorized{
			Success: false,
			Message: api.NewOptNilString("Email and password are required"),
		}, nil
	}

	// Simple validation - accept any non-empty credentials for demo
	// In production, you would integrate with your authentication system
	h.log.Info("login successful", "email", req.Email)

	return &api.AuthLoginOK{
		Success:      true,
		SessionToken: api.NewOptNilString("demo-session-token-" + req.Email),
		Message:      api.NewOptNilString("Login successful"),
	}, nil
}