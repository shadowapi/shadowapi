package executor

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/mail"
	"net/url"
	"strings"
	"time"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/shadowapi/shadowapi/backend/pkg/jobs"
)

// handleEmailFetch fetches emails via IMAP and stores them in external PostgreSQL
func (e *Executor) handleEmailFetch(ctx context.Context, data []byte) ([]byte, error) {
	var args jobs.EmailFetchJobArgs
	if err := json.Unmarshal(data, &args); err != nil {
		return nil, fmt.Errorf("failed to unmarshal job args: %w", err)
	}

	e.log.Info("starting email fetch job",
		"job_uuid", args.JobUUID,
		"pipeline_uuid", args.PipelineUUID,
		"email", args.Email,
		"batch_size", args.BatchSize,
		"last_uid", args.LastUID,
	)

	start := time.Now()
	result := jobs.EmailFetchResult{
		JobUUID:       args.JobUUID,
		SchedulerUUID: args.SchedulerUUID,
		PipelineUUID:  args.PipelineUUID,
		StartedAt:     start,
	}

	// Connect to IMAP
	imapClient, err := e.connectEmailIMAP(ctx, args)
	if err != nil {
		result.Success = false
		result.Error = fmt.Sprintf("IMAP connection failed: %v", err)
		result.CompletedAt = time.Now()
		result.DurationMs = time.Since(start).Milliseconds()
		e.log.Error("IMAP connection failed", "error", err)
		return json.Marshal(result)
	}
	defer imapClient.Logout()

	// Select mailbox
	mbox, err := imapClient.Select(args.MailboxName, true)
	if err != nil {
		result.Success = false
		result.Error = fmt.Sprintf("failed to select mailbox %s: %v", args.MailboxName, err)
		result.CompletedAt = time.Now()
		result.DurationMs = time.Since(start).Milliseconds()
		e.log.Error("mailbox selection failed", "mailbox", args.MailboxName, "error", err)
		return json.Marshal(result)
	}

	e.log.Debug("mailbox selected", "mailbox", args.MailboxName, "messages", mbox.Messages)

	// Get all message UIDs and filter to those > LastUID
	uids, err := imapClient.UidSearch(imap.NewSearchCriteria())
	if err != nil {
		result.Success = false
		result.Error = fmt.Sprintf("IMAP UID search failed: %v", err)
		result.CompletedAt = time.Now()
		result.DurationMs = time.Since(start).Milliseconds()
		e.log.Error("IMAP UID search failed", "error", err)
		return json.Marshal(result)
	}

	// Filter UIDs to only those > LastUID
	var filteredUIDs []uint32
	for _, uid := range uids {
		if uid > args.LastUID {
			filteredUIDs = append(filteredUIDs, uid)
		}
	}

	if len(filteredUIDs) == 0 {
		e.log.Info("no new messages found", "job_uuid", args.JobUUID, "last_uid", args.LastUID)
		result.Success = true
		result.MessagesFetched = 0
		result.MessagesStored = 0
		result.LastUID = args.LastUID // Keep the same UID
		result.CompletedAt = time.Now()
		result.DurationMs = time.Since(start).Milliseconds()
		return json.Marshal(result)
	}

	e.log.Info("found new messages to fetch", "count", len(filteredUIDs), "batch_size", args.BatchSize, "first_uid", filteredUIDs[0])

	// Limit to batch size
	if args.BatchSize > 0 && len(filteredUIDs) > args.BatchSize {
		filteredUIDs = filteredUIDs[:args.BatchSize]
	}
	result.MessagesFetched = len(filteredUIDs)

	// Connect to external PostgreSQL
	pool, err := e.connectEmailPostgres(ctx, args)
	if err != nil {
		result.Success = false
		result.Error = fmt.Sprintf("PostgreSQL connection failed: %v", err)
		result.CompletedAt = time.Now()
		result.DurationMs = time.Since(start).Milliseconds()
		e.log.Error("PostgreSQL connection failed", "error", err)
		return json.Marshal(result)
	}
	defer pool.Close()

	// Create mapper executor
	mapperExec := newJobMapperExecutor(args.MapperConfig)

	// Fetch and process messages using UID FETCH
	fetchSeqSet := new(imap.SeqSet)
	fetchSeqSet.AddNum(filteredUIDs...)

	section := &imap.BodySectionName{}
	items := []imap.FetchItem{section.FetchItem(), imap.FetchEnvelope, imap.FetchFlags, imap.FetchUid}

	messages := make(chan *imap.Message, 10)
	done := make(chan error, 1)
	go func() {
		done <- imapClient.UidFetch(fetchSeqSet, items, messages)
	}()

	var highestUID uint32 = args.LastUID
	for msg := range messages {
		// Track highest UID processed
		if msg.Uid > highestUID {
			highestUID = msg.Uid
		}

		// Convert IMAP message to source data map
		sourceData := e.imapToSourceData(msg, section)

		// Apply mapper transformations
		tableData, err := mapperExec.Execute(sourceData)
		if err != nil {
			e.log.Warn("mapper execution failed",
				"job_uuid", args.JobUUID,
				"uid", msg.Uid,
				"error", err,
			)
			result.ErrorCount++
			continue
		}

		// Insert into PostgreSQL for each target table
		for tableName, fields := range tableData {
			if err := e.insertEmailRow(ctx, pool, tableName, fields); err != nil {
				e.log.Warn("failed to insert row",
					"job_uuid", args.JobUUID,
					"table", tableName,
					"uid", msg.Uid,
					"error", err,
				)
				result.ErrorCount++
				continue
			}
			result.MessagesStored++
		}
	}

	if err := <-done; err != nil {
		e.log.Error("IMAP fetch error", "error", err)
	}

	result.Success = result.ErrorCount < result.MessagesFetched
	result.LastUID = highestUID
	result.CompletedAt = time.Now()
	result.DurationMs = time.Since(start).Milliseconds()

	e.log.Info("email fetch job completed",
		"job_uuid", args.JobUUID,
		"fetched", result.MessagesFetched,
		"stored", result.MessagesStored,
		"errors", result.ErrorCount,
		"last_uid", result.LastUID,
		"duration_ms", result.DurationMs,
	)

	return json.Marshal(result)
}

