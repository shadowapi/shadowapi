// mock-oidc is a lightweight OIDC provider for local development.
// It implements the custom oxoauth protocol expected by the backend:
//   - GET  /authorize      → redirects to backend login page
//   - GET  /login/callback → verifies HMAC, issues code, redirects to OAuth2 callback
//   - POST /oauth/token    → exchanges code for JWT tokens
//   - GET  /keys           → serves JWKS public key
//   - POST /oauth/revoke   → no-op, returns 200
package main

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// authRequest stores the state for an authorization request
type authRequest struct {
	ClientID      string
	RedirectURI   string
	State         string
	CodeChallenge string
	CreatedAt     time.Time
}

// authCode stores the state for an issued authorization code
type authCode struct {
	Subject       string
	Claims        map[string]interface{}
	RedirectURI   string
	State         string
	CodeChallenge string
	ClientID      string
	CreatedAt     time.Time
}

type server struct {
	privateKey     *rsa.PrivateKey
	keyID          string
	callbackSecret string
	issuerURL      string
	loginURL       string // backend login page URL

	mu           sync.RWMutex
	authRequests map[string]*authRequest // auth_request_id → request
	authCodes    map[string]*authCode    // code → auth code data
}

func main() {
	callbackSecret := os.Getenv("OIDC_CALLBACK_SECRET")
	if callbackSecret == "" {
		log.Fatal("OIDC_CALLBACK_SECRET is required")
	}

	issuerURL := os.Getenv("OIDC_ISSUER_URL")
	if issuerURL == "" {
		issuerURL = "http://oidc.localtest.me"
	}

	loginURL := os.Getenv("OIDC_LOGIN_URL")
	if loginURL == "" {
		loginURL = "http://api.localtest.me/api/v1/auth/login"
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "4444"
	}

	// Generate RSA key pair for JWT signing
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		log.Fatalf("failed to generate RSA key: %v", err)
	}

	s := &server{
		privateKey:     privateKey,
		keyID:          "mock-oidc-dev-key",
		callbackSecret: callbackSecret,
		issuerURL:      strings.TrimSuffix(issuerURL, "/"),
		loginURL:       loginURL,
		authRequests:   make(map[string]*authRequest),
		authCodes:      make(map[string]*authCode),
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /authorize", s.handleAuthorize)
	mux.HandleFunc("GET /login/callback", s.handleLoginCallback)
	mux.HandleFunc("POST /oauth/token", s.handleToken)
	mux.HandleFunc("GET /keys", s.handleJWKS)
	mux.HandleFunc("POST /oauth/revoke", s.handleRevoke)
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "ok")
	})

	log.Printf("mock-oidc starting on :%s (issuer: %s)", port, issuerURL)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}

// handleAuthorize handles GET /authorize
// Stores the auth request and redirects to the backend login page
func (s *server) handleAuthorize(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	clientID := q.Get("client_id")
	redirectURI := q.Get("redirect_uri")
	state := q.Get("state")
	codeChallenge := q.Get("code_challenge")

	if clientID == "" || redirectURI == "" || state == "" {
		http.Error(w, "missing required parameters", http.StatusBadRequest)
		return
	}

	// Generate auth_request_id
	authRequestID := generateRandomString(32)

	s.mu.Lock()
	s.authRequests[authRequestID] = &authRequest{
		ClientID:      clientID,
		RedirectURI:   redirectURI,
		State:         state,
		CodeChallenge: codeChallenge,
		CreatedAt:     time.Now(),
	}
	s.mu.Unlock()

	log.Printf("authorize: created auth_request_id=%s for client=%s", authRequestID[:8]+"...", clientID)

	// Redirect to backend login page
	loginRedirect := fmt.Sprintf("%s?auth_request_id=%s", s.loginURL, url.QueryEscape(authRequestID))
	http.Redirect(w, r, loginRedirect, http.StatusFound)
}

// handleLoginCallback handles GET /login/callback
// Verifies the HMAC signature, generates an authorization code, and redirects
func (s *server) handleLoginCallback(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	authRequestID := q.Get("auth_request_id")
	subject := q.Get("subject")
	claimsB64 := q.Get("claims")
	signature := q.Get("signature")

	if authRequestID == "" || subject == "" || claimsB64 == "" || signature == "" {
		http.Error(w, "missing required parameters", http.StatusBadRequest)
		return
	}

	// Verify HMAC signature
	message := authRequestID + "|" + subject + "|" + claimsB64
	mac := hmac.New(sha256.New, []byte(s.callbackSecret))
	mac.Write([]byte(message))
	expectedSig := hex.EncodeToString(mac.Sum(nil))

	if !hmac.Equal([]byte(signature), []byte(expectedSig)) {
		log.Printf("login/callback: HMAC verification failed for auth_request_id=%s", authRequestID[:8]+"...")
		http.Error(w, "invalid signature", http.StatusForbidden)
		return
	}

	// Decode claims
	claimsJSON, err := base64.URLEncoding.DecodeString(claimsB64)
	if err != nil {
		http.Error(w, "invalid claims encoding", http.StatusBadRequest)
		return
	}

	var claims map[string]interface{}
	if err := json.Unmarshal(claimsJSON, &claims); err != nil {
		http.Error(w, "invalid claims JSON", http.StatusBadRequest)
		return
	}

	// Look up auth request
	s.mu.Lock()
	authReq, ok := s.authRequests[authRequestID]
	if ok {
		delete(s.authRequests, authRequestID)
	}
	s.mu.Unlock()

	if !ok {
		http.Error(w, "unknown or expired auth_request_id", http.StatusBadRequest)
		return
	}

	// Generate authorization code
	code := generateRandomString(32)

	s.mu.Lock()
	s.authCodes[code] = &authCode{
		Subject:       subject,
		Claims:        claims,
		RedirectURI:   authReq.RedirectURI,
		State:         authReq.State,
		CodeChallenge: authReq.CodeChallenge,
		ClientID:      authReq.ClientID,
		CreatedAt:     time.Now(),
	}
	s.mu.Unlock()

	log.Printf("login/callback: issued code for subject=%s, redirecting to callback", subject)

	// Redirect to the OAuth2 callback (on the backend)
	redirectURL := fmt.Sprintf("%s?code=%s&state=%s", authReq.RedirectURI, url.QueryEscape(code), url.QueryEscape(authReq.State))
	http.Redirect(w, r, redirectURL, http.StatusFound)
}

