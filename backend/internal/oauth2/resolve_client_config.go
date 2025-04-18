package oauth2

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shadowapi/shadowapi/backend/internal/converter"
	"github.com/shadowapi/shadowapi/backend/pkg/query"
	"golang.org/x/oauth2"
	googleOAuth2 "golang.org/x/oauth2/google"
)

// GetClientConfig returns resolved OAuth2 Config for given oauth2_client uuid.
func GetClientConfig(ctx context.Context, dbp *pgxpool.Pool, clientID string) (*Config, error) {
	tx := query.New(dbp)
	pgID, err := converter.ConvertStringToPgUUID(clientID)
	if err != nil {
		slog.Error("convert client id", "error", err)
		return nil, fmt.Errorf("invalid client id")
	}

	row, err := tx.GetOauth2Client(ctx, pgID)
	if err != nil {
		slog.Error("query oauth2 client", "error", err)
		return nil, fmt.Errorf("failed to query oauth2 client")
	}
	return ResolveClientConfig(row.Oauth2Client)
}

// ResolveClientConfig builds *oauth2.Config from DB row.
func ResolveClientConfig(provider query.Oauth2Client) (*Config, error) {
	var base oauth2.Config

	switch strings.ToLower(provider.Provider) {
	case "gmail":
		base = oauth2.Config{
			ClientID:     provider.ClientID,
			ClientSecret: provider.Secret,
			Endpoint:     googleOAuth2.Endpoint,
			Scopes: []string{
				"https://mail.google.com/",
				"https://www.googleapis.com/auth/userinfo.email",
				"https://www.googleapis.com/auth/gmail.readonly",
			},
		}
	case "google":
		base = oauth2.Config{
			ClientID:     provider.ClientID,
			ClientSecret: provider.Secret,
			Endpoint:     googleOAuth2.Endpoint,
			Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email"},
		}
	default:
		return nil, fmt.Errorf("unknown provider %s", provider.Provider)
	}

	// redirect uri must match google console; default to localhost variant.
	base.RedirectURL = "http://localhost/api/v1/oauth2/callback"

	return &Config{
		Config:   base,
		Name:     provider.Name,
		Provider: provider.Provider,
		Secret:   provider.Secret,
	}, nil
}
