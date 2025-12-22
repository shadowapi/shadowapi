package dbauth

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/samber/do/v2"
	"golang.org/x/crypto/bcrypt"

	"github.com/shadowapi/shadowapi/backend/internal/auth"
	"github.com/shadowapi/shadowapi/backend/internal/converter"
	"github.com/shadowapi/shadowapi/backend/internal/db"
	"github.com/shadowapi/shadowapi/backend/internal/handler"
	"github.com/shadowapi/shadowapi/backend/pkg/api"
	"github.com/shadowapi/shadowapi/backend/pkg/query"
)

// DBUserManager implements UserManager interface using database
type DBUserManager struct {
	dbp *pgxpool.Pool
	log *slog.Logger
}

// Provide creates a new DBUserManager instance
func Provide(i do.Injector) (auth.UserManager, error) {
	return &DBUserManager{
		dbp: do.MustInvoke[*pgxpool.Pool](i),
		log: do.MustInvoke[*slog.Logger](i),
	}, nil
}

// CreateUser creates a new user in the database
func (m *DBUserManager) CreateUser(ctx context.Context, req *api.User) (*api.User, error) {
	return db.InTx(ctx, m.dbp, func(tx pgx.Tx) (*api.User, error) {
		// Generate a new UUID for the user.
		userUUID := uuid.Must(uuid.NewV7())

		// Hash the password with bcrypt
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			m.log.Error("failed to hash password", "error", err)
			return nil, handler.ErrWithCode(http.StatusInternalServerError, handler.E("failed to hash password"))
		}

		// Marshal Meta if provided.
		var metaBytes []byte
		if req.Meta.IsSet() && req.Meta.Value != nil {
			var err error
			metaBytes, err = json.Marshal(req.Meta.Value)
			if err != nil {
				return nil, handler.ErrWithCode(http.StatusInternalServerError, handler.E("failed to marshal user meta: %w", err))
			}
		}

		created, err := query.New(tx).CreateUser(ctx, query.CreateUserParams{
			UUID:      pgtype.UUID{Bytes: converter.UToBytes(userUUID), Valid: true},
			Email:     req.Email,
			Password:  string(hashedPassword),
			FirstName: req.FirstName,
			LastName:  req.LastName,
			IsEnabled: req.IsEnabled.Or(false),
			IsAdmin:   req.IsAdmin.Or(false),
			Meta:      metaBytes,
		})
		if err != nil {
			return nil, handler.ErrWithCode(http.StatusInternalServerError, handler.E("failed to create user: %w", err))
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

// GetUser retrieves a user by UUID from the database
func (m *DBUserManager) GetUser(ctx context.Context, uuidStr string) (*api.User, error) {
	userUUID, err := uuid.FromString(uuidStr)
	if err != nil {
		m.log.Error("failed to parse user UUID", "error", err)
		return nil, handler.ErrWithCode(http.StatusBadRequest, handler.E("invalid user UUID"))
	}

	user, err := query.New(m.dbp).GetUser(ctx, pgtype.UUID{Bytes: converter.UToBytes(userUUID), Valid: true})
	if err == pgx.ErrNoRows {
		return nil, handler.ErrWithCode(http.StatusNotFound, handler.E("user not found"))
	} else if err != nil {
		m.log.Error("failed to get user", "error", err)
		return nil, handler.ErrWithCode(http.StatusInternalServerError, handler.E("failed to get user"))
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

// UpdateUser updates an existing user in the database
func (m *DBUserManager) UpdateUser(ctx context.Context, req *api.User, uuidStr string) (*api.User, error) {
	userUUID, err := uuid.FromString(uuidStr)
	if err != nil {
		m.log.Error("failed to parse user UUID", "error", err)
		return nil, handler.ErrWithCode(http.StatusBadRequest, handler.E("invalid user UUID"))
	}

	return db.InTx(ctx, m.dbp, func(tx pgx.Tx) (*api.User, error) {
		pgUUID := pgtype.UUID{Bytes: converter.UToBytes(userUUID), Valid: true}

		// Fetch current user to preserve password if not provided
		currentUser, err := query.New(tx).GetUser(ctx, pgUUID)
		if err == pgx.ErrNoRows {
			return nil, handler.ErrWithCode(http.StatusNotFound, handler.E("user not found"))
		} else if err != nil {
			m.log.Error("failed to get current user", "error", err)
			return nil, handler.ErrWithCode(http.StatusInternalServerError, handler.E("failed to get current user"))
		}

		// Determine password: if new password provided, hash it; otherwise keep existing
		passwordToStore := currentUser.Password
		if req.Password != "" {
			hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
			if err != nil {
				m.log.Error("failed to hash password", "error", err)
				return nil, handler.ErrWithCode(http.StatusInternalServerError, handler.E("failed to hash password"))
			}
			passwordToStore = string(hashedPassword)
		}

		updateParams := query.UpdateUserParams{
			UUID:      pgUUID,
			Email:     req.Email,
			Password:  passwordToStore,
			FirstName: req.FirstName,
			LastName:  req.LastName,
			IsEnabled: req.IsEnabled.Or(false),
			IsAdmin:   req.IsAdmin.Or(false),
		}
		// Handle Meta field (if provided)
		if req.Meta.IsSet() && req.Meta.Value != nil {
			b, err := json.Marshal(req.Meta.Value)
			if err != nil {
				return nil, handler.ErrWithCode(http.StatusInternalServerError, handler.E("failed to marshal user meta: %w", err))
			}
			updateParams.Meta = b
		} else {
			updateParams.Meta = nil
		}

		if err := query.New(tx).UpdateUser(ctx, updateParams); err != nil {
			m.log.Error("failed to update user", "error", err)
			return nil, handler.ErrWithCode(http.StatusInternalServerError, handler.E("failed to update user"))
		}

		user, err := query.New(tx).GetUser(ctx, pgUUID)
		if err != nil {
			m.log.Error("failed to get updated user", "error", err)
			return nil, handler.ErrWithCode(http.StatusInternalServerError, handler.E("failed to get updated user"))
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

// DeleteUser deletes a user by UUID from the database
func (m *DBUserManager) DeleteUser(ctx context.Context, uuidStr string) error {
	userUUID, err := uuid.FromString(uuidStr)
	if err != nil {
		m.log.Error("failed to parse user UUID", "error", err)
		return handler.ErrWithCode(http.StatusBadRequest, handler.E("invalid user UUID"))
	}

	if err := query.New(m.dbp).DeleteUser(ctx, pgtype.UUID{Bytes: converter.UToBytes(userUUID), Valid: true}); err != nil {
		m.log.Error("failed to delete user", "error", err)
		return handler.ErrWithCode(http.StatusInternalServerError, handler.E("failed to delete user"))
	}
	return nil
}

// ListUsers returns a list of all users from the database
func (m *DBUserManager) ListUsers(ctx context.Context) ([]api.User, error) {
	users, err := query.New(m.dbp).ListUsers(ctx, query.ListUsersParams{
		Offset: 0,
		Limit:  10000, // TODO @reactima worry about paging later
	})
	if err != nil && err != pgx.ErrNoRows {
		m.log.Error("failed to list users", "error", err)
		return nil, handler.ErrWithCode(http.StatusInternalServerError, handler.E("failed to list users"))
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

// AuthenticateUser verifies email/password and returns the user if valid.
// With workspace-based architecture, users are global (unique email across system).
func (m *DBUserManager) AuthenticateUser(ctx context.Context, email, password string) (*api.User, error) {
	// Global email lookup - users are not scoped to workspaces
	user, err := query.New(m.dbp).GetUserByEmail(ctx, email)
	if err == pgx.ErrNoRows {
		m.log.Debug("authentication failed: user not found", "email", email)
		return nil, handler.ErrWithCode(http.StatusUnauthorized, handler.E("invalid credentials"))
	}
	if err != nil {
		m.log.Error("failed to get user by email", "error", err)
		return nil, handler.ErrWithCode(http.StatusInternalServerError, handler.E("authentication failed"))
	}

	if !user.IsEnabled {
		m.log.Debug("authentication failed: account disabled", "email", email)
		return nil, handler.ErrWithCode(http.StatusUnauthorized, handler.E("account disabled"))
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		m.log.Debug("authentication failed: invalid password", "email", email)
		return nil, handler.ErrWithCode(http.StatusUnauthorized, handler.E("invalid credentials"))
	}

	var meta api.UserMeta
	if len(user.Meta) > 0 {
		if err := json.Unmarshal(user.Meta, &meta); err != nil {
			meta = make(api.UserMeta)
		}
	} else {
		meta = make(api.UserMeta)
	}

	m.log.Debug("user authenticated successfully", "email", email, "uuid", user.UUID.String())

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
