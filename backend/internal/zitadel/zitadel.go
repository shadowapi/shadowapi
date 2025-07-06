package zitadel

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"golang.org/x/oauth2"

	"github.com/shadowapi/shadowapi/backend/internal/config"
)

// Client wraps OAuth2 config for Zitadel service user
// and provides helpers for token exchange and introspection.
type Client struct {
	cfg    *config.Config
	oauth2 *oauth2.Config
	client *http.Client
}

// Provide creates a new Client for dependency injection
func Provide(c *config.Config) *Client {
	oc := &oauth2.Config{
		ClientID:     c.Auth.Zitadel.ClientID,
		ClientSecret: c.Auth.Zitadel.ClientSecret,
		Endpoint: oauth2.Endpoint{
			AuthURL:  fmt.Sprintf("%s/oauth/v2/authorize", c.Auth.Zitadel.InstanceURL),
			TokenURL: fmt.Sprintf("%s/oauth/v2/token", c.Auth.Zitadel.InstanceURL),
		},
		RedirectURL: c.Auth.Zitadel.RedirectURI,
		Scopes:      []string{"urn:zitadel:iam:org:project:id:zitadel:aud"},
	}
	return &Client{cfg: c, oauth2: oc, client: oc.Client(context.Background(), nil)}
}

// ExchangeCode exchanges authorization code for tokens
func (c *Client) ExchangeCode(ctx context.Context, code string) (*oauth2.Token, error) {
	return c.oauth2.Exchange(ctx, code)
}

// IntrospectResponse describes subset of fields returned by /oauth/v2/introspect
// see https://docs.zitadel.com for full schema.
type IntrospectResponse struct {
	Active  bool   `json:"active"`
	Subject string `json:"sub"`
}

// Introspect validates an access token using the service user credentials
func (c *Client) Introspect(ctx context.Context, token string) (*IntrospectResponse, error) {
	form := url.Values{}
	form.Set("token", token)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("%s/oauth/v2/introspect", c.cfg.Auth.Zitadel.InstanceURL), strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(c.cfg.Auth.Zitadel.ClientID, c.cfg.Auth.Zitadel.ClientSecret)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("introspect status %d", resp.StatusCode)
	}
	var out IntrospectResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	return &out, nil
}
