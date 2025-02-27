package workers

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/shadowapi/shadowapi/backend/internal/config"
	"github.com/shadowapi/shadowapi/backend/internal/tg/storages"

	"github.com/go-faster/errors"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/peers"
	"github.com/gotd/td/telegram/updates"
	"github.com/gotd/td/tg"
	"go.uber.org/zap"
)

type Worker struct {
	sessionID    int32
	cfg          *config.Config
	db           *pgxpool.Pool
	tc           *telegram.Client
	dispatcher   *tg.UpdateDispatcher
	peersManager *peers.Manager
	gaps         *updates.Manager

	stop context.CancelFunc
}

func NewWorker(sessionID int32, pg *pgxpool.Pool, cfg config.Config) *Worker {
	logger := zap.L().Named("worker").With(zap.Int32("session_id", sessionID))
	dispatcher := tg.NewUpdateDispatcher()

	var h telegram.UpdateHandler
	w := &Worker{
		sessionID:  sessionID,
		cfg:        &cfg,
		db:         pg,
		dispatcher: &dispatcher,
	}

	w.tc = telegram.NewClient(cfg.Telegram.AppID, cfg.Telegram.AppHash, telegram.Options{
		Logger:         logger,
		SessionStorage: storages.NewSessionStorage(pg, int64(sessionID)),
		UpdateHandler: telegram.UpdateHandlerFunc(func(ctx context.Context, u tg.UpdatesClass) error {
			return h.Handle(ctx, u)
		}),
	})

	w.peersManager = peers.Options{
		Logger:  logger,
		Storage: storages.NewPeerStorage(int64(sessionID), pg),
		Cache:   storages.NewCacheStorage(int64(sessionID), pg),
	}.Build(w.tc.API())

	w.gaps = updates.New(updates.Config{
		Handler: dispatcher,
		Logger:  logger.Named("gaps"),
		Storage: storages.NewStateStorage(w.db, int64(sessionID)),
	})
	h = w.peersManager.UpdateHook(w.gaps)

	return w
}

func (w *Worker) Run(ctx context.Context) error {
	ctx, w.stop = context.WithCancel(ctx)
	return w.tc.Run(ctx, func(ctx context.Context) error {
		if err := w.peersManager.Init(ctx); err != nil {
			return err
		}

		u, err := w.peersManager.Self(ctx)
		if err != nil {
			return err
		}

		_, isBot := u.ToBot()
		if err := w.gaps.Run(ctx, w.tc.API(), u.ID(), updates.AuthOptions{IsBot: isBot}); err != nil {
			return errors.Wrap(err, "gaps")
		}

		return nil
	})
}

func (w *Worker) Stop() {
	w.stop()
}
