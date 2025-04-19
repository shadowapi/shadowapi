package oauth2

import (
	"context"
	"encoding/json"
	"github.com/shadowapi/shadowapi/backend/internal/converter"
	"log/slog"
	"time"

	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/oauth2"

	"github.com/shadowapi/shadowapi/backend/pkg/query"
)

type TokenStore struct {
	ctx context.Context
	dbp *pgxpool.Pool
	cfg *oauth2.Config

	tokenUUID        uuid.UUID
	refreshThreshold time.Duration
}

func NewTokenStore(
	ctx context.Context,
	cfg *oauth2.Config,
	dbp *pgxpool.Pool,
	tokenUUID uuid.UUID,
	refreshThreshold time.Duration,
) (*TokenStore, error) {
	return &TokenStore{
		ctx: ctx,
		cfg: cfg,
		dbp: dbp,

		tokenUUID:        tokenUUID,
		refreshThreshold: refreshThreshold,
	}, nil
}

func (t *TokenStore) Token() (*oauth2.Token, error) {
	token, err := t.loadToken()
	if err != nil {
		return nil, err
	}

	duration := token.Expiry.Sub(time.Now())
	duration = time.Duration(float64(duration) * 0.3)

	if duration <= t.refreshThreshold {
		token, err := t.cfg.TokenSource(t.ctx, token).Token()
		if err != nil {
			return nil, err
		}

		if err := t.saveToken(token); err != nil {
			slog.Error("failed saving token", "error", err)
			return nil, err
		}
		return token, nil
	}
	return token, nil
}

func (t *TokenStore) saveToken(token *oauth2.Token) error {
	tx := query.New(t.dbp)
	tokenData, err := json.Marshal(token)
	if err != nil {
		slog.Error("failed to marshal token", "error", err)
		return err
	}

	update := query.UpdateOauth2TokenParams{
		UUID:  converter.UuidToPgUUID(t.tokenUUID),
		Token: tokenData,
	}
	if err := tx.UpdateOauth2Token(t.ctx, update); err != nil {
		slog.Error("failed to update token", "error", err)
		return err
	}
	return nil
}

func (t *TokenStore) loadToken() (*oauth2.Token, error) {
	tx := query.New(t.dbp)
	tokenData, err := tx.GetOauth2TokenByUUID(t.ctx, converter.UuidToPgUUID(t.tokenUUID))
	if err != nil {
		slog.Error("failed to get token by uuid", "uuid", t.tokenUUID)
		return nil, err
	}

	var token oauth2.Token
	if err := json.Unmarshal([]byte(tokenData.Oauth2Token.Token), &token); err != nil {
		slog.Error("failed to unmarshal token", "error", err)
		return nil, err
	}
	return &token, nil
}
