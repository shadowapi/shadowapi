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
	"unicode/utf8"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/shadowapi/shadowapi/backend/pkg/jobs"
	"github.com/shadowapi/shadowapi/backend/pkg/transforms"
)

// parsedBody caches the result of parsing the email body to avoid re-parsing.
type parsedBody struct {
	parsed      bool
	body        string
	headers     map[string]string
	contentType string
}

// ensureParsed lazily parses the email body on first access.
func (pb *parsedBody) ensureParsed(msg *imap.Message, section *imap.BodySectionName) {
	if pb.parsed {
		return
	}
	pb.parsed = true
	pb.headers = make(map[string]string)

	if body := msg.GetBody(section); body != nil {
		if parsed, err := mail.ReadMessage(body); err == nil {
			bodyBytes, _ := io.ReadAll(parsed.Body)
			pb.body = string(bodyBytes)

			// Cache headers
			pb.headers["From"] = parsed.Header.Get("From")
			pb.headers["To"] = parsed.Header.Get("To")
			pb.headers["Cc"] = parsed.Header.Get("Cc")
			pb.contentType = parsed.Header.Get("Content-Type")
		}
	}
}

// fieldExtractor extracts a specific field from an IMAP message.
type fieldExtractor func(msg *imap.Message, section *imap.BodySectionName, pb *parsedBody) any

// fieldExtractors maps source field keys to their extraction functions.
var fieldExtractors = map[string]fieldExtractor{
	"message.uuid": func(msg *imap.Message, section *imap.BodySectionName, pb *parsedBody) any {
		return uuid.Must(uuid.NewV7()).String()
	},
	"message.type": func(msg *imap.Message, section *imap.BodySectionName, pb *parsedBody) any {
		return "email"
	},
	"message.format": func(msg *imap.Message, section *imap.BodySectionName, pb *parsedBody) any {
		pb.ensureParsed(msg, section)
		if pb.contentType != "" && strings.Contains(strings.ToLower(pb.contentType), "html") {
			return "html"
		}
		return "text"
	},
	"message.subject": func(msg *imap.Message, section *imap.BodySectionName, pb *parsedBody) any {
		if msg.Envelope != nil {
			return msg.Envelope.Subject
		}
		return ""
	},
	"message.sender": func(msg *imap.Message, section *imap.BodySectionName, pb *parsedBody) any {
		if msg.Envelope != nil && len(msg.Envelope.From) > 0 {
			return msg.Envelope.From[0].Address()
		}
		return ""
	},
	"message.recipients": func(msg *imap.Message, section *imap.BodySectionName, pb *parsedBody) any {
		if msg.Envelope == nil {
			return ""
		}
		var recipients []string
		for _, addr := range msg.Envelope.To {
			recipients = append(recipients, addr.Address())
		}
		for _, addr := range msg.Envelope.Cc {
			recipients = append(recipients, addr.Address())
		}
		return strings.Join(recipients, ",")
	},
	"message.body": func(msg *imap.Message, section *imap.BodySectionName, pb *parsedBody) any {
		pb.ensureParsed(msg, section)
		return pb.body
	},
	"message.created_at": func(msg *imap.Message, section *imap.BodySectionName, pb *parsedBody) any {
		if msg.Envelope != nil && !msg.Envelope.Date.IsZero() {
			return msg.Envelope.Date
		}
		return time.Time{}
	},
	"message.external_message_id": func(msg *imap.Message, section *imap.BodySectionName, pb *parsedBody) any {
		if msg.Envelope != nil {
			return msg.Envelope.MessageId
		}
		return ""
	},
	"message.reply_to": func(msg *imap.Message, section *imap.BodySectionName, pb *parsedBody) any {
		if msg.Envelope != nil && len(msg.Envelope.ReplyTo) > 0 {
			return msg.Envelope.ReplyTo[0].Address()
		}
		return ""
	},
	"message.in_reply_to": func(msg *imap.Message, section *imap.BodySectionName, pb *parsedBody) any {
		if msg.Envelope != nil {
			return msg.Envelope.InReplyTo
		}
		return ""
	},
	"message.header_from": func(msg *imap.Message, section *imap.BodySectionName, pb *parsedBody) any {
		pb.ensureParsed(msg, section)
		return pb.headers["From"]
	},
	"message.header_to": func(msg *imap.Message, section *imap.BodySectionName, pb *parsedBody) any {
		pb.ensureParsed(msg, section)
		return pb.headers["To"]
	},
	"message.header_cc": func(msg *imap.Message, section *imap.BodySectionName, pb *parsedBody) any {
		pb.ensureParsed(msg, section)
		return pb.headers["Cc"]
	},
	"message.content_type": func(msg *imap.Message, section *imap.BodySectionName, pb *parsedBody) any {
		pb.ensureParsed(msg, section)
		return pb.contentType
	},
}