// connectEmailIMAP establishes IMAP connection with XOAUTH2
func (e *Executor) connectEmailIMAP(ctx context.Context, args jobs.EmailFetchJobArgs) (*client.Client, error) {
	addr := fmt.Sprintf("%s:%d", args.IMAPHost, args.IMAPPort)

	c, err := client.DialTLS(addr, &tls.Config{
		ServerName: args.IMAPHost,
	})
	if err != nil {
		return nil, fmt.Errorf("TLS connection failed: %w", err)
	}

	auth := &xoauth2Auth{
		Username: args.Email,
		Token:    args.AccessToken,
	}
	if err := c.Authenticate(auth); err != nil {
		c.Logout()
		return nil, fmt.Errorf("authentication failed: %w", err)
	}

	return c, nil
}

// connectEmailPostgres creates a connection pool to external PostgreSQL
func (e *Executor) connectEmailPostgres(ctx context.Context, args jobs.EmailFetchJobArgs) (*pgxpool.Pool, error) {
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

	// Use minimal pool for this job
	config.MaxConns = 5
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

// imapToSourceData extracts email data into a flat map for mapper processing
func (e *Executor) imapToSourceData(msg *imap.Message, section *imap.BodySectionName) map[string]any {
	data := make(map[string]any)

	// Set default values
	data["message.type"] = "email"
	data["message.format"] = "text"

	if msg.Envelope != nil {
		data["message.subject"] = msg.Envelope.Subject

		if len(msg.Envelope.From) > 0 {
			data["message.sender"] = msg.Envelope.From[0].Address()
		}

		var recipients []string
		for _, addr := range msg.Envelope.To {
			recipients = append(recipients, addr.Address())
		}
		for _, addr := range msg.Envelope.Cc {
			recipients = append(recipients, addr.Address())
		}
		if len(recipients) > 0 {
			data["message.recipients"] = strings.Join(recipients, ",")
		}

		if !msg.Envelope.Date.IsZero() {
			data["message.created_at"] = msg.Envelope.Date
		}

		data["message.external_message_id"] = msg.Envelope.MessageId

		// Reply-to
		if len(msg.Envelope.ReplyTo) > 0 {
			data["message.reply_to"] = msg.Envelope.ReplyTo[0].Address()
		}

		// In-Reply-To header
		if msg.Envelope.InReplyTo != "" {
			data["message.in_reply_to"] = msg.Envelope.InReplyTo
		}
	}

	// Read body
	if body := msg.GetBody(section); body != nil {
		if parsed, err := mail.ReadMessage(body); err == nil {
			bodyBytes, _ := io.ReadAll(parsed.Body)
			data["message.body"] = string(bodyBytes)

			// Extract additional headers
			if from := parsed.Header.Get("From"); from != "" {
				data["message.header_from"] = from
			}
			if to := parsed.Header.Get("To"); to != "" {
				data["message.header_to"] = to
			}
			if cc := parsed.Header.Get("Cc"); cc != "" {
				data["message.header_cc"] = cc
			}
			if contentType := parsed.Header.Get("Content-Type"); contentType != "" {
				data["message.content_type"] = contentType
				if strings.Contains(strings.ToLower(contentType), "html") {
					data["message.format"] = "html"
				}
			}
		}
	}

	return data
}

// insertEmailRow inserts a row into the target PostgreSQL table
func (e *Executor) insertEmailRow(ctx context.Context, pool *pgxpool.Pool, tableName string, fields map[string]any) error {
	if len(fields) == 0 {
		return nil
	}

	var columns []string
	var placeholders []string
	var values []any
	i := 1
	for col, val := range fields {
		columns = append(columns, fmt.Sprintf(`"%s"`, col))
		placeholders = append(placeholders, fmt.Sprintf("$%d", i))
		values = append(values, val)
		i++
	}

	query := fmt.Sprintf(
		`INSERT INTO "%s" (%s) VALUES (%s) ON CONFLICT DO NOTHING`,
		tableName,
		strings.Join(columns, ", "),
		strings.Join(placeholders, ", "),
	)

	_, err := pool.Exec(ctx, query, values...)
	return err
}

// jobMapperExecutor applies mapper configuration to source data
type jobMapperExecutor struct {
	config jobs.MapperConfigData
}

func newJobMapperExecutor(config jobs.MapperConfigData) *jobMapperExecutor {
	return &jobMapperExecutor{config: config}
}

// Execute applies the mapper config to source data and returns mapped values by table
func (e *jobMapperExecutor) Execute(sourceData map[string]any) (map[string]map[string]any, error) {
	result := make(map[string]map[string]any)

	for _, mapping := range e.config.Mappings {
		if !mapping.IsEnabled {
			continue
		}

		// Build the field key from source entity and field
		fieldKey := mapping.SourceEntity + "." + mapping.SourceField
		value, exists := sourceData[fieldKey]
		if !exists {
			continue
		}

		// Apply transform if specified
		if mapping.Transform != nil {
			value = e.applyTransform(value, mapping.Transform)
		}

		// Store in result by table
		if result[mapping.TargetTable] == nil {
			result[mapping.TargetTable] = make(map[string]any)
		}
		result[mapping.TargetTable][mapping.TargetField] = value
	}

	return result, nil
}

// applyTransform applies a transformation to a value
func (e *jobMapperExecutor) applyTransform(value any, transform *jobs.MapperTransformData) any {
	str, ok := value.(string)
	if !ok {
		// Try to convert to string for basic transforms
		switch v := value.(type) {
		case time.Time:
			str = v.Format(time.RFC3339)
			ok = true
		case int, int64, float64:
			str = fmt.Sprintf("%v", v)
			ok = true
		}
	}

	if !ok {
		return value
	}

	switch transform.Type {
	case "lowercase":
		return strings.ToLower(str)
	case "uppercase":
		return strings.ToUpper(str)
	case "trim":
		return strings.TrimSpace(str)
	case "set":
		// Return static value from params
		if val, ok := transform.Params["value"]; ok {
			return val
		}
		return value
	default:
		return value
	}
}
