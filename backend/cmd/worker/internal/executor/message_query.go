package executor

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/shadowapi/shadowapi/backend/pkg/jobs"
)

// handleMessageQuery queries messages from external PostgreSQL and streams them as individual records.
// It queries all configured tables and combines the fields from each table into a single JSON record.
func (e *Executor) handleMessageQuery(ctx context.Context, data []byte, sendRecord RecordSender) ([]byte, error) {
	var args jobs.MessageQueryJobArgs
	if err := json.Unmarshal(data, &args); err != nil {
		return nil, fmt.Errorf("failed to unmarshal job args: %w", err)
	}

	// Set defaults
	if args.Limit <= 0 {
		args.Limit = 100
	}

	// Extract table names for logging
	tableNames := make([]string, len(args.Tables))
	for i, t := range args.Tables {
		tableNames[i] = t.Name
	}

	e.log.Info("starting message query job",
		"job_uuid", args.JobUUID,
		"workspace_slug", args.WorkspaceSlug,
		"limit", args.Limit,
		"tables", tableNames,
	)

	start := time.Now()
	result := jobs.MessageQueryResult{
		JobUUID:       args.JobUUID,
		WorkspaceSlug: args.WorkspaceSlug,
		TablesQueried: tableNames,
		StartedAt:     start,
	}

	if len(args.Tables) == 0 {
		result.Success = false
		result.Error = "no tables configured"
		result.CompletedAt = time.Now()
		result.DurationMs = time.Since(start).Milliseconds()
		e.log.Error("no tables configured")
		return json.Marshal(result)
	}

	// Connect to external PostgreSQL
	pool, err := e.connectMessageQueryPostgres(ctx, args)
	if err != nil {
		result.Success = false
		result.Error = fmt.Sprintf("PostgreSQL connection failed: %v", err)
		result.CompletedAt = time.Now()
		result.DurationMs = time.Since(start).Milliseconds()
		e.log.Error("PostgreSQL connection failed", "error", err)
		return json.Marshal(result)
	}
	defer pool.Close()

	var sequence int64 = 0

	// Query each configured table
	for _, table := range args.Tables {
		e.log.Debug("querying table", "table", table.Name, "fields_count", len(table.Fields))

		// Build field list for SELECT - use configured fields only
		fieldNames := make([]string, len(table.Fields))
		for i, f := range table.Fields {
			fieldNames[i] = fmt.Sprintf(`"%s"`, f.Name)
		}

		// If no fields configured, select all
		selectFields := "*"
		if len(fieldNames) > 0 {
			selectFields = fmt.Sprintf("%s", joinStrings(fieldNames, ", "))
		}

		query := fmt.Sprintf(
			`SELECT %s FROM "%s" LIMIT $1 OFFSET $2`,
			selectFields,
			table.Name,
		)

		rows, err := pool.Query(ctx, query, args.Limit, args.Offset)
		if err != nil {
			e.log.Warn("query failed for table", "table", table.Name, "error", err)
			result.ErrorCount++
			continue
		}

		// Get column descriptions
		fieldDescs := rows.FieldDescriptions()
		columns := make([]string, len(fieldDescs))
		for i, fd := range fieldDescs {
			columns[i] = string(fd.Name)
		}

		// Stream each row as a DataRecord
		for rows.Next() {
			values, err := rows.Values()
			if err != nil {
				e.log.Warn("failed to scan row", "table", table.Name, "error", err)
				result.ErrorCount++
				continue
			}

			// Convert row to map with table name prefix for clarity
			record := make(map[string]any)
			record["_table"] = table.Name // Include source table name
			for i, col := range columns {
				record[col] = values[i]
			}

			// Serialize record to JSON
			recordJSON, err := json.Marshal(record)
			if err != nil {
				e.log.Warn("failed to marshal record", "table", table.Name, "error", err)
				result.ErrorCount++
				continue
			}

			result.MessagesQueried++

			// Send as DataRecord
			if err := sendRecord(DataRecord{
				Subject:  args.NATSSubject,
				Data:     recordJSON,
				Sequence: sequence,
			}); err != nil {
				e.log.Warn("failed to send record", "sequence", sequence, "error", err)
				result.ErrorCount++
				continue
			}

			result.MessagesPublished++
			sequence++

			e.log.Debug("streamed message record",
				"job_uuid", args.JobUUID,
				"table", table.Name,
				"sequence", sequence,
			)
		}

		if err := rows.Err(); err != nil {
			e.log.Warn("rows iteration error", "table", table.Name, "error", err)
		}
		rows.Close()
	}

	result.Success = result.ErrorCount == 0 || result.MessagesPublished > 0
	result.CompletedAt = time.Now()
	result.DurationMs = time.Since(start).Milliseconds()

	e.log.Info("message query job completed",
		"job_uuid", args.JobUUID,
		"tables", tableNames,
		"queried", result.MessagesQueried,
		"published", result.MessagesPublished,
		"errors", result.ErrorCount,
		"duration_ms", result.DurationMs,
	)

	return json.Marshal(result)
}

// joinStrings joins a slice of strings with a separator
func joinStrings(strs []string, sep string) string {
	if len(strs) == 0 {
		return ""
	}
	result := strs[0]
	for i := 1; i < len(strs); i++ {
		result += sep + strs[i]
	}
	return result
}

// connectMessageQueryPostgres creates a connection pool to external PostgreSQL for message queries
func (e *Executor) connectMessageQueryPostgres(ctx context.Context, args jobs.MessageQueryJobArgs) (*pgxpool.Pool, error) {
	database := args.StorageDatabase
	if database == "" {
		database = "postgres"
	}

	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s",
		url.QueryEscape(args.StorageUser),
		url.QueryEscape(args.StoragePassword),
		args.StorageHost,
		args.StoragePort,
		database,
	)
	if args.StorageOptions != "" {
		connStr += "?" + args.StorageOptions
	}

	config, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		return nil, fmt.Errorf("invalid connection config: %w", err)
	}

	// Use minimal pool for query job
	config.MaxConns = 2
	config.MinConns = 1

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("connection failed: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping failed: %w", err)
	}

	return pool, nil
}

// Ensure pgx types are used
var _ pgx.Rows = (pgx.Rows)(nil)
