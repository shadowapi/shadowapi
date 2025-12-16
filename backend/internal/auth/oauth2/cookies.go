package oauth2

import (
	"net/http"
	"time"
)

// Cookie names
const (
	AccessTokenCookie  = "shadowapi_access_token"
	RefreshTokenCookie = "shadowapi_refresh_token"
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

