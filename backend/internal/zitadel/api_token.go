package zitadel

//
//import (
//	"context"
//	"crypto/rsa"
//	"encoding/json"
//	"encoding/pem"
//	"errors"
//	"io"
//	"net/http"
//	"net/url"
//	"os"
//	"strings"
//	"time"
//
//	"github.com/google/uuid"
//	"github.com/lestrrat-go/jwx/v2/jwa"
//	"github.com/lestrrat-go/jwx/v2/jwk"
//	"github.com/lestrrat-go/jwx/v2/jwt"
//	"golang.org/x/oauth2"
//)
//
//type apiKeyJSON struct {
//	KeyID    string `json:"keyId"`
//	Key      string `json:"key"`
//	ClientID string `json:"clientId"`
//}
//
//func loadAPIKey(p string) (*apiKeyJSON, *rsa.PrivateKey, error) {
//	f, err := os.Open(p)
//	if err != nil {
//		return nil, nil, err
//	}
//	defer f.Close()
//	raw, _ := io.ReadAll(f)
//	var k apiKeyJSON
//	if err := json.Unmarshal(raw, &k); err != nil {
//		return nil, nil, err
//	}
//	b, _ := pem.Decode([]byte(k.Key))
//	if b == nil {
//		return nil, nil, errors.New("pem decode")
//	}
//	key, err := jwk.ParseKey(b.Bytes, jwk.WithPEM(true))
//	if err != nil {
//		return nil, nil, err
//	}
//	var rsaKey *rsa.PrivateKey
//	if err := key.Raw(&rsaKey); err != nil {
//		return nil, nil, err
//	}
//	return &k, rsaKey, nil
//}
//
//func (c *Client) TokenWithJWTProfile(ctx context.Context, scope string) (*oauth2.Token, error) {
//	meta, priv, err := loadAPIKey(c.cfg.Auth.Zitadel.APIKeyFile)
//	if err != nil {
//		return nil, err
//	}
//	j, _ := jwt.NewBuilder().
//		Issuer(meta.ClientID).
//		Subject(meta.ClientID).
//		Audience([]string{c.cfg.Auth.Zitadel.InstanceURL + "/oauth/v2/token"}).
//		IssuedAt(time.Now()).
//		Expiration(time.Now().Add(5*time.Minute)).
//		Claim("jti", uuid.NewString()).
//		Build()
//	signed, _ := jwt.Sign(j, jwt.WithKey(jwa.RS256, priv))
//	v := url.Values{}
//	v.Set("grant_type", "urn:ietf:params:oauth:grant-type:jwt-bearer")
//	v.Set("assertion", string(signed))
//	if scope != "" {
//		v.Set("scope", scope)
//	}
//	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, c.cfg.Auth.Zitadel.InstanceURL+"/oauth/v2/token", strings.NewReader(v.Encode()))
//	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
//	res, err := c.client.Do(req)
//	if err != nil {
//		return nil, err
//	}
//	defer res.Body.Close()
//	var tok oauth2.Token
//	return &tok, json.NewDecoder(res.Body).Decode(&tok)
//}
