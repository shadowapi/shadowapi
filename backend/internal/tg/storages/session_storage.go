package storages

import (
	"context"
	"github.com/shadowapi/shadowapi/backend/internal/config"
	"github.com/shadowapi/shadowapi/backend/pkg/query"

	"github.com/go-faster/errors"
	"github.com/gotd/td/session"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"log/slog"
)

var _ session.Storage = &SessionStorage{}

type SessionStorage struct {
	ctx       context.Context
	cfg       *config.Config
	log       *slog.Logger
	dbp       *pgxpool.Pool
	sessionID int64
}

func NewSessionStorage(db *pgxpool.Pool, sessionID int64) *SessionStorage {
	return &SessionStorage{
		dbp:       db,
		sessionID: sessionID,
	}
}

//func ProvideSessionStorage(i do.Injector) (*SessionStorage, error) {
//	ctx := do.MustInvoke[context.Context](i)
//	cfg := do.MustInvoke[*config.Config](i)
//	dbp := do.MustInvoke[*pgxpool.Pool](i)
//	log := do.MustInvoke[*slog.Logger](i).With("service", "sessionstorage")
//
//	return &SessionStorage{
//		ctx:       ctx,
//		cfg:       cfg,
//		log:       log,
//		dbp:       dbp,
//		sessionID: 0,
//	}, nil
//}

func (s *SessionStorage) LoadSession(ctx context.Context) ([]byte, error) {
	tx := query.New(s.dbp)

	result, err := tx.TgGetSession(ctx, s.sessionID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, session.ErrNotFound
	}

	if err != nil {
		s.log.Error("failed to get user session", "session_id", s.sessionID, "error", err)
		return nil, err
	}

	return result.Session, nil
}

func (s *SessionStorage) StoreSession(ctx context.Context, data []byte) error {
	s.log = s.log.With("session_id", s.sessionID)
	tx := query.New(s.dbp)

	if err := tx.TgUpdateSession(ctx, query.TgUpdateSessionParams{
		Session: data,
		ID:      s.sessionID,
	}); err != nil {
		s.log.Error("failed to save session", "error", err)
		return err
	}

	return nil
}
