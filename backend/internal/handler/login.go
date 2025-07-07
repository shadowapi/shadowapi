package handler

import (
	"context"
	"errors"

	"github.com/gofrs/uuid"
	gouuid "github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"golang.org/x/crypto/bcrypt"

	"github.com/shadowapi/shadowapi/backend/internal/session"
	"github.com/shadowapi/shadowapi/backend/pkg/api"
)

// PlainLogin verifies email/password and returns user UUID on success.
func (h *Handler) PlainLogin(ctx context.Context, email, password string) (string, error) {
	var id pgtype.UUID
	var hash string
	err := h.dbp.QueryRow(ctx, `SELECT uuid, password FROM "user" WHERE email=$1 AND is_enabled=TRUE`, email).Scan(&id, &hash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", errors.New("invalid credentials")
		}
		return "", err
	}
	if bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) != nil {
		return "", errors.New("invalid credentials")
	}
	return uuid.UUID(id.Bytes).String(), nil
}

// SessionStatus implements session-status operation.
func (h *Handler) SessionStatus(ctx context.Context) (*api.SessionStatus, error) {
	id, ok := session.GetIdentity(ctx)
	if !ok {
		return &api.SessionStatus{Active: false}, nil
	}
	out := api.SessionStatus{Active: true}
	uid, err := gouuid.Parse(id.ID)
	if err == nil {
		out.SetUUID(api.NewOptUUID(uid))
	}
	return &out, nil
}
