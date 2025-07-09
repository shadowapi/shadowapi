package handlers

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"time"

	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/lestrrat-go/jwx/v2/jwt"

	zitadellog "github.com/shadowapi/shadowapi/backend/internal/auth/zitadel"
	"github.com/shadowapi/shadowapi/backend/internal/config"
	"github.com/shadowapi/shadowapi/backend/internal/handler"
	"github.com/shadowapi/shadowapi/backend/internal/session"
	"github.com/shadowapi/shadowapi/backend/internal/zitadel"
	"github.com/shadowapi/shadowapi/backend/pkg/query"
)

// AuthHandlers groups login related HTTP handlers.
type AuthHandlers struct {
	cfg      *config.Config
	log      *slog.Logger
	zitadel  *zitadel.Client
	handler  *handler.Handler
	sessions *session.Middleware
}

// NewAuthHandlers returns a new AuthHandlers instance.
func NewAuthHandlers(cfg *config.Config, log *slog.Logger, z *zitadel.Client, h *handler.Handler, s *session.Middleware) *AuthHandlers {
	return &AuthHandlers{cfg: cfg, log: log, zitadel: z, handler: h, sessions: s}
}

func newCodeVerifier() (string, string, error) {
	b := make([]byte, 43) // 256-bit entropy
	if _, err := rand.Read(b); err != nil {
		return "", "", err
	}
	verifier := base64.RawURLEncoding.EncodeToString(b)
	sum := sha256.Sum256([]byte(verifier))
	challenge := base64.RawURLEncoding.EncodeToString(sum[:])
	return verifier, challenge, nil
}

// ZitadelLogin initiates the ZITADEL PKCE flow.
func (h *AuthHandlers) ZitadelLogin(w http.ResponseWriter, r *http.Request) {
	ver, chal, err := newCodeVerifier()
	if err != nil {
		h.log.Error("pkce", "err", err)
		http.Error(w, "internal", http.StatusInternalServerError)
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     "sa_pkce",
		Value:    ver,
		Path:     "/auth",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   r.TLS != nil,
		MaxAge:   300,
	})
	u := h.cfg.Auth.Zitadel.InstanceURL + "/oauth/v2/authorize?" + url.Values{
		"client_id":             {h.cfg.Auth.Zitadel.Audience},
		"response_type":         {"code"},
		"scope":                 {"openid profile email"},
		"redirect_uri":          {h.cfg.Auth.Zitadel.RedirectURI},
		"state":                 {fmt.Sprintf("%d", time.Now().UnixNano())},
		"code_challenge":        {chal},
		"code_challenge_method": {"S256"},
	}.Encode()
	http.Redirect(w, r, u, http.StatusFound)
}

