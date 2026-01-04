package scheduler

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/robfig/cron/v3"
	"golang.org/x/oauth2"
	"log/slog"

	"github.com/shadowapi/shadowapi/backend/internal/converter"
	"github.com/shadowapi/shadowapi/backend/internal/metrics"
	"github.com/shadowapi/shadowapi/backend/internal/queue"
	"github.com/shadowapi/shadowapi/backend/internal/worker/monitor"
	"github.com/shadowapi/shadowapi/backend/internal/worker/subjects"
	"github.com/shadowapi/shadowapi/backend/pkg/api"
	"github.com/shadowapi/shadowapi/backend/pkg/jobs"
	"github.com/shadowapi/shadowapi/backend/pkg/query"
)

type MultiEmailScheduler struct {
	log        *slog.Logger
	dbp        *pgxpool.Pool
	queue      *queue.Queue
	cronParser cron.Parser
	interval   time.Duration
	maxBackoff time.Duration
	monitor    *monitor.WorkerMonitor
}

func NewMultiEmailScheduler(log *slog.Logger, dbp *pgxpool.Pool, queue *queue.Queue, monitor *monitor.WorkerMonitor) *MultiEmailScheduler {
	return &MultiEmailScheduler{
		log:        log,
		dbp:        dbp,
		queue:      queue,
		monitor:    monitor,
		cronParser: cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow),
		interval:   time.Minute,
		maxBackoff: 10 * time.Minute,
	}
}

func (s *MultiEmailScheduler) Start(ctx context.Context) {
	ticker := time.NewTicker(s.interval)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				s.run(ctx)
			case <-ctx.Done():
				s.log.Info("MultiEmailScheduler shutting down")
				return
			}
		}
	}()
}

func (s *MultiEmailScheduler) run(ctx context.Context) {
	queries := query.New(s.dbp)
	now := time.Now().UTC()

	// Get all enabled, unpaused email schedulers.
	schedulers, err := queries.GetSchedulers(ctx, query.GetSchedulersParams{
		OrderBy:        "created_at",
		OrderDirection: "asc",
		Offset:         0,
		Limit:          100,
		PipelineUuid:   "",
		IsEnabled:      1,
		IsPaused:       0,
	})
	if err != nil {
		s.log.Error("Failed fetching schedulers", "err", err)
		return
	}

	if len(schedulers) == 0 {
		s.log.Debug("No schedulers found")
		return
	}

	// Loop over each scheduler row.
	for _, sched := range schedulers {
		// If NextRun is set and still in the future, skip this scheduler.
		if sched.NextRun.Valid && sched.NextRun.Time.After(now) {
			s.log.Debug("MultiEmailScheduler Skipping scheduler", "schedulerUUID", sched.UUID.String(), "nextRun", sched.NextRun.Time)
			continue
		}

		// Build job args with all necessary data
		jobArgs, workspaceSlug, err := s.buildEmailFetchJobArgs(ctx, queries, sched)
		if err != nil {
			s.log.Error("Failed to build job args", "scheduler", sched.UUID.String(), "err", err)
			continue
		}

		s.log.Debug("MultiEmailScheduler Scheduling job", "schedulerUUID", sched.UUID.String(), "jobUUID", jobArgs.JobUUID)

		// Marshal and publish
		jobPayload, err := json.Marshal(jobArgs)
		if err != nil {
			s.log.Error("Failed to marshal job payload", "scheduler", sched.UUID.String(), "err", err)
			continue
		}

		headers := queue.Headers{"X-Job-ID": jobArgs.JobUUID}
		subject := subjects.JobSubject(workspaceSlug, subjects.JobTypeEmailOAuthFetch)

		err = s.queue.PublishWithHeaders(ctx, subject, headers, jobPayload)
		if err != nil {
			s.log.Error("Failed to publish job", "schedulerUUID", sched.UUID.String(), "pipelineUUID", sched.PipelineUuid.String(), "err", err)
			backoffDelay := s.calculateBackoff(sched)
			s.updateNextRun(ctx, queries, converter.UuidToPgUUID(sched.UUID), now.Add(backoffDelay))
			continue
		}

		// Create worker_jobs record for tracking
		s.createWorkerJobRecord(ctx, queries, jobArgs.JobUUID, sched.UUID.String(), subject)

		// Calculate the next run time.
		nextRun := s.nextRunTime(sched, now)
		// Update the scheduler record with the new run time.
		s.updateSchedulerRun(ctx, queries, converter.UuidToPgUUID(sched.UUID), now, nextRun)
		// Increase the scheduled jobs metric.
		metrics.JobScheduledTotal.WithLabelValues(sched.PipelineUuid.String(), "").Inc()
	}
}

