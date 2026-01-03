package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/shadowapi/shadowapi/backend/internal/converter"
	"github.com/shadowapi/shadowapi/backend/internal/queue"
	"github.com/shadowapi/shadowapi/backend/internal/worker/subjects"
	"github.com/shadowapi/shadowapi/backend/internal/workspace"
	"github.com/shadowapi/shadowapi/backend/pkg/api"
	"github.com/shadowapi/shadowapi/backend/pkg/jobs"
	"github.com/shadowapi/shadowapi/backend/pkg/query"
)

// StoragePostgresMessagesQuery triggers a job to query messages from a PostgreSQL storage.
// The job streams individual message records to NATS.
// POST /storage/postgres/{uuid}/messages/query
func (h *Handler) StoragePostgresMessagesQuery(ctx context.Context, req api.OptStoragePostgresMessagesQueryReq, params api.StoragePostgresMessagesQueryParams) (api.StoragePostgresMessagesQueryRes, error) {
	log := h.log.With("handler", "StoragePostgresMessagesQuery")

	// Get workspace from context
	workspaceSlug := workspace.GetWorkspaceSlug(ctx)
	if workspaceSlug == "" {
		// Default to "internal" workspace if not specified
		workspaceSlug = "internal"
	}

	// Get storage UUID from params
	storageUUIDStr := params.UUID.String()
	storageUUID, err := uuid.FromString(storageUUIDStr)
	if err != nil {
		log.Error("failed to parse storage uuid", "error", err)
		return nil, ErrWithCode(http.StatusBadRequest, E("invalid storage UUID"))
	}

	q := query.New(h.dbp)

	// Get storage
	storage, err := q.GetStorage(ctx, pgtype.UUID{Bytes: converter.UToBytes(storageUUID), Valid: true})
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrWithCode(http.StatusNotFound, E("storage not found"))
		}
		log.Error("failed to get storage", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to get storage"))
	}

	// Verify it's a postgres storage
	if storage.Storage.Type != "postgres" {
		return nil, ErrWithCode(http.StatusBadRequest, E("storage is not PostgreSQL type"))
	}

	// Parse storage settings including tables configuration
	var settings struct {
		Host     string `json:"host"`
		Port     string `json:"port"`
		User     string `json:"user"`
		Password string `json:"password"`
		Database string `json:"database"`
		Options  string `json:"options"`
		Tables   []struct {
			Name   string `json:"name"`
			Fields []struct {
				Name string `json:"name"`
				Type string `json:"type"`
			} `json:"fields"`
		} `json:"tables"`
	}
	if err := json.Unmarshal(storage.Storage.Settings, &settings); err != nil {
		log.Error("failed to parse storage settings", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("invalid storage settings"))
	}

	// Validate that tables are configured
	if len(settings.Tables) == 0 {
		return nil, ErrWithCode(http.StatusBadRequest, E("no tables configured for this storage"))
	}

	// Convert tables to job config format
	tables := make([]jobs.TableConfig, len(settings.Tables))
	for i, t := range settings.Tables {
		fields := make([]jobs.FieldConfig, len(t.Fields))
		for j, f := range t.Fields {
			fields[j] = jobs.FieldConfig{
				Name: f.Name,
				Type: f.Type,
			}
		}
		tables[i] = jobs.TableConfig{
			Name:   t.Name,
			Fields: fields,
		}
	}

	// Create job
	jobUUID := uuid.Must(uuid.NewV7())

	// Build NATS subject for data records
	dataSubject := subjects.DataSubject(workspaceSlug, "messages")

	// Build job args with defaults
	limit := 100
	offset := 0

	// Override with request values if provided
	if req.IsSet() {
		reqVal := req.Value
		if reqVal.Limit.IsSet() && reqVal.Limit.Value > 0 {
			limit = reqVal.Limit.Value
		}
		if reqVal.Offset.IsSet() && reqVal.Offset.Value >= 0 {
			offset = reqVal.Offset.Value
		}
	}

	jobArgs := jobs.MessageQueryJobArgs{
		JobUUID:         jobUUID.String(),
		WorkspaceSlug:   workspaceSlug,
		Limit:           limit,
		Offset:          offset,
		StorageHost:     settings.Host,
		StoragePort:     settings.Port,
		StorageUser:     settings.User,
		StoragePassword: settings.Password,
		StorageDatabase: settings.Database,
		StorageOptions:  settings.Options,
		Tables:          tables,
		NATSSubject:     dataSubject,
	}

	// Create worker_jobs record
	subject := subjects.JobSubject(workspaceSlug, subjects.JobTypeMessageQuery)
	_, err = q.CreateWorkerJob(ctx, query.CreateWorkerJobParams{
		UUID:          pgtype.UUID{Bytes: converter.UToBytes(jobUUID), Valid: true},
		SchedulerUuid: pgtype.UUID{Valid: false},
		JobUuid:       pgtype.UUID{Bytes: converter.UToBytes(jobUUID), Valid: true},
		Subject:       subject,
		Status:        "running",
		Data:          []byte(`{"type":"message_query","storage_uuid":"` + storageUUIDStr + `"}`),
		FinishedAt:    pgtype.Timestamptz{Valid: false},
	})
	if err != nil {
		log.Warn("failed to create worker_jobs record", "error", err)
	}

	// Publish job to NATS
	payload, err := json.Marshal(jobArgs)
	if err != nil {
		log.Error("failed to marshal job args", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to create query job"))
	}

	headers := queue.Headers{"X-Job-ID": jobUUID.String()}
	if err := h.queue.PublishWithHeaders(ctx, subject, headers, payload); err != nil {
		log.Error("failed to publish job", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to dispatch query job"))
	}

	// Extract table names for logging
	tableNames := make([]string, len(tables))
	for i, t := range tables {
		tableNames[i] = t.Name
	}

	log.Info("message query job created",
		"job_uuid", jobUUID.String(),
		"storage_uuid", storageUUIDStr,
		"workspace", workspaceSlug,
		"nats_subject", dataSubject,
		"tables", tableNames,
	)

	return &api.MessageQueryJob{
		UUID:          jobUUID.String(),
		StorageUUID:   storageUUIDStr,
		Status:        api.MessageQueryJobStatusPending,
		NatsSubject:   dataSubject,
		Limit:         api.NewOptInt(limit),
		Offset:        api.NewOptInt(offset),
		TablesQueried: tableNames,
		CreatedAt:     api.NewOptDateTime(time.Now().UTC()),
	}, nil
}
