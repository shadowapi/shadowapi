package storages

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/samber/do/v2"

	"github.com/go-faster/errors"
	"github.com/gotd/td/bin"
	"github.com/gotd/td/telegram/peers"
	"github.com/gotd/td/tg"
	"github.com/shadowapi/shadowapi/backend/internal/config"
	"github.com/shadowapi/shadowapi/backend/pkg/query"
	"log/slog"
)

var _ peers.Cache = &CacheStorage{}

type CacheStorage struct {
	ctx       context.Context
	cfg       *config.Config
	log       *slog.Logger
	dbp       *pgxpool.Pool
	sessionID int64
}

func NewCacheStorage(sessionID int64, pg *pgxpool.Pool) *CacheStorage {
	return &CacheStorage{
		sessionID: sessionID,
		dbp:       pg,
	}
}

func ProvideCacheStorage(i do.Injector) (*CacheStorage, error) {
	ctx := do.MustInvoke[context.Context](i)
	cfg := do.MustInvoke[*config.Config](i)
	dbp := do.MustInvoke[*pgxpool.Pool](i)
	log := do.MustInvoke[*slog.Logger](i).With("service", "cachestorage")

	return &CacheStorage{
		ctx:       ctx,
		cfg:       cfg,
		log:       log,
		dbp:       dbp,
		sessionID: 0,
	}, nil
}

func (c *CacheStorage) SaveUsers(ctx context.Context, users ...*tg.User) error {
	tx := query.New(c.dbp)

	for _, user := range users {
		b := bin.Buffer{}
		if err := user.Encode(&b); err != nil {
			return fmt.Errorf("failed to encode user: %w", err)
		}

		if err := tx.TgCreateCachedUser(ctx, query.TgCreateCachedUserParams{
			ID:        user.ID,
			FirstName: ToText(user.FirstName),
			LastName:  ToText(user.LastName),
			Username:  ToText(user.Username),
			Phone:     ToText(user.Phone),
			Raw:       b.Raw(),
			SessionID: c.sessionID,
		}); err != nil {
			return fmt.Errorf("failed to create cached user: %w", err)
		}
	}

	return nil
}

func (c *CacheStorage) SaveUserFulls(ctx context.Context, users ...*tg.UserFull) error {
	tx := query.New(c.dbp)

	for _, user := range users {
		b := bin.Buffer{}
		if err := user.Encode(&b); err != nil {
			return fmt.Errorf("failed to encode user: %w", err)
		}

		if err := tx.TgCreateCachedUser(ctx, query.TgCreateCachedUserParams{
			ID:        user.ID,
			SessionID: c.sessionID,
			RawFull:   b.Raw(),
		}); err != nil {
			return fmt.Errorf("failed to create cached full user: %w", err)
		}
	}

	return nil
}

func (c *CacheStorage) FindUser(ctx context.Context, id int64) (*tg.User, bool, error) {
	tx := query.New(c.dbp)

	result, err := tx.TgGetCachedUser(ctx, query.TgGetCachedUserParams{
		SessionID: c.sessionID,
		ID:        id,
	})

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, false, nil
	}

	if err != nil {
		return nil, false, err
	}

	if result.Raw == nil {
		return nil, false, nil
	}

	var user tg.User
	b := bin.Buffer{Buf: result.Raw}
	if err := user.Decode(&b); err != nil {
		return nil, false, fmt.Errorf("failed to decode user: %w", err)
	}

	return &user, true, nil
}

func (c *CacheStorage) FindUserFull(ctx context.Context, id int64) (*tg.UserFull, bool, error) {
	tx := query.New(c.dbp)

	result, err := tx.TgGetCachedUser(ctx, query.TgGetCachedUserParams{
		SessionID: c.sessionID,
		ID:        id,
	})

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, false, nil
	}

	if err != nil {
		return nil, false, err
	}

	if result.RawFull == nil {
		return nil, false, nil
	}

	var user tg.UserFull
	b := bin.Buffer{Buf: result.RawFull}
	if err := user.Decode(&b); err != nil {
		return nil, false, fmt.Errorf("failed to decode full user: %w", err)
	}

	return &user, true, nil
}

func (c *CacheStorage) SaveChats(ctx context.Context, chats ...*tg.Chat) error {
	tx := query.New(c.dbp)

	for _, chat := range chats {
		var b bin.Buffer
		if err := chat.Encode(&b); err != nil {
			return fmt.Errorf("failed to encode chat: %w", err)
		}

		if err := tx.TgCreateCachedChat(ctx, query.TgCreateCachedChatParams{
			ID:        chat.ID,
			Title:     ToText(chat.Title),
			Raw:       b.Raw(),
			SessionID: c.sessionID,
		}); err != nil {
			return fmt.Errorf("failed to create cached chat: %w", err)
		}
	}

	return nil
}

