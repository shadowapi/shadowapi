package workspace

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/shadowapi/shadowapi/backend/internal/converter"
)

var (
	// ErrWorkspaceRequired is returned when workspace context is missing but required
	ErrWorkspaceRequired = errors.New("workspace context required")
	// ErrInvalidWorkspace is returned when workspace UUID is invalid
	ErrInvalidWorkspace = errors.New("invalid workspace UUID")
)

// RequireWorkspaceUUID extracts the workspace UUID from context and converts it to pgtype.UUID.
// Returns an error if workspace context is missing or invalid.
// This enforces strict workspace filtering for all workspace-scoped resources.
func RequireWorkspaceUUID(ctx context.Context) (pgtype.UUID, error) {
	workspaceID := GetWorkspaceUUID(ctx)
	if workspaceID == "" {
		return pgtype.UUID{}, ErrWorkspaceRequired
	}

	pgUUID, err := converter.ConvertStringToPgUUID(workspaceID)
	if err != nil {
		return pgtype.UUID{}, ErrInvalidWorkspace
	}

	return pgUUID, nil
}

// GetWorkspaceUUIDOrNil returns workspace UUID if available, nil-valid pgtype.UUID otherwise.
// Use this for optional workspace filtering (e.g., super_admin access to all workspaces).
func GetWorkspaceUUIDOrNil(ctx context.Context) pgtype.UUID {
	workspaceID := GetWorkspaceUUID(ctx)
	if workspaceID == "" {
		return pgtype.UUID{Valid: false}
	}

	pgUUID, err := converter.ConvertStringToPgUUID(workspaceID)
	if err != nil {
		return pgtype.UUID{Valid: false}
	}

	return pgUUID
}
