package oauth2

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Claims represents the JWT claims we care about
type Claims struct {
	jwt.RegisteredClaims
	Scope     string `json:"scope,omitempty"`
	ClientID  string `json:"client_id,omitempty"`
	SessionID string `json:"sid,omitempty"`
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
	)

	return claims, nil
}

// Subject returns the subject (user ID) from the claims
func (c *Claims) GetUserID() string {
	return c.Subject
}
