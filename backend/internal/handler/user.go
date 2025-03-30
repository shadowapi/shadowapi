package handler

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/shadowapi/shadowapi/backend/internal/db"
	"github.com/shadowapi/shadowapi/backend/pkg/api"
	"github.com/shadowapi/shadowapi/backend/pkg/query"
)

// CreateUser implements createUser operation.
//
// Create a new user.
//
// POST /user
func (h *Handler) CreateUser(ctx context.Context, req *api.User) (*api.User, error) {
	return db.InTx(ctx, h.dbp, func(tx pgx.Tx) (*api.User, error) {
		// Generate a new UUID for the user.
		userUUID := uuid.Must(uuid.NewV7())

		// Marshal Meta if provided.
		var metaBytes []byte
		if req.Meta.IsSet() && req.Meta.Value != nil {
			var err error
			metaBytes, err = json.Marshal(req.Meta.Value)
			if err != nil {
				return nil, ErrWithCode(http.StatusInternalServerError, E("failed to marshal user meta", err.Error()))
			}
		}

		created, err := query.New(tx).CreateUser(ctx, query.CreateUserParams{
			UUID:      userUUID,
			Email:     req.Email,
			Password:  req.Password,
			FirstName: req.FirstName,
			LastName:  req.LastName,
			IsEnabled: req.IsEnabled.Or(false),
			IsAdmin:   req.IsAdmin.Or(false),
			Meta:      metaBytes,
		})
		if err != nil {
			return nil, ErrWithCode(http.StatusInternalServerError, E("failed to create user", err.Error()))
		}

		// Convert the stored meta bytes into an api.UserMeta.
		var meta api.UserMeta
		if len(created.Meta) > 0 {
			if err := json.Unmarshal(created.Meta, &meta); err != nil {
				// fallback to an empty meta if unmarshalling fails
				meta = make(api.UserMeta)
			}
		} else {
			meta = make(api.UserMeta)
		}

		out := api.User{
			UUID:      api.NewOptString(created.UUID.String()),
			Email:     created.Email,
			Password:  created.Password,
			FirstName: created.FirstName,
			LastName:  created.LastName,
			IsEnabled: api.NewOptBool(created.IsEnabled),
			IsAdmin:   api.NewOptBool(created.IsAdmin),
			Meta:      api.NewOptUserMeta(meta),
			CreatedAt: api.NewOptDateTime(created.CreatedAt.Time),
			UpdatedAt: api.NewOptDateTime(created.UpdatedAt.Time),
		}
		return &out, nil
	})
}

// DeleteUser implements deleteUser operation.
//
// Delete user.
//
// DELETE /user/{uuid}
func (h *Handler) DeleteUser(ctx context.Context, params api.DeleteUserParams) error {
	userUUID, err := uuid.FromString(params.UUID)
	if err != nil {
		h.log.Error("failed to parse user UUID", "error", err)
		return ErrWithCode(http.StatusBadRequest, E("invalid user UUID"))
	}
	if err := query.New(h.dbp).DeleteUser(ctx, userUUID); err != nil {
		h.log.Error("failed to delete user", "error", err)
		return ErrWithCode(http.StatusInternalServerError, E("failed to delete user"))
	}
	return nil
}

// GetUser implements getUser operation.
//
// Get user details.
//
// GET /user/{uuid}
func (h *Handler) GetUser(ctx context.Context, params api.GetUserParams) (*api.User, error) {
	userUUID, err := uuid.FromString(params.UUID)
	if err != nil {
		h.log.Error("failed to parse user UUID", "error", err)
		return nil, ErrWithCode(http.StatusBadRequest, E("invalid user UUID"))
	}
	user, err := query.New(h.dbp).GetUser(ctx, userUUID)
	if err == pgx.ErrNoRows {
		return nil, ErrWithCode(http.StatusNotFound, E("user not found"))
	} else if err != nil {
		h.log.Error("failed to get user", "error", err)
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

// ListUsers implements listUsers operation.
//
// List all users.
//
// GET /user
func (h *Handler) ListUsers(ctx context.Context) ([]api.User, error) {
	users, err := query.New(h.dbp).ListUsers(ctx, query.ListUsersParams{
		OffsetRecords: 0,
		LimitRecords:  100,
	})
	if err != nil && err != pgx.ErrNoRows {
		h.log.Error("failed to list users", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to list users"))
	}
	var result []api.User
	for _, u := range users {
		var meta api.UserMeta
		if len(u.Meta) > 0 {
			if err := json.Unmarshal(u.Meta, &meta); err != nil {
				meta = make(api.UserMeta)
			}
		} else {
			meta = make(api.UserMeta)
		}
		result = append(result, api.User{
			UUID:      api.NewOptString(u.UUID.String()),
			Email:     u.Email,
			Password:  u.Password,
			FirstName: u.FirstName,
			LastName:  u.LastName,
			IsEnabled: api.NewOptBool(u.IsEnabled),
			IsAdmin:   api.NewOptBool(u.IsAdmin),
			Meta:      api.NewOptUserMeta(meta),
			CreatedAt: api.NewOptDateTime(u.CreatedAt.Time),
			UpdatedAt: api.NewOptDateTime(u.UpdatedAt.Time),
		})
	}
	return result, nil
}

// UpdateUser implements updateUser operation.
//
// Update user details.
//
// PUT /user/{uuid}
func (h *Handler) UpdateUser(ctx context.Context, req *api.User, params api.UpdateUserParams) (*api.User, error) {
	userUUID, err := uuid.FromString(params.UUID)
	if err != nil {
		h.log.Error("failed to parse user UUID", "error", err)
		return nil, ErrWithCode(http.StatusBadRequest, E("invalid user UUID"))
	}
	return db.InTx(ctx, h.dbp, func(tx pgx.Tx) (*api.User, error) {
		updateParams := query.UpdateUserParams{
			UUID:      userUUID,
			Email:     req.Email,
			Password:  req.Password,
			FirstName: req.FirstName,
			LastName:  req.LastName,
			IsEnabled: req.IsEnabled.Or(false),
			IsAdmin:   req.IsAdmin.Or(false),
		}
		// Handle Meta field (if provided)
		if req.Meta.IsSet() && req.Meta.Value != nil {
			b, err := json.Marshal(req.Meta.Value)
			if err != nil {
				return nil, ErrWithCode(http.StatusInternalServerError, E("failed to marshal user meta", err.Error()))
			}
			updateParams.Meta = b
		} else {
			updateParams.Meta = nil
		}
		if err := query.New(tx).UpdateUser(ctx, updateParams); err != nil {
			h.log.Error("failed to update user", "error", err)
			return nil, ErrWithCode(http.StatusInternalServerError, E("failed to update user"))
		}
		user, err := query.New(tx).GetUser(ctx, userUUID)
		if err != nil {
			h.log.Error("failed to get updated user", "error", err)
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
	})
}
