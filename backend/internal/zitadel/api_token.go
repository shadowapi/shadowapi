package zitadel

import (
	"context"
	"crypto/rsa"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"golang.org/x/oauth2"
)

type apiKeyJSON struct {
	KeyID    string `json:"keyId"`
	Key      string `json:"key"`
	ClientID string `json:"clientId"`
}

func loadAPIKey(p string) (*apiKeyJSON, *rsa.PrivateKey, error) {
	f, err := os.Open(p)
	if err != nil {
		return nil, nil, err
	}
	defer f.Close()
	raw, err := io.ReadAll(f)
	if err != nil {
		return nil, nil, err
	}
	var k apiKeyJSON
	if err = json.Unmarshal(raw, &k); err != nil {
		return nil, nil, err
	}
	block, _ := pem.Decode([]byte(k.Key))
	if block == nil {
		return nil, nil, errors.New("pem decode failed")
	}
	jwkKey, err := jwk.ParseKey(block.Bytes, jwk.WithPEM(true))
	if err != nil {
		return nil, nil, err
	}
	var priv *rsa.PrivateKey
	if err = jwkKey.Raw(&priv); err != nil {
		return nil, nil, err
	}
	return &k, priv, nil
}

// TokenWithJWTProfile issues an access token for backend-to-backend calls.
func (c *Client) TokenWithJWTProfile(ctx context.Context, scope string) (*oauth2.Token, error) {
	if c.cfg.Auth.Zitadel.APIKeyFile == "" {
		return nil, errors.New("zitadel.api_key_file not configured")
	}
	meta, priv, err := loadAPIKey(c.cfg.Auth.Zitadel.APIKeyFile)
	if err != nil {
		return nil, err
	}
	audience := c.cfg.Auth.Zitadel.InstanceURL + "/oauth/v2/token"
	claims, err := jwt.NewBuilder().
		Issuer(meta.ClientID).
		Subject(meta.ClientID).
		Audience([]string{audience}).
		IssuedAt(time.Now()).
		Expiration(time.Now().Add(5*time.Minute)).
		Claim("jti", uuid.NewString()).
		Build()
	if err != nil {
		return nil, err
	}
	signed, err := jwt.Sign(claims, jwt.WithKey(jwa.RS256, priv))
	if err != nil {
		return nil, err
	}
	form := url.Values{
		"grant_type": {"urn:ietf:params:oauth:grant-type:jwt-bearer"},
		"assertion":  {string(signed)},
	}
	if scope != "" {
		form.Set("scope", scope)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, audience, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	res, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(res.Body)
		return nil, fmt.Errorf("token status %d: %s", res.StatusCode, body)
	}
	var tok oauth2.Token
	return &tok, json.NewDecoder(res.Body).Decode(&tok)
}
