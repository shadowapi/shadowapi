package auth

import (
	"context"

	"github.com/shadowapi/shadowapi/backend/pkg/api"
)

// UserManager defines the interface for user management operations
type UserManager interface {
	// CreateUser creates a new user
	CreateUser(ctx context.Context, user *api.User) (*api.User, error)

	// GetUser retrieves a user by UUID
	GetUser(ctx context.Context, uuid string) (*api.User, error)

	// UpdateUser updates an existing user
	UpdateUser(ctx context.Context, user *api.User, uuid string) (*api.User, error)

	// DeleteUser deletes a user by UUID
	DeleteUser(ctx context.Context, uuid string) error

	// ListUsers returns a list of all users
	ListUsers(ctx context.Context) ([]api.User, error)
}