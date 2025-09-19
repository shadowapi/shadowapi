package handler

import (
	"context"
	"github.com/shadowapi/shadowapi/backend/pkg/api"
)

// GetProfile implements getProfile operation.
func (h *Handler) GetProfile(ctx context.Context) (*api.User, error) {
	return &api.User{}, nil
}

// UpdateProfile implements updateProfile operation.
func (h *Handler) UpdateProfile(ctx context.Context, req *api.UserProfile) (*api.User, error) {
	return &api.User{}, nil
}
