package handler

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/shadowapi/shadowapi/backend/internal/converter"
	"github.com/shadowapi/shadowapi/backend/internal/session"
	"github.com/shadowapi/shadowapi/backend/pkg/api"
	"github.com/shadowapi/shadowapi/backend/pkg/query"
)

// GetProfile returns current user profile.
func (h *Handler) GetProfile(ctx context.Context) (*api.User, error) {
	ident, ok := session.GetIdentity(ctx)
	if !ok {
		return nil, ErrWithCode(http.StatusUnauthorized, E("unauthorized"))
	}
	uid, err := uuid.FromString(ident.ID)
	if err != nil {
		h.log.Error("invalid identity uuid", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("invalid identity"))
	}
	user, err := query.New(h.dbp).GetUser(ctx, pgtype.UUID{Bytes: converter.UToBytes(uid), Valid: true})
	if err == pgx.ErrNoRows {
		return nil, ErrWithCode(http.StatusNotFound, E("user not found"))
	} else if err != nil {
		h.log.Error("get user", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to get user"))
	}
	var meta api.UserMeta
	if len(user.Meta) > 0 {
		if err := json.Unmarshal(user.Meta, &meta); err != nil {
			meta = make(api.UserMeta)
		}
	} else {
		meta = make(api.UserMeta)
	}
	out := api.User{
		UUID:      api.NewOptString(user.UUID.String()),
		Email:     user.Email,
		Password:  user.Password,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		IsEnabled: api.NewOptBool(user.IsEnabled),
		IsAdmin:   api.NewOptBool(user.IsAdmin),
		Meta:      api.NewOptUserMeta(meta),
		CreatedAt: api.NewOptDateTime(user.CreatedAt.Time),
		UpdatedAt: api.NewOptDateTime(user.UpdatedAt.Time),
	}
	return &out, nil
}

// UpdateProfile updates current user's first and last name.
func (h *Handler) UpdateProfile(ctx context.Context, req *api.ProfileUpdate) (*api.User, error) {
	ident, ok := session.GetIdentity(ctx)
	if !ok {
		return nil, ErrWithCode(http.StatusUnauthorized, E("unauthorized"))
	}
	uid, err := uuid.FromString(ident.ID)
	if err != nil {
		h.log.Error("invalid identity uuid", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("invalid identity"))
	}
	if err := query.New(h.dbp).UpdateUserName(ctx, query.UpdateUserNameParams{
		FirstName: req.FirstName,
		LastName:  req.LastName,
		UUID:      pgtype.UUID{Bytes: converter.UToBytes(uid), Valid: true},
	}); err != nil {
		h.log.Error("update profile", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to update profile"))
	}
	user, err := query.New(h.dbp).GetUser(ctx, pgtype.UUID{Bytes: converter.UToBytes(uid), Valid: true})
	if err != nil {
		h.log.Error("get updated user", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to get updated user"))
	}
	var meta api.UserMeta
	if len(user.Meta) > 0 {
		if err := json.Unmarshal(user.Meta, &meta); err != nil {
			meta = make(api.UserMeta)
		}
	} else {
		meta = make(api.UserMeta)
	}
	out := api.User{
		UUID:      api.NewOptString(user.UUID.String()),
		Email:     user.Email,
		Password:  user.Password,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		IsEnabled: api.NewOptBool(user.IsEnabled),
		IsAdmin:   api.NewOptBool(user.IsAdmin),
		Meta:      api.NewOptUserMeta(meta),
		CreatedAt: api.NewOptDateTime(user.CreatedAt.Time),
		UpdatedAt: api.NewOptDateTime(user.UpdatedAt.Time),
	}
	return &out, nil
}
