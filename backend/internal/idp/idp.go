package idp

import (
    "context"

    "github.com/shadowapi/shadowapi/backend/internal/config"
    "github.com/shadowapi/shadowapi/backend/internal/zitadel"
    "golang.org/x/oauth2"
)

// Provider defines minimal operations needed for login via an external IDP
type Provider interface {
    // Name of the provider (e.g., "zitadel", "auth0")
    Name() string
    // AuthCodeURL builds an authorization URL, optional extra options
    AuthCodeURL(state string, extra ...oauth2.AuthCodeOption) string
    // Exchange exchanges an auth code for tokens; redirectURI must match the one used at AuthCodeURL
    Exchange(ctx context.Context, code, redirectURI string) (*oauth2.Token, error)
}

// NewProvider constructs an IDP provider based on configuration
// Currently supports "zitadel". Other providers (e.g., Auth0) can be added later.
func NewProvider(cfg *config.Config) Provider {
    // Default/fallback to Zitadel if selected or unspecified
    // (switch later when additional providers are implemented)
    zc := zitadel.Provide(cfg)
    return &zitadelProvider{c: zc}
}

type zitadelProvider struct{ c *zitadel.Client }

func (z *zitadelProvider) Name() string { return "zitadel" }
func (z *zitadelProvider) AuthCodeURL(state string, extra ...oauth2.AuthCodeOption) string {
    return z.c.AuthCodeURL(state, extra...)
}
func (z *zitadelProvider) Exchange(ctx context.Context, code, redirectURI string) (*oauth2.Token, error) {
    return z.c.ExchangeCode(ctx, code, "", redirectURI)
}
