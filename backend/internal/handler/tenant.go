package handler

import (
	"context"
	"net/http"

	gofrsUUID "github.com/gofrs/uuid"
	googleUUID "github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/shadowapi/shadowapi/backend/internal/auth/oauth2"
	"github.com/shadowapi/shadowapi/backend/pkg/api"
	"github.com/shadowapi/shadowapi/backend/pkg/query"
)

const SharedSessionCookie = "shadowapi_shared_session"

// ListTenants implements api.Handler.
func (h *Handler) ListTenants(ctx context.Context, params api.ListTenantsParams) ([]api.Tenant, error) {
	log := h.log.With("handler", "ListTenants")

	limit := int32(100)
	offset := int32(0)
	if params.Limit.IsSet() {
		limit = int32(params.Limit.Value)
	}
	if params.Offset.IsSet() {
		offset = int32(params.Offset.Value)
	}

	tenants, err := query.New(h.dbp).ListTenants(ctx, query.ListTenantsParams{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil && err != pgx.ErrNoRows {
		log.Error("failed to list tenants", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to list tenants"))
	}

	result := make([]api.Tenant, len(tenants))
	for i, t := range tenants {
		result[i] = tenantToAPI(t)
	}
	return result, nil
}

// CreateTenant implements api.Handler.
func (h *Handler) CreateTenant(ctx context.Context, req *api.Tenant) (*api.Tenant, error) {
	log := h.log.With("handler", "CreateTenant")

	tenantUUID := gofrsUUID.Must(gofrsUUID.NewV7())

	created, err := query.New(h.dbp).CreateTenant(ctx, query.CreateTenantParams{
		UUID:        pgtype.UUID{Bytes: tenantUUID, Valid: true},
		Name:        req.Name,
		DisplayName: req.DisplayName,
		IsEnabled:   req.IsEnabled.Or(true),
		Settings:    []byte(`{}`),
	})
	if err != nil {
		log.Error("failed to create tenant", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to create tenant"))
	}

	result := tenantToAPI(created)
	return &result, nil
}

// GetTenant implements api.Handler.
func (h *Handler) GetTenant(ctx context.Context, params api.GetTenantParams) (*api.Tenant, error) {
	log := h.log.With("handler", "GetTenant")

	tenantUUID, err := gofrsUUID.FromString(params.UUID.String())
	if err != nil {
		return nil, ErrWithCode(http.StatusBadRequest, E("invalid UUID"))
	}

	tenant, err := query.New(h.dbp).GetTenant(ctx, pgtype.UUID{Bytes: tenantUUID, Valid: true})
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrWithCode(http.StatusNotFound, E("tenant not found"))
		}
		log.Error("failed to get tenant", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to get tenant"))
	}

	result := tenantToAPI(tenant)
	return &result, nil
}

// UpdateTenant implements api.Handler.
func (h *Handler) UpdateTenant(ctx context.Context, req *api.Tenant, params api.UpdateTenantParams) (*api.Tenant, error) {
	log := h.log.With("handler", "UpdateTenant")

	tenantUUID, err := gofrsUUID.FromString(params.UUID.String())
	if err != nil {
		return nil, ErrWithCode(http.StatusBadRequest, E("invalid UUID"))
	}

	err = query.New(h.dbp).UpdateTenant(ctx, query.UpdateTenantParams{
		UUID:        pgtype.UUID{Bytes: tenantUUID, Valid: true},
		DisplayName: req.DisplayName,
		IsEnabled:   req.IsEnabled.Or(true),
		Settings:    []byte(`{}`),
	})
	if err != nil {
		log.Error("failed to update tenant", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to update tenant"))
	}

	// Fetch updated tenant
	updated, err := query.New(h.dbp).GetTenant(ctx, pgtype.UUID{Bytes: tenantUUID, Valid: true})
	if err != nil {
		log.Error("failed to get updated tenant", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to get updated tenant"))
	}

	result := tenantToAPI(updated)
	return &result, nil
}

// DeleteTenant implements api.Handler.
func (h *Handler) DeleteTenant(ctx context.Context, params api.DeleteTenantParams) error {
	log := h.log.With("handler", "DeleteTenant")

	tenantUUID, err := gofrsUUID.FromString(params.UUID.String())
	if err != nil {
		return ErrWithCode(http.StatusBadRequest, E("invalid UUID"))
	}

	err = query.New(h.dbp).DeleteTenant(ctx, pgtype.UUID{Bytes: tenantUUID, Valid: true})
	if err != nil {
		log.Error("failed to delete tenant", "error", err)
		return ErrWithCode(http.StatusInternalServerError, E("failed to delete tenant"))
	}

	return nil
}

// CheckTenantExists implements api.Handler.
func (h *Handler) CheckTenantExists(ctx context.Context, params api.CheckTenantExistsParams) (*api.TenantCheck, error) {
	log := h.log.With("handler", "CheckTenantExists")

	tenant, err := query.New(h.dbp).GetTenantByName(ctx, params.Name)
	if err != nil {
		if err == pgx.ErrNoRows {
			return &api.TenantCheck{Exists: false}, nil
		}
		log.Error("failed to check tenant", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to check tenant"))
	}

	return &api.TenantCheck{
		Exists:      true,
		DisplayName: api.NewOptString(tenant.DisplayName),
	}, nil
}

// ListAuthenticatedTenants implements api.Handler.
func (h *Handler) ListAuthenticatedTenants(ctx context.Context) ([]api.AuthenticatedTenant, error) {
	log := h.log.With("handler", "ListAuthenticatedTenants")

	// Get session ID from shared cookie (set in context by middleware)
	sessionID, ok := ctx.Value(oauth2.SharedSessionContextKey).(string)
	if !ok || sessionID == "" {
		// No shared session cookie, return empty list
		return []api.AuthenticatedTenant{}, nil
	}

	sessions, err := query.New(h.dbp).GetTenantSessionsBySessionID(ctx, sessionID)
	if err != nil && err != pgx.ErrNoRows {
		log.Error("failed to get tenant sessions", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to get sessions"))
	}

	result := make([]api.AuthenticatedTenant, len(sessions))
	for i, s := range sessions {
		var tenantUUID googleUUID.UUID
		if s.TenantUuid != nil {
			tenantUUID = googleUUID.UUID(s.TenantUuid.Bytes())
		}
		result[i] = api.AuthenticatedTenant{
			TenantUUID:        tenantUUID,
			TenantName:        s.TenantName,
			TenantDisplayName: s.TenantDisplayName,
			UserEmail:         s.UserEmail,
		}
		if s.LastAccessedAt.Valid {
			result[i].LastAccessedAt = api.NewOptDateTime(s.LastAccessedAt.Time)
		}
	}
	return result, nil
}

// gofrsToGoogleUUID converts gofrs/uuid.UUID to google/uuid.UUID
func gofrsToGoogleUUID(u gofrsUUID.UUID) googleUUID.UUID {
	return googleUUID.UUID(u.Bytes())
}

func tenantToAPI(t query.Tenant) api.Tenant {
	result := api.Tenant{
		UUID:        api.NewOptUUID(gofrsToGoogleUUID(t.UUID)),
		Name:        t.Name,
		DisplayName: t.DisplayName,
		IsEnabled:   api.NewOptBool(t.IsEnabled),
	}
	if t.CreatedAt.Valid {
		result.CreatedAt = api.NewOptDateTime(t.CreatedAt.Time)
	}
	if t.UpdatedAt.Valid {
		result.UpdatedAt = api.NewOptDateTime(t.UpdatedAt.Time)
	}
	return result
}
