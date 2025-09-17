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
func (h *Handler) AuthLogin(ctx context.Context, req *api.AuthLoginReq) (api.AuthLoginRes, error) {
	result, err := h.userManager.Login(ctx, req.Email, req.Password)
	if err != nil {
		return nil, err
	}

	if !result.Success {
		return &api.Error{
			Status: api.NewOptInt64(401),
			Detail: api.NewOptString("Invalid email or password"),
		}, nil
	}

	return &api.AuthLoginOK{
		Success:      api.NewOptBool(true),
		SessionToken: api.NewOptString(result.SessionToken),
	}, nil
}
