package executor

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/shadowapi/shadowapi/backend/pkg/jobs"
)

// handleTestConnectionPostgres tests PostgreSQL database connectivity.
func (e *Executor) handleTestConnectionPostgres(ctx context.Context, data []byte) ([]byte, error) {
	var args jobs.TestConnectionPostgresJobArgs
	if err := json.Unmarshal(data, &args); err != nil {
		return nil, fmt.Errorf("failed to unmarshal job args: %w", err)
	}

	e.log.Debug("testing PostgreSQL connection",
		"storage_uuid", args.StorageUUID,
		"host", args.Host,
		"port", args.Port,
	)

	start := time.Now()
	result := jobs.TestConnectionResult{
		JobUUID:      args.JobUUID,
		ResourceType: "postgres",
		ResourceUUID: args.StorageUUID,
		TestedAt:     start,
	}

	// Build connection string
	connStr := buildPostgresConnString(args)

	// Test connection with timeout
	testCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	err := e.testPostgresConnection(testCtx, connStr)

	result.DurationMs = time.Since(start).Milliseconds()

	if err != nil {
		result.Success = false
		result.ErrorCode, result.ErrorMessage = categorizePostgresError(err)
		result.ErrorDetails = err.Error()
		e.log.Error("PostgreSQL connection test failed",
			"storage_uuid", args.StorageUUID,
			"error", err,
		)
	} else {
		result.Success = true
		result.Details = map[string]any{
			"host": args.Host,
			"port": args.Port,
		}
		e.log.Info("PostgreSQL connection test succeeded",
			"storage_uuid", args.StorageUUID,
		)
	}

	return json.Marshal(result)
}

// buildPostgresConnString constructs a PostgreSQL connection URI from job args.
func buildPostgresConnString(args jobs.TestConnectionPostgresJobArgs) string {
	database := args.Database
	if database == "" {
		database = "postgres"
	}

	// Format: postgres://user:password@host:port/database?options
	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s",
		url.QueryEscape(args.User),
		url.QueryEscape(args.Password),
		args.Host,
		args.Port,
		database,
	)

	if args.Options != "" {
		connStr += "?" + args.Options
	}

	return connStr
}

// testPostgresConnection attempts to connect and ping the database.
func (e *Executor) testPostgresConnection(ctx context.Context, connStr string) error {
	config, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		return fmt.Errorf("invalid connection config: %w", err)
	}

	// Use minimal pool for testing
	config.MaxConns = 1
	config.MinConns = 0

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return fmt.Errorf("connection failed: %w", err)
	}
	defer pool.Close()

	// Ping to verify connection
	if err := pool.Ping(ctx); err != nil {
		return fmt.Errorf("ping failed: %w", err)
	}

	return nil
}

// categorizePostgresError converts PostgreSQL errors to error codes.
func categorizePostgresError(err error) (code string, message string) {
	errStr := strings.ToLower(err.Error())

	switch {
	case strings.Contains(errStr, "password authentication failed"):
		return jobs.ErrorCodeInvalidCredentials, "Invalid username or password"
	case strings.Contains(errStr, "connection refused"):
		return jobs.ErrorCodeConnectionRefused, "Connection refused - check host and port"
	case strings.Contains(errStr, "timeout") || strings.Contains(errStr, "deadline exceeded"):
		return jobs.ErrorCodeConnectionTimeout, "Connection timed out"
	case strings.Contains(errStr, "no such host") || strings.Contains(errStr, "lookup"):
		return jobs.ErrorCodeDNSFailure, "Host not found - check hostname"
	case strings.Contains(errStr, "network is unreachable"):
		return jobs.ErrorCodeHostUnreachable, "Host is unreachable"
	case strings.Contains(errStr, "ssl") || strings.Contains(errStr, "tls"):
		return jobs.ErrorCodeSSLRequired, "SSL/TLS configuration error"
	default:
		return jobs.ErrorCodeUnknown, "PostgreSQL connection test failed"
	}
}