// handleToken handles POST /oauth/token
// Exchanges authorization code or refresh token for JWT tokens
func (s *server) handleToken(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid form data", http.StatusBadRequest)
		return
	}

	grantType := r.FormValue("grant_type")

	switch grantType {
	case "authorization_code":
		s.handleAuthCodeExchange(w, r)
	case "refresh_token":
		s.handleRefreshTokenExchange(w, r)
	default:
		http.Error(w, "unsupported grant_type", http.StatusBadRequest)
	}
}

func (s *server) handleAuthCodeExchange(w http.ResponseWriter, r *http.Request) {
	code := r.FormValue("code")
	codeVerifier := r.FormValue("code_verifier")
	clientID := r.FormValue("client_id")

	if code == "" || clientID == "" {
		http.Error(w, "missing code or client_id", http.StatusBadRequest)
		return
	}

	// Look up and consume the code
	s.mu.Lock()
	authCode, ok := s.authCodes[code]
	if ok {
		delete(s.authCodes, code)
	}
	s.mu.Unlock()

	if !ok {
		http.Error(w, "invalid or expired code", http.StatusBadRequest)
		return
	}

	// Verify code challenge (PKCE S256)
	if authCode.CodeChallenge != "" && codeVerifier != "" {
		h := sha256.Sum256([]byte(codeVerifier))
		computedChallenge := base64.RawURLEncoding.EncodeToString(h[:])
		if computedChallenge != authCode.CodeChallenge {
			http.Error(w, "invalid code_verifier", http.StatusBadRequest)
			return
		}
	}

	// Generate tokens
	accessToken, err := s.generateAccessToken(authCode.Subject, authCode.Claims, 3600)
	if err != nil {
		http.Error(w, "failed to generate access token", http.StatusInternalServerError)
		return
	}

	refreshToken := generateRandomString(48)

	log.Printf("token: issued access_token for subject=%s", authCode.Subject)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"access_token":  accessToken,
		"token_type":    "Bearer",
		"expires_in":    3600,
		"refresh_token": refreshToken,
		"scope":         "openid offline_access profile email",
	})
}

func (s *server) handleRefreshTokenExchange(w http.ResponseWriter, r *http.Request) {
	refreshToken := r.FormValue("refresh_token")
	if refreshToken == "" {
		http.Error(w, "missing refresh_token", http.StatusBadRequest)
		return
	}

	// For mock: generate a new access token with a generic subject
	// In production, the refresh token would be looked up to find the subject
	accessToken, err := s.generateAccessToken("mock-user", nil, 3600)
	if err != nil {
		http.Error(w, "failed to generate access token", http.StatusInternalServerError)
		return
	}

	newRefreshToken := generateRandomString(48)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"access_token":  accessToken,
		"token_type":    "Bearer",
		"expires_in":    3600,
		"refresh_token": newRefreshToken,
		"scope":         "openid offline_access profile email",
	})
}

// handleJWKS handles GET /keys
// Returns the public key in JWKS format
func (s *server) handleJWKS(w http.ResponseWriter, _ *http.Request) {
	pub := &s.privateKey.PublicKey

	jwks := map[string]interface{}{
		"keys": []map[string]interface{}{
			{
				"kty": "RSA",
				"use": "sig",
				"kid": s.keyID,
				"alg": "RS256",
				"n":   base64.RawURLEncoding.EncodeToString(pub.N.Bytes()),
				"e":   base64.RawURLEncoding.EncodeToString(big.NewInt(int64(pub.E)).Bytes()),
			},
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "public, max-age=3600")
	json.NewEncoder(w).Encode(jwks)
}

// handleRevoke handles POST /oauth/revoke (no-op for mock)
func (s *server) handleRevoke(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}

// generateAccessToken creates a signed JWT
func (s *server) generateAccessToken(subject string, claims map[string]interface{}, expiresIn int) (string, error) {
	now := time.Now()

	jwtClaims := jwt.MapClaims{
		"iss": s.issuerURL,
		"sub": subject,
		"iat": now.Unix(),
		"exp": now.Add(time.Duration(expiresIn) * time.Second).Unix(),
		"aud": "meshpump-spa",
	}

	// Merge user claims
	for k, v := range claims {
		jwtClaims[k] = v
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwtClaims)
	token.Header["kid"] = s.keyID

	return token.SignedString(s.privateKey)
}

func generateRandomString(n int) string {
	b := make([]byte, n)
	rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)
}