// collectRequiredFields pre-computes which fields need extraction based on mapper config.
func collectRequiredFields(config jobs.MapperConfigData) map[string]bool {
	required := make(map[string]bool)

	for _, m := range config.Mappings {
		if !m.IsEnabled {
			continue
		}

		// Add the primary source field
		fieldKey := m.SourceEntity + "." + m.SourceField
		required[fieldKey] = true

		// Check for concat transform which may reference other fields
		if m.Transform != nil && m.Transform.Type == "concat" {
			if parts, ok := transforms.GetParamParts(m.Transform.Params); ok {
				for _, part := range parts {
					if part.Type == "field" {
						required[part.Value] = true
					}
				}
			}
		}
	}

	return required
}

// timestampBounds holds the min/max timestamps from external storage
type timestampBounds struct {
	MaxCreatedAt time.Time
	MinCreatedAt time.Time
	RowCount     int64
}

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
		"sync_state", args.SyncState,
	)

	start := time.Now()
	result := jobs.EmailFetchResult{
		JobUUID:       args.JobUUID,
		SchedulerUUID: args.SchedulerUUID,
		PipelineUUID:  args.PipelineUUID,
		StartedAt:     start,
		NewSyncState:  args.SyncState, // Default to current state
	}

	// Connect to external PostgreSQL first to query timestamp bounds
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

	// Query external storage for timestamp bounds
	var storageBounds *timestampBounds
	if args.TargetTableName != "" && args.TimestampColumn != "" {
		storageBounds, err = e.queryStorageTimestampBounds(ctx, pool, args.TargetTableName, args.TimestampColumn)
		if err != nil {
			e.log.Warn("failed to query storage bounds, proceeding without", "error", err)
			storageBounds = nil
		} else {
			e.log.Debug("storage timestamp bounds",
				"max_created_at", storageBounds.MaxCreatedAt,
				"min_created_at", storageBounds.MinCreatedAt,
				"row_count", storageBounds.RowCount,
			)
		}
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

	// Build search criteria based on sync state and storage bounds
	searchCriteria := e.buildSearchCriteria(args, storageBounds)

	// Search for messages matching criteria
	uids, err := imapClient.UidSearch(searchCriteria)
	if err != nil {
		result.Success = false
		result.Error = fmt.Sprintf("IMAP search failed: %v", err)
		result.CompletedAt = time.Now()
		result.DurationMs = time.Since(start).Milliseconds()
		e.log.Error("IMAP search failed", "error", err)
		return json.Marshal(result)
	}

	if len(uids) == 0 {
		e.log.Info("no messages found matching criteria", "job_uuid", args.JobUUID, "sync_state", args.SyncState)
		result.Success = true
		result.MessagesFetched = 0
		result.MessagesStored = 0
		// Determine new sync state when no messages found
		result.NewSyncState = e.determineNewSyncState(args, storageBounds, time.Time{}, true)
		result.CompletedAt = time.Now()
		result.DurationMs = time.Since(start).Milliseconds()
		return json.Marshal(result)
	}

	e.log.Info("found messages to fetch", "count", len(uids), "batch_size", args.BatchSize, "sync_state", args.SyncState)

	// Sort UIDs based on sync direction
	// For recent sync: fetch newest first (descending)
	// For historical sync: fetch oldest first (ascending, then reverse to go back in time)
	sortedUIDs := e.sortUIDsForSyncDirection(uids, args.SyncState)

	// Limit to batch size
	if args.BatchSize > 0 && len(sortedUIDs) > args.BatchSize {
		sortedUIDs = sortedUIDs[:args.BatchSize]
	}
	result.MessagesFetched = len(sortedUIDs)

	// Pre-compute required fields from mapper config (lazy extraction optimization)
	requiredFields := collectRequiredFields(args.MapperConfig)

	// Create mapper executor
	mapperExec := newJobMapperExecutor(args.MapperConfig)

	// Fetch and process messages using UID FETCH
	fetchSeqSet := new(imap.SeqSet)
	fetchSeqSet.AddNum(sortedUIDs...)

	section := &imap.BodySectionName{}
	items := []imap.FetchItem{section.FetchItem(), imap.FetchEnvelope, imap.FetchFlags, imap.FetchUid}

	messages := make(chan *imap.Message, 10)
	done := make(chan error, 1)
	go func() {
		done <- imapClient.UidFetch(fetchSeqSet, items, messages)
	}()

	var newestTimestamp, oldestTimestamp time.Time
	var highestUID uint32 = args.LastUID

	for msg := range messages {
		// Track highest UID processed (for backward compatibility)
		if msg.Uid > highestUID {
			highestUID = msg.Uid
		}

		// Track timestamps for sync progress
		if msg.Envelope != nil && !msg.Envelope.Date.IsZero() {
			msgTime := msg.Envelope.Date
			if newestTimestamp.IsZero() || msgTime.After(newestTimestamp) {
				newestTimestamp = msgTime
			}
			if oldestTimestamp.IsZero() || msgTime.Before(oldestTimestamp) {
				oldestTimestamp = msgTime
			}
		}

		// Convert IMAP message to source data map (only extracts required fields)
		sourceData := e.imapToSourceData(msg, section, requiredFields)

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
	result.NewestTimestamp = newestTimestamp
	result.OldestTimestamp = oldestTimestamp
	result.NewSyncState = e.determineNewSyncState(args, storageBounds, oldestTimestamp, false)
	result.CompletedAt = time.Now()
	result.DurationMs = time.Since(start).Milliseconds()

	e.log.Info("email fetch job completed",
		"job_uuid", args.JobUUID,
		"fetched", result.MessagesFetched,
		"stored", result.MessagesStored,
		"errors", result.ErrorCount,
		"newest_timestamp", result.NewestTimestamp,
		"oldest_timestamp", result.OldestTimestamp,
		"new_sync_state", result.NewSyncState,
		"duration_ms", result.DurationMs,
	)

	return json.Marshal(result)
}

// queryStorageTimestampBounds queries the external storage for min/max timestamps
func (e *Executor) queryStorageTimestampBounds(ctx context.Context, pool *pgxpool.Pool, tableName, timestampColumn string) (*timestampBounds, error) {
	query := fmt.Sprintf(`
		SELECT
			COALESCE(MAX("%s"), '1970-01-01'::timestamptz) as max_ts,
			COALESCE(MIN("%s"), '1970-01-01'::timestamptz) as min_ts,
			COUNT(*) as row_count
		FROM "%s"
	`, timestampColumn, timestampColumn, tableName)

	var bounds timestampBounds
	err := pool.QueryRow(ctx, query).Scan(&bounds.MaxCreatedAt, &bounds.MinCreatedAt, &bounds.RowCount)
	if err != nil {
		return nil, fmt.Errorf("failed to query timestamp bounds: %w", err)
	}

	// If row_count is 0, reset timestamps to zero values
	if bounds.RowCount == 0 {
		bounds.MaxCreatedAt = time.Time{}
		bounds.MinCreatedAt = time.Time{}
	}

	return &bounds, nil
}

// buildSearchCriteria builds IMAP search criteria based on sync state
func (e *Executor) buildSearchCriteria(args jobs.EmailFetchJobArgs, bounds *timestampBounds) *imap.SearchCriteria {
	criteria := imap.NewSearchCriteria()

	switch args.SyncState {
	case "initial", "sync_recent", "":
		// Fetch recent messages: those newer than what we have in storage
		if bounds != nil && !bounds.MaxCreatedAt.IsZero() {
			// Add 1 second to avoid re-fetching the exact same timestamp
			criteria.Since = bounds.MaxCreatedAt.Add(time.Second)
		}
		// If bounds is nil or MaxCreatedAt is zero, fetch all (no date filter)

	case "sync_historical":
		// Fetch older messages: those older than what we have in storage
		if bounds != nil && !bounds.MinCreatedAt.IsZero() {
			// Subtract 1 second to avoid re-fetching the exact same timestamp
			criteria.Before = bounds.MinCreatedAt.Add(-time.Second)
		}
		// Apply cutoff date if set
		if !args.CutoffDate.IsZero() {
			criteria.Since = args.CutoffDate
		}

	case "sync_complete":
		// In sync_complete state, only check for new messages (same as sync_recent)
		if bounds != nil && !bounds.MaxCreatedAt.IsZero() {
			criteria.Since = bounds.MaxCreatedAt.Add(time.Second)
		}
	}

	return criteria
}

// sortUIDsForSyncDirection sorts UIDs based on sync direction
func (e *Executor) sortUIDsForSyncDirection(uids []uint32, syncState string) []uint32 {
	sorted := make([]uint32, len(uids))
	copy(sorted, uids)

	switch syncState {
	case "sync_historical":
		// For historical sync, we want oldest first (which means highest UIDs first typically)
		// But since we're going backwards, we want the oldest messages we haven't synced yet
		// UIDs are generally ascending with time, so we sort descending to get newest first
		// then process to go back in time
		for i, j := 0, len(sorted)-1; i < j; i, j = i+1, j-1 {
			sorted[i], sorted[j] = sorted[j], sorted[i]
		}
	default:
		// For recent sync, sort descending to get newest first
		for i, j := 0, len(sorted)-1; i < j; i, j = i+1, j-1 {
			sorted[i], sorted[j] = sorted[j], sorted[i]
		}
	}

	return sorted
}

// determineNewSyncState determines the next sync state based on results
func (e *Executor) determineNewSyncState(args jobs.EmailFetchJobArgs, bounds *timestampBounds, oldestFetched time.Time, noMessages bool) string {
	switch args.SyncState {
	case "initial", "":
		// From initial, move to sync_recent
		if noMessages {
			// No messages at all, go to historical
			return "sync_historical"
		}
		return "sync_recent"

	case "sync_recent":
		if noMessages {
			// No new messages, move to historical sync
			return "sync_historical"
		}
		// Stay in sync_recent while there are new messages
		return "sync_recent"

	case "sync_historical":
		if noMessages {
			// No more historical messages, sync is complete
			return "sync_complete"
		}
		// Check if we've reached the cutoff date
		if !args.CutoffDate.IsZero() && !oldestFetched.IsZero() && oldestFetched.Before(args.CutoffDate) {
			return "sync_complete"
		}
		// Continue historical sync
		return "sync_historical"

	case "sync_complete":
		// Stay in sync_complete (periodic checks will find new messages if any)
		return "sync_complete"
	}

	return args.SyncState
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

// imapToSourceData extracts only the required fields from an IMAP message.
// This is a lazy extraction optimization - fields not in requiredFields are not extracted,
// avoiding expensive operations like body parsing when not needed.
func (e *Executor) imapToSourceData(msg *imap.Message, section *imap.BodySectionName, requiredFields map[string]bool) map[string]any {
	data := make(map[string]any)
	pb := &parsedBody{} // Lazy body parsing

	for fieldKey := range requiredFields {
		if extractor, ok := fieldExtractors[fieldKey]; ok {
			if value := extractor(msg, section, pb); value != nil {
				data[fieldKey] = value
			}
		}
	}

	return data
}

// sanitizeUTF8 removes invalid UTF-8 byte sequences from a string.
// This is necessary because email content may have mixed encodings or
// corrupted data that PostgreSQL will reject.
func sanitizeUTF8(s string) string {
	if utf8.ValidString(s) {
		return s
	}
	// Build a new string with only valid UTF-8 runes
	var b strings.Builder
	b.Grow(len(s))
	for i := 0; i < len(s); {
		r, size := utf8.DecodeRuneInString(s[i:])
		if r == utf8.RuneError && size == 1 {
			// Invalid byte, replace with replacement character
			b.WriteRune('\uFFFD')
			i++
		} else {
			b.WriteRune(r)
			i += size
		}
	}
	return b.String()
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
		// Sanitize string values to ensure valid UTF-8
		if s, ok := val.(string); ok {
			val = sanitizeUTF8(s)
		}
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

		// Apply transform if specified (pass sourceData for transforms like concat)
		if mapping.Transform != nil {
			value = e.applyTransform(value, mapping.Transform, sourceData)
		}

		// Store in result by table
		if result[mapping.TargetTable] == nil {
			result[mapping.TargetTable] = make(map[string]any)
		}
		result[mapping.TargetTable][mapping.TargetField] = value
	}

	return result, nil
}

// applyTransform applies a transformation using the shared transforms package.
// This provides full parity with the backend mapper (all 9 transform types).
func (e *jobMapperExecutor) applyTransform(value any, transform *jobs.MapperTransformData, sourceData map[string]any) any {
	if transform == nil {
		return value
	}

	t := transforms.Transform{
		Type:   transform.Type,
		Params: transform.Params,
	}

	result, err := transforms.Apply(value, t, sourceData)
	if err != nil {
		// Graceful degradation: return original value on transform error
		return value
	}
	return result
}
