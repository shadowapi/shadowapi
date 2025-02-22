package telegram

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"github.com/shadowapi/shadowapi/backend/internal/tg/storages"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/updates"
)

type Option func(t *telegram.Options)

func WithSessionStorage(session int64, p *pgxpool.Pool) Option {
	return func(t *telegram.Options) {
		t.SessionStorage = storages.NewSessionStorage(p, session)
	}
}

func WithLogger(l *zap.Logger) Option {
	return func(t *telegram.Options) {
		t.Logger = l
	}
}

func WithUpdateManager(u *updates.Manager) Option {
	return func(t *telegram.Options) {
		t.UpdateHandler = u
	}
}