// buildEmailFetchJobArgs constructs complete job payload with all necessary data
func (s *MultiEmailScheduler) buildEmailFetchJobArgs(
	ctx context.Context,
	queries *query.Queries,
	sched query.GetSchedulersRow,
) (*jobs.EmailFetchJobArgs, string, error) {
	jobUUID := uuid.Must(uuid.NewV7()).String()

	// Load pipeline with datasource and storage
	pipelineDetails, err := queries.GetPipelineWithDetails(ctx, converter.UuidToPgUUID(*sched.PipelineUuid))
	if err != nil {
		return nil, "", fmt.Errorf("failed to load pipeline: %w", err)
	}

	// Check if pipeline is enabled and datasource type is email_oauth
	if !pipelineDetails.PipelineEnabled {
		return nil, "", fmt.Errorf("pipeline is disabled")
	}

	if pipelineDetails.DatasourceType != "email_oauth" {
		return nil, "", fmt.Errorf("unsupported datasource type: %s (expected email_oauth)", pipelineDetails.DatasourceType)
	}

	if pipelineDetails.StorageType != "postgres" {
		return nil, "", fmt.Errorf("unsupported storage type: %s (expected postgres)", pipelineDetails.StorageType)
	}

	// Parse datasource settings to get oauth2_token_uuid and email
	var dsSettings struct {
		OAuth2TokenUUID string `json:"oauth2_token_uuid"`
		Email           string `json:"email"`
		MailboxName     string `json:"mailbox_name,omitempty"`
	}
	if err := json.Unmarshal(pipelineDetails.DatasourceSettings, &dsSettings); err != nil {
		return nil, "", fmt.Errorf("failed to parse datasource settings: %w", err)
	}

	if dsSettings.OAuth2TokenUUID == "" {
		return nil, "", fmt.Errorf("datasource missing oauth2_token_uuid - OAuth authentication required")
	}

	// Load OAuth2 token with client details
	tokenUUID, err := converter.ConvertStringToPgUUID(dsSettings.OAuth2TokenUUID)
	if err != nil {
		return nil, "", fmt.Errorf("invalid oauth2_token_uuid: %w", err)
	}

	tokenRow, err := queries.GetOauth2TokenWithClient(ctx, tokenUUID)
	if err != nil {
		return nil, "", fmt.Errorf("failed to load OAuth2 token: %w", err)
	}

	// Parse token data
	var token oauth2.Token
	if err := json.Unmarshal(tokenRow.Token, &token); err != nil {
		return nil, "", fmt.Errorf("failed to parse token: %w", err)
	}

	// Check if token is expired - skip job if expired (token refresh scheduler will handle it)
	if token.Expiry.Before(time.Now()) {
		// If expired less than 30 minutes ago, skip silently (pending token refresh)
		// If expired more than 30 minutes ago, return error (token refresh likely failed)
		if token.Expiry.After(time.Now().Add(-30 * time.Minute)) {
			return nil, "", fmt.Errorf("OAuth2 token expired - awaiting refresh (expired at %v)", token.Expiry)
		}
		return nil, "", fmt.Errorf("OAuth2 token expired at %v - token refresh may have failed", token.Expiry)
	}

	// Parse storage settings
	var storageSettings struct {
		Host     string `json:"host"`
		Port     string `json:"port"`
		User     string `json:"user"`
		Password string `json:"password"`
		Database string `json:"database"`
		Options  string `json:"options"`
	}
	if err := json.Unmarshal(pipelineDetails.StorageSettings, &storageSettings); err != nil {
		return nil, "", fmt.Errorf("failed to parse storage settings: %w", err)
	}

	// Validate storage settings
	if storageSettings.Host == "" || storageSettings.Port == "" || storageSettings.User == "" {
		return nil, "", fmt.Errorf("incomplete PostgreSQL storage settings")
	}

	// Extract mapper config from pipeline flow
	mapperConfig, err := s.extractMapperConfig(pipelineDetails.PipelineFlow)
	if err != nil {
		return nil, "", fmt.Errorf("failed to extract mapper config: %w", err)
	}

	// Get last processed UID for incremental fetch (DEPRECATED)
	lastUID := uint32(sched.LastUid)

	// Extract sync state fields
	syncState := "initial"
	if sched.SyncState.Valid && sched.SyncState.String != "" {
		syncState = sched.SyncState.String
	}

	var lastSyncTimestamp, oldestSyncTimestamp, cutoffDate time.Time
	if sched.LastSyncTimestamp.Valid {
		lastSyncTimestamp = sched.LastSyncTimestamp.Time
	}
	if sched.OldestSyncTimestamp.Valid {
		oldestSyncTimestamp = sched.OldestSyncTimestamp.Time
	}
	if sched.CutoffDate.Valid {
		cutoffDate = sched.CutoffDate.Time
	}

	// Extract target table name from mapper config
	targetTableName := extractTargetTableName(mapperConfig)
	timestampColumn := "created_at" // Default column name

	// Resolve IMAP server
	imapHost, imapPort := resolveIMAPServer(pipelineDetails.DatasourceProvider, "", 0)

	mailboxName := dsSettings.MailboxName
	if mailboxName == "" {
		mailboxName = "INBOX"
	}

	batchSize := int(sched.BatchSize)
	if batchSize <= 0 {
		batchSize = 100
	}

	return &jobs.EmailFetchJobArgs{
		JobUUID:       jobUUID,
		SchedulerUUID: sched.UUID.String(),
		PipelineUUID:  sched.PipelineUuid.String(),

		Email:       dsSettings.Email,
		Provider:    pipelineDetails.DatasourceProvider,
		AccessToken: token.AccessToken,

		IMAPHost: imapHost,
		IMAPPort: imapPort,

		LastUID:     lastUID,
		BatchSize:   batchSize,
		MailboxName: mailboxName,

		// Timestamp-based sync tracking
		SyncState:           syncState,
		LastSyncTimestamp:   lastSyncTimestamp,
		OldestSyncTimestamp: oldestSyncTimestamp,
		CutoffDate:          cutoffDate,
		TargetTableName:     targetTableName,
		TimestampColumn:     timestampColumn,

		MapperConfig: *mapperConfig,

		StorageHost:     storageSettings.Host,
		StoragePort:     storageSettings.Port,
		StorageUser:     storageSettings.User,
		StoragePassword: storageSettings.Password,
		StorageDatabase: storageSettings.Database,
		StorageOptions:  storageSettings.Options,
	}, pipelineDetails.WorkspaceSlug, nil
}