func (c *CacheStorage) SaveChatFulls(ctx context.Context, chats ...*tg.ChatFull) error {
	tx := query.New(c.dbp)

	for _, chat := range chats {
		var b bin.Buffer
		if err := chat.Encode(&b); err != nil {
			return fmt.Errorf("failed to encode chat: %w", err)
		}

		if err := tx.TgCreateCachedChat(ctx, query.TgCreateCachedChatParams{
			ID:        chat.ID,
			SessionID: c.sessionID,
			RawFull:   b.Raw(),
		}); err != nil {
			return fmt.Errorf("failed to create cached full chat: %w", err)
		}
	}

	return nil
}

func (c *CacheStorage) FindChat(ctx context.Context, id int64) (*tg.Chat, bool, error) {
	tx := query.New(c.dbp)

	res, err := tx.TgGetCachedChat(ctx, query.TgGetCachedChatParams{
		ID:        id,
		SessionID: c.sessionID,
	})

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, false, nil
	}

	if err != nil {
		return nil, false, fmt.Errorf("failed to get cached chat: %w", err)
	}

	if res.Raw == nil {
		return nil, false, nil
	}

	var b = bin.Buffer{
		Buf: res.Raw,
	}

	var chat tg.Chat
	if err := chat.Decode(&b); err != nil {
		return nil, false, fmt.Errorf("failed to decode cached chat: %w", err)
	}

	return &chat, true, nil
}

func (c *CacheStorage) FindChatFull(ctx context.Context, id int64) (*tg.ChatFull, bool, error) {
	tx := query.New(c.dbp)

	res, err := tx.TgGetCachedChat(ctx, query.TgGetCachedChatParams{
		ID:        id,
		SessionID: c.sessionID,
	})

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, false, nil
	}

	if err != nil {
		return nil, false, fmt.Errorf("failed to get cached chat: %w", err)
	}

	if res.RawFull == nil {
		return nil, false, nil
	}

	var b = bin.Buffer{
		Buf: res.RawFull,
	}

	var chat tg.ChatFull
	if err := chat.Decode(&b); err != nil {
		return nil, false, fmt.Errorf("failed to decode cached full chat: %w", err)
	}

	return &chat, true, nil
}

func (c *CacheStorage) SaveChannels(ctx context.Context, channels ...*tg.Channel) error {
	tx := query.New(c.dbp)

	for _, channel := range channels {
		var b bin.Buffer
		if err := channel.Encode(&b); err != nil {
			return fmt.Errorf("failed to encode channel: %w", err)
		}

		if err := tx.TgCreateCachedChannel(ctx, query.TgCreateCachedChannelParams{
			ID:        channel.ID,
			Title:     ToText(channel.Title),
			Username:  ToText(channel.Username),
			Broadcast: ToBool(channel.Broadcast),
			Forum:     ToBool(channel.Forum),
			Megagroup: ToBool(channel.Megagroup),
			Raw:       b.Raw(),
			SessionID: c.sessionID,
		}); err != nil {
			return fmt.Errorf("failed to create cached channel: %w", err)
		}
	}

	return nil
}

func (c *CacheStorage) SaveChannelFulls(ctx context.Context, channels ...*tg.ChannelFull) error {
	tx := query.New(c.dbp)

	for _, channel := range channels {
		var b bin.Buffer
		if err := channel.Encode(&b); err != nil {
			return fmt.Errorf("failed to encode full channel: %w", err)
		}

		if err := tx.TgCreateCachedChannel(ctx, query.TgCreateCachedChannelParams{
			ID:        channel.ID,
			RawFull:   b.Raw(),
			SessionID: c.sessionID,
		}); err != nil {
			return fmt.Errorf("failed to create cached full channel: %w", err)
		}
	}

	return nil
}

func (c *CacheStorage) FindChannel(ctx context.Context, id int64) (*tg.Channel, bool, error) {
	tx := query.New(c.dbp)

	res, err := tx.TgGetCachedChannel(ctx, query.TgGetCachedChannelParams{
		ID:        id,
		SessionID: c.sessionID,
	})

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, false, nil
	}

	if err != nil {
		return nil, false, fmt.Errorf("failed to get cached channel: %w", err)
	}

	if res.Raw == nil {
		return nil, false, nil
	}

	var b = bin.Buffer{
		Buf: res.Raw,
	}

	var channel tg.Channel
	if err := channel.Decode(&b); err != nil {
		return nil, false, fmt.Errorf("failed to decode cached channel: %w", err)
	}

	return &channel, true, nil
}

func (c *CacheStorage) FindChannelFull(ctx context.Context, id int64) (*tg.ChannelFull, bool, error) {
	tx := query.New(c.dbp)

	res, err := tx.TgGetCachedChannel(ctx, query.TgGetCachedChannelParams{
		ID:        id,
		SessionID: c.sessionID,
	})

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, false, nil
	}

	if err != nil {
		return nil, false, fmt.Errorf("failed to get cached channel: %w", err)
	}

	if res.RawFull == nil {
		return nil, false, nil
	}

	var b = bin.Buffer{
		Buf: res.RawFull,
	}

	var channel tg.ChannelFull
	if err := channel.Decode(&b); err != nil {
		return nil, false, fmt.Errorf("failed to decode cached full channel: %w", err)
	}

	return &channel, true, nil
}
