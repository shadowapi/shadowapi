package executor

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/emersion/go-imap/client"

	"github.com/shadowapi/shadowapi/backend/pkg/jobs"
)

// xoauth2Auth implements the XOAUTH2 authentication mechanism for IMAP.
type xoauth2Auth struct {
	Username string
	Token    string
}

// Start creates the initial response for the XOAUTH2 authentication.
func (a *xoauth2Auth) Start() (string, []byte, error) {
	s := fmt.Sprintf("user=%s\x01auth=Bearer %s\x01\x01", a.Username, a.Token)
	return "XOAUTH2", []byte(s), nil
}

// Next is not needed for XOAUTH2.
func (a *xoauth2Auth) Next(challenge []byte) ([]byte, error) {
	return nil, nil
}

// handleTestConnectionEmailOAuth tests IMAP connectivity using OAuth2 credentials.
func (e *Executor) handleTestConnectionEmailOAuth(ctx context.Context, data []byte) ([]byte, error) {
	var args jobs.TestConnectionEmailOAuthJobArgs
	if err := json.Unmarshal(data, &args); err != nil {
		return nil, fmt.Errorf("failed to unmarshal job args: %w", err)
	}

	e.log.Debug("testing email OAuth connection",
		"datasource_uuid", args.DatasourceUUID,
		"email", args.Email,
		"provider", args.Provider,
	)

	start := time.Now()
	result := jobs.TestConnectionResult{
		JobUUID:      args.JobUUID,
		ResourceType: "email_oauth",
		ResourceUUID: args.DatasourceUUID,
		TestedAt:     start,
	}

	// Resolve IMAP server from provider
	host, port := resolveIMAPServer(args.Provider, args.IMAPHost, args.IMAPPort)

	// Test IMAP connection
	err := e.testIMAPConnection(ctx, host, port, args.Email, args.AccessToken)

	result.DurationMs = time.Since(start).Milliseconds()

	if err != nil {
		result.Success = false
		result.ErrorCode, result.ErrorMessage = categorizeIMAPError(err)
		result.ErrorDetails = err.Error()
		e.log.Error("email OAuth connection test failed",
			"datasource_uuid", args.DatasourceUUID,
			"error", err,
		)
	} else {
		result.Success = true
		result.Details = map[string]any{
			"imap_host": host,
			"imap_port": port,
		}
		e.log.Info("email OAuth connection test succeeded",
			"datasource_uuid", args.DatasourceUUID,
		)
	}

	return json.Marshal(result)
}

// resolveIMAPServer returns IMAP host and port based on provider.
func resolveIMAPServer(provider, customHost string, customPort int) (string, int) {
	if customHost != "" && customPort > 0 {
		return customHost, customPort
	}

	switch strings.ToLower(provider) {
	case "google", "gmail":
		return "imap.gmail.com", 993
	default:
		// Fallback to Gmail if provider not recognized
		return "imap.gmail.com", 993
	}
}

// testIMAPConnection performs the actual IMAP connectivity test.
func (e *Executor) testIMAPConnection(ctx context.Context, host string, port int, email, accessToken string) error {
	addr := fmt.Sprintf("%s:%d", host, port)

	// Connect with TLS
	c, err := client.DialTLS(addr, &tls.Config{
		ServerName: host,
	})
	if err != nil {
		return fmt.Errorf("connection failed: %w", err)
	}
	defer c.Logout()

	// Authenticate with XOAUTH2
	auth := &xoauth2Auth{
		Username: email,
		Token:    accessToken,
	}
	if err := c.Authenticate(auth); err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}

	// Select INBOX to verify access (read-only)
	_, err = c.Select("INBOX", true)
	if err != nil {
		return fmt.Errorf("mailbox access failed: %w", err)
	}

	return nil
}

// categorizeIMAPError converts IMAP errors to error codes.
func categorizeIMAPError(err error) (code string, message string) {
	errStr := strings.ToLower(err.Error())

	switch {
	case strings.Contains(errStr, "authentication failed") || strings.Contains(errStr, "invalid credentials"):
		return jobs.ErrorCodeAuthFailed, "IMAP authentication failed - check OAuth2 token"
	case strings.Contains(errStr, "connection refused"):
		return jobs.ErrorCodeConnectionRefused, "Connection refused by IMAP server"
	case strings.Contains(errStr, "timeout") || strings.Contains(errStr, "deadline exceeded"):
		return jobs.ErrorCodeConnectionTimeout, "Connection timed out"
	case strings.Contains(errStr, "no such host") || strings.Contains(errStr, "lookup"):
		return jobs.ErrorCodeDNSFailure, "DNS resolution failed - check host name"
	case strings.Contains(errStr, "certificate") || strings.Contains(errStr, "tls"):
		return jobs.ErrorCodeSSLRequired, "TLS/SSL connection error"
	case strings.Contains(errStr, "network is unreachable"):
		return jobs.ErrorCodeHostUnreachable, "Host is unreachable"
	default:
		return jobs.ErrorCodeUnknown, "IMAP connection test failed"
	}
}
