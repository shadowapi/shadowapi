package handler

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/shadowapi/shadowapi/backend/internal/converter"
	"github.com/shadowapi/shadowapi/backend/internal/db"
	"github.com/shadowapi/shadowapi/backend/internal/session"
	"github.com/shadowapi/shadowapi/backend/pkg/api"
	"github.com/shadowapi/shadowapi/backend/pkg/query"
)

// GetProfile implements getProfile operation.
func (h *Handler) GetProfile(ctx context.Context) (*api.User, error) {
	ident, ok := session.GetIdentity(ctx)
	if !ok {
		return nil, ErrWithCode(http.StatusUnauthorized, E("unauthorized"))
	}
	uid, err := uuid.FromString(ident.ID)
	if err != nil {
		return nil, ErrWithCode(http.StatusBadRequest, E("invalid user id"))
	}
	user, err := query.New(h.dbp).GetUser(ctx, pgtype.UUID{Bytes: converter.UToBytes(uid), Valid: true})
	if err == pgx.ErrNoRows {
		return nil, ErrWithCode(http.StatusNotFound, E("user not found"))
	} else if err != nil {
		h.log.Error("get profile", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to fetch user"))
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

// UpdateProfile implements updateProfile operation.
func (h *Handler) UpdateProfile(ctx context.Context, req *api.UserProfile) (*api.User, error) {
	ident, ok := session.GetIdentity(ctx)
	if !ok {
		return nil, ErrWithCode(http.StatusUnauthorized, E("unauthorized"))
	}
	uid, err := uuid.FromString(ident.ID)
	if err != nil {
		return nil, ErrWithCode(http.StatusBadRequest, E("invalid user id"))
	}
	return db.InTx(ctx, h.dbp, func(tx pgx.Tx) (*api.User, error) {
		err := query.New(tx).UpdateUserNames(ctx, query.UpdateUserNamesParams{
			UUID:      pgtype.UUID{Bytes: converter.UToBytes(uid), Valid: true},
			FirstName: req.FirstName,
			LastName:  req.LastName,
		})
		if err != nil {
			h.log.Error("update profile", "error", err)
			return nil, ErrWithCode(http.StatusInternalServerError, E("failed to update profile"))
		}
		user, err := query.New(tx).GetUser(ctx, pgtype.UUID{Bytes: converter.UToBytes(uid), Valid: true})
		if err != nil {
			h.log.Error("fetch updated user", "error", err)
			return nil, ErrWithCode(http.StatusInternalServerError, E("failed to fetch user"))
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
			FirstName: user.FirstName,
			LastName:  user.LastName,
			IsEnabled: api.NewOptBool(user.IsEnabled),
			IsAdmin:   api.NewOptBool(user.IsAdmin),
			Meta:      api.NewOptUserMeta(meta),
			CreatedAt: api.NewOptDateTime(user.CreatedAt.Time),
			UpdatedAt: api.NewOptDateTime(user.UpdatedAt.Time),
		}
		return &out, nil
	})
}
