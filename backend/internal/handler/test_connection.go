package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/nats-io/nats.go/jetstream"
	"golang.org/x/oauth2"

	"github.com/shadowapi/shadowapi/backend/internal/converter"
	"github.com/shadowapi/shadowapi/backend/internal/jobstore"
	"github.com/shadowapi/shadowapi/backend/internal/queue"
	"github.com/shadowapi/shadowapi/backend/internal/worker/subjects"
	"github.com/shadowapi/shadowapi/backend/pkg/api"
	"github.com/shadowapi/shadowapi/backend/pkg/jobs"
	"github.com/shadowapi/shadowapi/backend/pkg/query"
)

// DatasourceEmailOAuthTest initiates a connection test for an OAuth email datasource.
// POST /datasource/email_oauth/{uuid}/test
func (h *Handler) DatasourceEmailOAuthTest(ctx context.Context, params api.DatasourceEmailOAuthTestParams) (*api.TestConnectionJob, error) {
	log := h.log.With("handler", "DatasourceEmailOAuthTest")

	// Parse datasource UUID
	dsUUID, err := uuid.FromString(params.UUID)
	if err != nil {
		log.Error("failed to parse datasource uuid", "error", err)
		return nil, ErrWithCode(http.StatusBadRequest, E("invalid datasource UUID"))
	}

	q := query.New(h.dbp)

	// Get datasource
	ds, err := q.GetDatasource(ctx, pgtype.UUID{Bytes: converter.UToBytes(dsUUID), Valid: true})
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrWithCode(http.StatusNotFound, E("datasource not found"))
		}
		log.Error("failed to get datasource", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to get datasource"))
	}

	// Verify it's an email_oauth datasource
	if ds.Datasource.Type != "email_oauth" {
		return nil, ErrWithCode(http.StatusBadRequest, E("datasource is not email_oauth type"))
	}

	// Parse datasource settings
	var settings struct {
		Email           string `json:"email"`
		Provider        string `json:"provider"`
		OAuth2TokenUUID string `json:"oauth2_token_uuid"`
	}
	if err := json.Unmarshal(ds.Datasource.Settings, &settings); err != nil {
		log.Error("failed to parse datasource settings", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("invalid datasource settings"))
	}

	// Check if OAuth2 token is configured
	if settings.OAuth2TokenUUID == "" {
		return nil, ErrWithCode(http.StatusBadRequest, E("no OAuth2 token configured for this datasource"))
	}

	// Get OAuth2 token
	tokenUUID, err := uuid.FromString(settings.OAuth2TokenUUID)
	if err != nil {
		log.Error("failed to parse token uuid", "error", err)
		return nil, ErrWithCode(http.StatusBadRequest, E("invalid OAuth2 token UUID"))
	}

	tokenRow, err := q.GetOauth2TokenByUUID(ctx, pgtype.UUID{Bytes: converter.UToBytes(tokenUUID), Valid: true})
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrWithCode(http.StatusBadRequest, E("OAuth2 token not found"))
		}
		log.Error("failed to get oauth2 token", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to get OAuth2 token"))
	}

	// Parse token data
	var storedToken oauth2.Token
	if err := json.Unmarshal(tokenRow.Oauth2Token.Token, &storedToken); err != nil {
		log.Error("failed to parse token data", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("invalid OAuth2 token data"))
	}

	// Create job record in KV store
	jobUUID := uuid.Must(uuid.NewV7())
	err = h.jobStore.Put(ctx, &jobstore.JobResult{
		UUID:         jobUUID.String(),
		ResourceType: "email_oauth",
		ResourceUUID: params.UUID,
		Status:       "pending",
		CreatedAt:    time.Now().UTC(),
	})
	if err != nil {
		log.Error("failed to create test connection job", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to create test job"))
	}

	// Also create worker_jobs record for Active Jobs tracking
	subject := subjects.GlobalJobSubject(subjects.JobTypeTestConnectionEmailOAuth)
	_, err = q.CreateWorkerJob(ctx, query.CreateWorkerJobParams{
		UUID:          pgtype.UUID{Bytes: converter.UToBytes(jobUUID), Valid: true},
		SchedulerUuid: pgtype.UUID{Valid: false}, // No scheduler for test connections
		JobUuid:       pgtype.UUID{Bytes: converter.UToBytes(jobUUID), Valid: true},
		Subject:       subject,
		Status:        "running",
		Data:          []byte(`{"type":"test_connection","resource_type":"email_oauth","resource_uuid":"` + params.UUID + `"}`),
		FinishedAt:    pgtype.Timestamptz{Valid: false},
	})
	if err != nil {
		log.Warn("failed to create worker_jobs record", "error", err)
		// Don't fail the request, the job was already created in KV store
	}

	// Build job args
	jobArgs := jobs.TestConnectionEmailOAuthJobArgs{
		JobUUID:        jobUUID.String(),
		DatasourceUUID: params.UUID,
		Email:          settings.Email,
		Provider:       settings.Provider,
		AccessToken:    storedToken.AccessToken,
	}

	// Publish job to NATS
	payload, err := json.Marshal(jobArgs)
	if err != nil {
		log.Error("failed to marshal job args", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to create test job"))
	}

	headers := queue.Headers{"X-Job-ID": jobUUID.String()}
	if err := h.queue.PublishWithHeaders(ctx, subject, headers, payload); err != nil {
		log.Error("failed to publish job", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to dispatch test job"))
	}

	log.Info("email OAuth connection test job created",
		"job_uuid", jobUUID.String(),
		"datasource_uuid", params.UUID,
	)

	return &api.TestConnectionJob{
		UUID:         jobUUID.String(),
		ResourceType: api.TestConnectionJobResourceTypeEmailOAuth,
		ResourceUUID: params.UUID,
		Status:       api.TestConnectionJobStatusPending,
	}, nil
}

// StoragePostgresTest initiates a connection test for a PostgreSQL storage.
// POST /storage/postgres/{uuid}/test
func (h *Handler) StoragePostgresTest(ctx context.Context, params api.StoragePostgresTestParams) (*api.TestConnectionJob, error) {
	log := h.log.With("handler", "StoragePostgresTest")

	// Parse storage UUID
	storageUUID, err := uuid.FromString(params.UUID)
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

	// Parse settings
	var settings struct {
		Host     string `json:"host"`
		Port     string `json:"port"`
		User     string `json:"user"`
		Password string `json:"password"`
		Database string `json:"database"`
		Options  string `json:"options"`
	}
	if err := json.Unmarshal(storage.Storage.Settings, &settings); err != nil {
		log.Error("failed to parse storage settings", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("invalid storage settings"))
	}

	// Create job record in KV store
	jobUUID := uuid.Must(uuid.NewV7())
	err = h.jobStore.Put(ctx, &jobstore.JobResult{
		UUID:         jobUUID.String(),
		ResourceType: "postgres",
		ResourceUUID: params.UUID,
		Status:       "pending",
		CreatedAt:    time.Now().UTC(),
	})
	if err != nil {
		log.Error("failed to create test connection job", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to create test job"))
	}

	// Also create worker_jobs record for Active Jobs tracking
	subject := subjects.GlobalJobSubject(subjects.JobTypeTestConnectionPostgres)
	_, err = q.CreateWorkerJob(ctx, query.CreateWorkerJobParams{
		UUID:          pgtype.UUID{Bytes: converter.UToBytes(jobUUID), Valid: true},
		SchedulerUuid: pgtype.UUID{Valid: false}, // No scheduler for test connections
		JobUuid:       pgtype.UUID{Bytes: converter.UToBytes(jobUUID), Valid: true},
		Subject:       subject,
		Status:        "running",
		Data:          []byte(`{"type":"test_connection","resource_type":"postgres","resource_uuid":"` + params.UUID + `"}`),
		FinishedAt:    pgtype.Timestamptz{Valid: false},
	})
	if err != nil {
		log.Warn("failed to create worker_jobs record", "error", err)
		// Don't fail the request, the job was already created in KV store
	}

	// Build job args
	jobArgs := jobs.TestConnectionPostgresJobArgs{
		JobUUID:     jobUUID.String(),
		StorageUUID: params.UUID,
		Host:        settings.Host,
		Port:        settings.Port,
		User:        settings.User,
		Password:    settings.Password,
		Database:    settings.Database,
		Options:     settings.Options,
	}

	// Publish job to NATS
	payload, err := json.Marshal(jobArgs)
	if err != nil {
		log.Error("failed to marshal job args", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to create test job"))
	}

	headers := queue.Headers{"X-Job-ID": jobUUID.String()}
	if err := h.queue.PublishWithHeaders(ctx, subject, headers, payload); err != nil {
		log.Error("failed to publish job", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to dispatch test job"))
	}

	log.Info("PostgreSQL connection test job created",
		"job_uuid", jobUUID.String(),
		"storage_uuid", params.UUID,
	)

	return &api.TestConnectionJob{
		UUID:         jobUUID.String(),
		ResourceType: api.TestConnectionJobResourceTypePostgres,
		ResourceUUID: params.UUID,
		Status:       api.TestConnectionJobStatusPending,
	}, nil
}

// StoragePostgresTestInline tests a PostgreSQL connection using inline parameters (without saving).
// POST /storage/postgres/test
func (h *Handler) StoragePostgresTestInline(ctx context.Context, req *api.StoragePostgresTestRequest) (*api.TestConnectionJob, error) {
	log := h.log.With("handler", "StoragePostgresTestInline")

	// Extract and validate required fields
	host := req.Host.Or("")
	port := req.Port.Or("5432")
	user := req.User.Or("")
	password := req.Password.Or("")
	database := req.Database.Or("postgres")
	options := req.Options.Or("")

	if host == "" {
		return nil, ErrWithCode(http.StatusBadRequest, E("host is required"))
	}
	if user == "" {
		return nil, ErrWithCode(http.StatusBadRequest, E("user is required"))
	}

	// Create job record in KV store
	jobUUID := uuid.Must(uuid.NewV7())
	err := h.jobStore.Put(ctx, &jobstore.JobResult{
		UUID:         jobUUID.String(),
		ResourceType: "postgres",
		ResourceUUID: "", // No resource UUID since not saved yet
		Status:       "pending",
		CreatedAt:    time.Now().UTC(),
	})
	if err != nil {
		log.Error("failed to create test connection job", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to create test job"))
	}

	// Also create worker_jobs record for Active Jobs tracking
	subject := subjects.GlobalJobSubject(subjects.JobTypeTestConnectionPostgres)
	q := query.New(h.dbp)
	_, err = q.CreateWorkerJob(ctx, query.CreateWorkerJobParams{
		UUID:          pgtype.UUID{Bytes: converter.UToBytes(jobUUID), Valid: true},
		SchedulerUuid: pgtype.UUID{Valid: false}, // No scheduler for test connections
		JobUuid:       pgtype.UUID{Bytes: converter.UToBytes(jobUUID), Valid: true},
		Subject:       subject,
		Status:        "running",
		Data:          []byte(`{"type":"test_connection","resource_type":"postgres","inline":true}`),
		FinishedAt:    pgtype.Timestamptz{Valid: false},
	})
	if err != nil {
		log.Warn("failed to create worker_jobs record", "error", err)
		// Don't fail the request, the job was already created in KV store
	}

	// Build job args
	jobArgs := jobs.TestConnectionPostgresJobArgs{
		JobUUID:     jobUUID.String(),
		StorageUUID: "", // Empty - not saved yet
		Host:        host,
		Port:        port,
		User:        user,
		Password:    password,
		Database:    database,
		Options:     options,
	}

	// Publish job to NATS
	payload, err := json.Marshal(jobArgs)
	if err != nil {
		log.Error("failed to marshal job args", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to create test job"))
	}

	headers := queue.Headers{"X-Job-ID": jobUUID.String()}
	if err := h.queue.PublishWithHeaders(ctx, subject, headers, payload); err != nil {
		log.Error("failed to publish job", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to dispatch test job"))
	}

	log.Info("PostgreSQL inline connection test job created", "job_uuid", jobUUID.String())

	return &api.TestConnectionJob{
		UUID:         jobUUID.String(),
		ResourceType: api.TestConnectionJobResourceTypePostgres,
		ResourceUUID: "", // Empty for inline test
		Status:       api.TestConnectionJobStatusPending,
	}, nil
}

// TestConnectionJobGet retrieves the status and result of a test connection job.
// GET /test-connection-job/{uuid}
func (h *Handler) TestConnectionJobGet(ctx context.Context, params api.TestConnectionJobGetParams) (*api.TestConnectionJob, error) {
	log := h.log.With("handler", "TestConnectionJobGet")

	// Get job from KV store
	job, err := h.jobStore.Get(ctx, params.UUID)
	if err != nil {
		if err == jetstream.ErrKeyNotFound {
			return nil, ErrWithCode(http.StatusNotFound, E("test connection job not found or expired"))
		}
		log.Error("failed to get test connection job", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to get test connection job"))
	}

	// Build response
	resp := &api.TestConnectionJob{
		UUID:         job.UUID,
		ResourceUUID: job.ResourceUUID,
		Status:       api.TestConnectionJobStatus(job.Status),
		CreatedAt:    api.NewOptDateTime(job.CreatedAt),
	}

	// Set resource type
	switch job.ResourceType {
	case "email_oauth":
		resp.ResourceType = api.TestConnectionJobResourceTypeEmailOAuth
	case "postgres":
		resp.ResourceType = api.TestConnectionJobResourceTypePostgres
	}

	// Set completed_at if present
	if !job.CompletedAt.IsZero() {
		resp.CompletedAt = api.NewOptDateTime(job.CompletedAt)
	}

	// Parse and set result if present
	if job.Result != nil {
		var result jobs.TestConnectionResult
		if err := json.Unmarshal(job.Result, &result); err == nil {
			resp.Result = api.NewOptTestConnectionResult(api.TestConnectionResult{
				Success:      result.Success,
				ErrorCode:    api.NewOptTestConnectionResultErrorCode(api.TestConnectionResultErrorCode(result.ErrorCode)),
				ErrorMessage: api.NewOptString(result.ErrorMessage),
				ErrorDetails: api.NewOptString(result.ErrorDetails),
				DurationMs:   api.NewOptInt(int(result.DurationMs)),
				TestedAt:     api.NewOptDateTime(result.TestedAt),
			})
		}
	}

	return resp, nil
}
