package zitadel

import (
	"bytes"
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/samber/do/v2"

	"github.com/shadowapi/shadowapi/backend/internal/auth"
	"github.com/shadowapi/shadowapi/backend/internal/config"
	"github.com/shadowapi/shadowapi/backend/internal/handler"
	"github.com/shadowapi/shadowapi/backend/pkg/api"
)

// ZitadelUserManager implements UserManager interface using Zitadel Management API
type ZitadelUserManager struct {
	cfg        *config.Config
	log        *slog.Logger
	httpClient *http.Client

	// JWT authentication fields
	privateKey *rsa.PrivateKey
	keyID      string

	// Token caching
	tokenMutex  sync.RWMutex
	accessToken string
	tokenExpiry time.Time
}

// Provide creates a new ZitadelUserManager instance
func Provide(i do.Injector) (auth.UserManager, error) {
	cfg := do.MustInvoke[*config.Config](i)
	log := do.MustInvoke[*slog.Logger](i)

	zm := &ZitadelUserManager{
		cfg:        cfg,
		log:        log,
		httpClient: &http.Client{},
	}

	// Debug configuration
	log.Debug("DEBUG: Zitadel Manager initialized", "management_url", cfg.Auth.Zitadel.ManagementURL, "instance_url", cfg.Auth.Zitadel.InstanceURL)

	// Load private key for JWT authentication
	if err := zm.loadPrivateKey(); err != nil {
		return nil, handler.E("failed to load private key: %w", err)
	}

	return zm, nil
}

// ZitadelUser represents a user in Zitadel API format
type ZitadelUser struct {
	ID       string `json:"id,omitempty"`
	Username string `json:"userName,omitempty"`
	Profile  struct {
		FirstName   string `json:"firstName,omitempty"`
		LastName    string `json:"lastName,omitempty"`
		DisplayName string `json:"displayName,omitempty"`
	} `json:"profile,omitempty"`
	Email struct {
		Email      string `json:"email,omitempty"`
		IsVerified bool   `json:"isVerified,omitempty"`
	} `json:"email,omitempty"`
	State string `json:"state,omitempty"`
}

// ZitadelCreateUserRequest represents request to create user in Zitadel
type ZitadelCreateUserRequest struct {
	Username string `json:"userName"`
	Profile  struct {
		FirstName   string `json:"firstName"`
		LastName    string `json:"lastName"`
		DisplayName string `json:"displayName"`
	} `json:"profile"`
	Email struct {
		Email      string `json:"email"`
		IsVerified bool   `json:"isVerified"`
	} `json:"email"`
	Password struct {
		Password       string `json:"password"`
		ChangeRequired bool   `json:"changeRequired"`
	} `json:"password"`
}

// ZitadelKeyFile represents the structure of Zitadel service user key file
type ZitadelKeyFile struct {
	Type   string `json:"type"`
	KeyID  string `json:"keyId"`
	Key    string `json:"key"`
	UserID string `json:"userId"`
}

// TokenResponse represents Zitadel OAuth token response
type TokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

// loadPrivateKey loads the private key from the configured key file
func (m *ZitadelUserManager) loadPrivateKey() error {
	keyPath := m.cfg.Auth.Zitadel.ServiceUserKeyPath
	m.log.Debug("loading Zitadel service user key", "path", keyPath)

	if keyPath == "" {
		m.log.Error("service user key path not configured")
		return handler.E("service user key path not configured")
	}

	// Read the key file
	keyData, err := os.ReadFile(keyPath)
	if err != nil {
		m.log.Error("failed to read key file", "error", err, "path", keyPath)
		return handler.E("failed to read key file: %w", err)
	}

	// Parse the JSON key file
	var keyFile ZitadelKeyFile
	if err := json.Unmarshal(keyData, &keyFile); err != nil {
		m.log.Error("failed to parse key file", "error", err)
		return handler.E("failed to parse key file: %w", err)
	}

	m.log.Debug("parsed key file", "keyId", keyFile.KeyID, "userId", keyFile.UserID, "type", keyFile.Type)

	// Parse the private key
	block, _ := pem.Decode([]byte(keyFile.Key))
	if block == nil {
		m.log.Error("failed to decode PEM block")
		return handler.E("failed to decode PEM block")
	}

	privateKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		// Try PKCS1 format
		privateKey, err = x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			m.log.Error("failed to parse private key", "error", err)
			return handler.E("failed to parse private key: %w", err)
		}
	}

	rsaKey, ok := privateKey.(*rsa.PrivateKey)
	if !ok {
		m.log.Error("private key is not RSA key")
		return handler.E("private key is not RSA key")
	}

	m.privateKey = rsaKey
	m.keyID = keyFile.KeyID

	m.log.Info("successfully loaded Zitadel service user key", "keyId", keyFile.KeyID)
	return nil
}

