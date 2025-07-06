package auth

import (
	"context"
	"log/slog"

	"github.com/samber/do/v2"

	"github.com/shadowapi/shadowapi/backend/pkg/api"
)

type Auth struct {
	log *slog.Logger
}

// Provide returns the authenticator instance
func Provide(i do.Injector) (*Auth, error) {
	// keep log case of debugging ogen
	return &Auth{
		log: do.MustInvoke[*slog.Logger](i),
	}, nil
}

// HandleBearerAuth checks the Bearer token
// this is just a passthrough as we use session middleware instead
// keep it so default HandleBearerAuth wont be triggered
func (a *Auth) HandleBearerAuth(
	ctx context.Context,
	op api.OperationName,
	t api.BearerAuth,
) (context.Context, error) {
	return ctx, nil
}

// HandleZitadelCookieAuth handles session cookie authentication
// this is just a passthrough as we use session middleware instead
// keep it so default HandleZitadelCookieAuth wont be triggered
func (a *Auth) HandleZitadelCookieAuth(
	ctx context.Context, operationName api.OperationName, t api.ZitadelCookieAuth,
) (context.Context, error) {
	return ctx, nil
}
