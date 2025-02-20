package oauth2

import (
	"context"
	"net/http"

	"golang.org/x/oauth2"
)

type Config struct {
	oauth2.Config
	Name     string
	Provider string
	Secret   string
	Token    string
}

// Client returns OAuth2 client with cached Token store
func (c *Config) Client(ctx context.Context, t *oauth2.Token) *http.Client {
	return oauth2.NewClient(ctx, c.TokenSource(ctx, t))
}

func (c *Config) Close() error {
	return nil
}