// extractTargetTableName extracts the first target table name from mapper config
func extractTargetTableName(config *jobs.MapperConfigData) string {
	if config == nil || len(config.Mappings) == 0 {
		return ""
	}
	// Return the first target table found
	for _, m := range config.Mappings {
		if m.TargetTable != "" {
			return m.TargetTable
		}
	}
	return ""
}

// extractMapperConfig extracts mapper configuration from pipeline flow JSON
func (s *MultiEmailScheduler) extractMapperConfig(flowJSON []byte) (*jobs.MapperConfigData, error) {
	if len(flowJSON) == 0 {
		return nil, fmt.Errorf("pipeline flow is empty")
	}

	var flow api.PipelineFlow
	if err := json.Unmarshal(flowJSON, &flow); err != nil {
		return nil, fmt.Errorf("failed to parse pipeline flow: %w", err)
	}

	// Find the mapper node in the flow
	for _, node := range flow.Nodes {
		if node.Type == "mapper" {
			if config, ok := node.Data.Config.Get(); ok {
				// The config is stored as PipelineNodeDataConfig which wraps the mapper config
				configJSON, err := json.Marshal(config)
				if err != nil {
					return nil, fmt.Errorf("failed to marshal mapper config: %w", err)
				}

				var mapperConfig api.MapperConfig
				if err := json.Unmarshal(configJSON, &mapperConfig); err != nil {
					return nil, fmt.Errorf("failed to parse mapper config: %w", err)
				}

				// Convert to jobs.MapperConfigData
				result := &jobs.MapperConfigData{
					Version: mapperConfig.Version.Or("1.0"),
				}
				for _, m := range mapperConfig.Mappings {
					mapping := jobs.MapperMappingData{
						SourceEntity: string(m.SourceEntity),
						SourceField:  m.SourceField,
						TargetTable:  m.TargetTable,
						TargetField:  m.TargetField,
						IsEnabled:    m.IsEnabled.Or(true),
					}
					if m.Transform.IsSet() {
						mapping.Transform = &jobs.MapperTransformData{
							Type: string(m.Transform.Value.Type),
						}
						if m.Transform.Value.Params.IsSet() {
							mapping.Transform.Params = make(map[string]any)
							for k, v := range m.Transform.Value.Params.Value {
								var val any
								json.Unmarshal(v, &val)
								mapping.Transform.Params[k] = val
							}
						}
					}
					result.Mappings = append(result.Mappings, mapping)
				}
				return result, nil
			}
		}
	}

	return nil, fmt.Errorf("no mapper node found in pipeline flow")
}

