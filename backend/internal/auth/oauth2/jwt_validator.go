package oauth2

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// ClaimsExt represents the extra claims embedded in the JWT's ext field
type ClaimsExt struct {
	TenantUUID string `json:"tenant_uuid,omitempty"`
	TenantName string `json:"tenant_name,omitempty"`
}

// Claims represents the JWT claims we care about
type Claims struct {
	jwt.RegisteredClaims
	Scope    string    `json:"scope,omitempty"`
	ClientID string    `json:"client_id,omitempty"`
	SessionID string   `json:"sid,omitempty"`
	Ext      ClaimsExt `json:"ext,omitempty"` // Hydra puts custom claims here
}

// TenantUUID returns the tenant UUID from ext claims
func (c *Claims) TenantUUID() string {
	return c.Ext.TenantUUID
}

// TenantName returns the tenant name from ext claims
func (c *Claims) TenantName() string {
	return c.Ext.TenantName
}

// JWTValidator validates JWT tokens using JWKS
type JWTValidator struct {
	cache  *JWKSCache
	issuer string
	log    *slog.Logger
}

// NewJWTValidator creates a new JWT validator
func NewJWTValidator(cache *JWKSCache, issuer string, log *slog.Logger) *JWTValidator {
	return &JWTValidator{
		cache:  cache,
		issuer: issuer,
		log:    log,
	}
}

// Validate parses and validates a JWT token
func (v *JWTValidator) Validate(ctx context.Context, tokenString string) (*Claims, error) {
	// Parse the token with custom key function
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Verify signing method
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		// Get the key ID from the token header
		kid, ok := token.Header["kid"].(string)
		if !ok {
			return nil, fmt.Errorf("token header missing kid")
		}

		// Fetch the public key from JWKS cache
		return v.cache.GetKey(ctx, kid)
	})
	if err != nil {
		v.log.Debug("JWT parse error", "error", err)
		return nil, fmt.Errorf("parse token: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token claims")
	}

	// Validate issuer
	if v.issuer != "" && claims.Issuer != v.issuer {
		return nil, fmt.Errorf("invalid issuer: expected %q, got %q", v.issuer, claims.Issuer)
	}

	// Validate expiration
	if claims.ExpiresAt != nil && claims.ExpiresAt.Time.Before(time.Now()) {
		return nil, fmt.Errorf("token expired")
	}

	v.log.Debug("JWT validated successfully",
		"subject", claims.Subject,
		"client_id", claims.ClientID,
		"expires_at", claims.ExpiresAt,
		"tenant_uuid", claims.TenantUUID(),
		"tenant_name", claims.TenantName(),
	)

	return claims, nil
}

// Subject returns the subject (user ID) from the claims
func (c *Claims) GetUserID() string {
	return c.Subject
}
