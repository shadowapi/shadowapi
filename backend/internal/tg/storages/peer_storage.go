package storages

import (
	"context"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shadowapi/shadowapi/backend/internal/config"
	"github.com/shadowapi/shadowapi/backend/pkg/query"

	"github.com/go-faster/errors"
	"github.com/gotd/td/telegram/peers"
	"github.com/jackc/pgx/v5"
)

var _ peers.Storage = &PeerStorage{}

type PeerStorage struct {
	ctx context.Context
	cfg *config.Config
	log *slog.Logger
	dbp *pgxpool.Pool

	sessionID int64
}

func NewPeerStorage(sessionID int64, db *pgxpool.Pool) *PeerStorage {
	return &PeerStorage{
		dbp:       db,
		sessionID: sessionID,
	}
}

// Provide session middleware instance for the dependency injector
//func ProvidePeerStorage(i do.Injector) (*PeerStorage, error) {
//	ctx := do.MustInvoke[context.Context](i)
//	cfg := do.MustInvoke[*config.Config](i)
//	dbp := do.MustInvoke[*pgxpool.Pool](i)
//	log := do.MustInvoke[*slog.Logger](i).With("service", "peerstorage")
//
//	return &PeerStorage{
//		ctx,
//		cfg,
//		log,
//		dbp,
//		0,
//	}, nil
//}

func (p *PeerStorage) Save(ctx context.Context, key peers.Key, value peers.Value) error {
	tx := query.New(p.dbp)

	return tx.TgSavePeer(ctx, query.TgSavePeerParams{
		SessionID:  p.sessionID,
		ID:         key.ID,
		PeerType:   key.Prefix,
		AccessHash: ToInt8(value.AccessHash),
	})
}

func (p *PeerStorage) Find(ctx context.Context, key peers.Key) (value peers.Value, found bool, err error) {
	tx := query.New(p.dbp)

	peer, err := tx.TgFindPeer(ctx, query.TgFindPeerParams{
		SessionID: p.sessionID,
		ID:        key.ID,
		PeerType:  key.Prefix,
	})

	if err != nil && errors.Is(err, pgx.ErrNoRows) {
		return peers.Value{}, false, nil
	}

	if err != nil {
		return peers.Value{}, false, err
	}

	return peers.Value{
		AccessHash: peer.AccessHash.Int64,
	}, true, nil
}

func (p *PeerStorage) SavePhone(ctx context.Context, phone string, key peers.Key) error {
	tx := query.New(p.dbp)

	return tx.TgSavePeerUserPhone(ctx, query.TgSavePeerUserPhoneParams{
		SessionID: p.sessionID,
		ID:        key.ID,
		Phone:     ToText(phone),
	})
}

func (p *PeerStorage) FindPhone(ctx context.Context, phone string) (key peers.Key, value peers.Value, found bool, err error) {
	tx := query.New(p.dbp)

	peer, err := tx.TgFindPeerUserByPhone(ctx, query.TgFindPeerUserByPhoneParams{
		SessionID: p.sessionID,
		Phone:     ToText(phone),
	})

	if err != nil && errors.Is(err, pgx.ErrNoRows) {
		return peers.Key{}, peers.Value{}, false, nil
	}

	if err != nil {
		return peers.Key{}, peers.Value{}, false, err
	}

	return peers.Key{
			ID: peer.ID,
		}, peers.Value{
			AccessHash: peer.AccessHash.Int64,
		}, true, nil
}

func (p *PeerStorage) GetContactsHash(ctx context.Context) (int64, error) {
	tx := query.New(p.dbp)

	session, err := tx.TgGetSession(ctx, p.sessionID)
	if err != nil {
		return 0, err
	}

	if session.ContactsHash.Valid {
		return session.ContactsHash.Int64, nil
	}

	return 0, nil
}

func (p *PeerStorage) SaveContactsHash(ctx context.Context, hash int64) error {
	tx := query.New(p.dbp)

	return tx.TgUpdateSession(ctx, query.TgUpdateSessionParams{
		ID:           p.sessionID,
		ContactsHash: ToInt8(hash),
	})
}
