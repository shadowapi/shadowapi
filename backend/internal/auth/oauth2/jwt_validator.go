package oauth2

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// ExtClaims holds custom claims placed in the "ext" field by Hydra
type ExtClaims struct {
	WorkspaceID   string `json:"workspace_id,omitempty"`
	WorkspaceSlug string `json:"workspace_slug,omitempty"`
}

// Claims represents the JWT claims we care about
// Workspace context is stored in JWT claims and set during workspace switch.
// Hydra v2.x puts session.access_token claims under the "ext" field.
type Claims struct {
	jwt.RegisteredClaims
	Scope         string    `json:"scope,omitempty"`
	ClientID      string    `json:"client_id,omitempty"`
	SessionID     string    `json:"sid,omitempty"`
	Ext           ExtClaims `json:"ext,omitempty"` // Hydra stores custom claims here
	WorkspaceID   string    `json:"-"`             // Populated from Ext after parsing
	WorkspaceSlug string    `json:"-"`             // Populated from Ext after parsing
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

	// Copy workspace claims from Ext to top-level fields for easier access
	claims.WorkspaceID = claims.Ext.WorkspaceID
	claims.WorkspaceSlug = claims.Ext.WorkspaceSlug

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
		"workspace_id", claims.WorkspaceID,
		"workspace_slug", claims.WorkspaceSlug,
		"expires_at", claims.ExpiresAt,
	)

	return claims, nil
}

// Subject returns the subject (user ID) from the claims
func (c *Claims) GetUserID() string {
	return c.Subject
}