// resolveIMAPServer returns IMAP host and port based on provider
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

// createWorkerJobRecord creates a tracking record for the job
func (s *MultiEmailScheduler) createWorkerJobRecord(
	ctx context.Context,
	queries *query.Queries,
	jobUUID string,
	schedulerUUID string,
	subject string,
) {
	jobUUIDParsed, err := uuid.FromString(jobUUID)
	if err != nil {
		s.log.Error("Failed to parse job UUID", "jobUUID", jobUUID, "error", err)
		return
	}
	schedUUIDParsed, err := uuid.FromString(schedulerUUID)
	if err != nil {
		s.log.Error("Failed to parse scheduler UUID", "schedulerUUID", schedulerUUID, "error", err)
		return
	}

	_, err = queries.CreateWorkerJob(ctx, query.CreateWorkerJobParams{
		UUID:          converter.UuidToPgUUID(uuid.Must(uuid.NewV7())),
		SchedulerUuid: pgtype.UUID{Bytes: schedUUIDParsed, Valid: true},
		JobUuid:       converter.UuidToPgUUID(jobUUIDParsed),
		Subject:       subject,
		Status:        "pending",
		Data:          nil,
		FinishedAt:    pgtype.Timestamptz{Valid: false},
	})
	if err != nil {
		s.log.Error("Failed to create worker job record", "error", err)
	}
}

func (s *MultiEmailScheduler) nextRunTime(sch query.GetSchedulersRow, now time.Time) time.Time {
	if sch.ScheduleType == "cron" {
		schedule, err := s.cronParser.Parse(sch.CronExpression.String)
		if err == nil {
			return schedule.Next(now)
		}
	}
	return now.Add(24 * time.Hour)
}

func (s *MultiEmailScheduler) updateSchedulerRun(ctx context.Context, queries *query.Queries, id pgtype.UUID, lastRun, nextRun time.Time) {
	err := queries.UpdateScheduler(ctx, query.UpdateSchedulerParams{
		CronExpression: pgtype.Text{String: "", Valid: false},
		RunAt:          pgtype.Timestamptz{Time: lastRun, Valid: true},
		Timezone:       "UTC",
		NextRun:        pgtype.Timestamptz{Time: nextRun, Valid: true},
		LastRun:        pgtype.Timestamptz{Time: lastRun, Valid: true},
		LastUid:        0, // Preserve existing value via COALESCE
		IsEnabled:      true,
		IsPaused:       false,
		UUID:           id,
		BatchSize:      100,
		CutoffDate:     pgtype.Timestamptz{Valid: false}, // Preserve existing value
	})
	if err != nil {
		s.log.Error("Failed to update scheduler run times", "error", err)
	}
}

func (s *MultiEmailScheduler) calculateBackoff(sch query.GetSchedulersRow) time.Duration {
	baseDelay := 5 * time.Minute
	backoff := baseDelay
	if backoff > s.maxBackoff {
		backoff = s.maxBackoff
	}
	return backoff
}

func (s *MultiEmailScheduler) updateNextRun(ctx context.Context, queries *query.Queries, id pgtype.UUID, nextRun time.Time) {
	err := queries.UpdateScheduler(ctx, query.UpdateSchedulerParams{
		CronExpression: pgtype.Text{String: "", Valid: false},
		RunAt:          pgtype.Timestamptz{Time: time.Now().UTC(), Valid: true},
		Timezone:       "UTC",
		NextRun:        pgtype.Timestamptz{Time: nextRun, Valid: true},
		LastRun:        pgtype.Timestamptz{Valid: false},
		LastUid:        0, // Preserve existing value via COALESCE
		IsEnabled:      true,
		IsPaused:       false,
		UUID:           id,
		BatchSize:      100,
		CutoffDate:     pgtype.Timestamptz{Valid: false}, // Preserve existing value
	})
	if err != nil {
		s.log.Error("Failed to update scheduler next run", "error", err)
	}
}
