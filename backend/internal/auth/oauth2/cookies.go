package oauth2

import (
	"fmt"
	"net/http"
	"time"
)

// Cookie names
const (
	AccessTokenCookie  = "shadowapi_access_token"
	RefreshTokenCookie = "shadowapi_refresh_token"
	WorkspaceCookie    = "shadowapi_workspace"
)

// CookieConfig holds configuration for auth cookies
type CookieConfig struct {
	Domain   string
	Secure   bool
	SameSite http.SameSite
}

// SetTokenCookies sets HttpOnly cookies for access and refresh tokens
func SetTokenCookies(w http.ResponseWriter, cfg CookieConfig, accessToken, refreshToken string, accessTTL, refreshTTL time.Duration) {
	// Access token cookie - sent with all API requests
	http.SetCookie(w, &http.Cookie{
		Name:     AccessTokenCookie,
		Value:    accessToken,
		Path:     "/api",
		Domain:   cfg.Domain,
		MaxAge:   int(accessTTL.Seconds()),
		HttpOnly: true,
		Secure:   cfg.Secure,
		SameSite: cfg.SameSite,
	})

	// Refresh token cookie - only sent to refresh endpoint
	http.SetCookie(w, &http.Cookie{
		Name:     RefreshTokenCookie,
		Value:    refreshToken,
		Path:     "/api/v1/auth/oauth2",
		Domain:   cfg.Domain,
		MaxAge:   int(refreshTTL.Seconds()),
		HttpOnly: true,
		Secure:   cfg.Secure,
		SameSite: cfg.SameSite,
	})
}

// ClearTokenCookies expires both token cookies
func ClearTokenCookies(w http.ResponseWriter, cfg CookieConfig) {
	http.SetCookie(w, &http.Cookie{
		Name:     AccessTokenCookie,
		Value:    "",
		Path:     "/api",
		Domain:   cfg.Domain,
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   cfg.Secure,
		SameSite: cfg.SameSite,
	})

	http.SetCookie(w, &http.Cookie{
		Name:     RefreshTokenCookie,
		Value:    "",
		Path:     "/api/v1/auth/oauth2",
		Domain:   cfg.Domain,
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   cfg.Secure,
		SameSite: cfg.SameSite,
	})
}

// GetAccessTokenFromCookie extracts the access token from the request cookie
func GetAccessTokenFromCookie(r *http.Request) (string, error) {
	cookie, err := r.Cookie(AccessTokenCookie)
	if err != nil {
		return "", err
	}
	return cookie.Value, nil
}

// GetRefreshTokenFromCookie extracts the refresh token from the request cookie
func GetRefreshTokenFromCookie(r *http.Request) (string, error) {
	cookie, err := r.Cookie(RefreshTokenCookie)
	if err != nil {
		return "", err
	}
	return cookie.Value, nil
}

// SetWorkspaceCookie sets the workspace slug cookie (not HttpOnly so frontend can read it)
func SetWorkspaceCookie(w http.ResponseWriter, cfg CookieConfig, slug string) {
	http.SetCookie(w, &http.Cookie{
		Name:     WorkspaceCookie,
		Value:    slug,
		Path:     "/",
		Domain:   cfg.Domain,
		MaxAge:   30 * 24 * 3600, // 30 days
		HttpOnly: false,          // Frontend needs to read this
		Secure:   cfg.Secure,
		SameSite: cfg.SameSite,
	})
}

// ClearWorkspaceCookie expires the workspace cookie
func ClearWorkspaceCookie(w http.ResponseWriter, cfg CookieConfig) {
	http.SetCookie(w, &http.Cookie{
		Name:     WorkspaceCookie,
		Value:    "",
		Path:     "/",
		Domain:   cfg.Domain,
		MaxAge:   -1,
		HttpOnly: false,
		Secure:   cfg.Secure,
		SameSite: cfg.SameSite,
	})
}

// GetWorkspaceSlugFromCookie reads the workspace slug from the request cookie
func GetWorkspaceSlugFromCookie(r *http.Request) (string, error) {
	cookie, err := r.Cookie(WorkspaceCookie)
	if err != nil {
		return "", err
	}
	return cookie.Value, nil
}

// BuildWorkspaceCookieHeader builds a Set-Cookie header string for the workspace cookie
func BuildWorkspaceCookieHeader(cfg CookieConfig, slug string) string {
	secure := ""
	if cfg.Secure {
		secure = "; Secure"
	}
	return fmt.Sprintf("%s=%s; Path=/; Domain=%s; Max-Age=%d; SameSite=Lax%s",
		WorkspaceCookie,
		slug,
		cfg.Domain,
		30*24*3600,
		secure,
	)
}

// BuildClearWorkspaceCookieHeader builds a Set-Cookie header string to clear the workspace cookie
func BuildClearWorkspaceCookieHeader(cfg CookieConfig) string {
	secure := ""
	if cfg.Secure {
		secure = "; Secure"
	}
	return fmt.Sprintf("%s=; Path=/; Domain=%s; Max-Age=0; SameSite=Lax%s",
		WorkspaceCookie,
		cfg.Domain,
		secure,
	)
}

