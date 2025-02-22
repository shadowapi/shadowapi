package telegram

import (
	"context"
	"fmt"

	"github.com/go-faster/errors"
	"github.com/gotd/contrib/bg"
	"github.com/gotd/td/session"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/auth"
	"github.com/gotd/td/tg"
)

// Telegram wraps gotd Telegram API implementation for convenient use in REST
type Telegram struct {
	client *telegram.Client
	stop   bg.StopFunc
}

func New(id int, hash string, opts ...Option) *Telegram {
	options := telegram.Options{}
	for _, apply := range opts {
		apply(&options)
	}

	if options.SessionStorage == nil {
		options.SessionStorage = &session.StorageMemory{}
	}

	t := &Telegram{
		client: telegram.NewClient(id, hash, options),
	}

	return t
}

func (t *Telegram) run(ctx context.Context, cb func(ctx context.Context) error) error {
	if t.stop != nil {
		return cb(ctx)
	}

	return t.client.Run(ctx, cb)
}

func (t *Telegram) Start() error {
	stop, err := bg.Connect(t.client)
	if err != nil {
		return err
	}
	t.stop = stop

	return nil
}

func (t *Telegram) Stop() error {
	return t.stop()
}

// SendCode starts process of authorization for phone number
// returns information needed to proceed with it
func (t *Telegram) SendCode(ctx context.Context, phone string) (*tg.AuthSentCode, error) {
	var result *tg.AuthSentCode
	if err := t.run(ctx, func(ctx context.Context) error {
		client := t.client.Auth()
		sentCode, err := client.SendCode(ctx, phone, auth.SendCodeOptions{})
		if err != nil {
			return err
		}

		switch s := sentCode.(type) {
		case *tg.AuthSentCode:
			result = s
		default:
			return errors.Errorf("unexpected sent code type: %T", sentCode)
		}

		return nil
	}); err != nil {
		return nil, errors.Wrap(err, "failed to send code")
	}

	return result, nil
}

// SignIn finishes authorization process. Requires phone, code from SMS and hash from SendCode, password is required
// only when 2FA is on.
// TODO: Security issue. Need to accept password hash instead.
func (t *Telegram) SignIn(ctx context.Context, phone, code, password, hash string) (*tg.AuthAuthorization, error) {
	var result *tg.AuthAuthorization

	if err := t.run(ctx, func(ctx context.Context) error {
		var err error

		client := t.client.Auth()
		result, err = client.SignIn(ctx, phone, code, hash)
		if err == nil {
			return nil
		}

		if !errors.Is(err, auth.ErrPasswordAuthNeeded) {
			return err
		}

		result, err = client.Password(ctx, password)
		if err != nil {
			return errors.Wrap(err, "sign in with password")
		}

		return nil
	}); err != nil {
		return nil, errors.Wrap(err, "failed to sign in")
	}

	return result, nil
}

// Self returns user under which client is authorized
func (t *Telegram) Self(ctx context.Context) (*tg.User, error) {
	var user *tg.User

	if err := t.run(ctx, func(ctx context.Context) error {
		var err error

		user, err = t.client.Self(ctx)
		if err != nil {
			return err
		}

		return nil
	}); err != nil {
		return nil, errors.Wrap(err, "failed to get self")
	}

	return user, nil
}

func (t *Telegram) GetUser(ctx context.Context, id int64) (*tg.User, error) {
	var user *tg.User

	if err := t.run(ctx, func(ctx context.Context) error {
		var err error

		users, err := t.client.API().UsersGetUsers(ctx, []tg.InputUserClass{
			&tg.InputUser{
				UserID: id,
			},
		})

		if err != nil {
			return err
		}

		for _, item := range users {
			if item.GetID() == id {
				user = item.(*tg.User)
			}
		}

		return nil
	}); err != nil {
		return nil, errors.Wrap(err, "failed to get user")
	}

	if user == nil {
		return nil, fmt.Errorf("can't find user with id: %d", id)
	}

	return user, nil
}

// API returns telegram API
func (t *Telegram) API() (*tg.Client, error) {
	if t.client == nil {
		return nil, fmt.Errorf("telegram client isn't initialized")
	}

	return t.client.API(), nil
}
