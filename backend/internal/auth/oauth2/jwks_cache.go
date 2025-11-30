package oauth2

import (
	"context"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log/slog"
	"math/big"
	"net/http"
	"sync"
	"time"
)

// JWK represents a JSON Web Key
type JWK struct {
	Kty string `json:"kty"` // Key type (RSA, EC, etc.)
	Kid string `json:"kid"` // Key ID
	Use string `json:"use"` // Key use (sig, enc)
	Alg string `json:"alg"` // Algorithm (RS256, etc.)
	N   string `json:"n"`   // RSA modulus
	E   string `json:"e"`   // RSA exponent
}

// JWKS represents a JSON Web Key Set
type JWKS struct {
	Keys []JWK `json:"keys"`
}

// JWKSCache provides thread-safe caching of JWKS keys
type JWKSCache struct {
	mu          sync.RWMutex
	keys        map[string]*rsa.PublicKey
	lastFetched time.Time
	ttl         time.Duration
	jwksURL     string
	httpClient  *http.Client
	log         *slog.Logger
}

// NewJWKSCache creates a new JWKS cache
func NewJWKSCache(jwksURL string, ttl time.Duration, log *slog.Logger) *JWKSCache {
	return &JWKSCache{
		keys:       make(map[string]*rsa.PublicKey),
		ttl:        ttl,
		jwksURL:    jwksURL,
		httpClient: &http.Client{Timeout: 10 * time.Second},
		log:        log,
	}
}

// GetKey retrieves a public key by key ID, refreshing the cache if needed
func (c *JWKSCache) GetKey(ctx context.Context, kid string) (*rsa.PublicKey, error) {
	c.mu.RLock()
	if time.Since(c.lastFetched) < c.ttl {
		if key, ok := c.keys[kid]; ok {
			c.mu.RUnlock()
			return key, nil
		}
	}
	c.mu.RUnlock()

	// Need to refresh the cache
	if err := c.refresh(ctx); err != nil {
		return nil, fmt.Errorf("failed to refresh JWKS: %w", err)
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	key, ok := c.keys[kid]
	if !ok {
		return nil, fmt.Errorf("key with kid %q not found in JWKS", kid)
	}
	return key, nil
}

// refresh fetches the JWKS from the remote endpoint and updates the cache
func (c *JWKSCache) refresh(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Double-check after acquiring write lock
	if time.Since(c.lastFetched) < c.ttl && len(c.keys) > 0 {
		return nil
	}

	c.log.Debug("refreshing JWKS cache", "url", c.jwksURL)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.jwksURL, nil)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("fetch JWKS: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("JWKS endpoint returned status %d", resp.StatusCode)
	}

	var jwks JWKS
	if err := json.NewDecoder(resp.Body).Decode(&jwks); err != nil {
		return fmt.Errorf("decode JWKS: %w", err)
	}

	newKeys := make(map[string]*rsa.PublicKey)
	for _, jwk := range jwks.Keys {
		if jwk.Kty != "RSA" {
			c.log.Debug("skipping non-RSA key", "kid", jwk.Kid, "kty", jwk.Kty)
			continue
		}

		pubKey, err := parseRSAPublicKey(jwk)
		if err != nil {
			c.log.Warn("failed to parse RSA public key", "kid", jwk.Kid, "error", err)
			continue
		}

		newKeys[jwk.Kid] = pubKey
		c.log.Debug("loaded JWKS key", "kid", jwk.Kid, "alg", jwk.Alg)
	}

	if len(newKeys) == 0 {
		return fmt.Errorf("no valid RSA keys found in JWKS")
	}

	c.keys = newKeys
	c.lastFetched = time.Now()
	c.log.Info("JWKS cache refreshed", "keys_count", len(newKeys))

	return nil
}

// parseRSAPublicKey converts a JWK to an RSA public key
func parseRSAPublicKey(jwk JWK) (*rsa.PublicKey, error) {
	// Decode the modulus (n)
	nBytes, err := base64.RawURLEncoding.DecodeString(jwk.N)
	if err != nil {
		return nil, fmt.Errorf("decode modulus: %w", err)
	}
	n := new(big.Int).SetBytes(nBytes)

	// Decode the exponent (e)
	eBytes, err := base64.RawURLEncoding.DecodeString(jwk.E)
	if err != nil {
		return nil, fmt.Errorf("decode exponent: %w", err)
	}

	// Convert exponent bytes to int
	var e int
	for _, b := range eBytes {
		e = e<<8 + int(b)
	}

	return &rsa.PublicKey{N: n, E: e}, nil
}
