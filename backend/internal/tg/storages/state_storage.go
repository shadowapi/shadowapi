package storages

import (
	"context"
	"fmt"

	"github.com/go-faster/errors"
	"github.com/gotd/td/telegram/updates"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	//"github.com/samber/do/v2"
	"github.com/shadowapi/shadowapi/backend/internal/config"
	"github.com/shadowapi/shadowapi/backend/pkg/query"
	"log/slog"
)

// TODO !!! investigate int <> int64 conversion safety

var _ updates.StateStorage = &StateStorage{}

type StateStorage struct {
	ctx       context.Context
	cfg       *config.Config
	log       *slog.Logger
	dbp       *pgxpool.Pool
	sessionID int64
}

func NewStateStorage(db *pgxpool.Pool, sessionID int64) *StateStorage {
	return &StateStorage{
		dbp:       db,
		sessionID: sessionID,
	}
}

//func Provide(i do.Injector) (*StateStorage, error) {
//	ctx := do.MustInvoke[context.Context](i)
//	cfg := do.MustInvoke[*config.Config](i)
//	dbp := do.MustInvoke[*pgxpool.Pool](i)
//	log := do.MustInvoke[*slog.Logger](i).With("service", "statestorage")
//
//	return &StateStorage{
//		ctx:       ctx,
//		cfg:       cfg,
//		log:       log,
//		dbp:       dbp,
//		sessionID: 0,
//	}, nil
//}

func (s *StateStorage) GetState(ctx context.Context, _ int64) (updates.State, bool, error) {
	q := query.New(s.dbp)
	res, err := q.TgGetState(ctx, s.sessionID)
	if err != nil && errors.Is(err, pgx.ErrNoRows) {
		return updates.State{}, false, nil
	}
	if err != nil {
		return updates.State{}, false, err
	}
	return updates.State{
		Pts:  int(res.Pts),
		Qts:  int(res.Qts),
		Date: int(res.Date),
		Seq:  int(res.Seq),
	}, true, nil
}

func (s *StateStorage) SetState(ctx context.Context, _ int64, state updates.State) error {
	q := query.New(s.dbp)
	return q.TgUpsertState(ctx, query.TgUpsertStateParams{
		ID:   s.sessionID,
		Pts:  int64(state.Pts),
		Qts:  int64(state.Qts),
		Date: int64(state.Date),
		Seq:  int64(state.Seq),
	})
}

func (s *StateStorage) SetPts(ctx context.Context, _ int64, pts int) error {
	q := query.New(s.dbp)
	return q.TgUpdatePts(ctx, query.TgUpdatePtsParams{
		ID:  s.sessionID,
		Pts: int64(pts),
	})
}

func (s *StateStorage) SetQts(ctx context.Context, _ int64, qts int) error {
	q := query.New(s.dbp)
	return q.TgUpdateQts(ctx, query.TgUpdateQtsParams{
		ID:  s.sessionID,
		Qts: int64(qts),
	})
}

func (s *StateStorage) SetDate(ctx context.Context, _ int64, date int) error {
	q := query.New(s.dbp)
	return q.TgUpdateDate(ctx, query.TgUpdateDateParams{
		ID:   s.sessionID,
		Date: int64(date),
	})
}

func (s *StateStorage) SetSeq(ctx context.Context, _ int64, seq int) error {
	q := query.New(s.dbp)
	return q.TgUpdateSeq(ctx, query.TgUpdateSeqParams{
		ID:  s.sessionID,
		Seq: int64(seq),
	})
}

func (s *StateStorage) SetDateSeq(ctx context.Context, _ int64, date, seq int) error {
	tx, err := s.dbp.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	q := query.New(tx)
	if err = q.TgUpdateDate(ctx, query.TgUpdateDateParams{
		ID:   s.sessionID,
		Date: int64(date),
	}); err != nil {
		_ = tx.Rollback(ctx)
		return err
	}
	if err = q.TgUpdateSeq(ctx, query.TgUpdateSeqParams{
		ID:  s.sessionID,
		Seq: int64(seq),
	}); err != nil {
		_ = tx.Rollback(ctx)
		return err
	}
	return tx.Commit(ctx)
}

func (s *StateStorage) GetChannelPts(ctx context.Context, _, channelID int64) (int, bool, error) {
	q := query.New(s.dbp)
	res, err := q.TgGetPeerChannel(ctx, query.TgGetPeerChannelParams{
		SessionID: s.sessionID,
		ID:        channelID,
	})
	if err != nil && errors.Is(err, pgx.ErrNoRows) {
		return 0, false, nil
	}
	if err != nil {
		return 0, false, err
	}
	return int(res.Pts), true, nil
}

func (s *StateStorage) SetChannelPts(ctx context.Context, _, channelID int64, pts int) error {
	tx, err := s.dbp.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	q := query.New(tx)
	if err = q.TgSavePeer(ctx, query.TgSavePeerParams{
		ID:        channelID,
		SessionID: s.sessionID,
		PeerType:  "channel_",
	}); err != nil {
		_ = tx.Rollback(ctx)
		return fmt.Errorf("cannot save peer: %w", err)
	}
	if err = q.TgSetPeerChannelState(ctx, query.TgSetPeerChannelStateParams{
		ID:        channelID,
		SessionID: s.sessionID,
		Pts:       int64(pts),
	}); err != nil {
		_ = tx.Rollback(ctx)
		return fmt.Errorf("cannot set channel pts: %w", err)
	}
	return tx.Commit(ctx)
}

func (s *StateStorage) ForEachChannels(ctx context.Context, _ int64, f func(context.Context, int64, int) error) error {
	q := query.New(s.dbp)
	channels, err := q.TgGetPeersChannels(ctx, s.sessionID)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return err
	}
	for _, ch := range channels {
		if err := f(ctx, ch.ID, int(ch.Pts)); err != nil {
			return err
		}
	}
	return nil
}
