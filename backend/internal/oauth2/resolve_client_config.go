package oauth2

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/oauth2"

	googleOAuth2 "golang.org/x/oauth2/google"

	"github.com/shadowapi/shadowapi/backend/pkg/query"
)

// GetClientConfig returns the oauth2 client for the given provider.
func GetClientConfig(ctx context.Context, dbp *pgxpool.Pool, clientID string) (*Config, error) {
	tx := query.New(dbp)

	provider, err := tx.GetOauth2Client(ctx, clientID)
	if err != nil {
		slog.Error("query oauth2 client", "error", err)
		return nil, fmt.Errorf("failed to query oauth2 client")
	}

	return ResolveClientConfig(provider)
}

// ResolveClientConfig returns the oauth2 client for the given provider.
func ResolveClientConfig(provider query.Oauth2Client) (cfg *Config, err error) {
	switch strings.ToLower(provider.Provider) {
	case "gmail":
		cfg = &Config{
			Config: oauth2.Config{
				ClientID:     provider.ID,
				ClientSecret: provider.Secret,
				Endpoint:     googleOAuth2.Endpoint,
				Scopes: []string{
					"https://mail.google.com/",
				},
			},
		}
	case "google":
		cfg = &Config{
			Config: oauth2.Config{
				ClientID:     provider.ID,
				ClientSecret: provider.Secret,
				Endpoint:     googleOAuth2.Endpoint,
				Scopes: []string{
					"https://www.googleapis.com/auth/userinfo.email",
				},
			},
		}
	default:
		return nil, fmt.Errorf("unknown provider %s", provider.Provider)
	}

	cfg.RedirectURL = "http://localtest.me/api/v1/oauth2/callback"
	cfg.Name = provider.Name
	cfg.Provider = provider.Provider
	return
}