// AuthCallback exchanges the code for a token and sets session cookies.
func (h *AuthHandlers) AuthCallback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "missing code", http.StatusBadRequest)
		return
	}
	ver, _ := r.Cookie("sa_pkce")

	tok, err := h.zitadel.ExchangeCode(r.Context(), code, func() string {
		if ver != nil {
			return ver.Value
		}
		return ""
	}())
	if err != nil {
		zitadellog.LogExchangeError(h.log, err)
		http.Error(w, "exchange failed", http.StatusInternalServerError)
		return
	}
	h.log.Debug("code exchanged", "access_exp", tok.Expiry)

	rawID, ok := tok.Extra("id_token").(string)
	if !ok || rawID == "" {
		h.log.Warn("exchange: no id_token")
		http.Error(w, "id_token missing", http.StatusUnauthorized)
		return
	}
	h.log.Info("rawID", "value", rawID)

	idToken, err := jwt.ParseString(rawID, jwt.WithVerify(false))
	if err != nil {
		h.log.Error("id_token parse", "err", err)
		http.Error(w, "token parse failed", http.StatusUnauthorized)
		return
	}
	h.log.Info("idToken", "value", idToken)

	subject := idToken.Subject()
	if subject == "" {
		http.Error(w, "token missing sub", http.StatusUnauthorized)
		return
	}
	emailAny, _ := idToken.Get("email")
	email, _ := emailAny.(string)
	if email == "" {
		email = fmt.Sprintf("%s@zitadel.local", subject)
	}

	q := query.New(h.handler.DB())
	user, errUser := q.GetUserByZitadelSubject(
		r.Context(),
		pgtype.Text{String: subject, Valid: true},
	)
	if errors.Is(errUser, pgx.ErrNoRows) {
		row := h.handler.DB().QueryRow(r.Context(), getUserByEmailQuery, email)
		errEmail := row.Scan(
			&user.UUID,
			&user.Email,
			&user.Password,
			&user.FirstName,
			&user.LastName,
			&user.IsEnabled,
			&user.IsAdmin,
			&user.ZitadelSubject,
			&user.Meta,
			&user.CreatedAt,
			&user.UpdatedAt,
		)
		switch {
		case errors.Is(errEmail, pgx.ErrNoRows):
			uuidv7 := uuid.Must(uuid.NewV7())
			user, errUser = q.CreateUser(r.Context(), query.CreateUserParams{
				UUID:           pgtype.UUID{Bytes: uuidv7, Valid: true},
				Email:          email,
				Password:       "",
				FirstName:      "",
				LastName:       "",
				IsEnabled:      false,
				IsAdmin:        false,
				ZitadelSubject: pgtype.Text{String: subject, Valid: true},
				Meta:           []byte(`{}`),
			})
		case errEmail != nil:
			errUser = errEmail
		default:
			if !user.ZitadelSubject.Valid {
				user.ZitadelSubject = pgtype.Text{String: subject, Valid: true}
				errUser = q.UpdateUser(r.Context(), query.UpdateUserParams{
					Email:          user.Email,
					Password:       user.Password,
					FirstName:      user.FirstName,
					LastName:       user.LastName,
					IsEnabled:      user.IsEnabled,
					IsAdmin:        user.IsAdmin,
					ZitadelSubject: user.ZitadelSubject,
					Meta:           user.Meta,
					UUID:           pgtype.UUID{Bytes: user.UUID, Valid: true},
				})
			} else {
				errUser = nil
			}
		}
	}
	if errUser != nil {
		h.log.Error("user upsert", "err", errUser)
		http.Error(w, "user store failed", http.StatusInternalServerError)
		return
	}
	h.log.Info("user upserted", "uid", user.UUID, "email", user.Email, "enabled", user.IsEnabled)

	http.SetCookie(w, &http.Cookie{
		Name:     "zitadel_access_token",
		Value:    tok.AccessToken,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	if !user.IsEnabled {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	token := uuid.Must(uuid.NewV7()).String()
	h.sessions.AddSession(token, user.UUID.String())
	h.log.Debug("session created", "uid", user.UUID, "token", token)

	http.SetCookie(w, &http.Cookie{
		Name:     "sa_session",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	http.Redirect(w, r, "/", http.StatusFound)
}

// PlainLogin verifies email/password and sets session cookie.
func (h *AuthHandlers) PlainLogin(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	userID, err := h.handler.PlainLogin(r.Context(), req.Email, req.Password)
	if err != nil {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}
	token := uuid.Must(uuid.NewV7()).String()
	h.sessions.AddSession(token, userID)
	http.SetCookie(w, &http.Cookie{
		Name:     "sa_session",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
	_ = json.NewEncoder(w).Encode(map[string]bool{"active": true})
}

// Logout invalidates local session and redirects to the appropriate logout flow.
func (h *AuthHandlers) Logout(w http.ResponseWriter, r *http.Request) {
	sessCookie, errSess := r.Cookie("sa_session")
	if errSess == nil {
		h.sessions.DeleteSession(sessCookie.Value)
	}
	http.SetCookie(w, &http.Cookie{
		Name:   "sa_session",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	})

	_, zitadelCookieErr := r.Cookie("zitadel_access_token")
	http.SetCookie(w, &http.Cookie{
		Name:   "zitadel_access_token",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	})

	if zitadelCookieErr == nil {
		scheme := "http"
		if r.TLS != nil {
			scheme = "https"
		}
		redirect := fmt.Sprintf("%s://%s/logout/callback", scheme, r.Host)
		target := fmt.Sprintf("%s/oidc/v1/end_session?post_logout_redirect_uri=%s", h.cfg.Auth.Zitadel.InstanceURL, url.QueryEscape(redirect))
		http.Redirect(w, r, target, http.StatusFound)
		return
	}

	http.Redirect(w, r, "/logout/callback", http.StatusFound)
}

// LogoutCallback clears session cookie.
func (h *AuthHandlers) LogoutCallback(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("sa_session")
	if err == nil {
		h.sessions.DeleteSession(cookie.Value)
	}
	http.SetCookie(w, &http.Cookie{
		Name:   "sa_session",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	})
	http.SetCookie(w, &http.Cookie{
		Name:   "zitadel_access_token",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	})
	http.Redirect(w, r, "/", http.StatusFound)
}

const getUserByEmailQuery = `SELECT
    uuid, email, password, first_name, last_name, is_enabled, is_admin, zitadel_subject, meta, created_at, updated_at
FROM "user" WHERE email=$1`