// createJWT creates a signed JWT for service user authentication
func (m *ZitadelUserManager) createJWT() (string, error) {
	now := time.Now().UTC()

	// Use the instance URL for audience
	audience := m.cfg.Auth.Zitadel.ManagementURL
	if m.cfg.Auth.Zitadel.InstanceURL != "" {
		audience = m.cfg.Auth.Zitadel.InstanceURL
	}
	audience = strings.TrimSuffix(audience, "/")

	claims := jwt.MapClaims{
		"iss": m.cfg.Auth.Zitadel.ServiceUserID,
		"sub": m.cfg.Auth.Zitadel.ServiceUserID,
		"aud": audience,
		"iat": now.Unix(),
		"exp": now.Add(5 * time.Minute).Unix(), // 5 minutes expiry
	}

	m.log.Debug("creating JWT",
		"iss", claims["iss"],
		"aud", claims["aud"],
		"keyId", m.keyID)

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = m.keyID

	signedToken, err := token.SignedString(m.privateKey)
	if err != nil {
		m.log.Error("failed to sign JWT", "error", err)
		return "", err
	}

	return signedToken, nil
}

// getAccessToken exchanges JWT for access token
func (m *ZitadelUserManager) getAccessToken(ctx context.Context) (string, int, error) {
	// Create JWT assertion
	assertion, err := m.createJWT()
	if err != nil {
		m.log.Error("failed to create JWT", "error", err)
		return "", 0, handler.E("failed to create JWT: %w", err)
	}

	// Prepare token request
	data := url.Values{
		"grant_type": {"urn:ietf:params:oauth:grant-type:jwt-bearer"},
		"assertion":  {assertion},
		"scope":      {"openid profile urn:zitadel:iam:org:project:id:zitadel:aud"},
	}

	// Always use ManagementURL for the actual request
	m.log.Debug("DEBUG: ManagementURL from config", "management_url", m.cfg.Auth.Zitadel.ManagementURL)
	tokenURL := strings.TrimSuffix(m.cfg.Auth.Zitadel.ManagementURL, "/") + "/oauth/v2/token"
	m.log.Debug("requesting access token", "url", tokenURL)

	req, err := http.NewRequestWithContext(ctx, "POST", tokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		m.log.Error("failed to create token request", "error", err)
		return "", 0, handler.E("failed to create token request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	// Set Host header to external domain if InstanceURL is configured
	if m.cfg.Auth.Zitadel.InstanceURL != "" {
		if u, err := url.Parse(m.cfg.Auth.Zitadel.InstanceURL); err == nil {
			req.Host = u.Host
		}
	}

	resp, err := m.httpClient.Do(req)
	if err != nil {
		m.log.Error("failed to make token request", "error", err, "url", tokenURL)
		return "", 0, handler.E("failed to make token request: %w", err)
	}
	defer resp.Body.Close()

	// Read response body for debugging
	var bodyBytes []byte
	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ = io.ReadAll(resp.Body)
		m.log.Error("token request failed",
			"status", resp.StatusCode,
			"response", string(bodyBytes),
			"url", tokenURL)
		return "", 0, handler.E("token request failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var tokenResp TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		m.log.Error("failed to decode token response", "error", err)
		return "", 0, handler.E("failed to decode token response: %w", err)
	}

	m.log.Debug("successfully obtained access token", "expires_in", tokenResp.ExpiresIn)
	return tokenResp.AccessToken, tokenResp.ExpiresIn, nil
}

// getAuthToken returns a cached or fresh access token for Zitadel Management API
func (m *ZitadelUserManager) getAuthToken(ctx context.Context) (string, error) {
	m.tokenMutex.RLock()
	// Check if we have a valid cached token (with 1 minute buffer)
	if m.accessToken != "" && time.Now().Add(time.Minute).Before(m.tokenExpiry) {
		token := m.accessToken
		m.tokenMutex.RUnlock()
		return token, nil
	}
	m.tokenMutex.RUnlock()

	// Need to get a new token
	m.tokenMutex.Lock()
	defer m.tokenMutex.Unlock()

	// Double-check in case another goroutine already refreshed the token
	if m.accessToken != "" && time.Now().Add(time.Minute).Before(m.tokenExpiry) {
		return m.accessToken, nil
	}

	// Get new access token
	token, expiresIn, err := m.getAccessToken(ctx)
	if err != nil {
		return "", err
	}

	// Cache the token
	m.accessToken = token
	m.tokenExpiry = time.Now().Add(time.Duration(expiresIn) * time.Second)

	return token, nil
}

// makeRequest makes an authenticated request to Zitadel Management API
func (m *ZitadelUserManager) makeRequest(ctx context.Context, method, path string, body interface{}) (*http.Response, error) {
	var reqBody *bytes.Buffer
	var bodyStr string
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			m.log.Error("failed to marshal request body", "error", err)
			return nil, handler.E("failed to marshal request body: %w", err)
		}
		bodyStr = string(jsonBody)
		reqBody = bytes.NewBuffer(jsonBody)
	} else {
		reqBody = bytes.NewBuffer(nil)
	}

	apiURL := m.cfg.Auth.Zitadel.ManagementURL + "/management/v1" + path
	m.log.Debug("making Zitadel API request",
		"method", method,
		"url", apiURL,
		"body", bodyStr)

	req, err := http.NewRequestWithContext(ctx, method, apiURL, reqBody)
	if err != nil {
		m.log.Error("failed to create request", "error", err, "url", apiURL)
		return nil, handler.E("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	// Set Host header to external domain if InstanceURL is configured
	if m.cfg.Auth.Zitadel.InstanceURL != "" {
		if u, err := url.Parse(m.cfg.Auth.Zitadel.InstanceURL); err == nil {
			req.Host = u.Host
		}
	}

	// Add authentication
	token, err := m.getAuthToken(ctx)
	if err != nil {
		m.log.Error("failed to get auth token", "error", err)
		return nil, handler.E("failed to get auth token: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)

	m.log.Debug("request headers", "headers", req.Header)

	resp, err := m.httpClient.Do(req)
	if err != nil {
		m.log.Error("HTTP request failed", "error", err, "url", apiURL)
		return nil, err
	}

	m.log.Debug("received response", "status", resp.StatusCode, "url", apiURL)
	return resp, nil
}

// convertZitadelToAPI converts Zitadel user format to api.User
func (m *ZitadelUserManager) convertZitadelToAPI(zu *ZitadelUser) *api.User {
	// Default to enabled=true for new users (Zitadel creates them as active by default)
	isEnabled := true
	if zu.State != "" && zu.State != "USER_STATE_ACTIVE" {
		isEnabled = false
	}

	return &api.User{
		UUID:      api.NewOptString(zu.ID),
		Email:     zu.Email.Email,
		FirstName: zu.Profile.FirstName,
		LastName:  zu.Profile.LastName,
		IsEnabled: api.NewOptBool(isEnabled),
		IsAdmin:   api.NewOptBool(false), // TODO: Determine from Zitadel roles
		Meta:      api.NewOptUserMeta(make(api.UserMeta)),
		// TODO: Add CreatedAt, UpdatedAt from Zitadel response
	}
}

// convertAPIToZitadel converts api.User to Zitadel create request format
func (m *ZitadelUserManager) convertAPIToZitadel(user *api.User) *ZitadelCreateUserRequest {
	req := &ZitadelCreateUserRequest{
		Username: fmt.Sprintf("user_%s", user.Email), // Use email-based username
	}

	req.Profile.FirstName = user.FirstName
	req.Profile.LastName = user.LastName
	req.Profile.DisplayName = fmt.Sprintf("%s %s", user.FirstName, user.LastName)

	req.Email.Email = user.Email
	req.Email.IsVerified = false

	req.Password.Password = user.Password
	req.Password.ChangeRequired = false

	return req
}

// CreateUser creates a new user via Zitadel Management API
func (m *ZitadelUserManager) CreateUser(ctx context.Context, user *api.User) (*api.User, error) {
	m.log.Info("creating user",
		"email", user.Email,
		"firstName", user.FirstName,
		"lastName", user.LastName)

	req := m.convertAPIToZitadel(user)
	m.log.Debug("converted user to Zitadel format", "request", req)

	resp, err := m.makeRequest(ctx, "POST", "/users/human", req)
	if err != nil {
		m.log.Error("failed to create user in Zitadel", "error", err)
		return nil, handler.ErrWithCode(http.StatusInternalServerError, handler.E("failed to create user in Zitadel: %w", err))
	}
	defer resp.Body.Close()

	// Read response body for debugging
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		m.log.Error("failed to read response body", "error", err)
		return nil, handler.ErrWithCode(http.StatusInternalServerError, handler.E("failed to read response: %w", err))
	}

	m.log.Debug("Zitadel API response",
		"status", resp.StatusCode,
		"body", string(bodyBytes))

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		m.log.Error("Zitadel API error",
			"status", resp.StatusCode,
			"response", string(bodyBytes))
		return nil, handler.ErrWithCode(resp.StatusCode, handler.E("Zitadel API returned status %d: %s", resp.StatusCode, string(bodyBytes)))
	}

	var zitadelUser ZitadelUser
	if err := json.Unmarshal(bodyBytes, &zitadelUser); err != nil {
		m.log.Error("failed to decode Zitadel response", "error", err, "body", string(bodyBytes))
		return nil, handler.ErrWithCode(http.StatusInternalServerError, handler.E("failed to decode Zitadel response: %w", err))
	}

	m.log.Info("user created successfully", "userId", zitadelUser.ID, "email", user.Email)
	return m.convertZitadelToAPI(&zitadelUser), nil
}

// GetUser retrieves a user by UUID via Zitadel Management API
func (m *ZitadelUserManager) GetUser(ctx context.Context, uuid string) (*api.User, error) {
	resp, err := m.makeRequest(ctx, "GET", "/users/"+uuid, nil)
	if err != nil {
		return nil, handler.ErrWithCode(http.StatusInternalServerError, handler.E("failed to get user from Zitadel: %w", err))
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, handler.ErrWithCode(http.StatusNotFound, handler.E("user not found"))
	}

	if resp.StatusCode != http.StatusOK {
		return nil, handler.ErrWithCode(resp.StatusCode, handler.E("Zitadel API returned status %d", resp.StatusCode))
	}

	var zitadelUser ZitadelUser
	if err := json.NewDecoder(resp.Body).Decode(&zitadelUser); err != nil {
		return nil, handler.ErrWithCode(http.StatusInternalServerError, handler.E("failed to decode Zitadel response: %w", err))
	}

	return m.convertZitadelToAPI(&zitadelUser), nil
}

// UpdateUser updates an existing user via Zitadel Management API
func (m *ZitadelUserManager) UpdateUser(ctx context.Context, user *api.User, uuid string) (*api.User, error) {
	// TODO: Implement Zitadel user update
	// 1. Use service user authentication
	// 2. Call Zitadel Management API /v1/users/{user_id}
	// 3. Convert response to api.User format
	return nil, &NotImplementedError{Operation: "UpdateUser"}
}

// DeleteUser deletes a user by UUID via Zitadel Management API
func (m *ZitadelUserManager) DeleteUser(ctx context.Context, uuid string) error {
	// TODO: Implement Zitadel user deletion
	// 1. Use service user authentication
	// 2. Call Zitadel Management API DELETE /v1/users/{user_id}
	return &NotImplementedError{Operation: "DeleteUser"}
}

// ListUsers returns a list of all users via Zitadel Management API
func (m *ZitadelUserManager) ListUsers(ctx context.Context) ([]api.User, error) {
	// TODO: Implement Zitadel user listing
	// 1. Use service user authentication
	// 2. Call Zitadel Management API /v1/users
	// 3. Convert response to []api.User format
	// 4. Handle pagination
	return nil, &NotImplementedError{Operation: "ListUsers"}
}

// NotImplementedError represents an operation that is not yet implemented
type NotImplementedError struct {
	Operation string
}

func (e *NotImplementedError) Error() string {
	return "zitadel." + e.Operation + " is not yet implemented"
}

