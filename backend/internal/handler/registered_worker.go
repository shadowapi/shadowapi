package handler

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"time"

	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/shadowapi/shadowapi/backend/pkg/api"
	"github.com/shadowapi/shadowapi/backend/pkg/query"
)

// ListRegisteredWorkers implements listRegisteredWorkers operation.
//
// GET /workers
func (h *Handler) ListRegisteredWorkers(ctx context.Context) (api.ListRegisteredWorkersRes, error) {
	q := query.New(h.dbp)

	workers, err := q.ListRegisteredWorkers(ctx, query.ListRegisteredWorkersParams{
		Column1: 0, // No limit
		Offset:  0,
	})
	if err != nil {
		return nil, err
	}

	result := make([]api.RegisteredWorker, 0, len(workers))
	for _, w := range workers {
		apiWorker, err := h.mapWorkerToAPI(ctx, q, w)
		if err != nil {
			h.log.Error("failed to map worker", "uuid", w.UUID, "error", err)
			continue
		}
		result = append(result, apiWorker)
	}

	res := api.ListRegisteredWorkersOKApplicationJSON(result)
	return &res, nil
}

// GetRegisteredWorker implements getRegisteredWorker operation.
//
// GET /workers/{uuid}
func (h *Handler) GetRegisteredWorker(ctx context.Context, params api.GetRegisteredWorkerParams) (api.GetRegisteredWorkerRes, error) {
	workerUUID, err := uuid.FromString(params.UUID)
	if err != nil {
		return nil, ErrWithCode(400, E("invalid UUID format"))
	}

	q := query.New(h.dbp)
	worker, err := q.GetRegisteredWorker(ctx, workerUUID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrWithCode(404, E("worker not found"))
		}
		return nil, err
	}

	apiWorker, err := h.mapWorkerToAPI(ctx, q, worker)
	if err != nil {
		return nil, err
	}

	return &apiWorker, nil
}

// UpdateRegisteredWorker implements updateRegisteredWorker operation.
//
// PUT /workers/{uuid}
func (h *Handler) UpdateRegisteredWorker(ctx context.Context, req *api.RegisteredWorker, params api.UpdateRegisteredWorkerParams) (api.UpdateRegisteredWorkerRes, error) {
	workerUUID, err := uuid.FromString(params.UUID)
	if err != nil {
		return nil, ErrWithCode(400, E("invalid UUID format"))
	}

	q := query.New(h.dbp)

	// Check if worker exists
	_, err = q.GetRegisteredWorker(ctx, workerUUID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrWithCode(404, E("worker not found"))
		}
		return nil, err
	}

	// Convert labels to JSON
	var labelsJSON []byte
	if labels, ok := req.Labels.Get(); ok {
		labelsJSON, err = json.Marshal(labels)
		if err != nil {
			return nil, ErrWithCode(400, E("invalid labels format"))
		}
	} else {
		labelsJSON = []byte("{}")
	}

	// Get version
	var version pgtype.Text
	if v, ok := req.Version.Get(); ok {
		version = pgtype.Text{String: v, Valid: true}
	}

	// Get is_global
	isGlobal := false
	if g, ok := req.IsGlobal.Get(); ok {
		isGlobal = g
	}

	// Update worker
	worker, err := q.UpdateRegisteredWorker(ctx, query.UpdateRegisteredWorkerParams{
		UUID:     workerUUID,
		Name:     req.Name,
		IsGlobal: isGlobal,
		Version:  version,
		Labels:   labelsJSON,
	})
	if err != nil {
		return nil, err
	}

	// Update workspace assignments if provided
	if len(req.WorkspaceUuids) > 0 && !isGlobal {
		// Get current assignments
		links, err := q.ListWorkerWorkspaceLinks(ctx, &workerUUID)
		if err != nil {
			return nil, err
		}

		// Build set of current workspace UUIDs
		currentWorkspaces := make(map[string]bool)
		for _, link := range links {
			if link.WorkspaceUUID != nil {
				currentWorkspaces[link.WorkspaceUUID.String()] = true
			}
		}

		// Build set of desired workspace UUIDs
		desiredWorkspaces := make(map[string]bool)
		for _, wsUUID := range req.WorkspaceUuids {
			desiredWorkspaces[wsUUID] = true
		}

		// Remove workspaces no longer needed
		for wsUUIDStr := range currentWorkspaces {
			if !desiredWorkspaces[wsUUIDStr] {
				wsUUID, _ := uuid.FromString(wsUUIDStr)
				err := q.RemoveWorkerWorkspace(ctx, query.RemoveWorkerWorkspaceParams{
					WorkerUUID:    &workerUUID,
					WorkspaceUUID: &wsUUID,
				})
				if err != nil {
					h.log.Error("failed to remove workspace assignment", "error", err)
				}
			}
		}

		// Add new workspaces
		for wsUUIDStr := range desiredWorkspaces {
			if !currentWorkspaces[wsUUIDStr] {
				wsUUID, err := uuid.FromString(wsUUIDStr)
				if err != nil {
					continue
				}
				linkUUID := uuid.Must(uuid.NewV7())
				err = q.AddWorkerWorkspace(ctx, query.AddWorkerWorkspaceParams{
					UUID:          linkUUID,
					WorkerUUID:    &workerUUID,
					WorkspaceUUID: &wsUUID,
				})
				if err != nil {
					h.log.Error("failed to add workspace assignment", "error", err)
				}
			}
		}
	}

	apiWorker, err := h.mapWorkerToAPI(ctx, q, worker)
	if err != nil {
		return nil, err
	}

	return &apiWorker, nil
}

// DeleteRegisteredWorker implements deleteRegisteredWorker operation.
//
// DELETE /workers/{uuid}
func (h *Handler) DeleteRegisteredWorker(ctx context.Context, params api.DeleteRegisteredWorkerParams) (api.DeleteRegisteredWorkerRes, error) {
	workerUUID, err := uuid.FromString(params.UUID)
	if err != nil {
		return nil, ErrWithCode(400, E("invalid UUID format"))
	}

	q := query.New(h.dbp)

	// Check if worker exists
	_, err = q.GetRegisteredWorker(ctx, workerUUID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrWithCode(404, E("worker not found"))
		}
		return nil, err
	}

	if err := q.DeleteRegisteredWorker(ctx, workerUUID); err != nil {
		return nil, err
	}
	return &api.DeleteRegisteredWorkerOK{}, nil
}

// ListWorkerEnrollmentTokens implements listWorkerEnrollmentTokens operation.
//
// GET /workers/enrollment-tokens
func (h *Handler) ListWorkerEnrollmentTokens(ctx context.Context) (api.ListWorkerEnrollmentTokensRes, error) {
	q := query.New(h.dbp)

	tokens, err := q.ListEnrollmentTokens(ctx, query.ListEnrollmentTokensParams{
		Column1: 0, // No limit
		Offset:  0,
	})
	if err != nil {
		return nil, err
	}

	result := make([]api.WorkerEnrollmentToken, 0, len(tokens))
	for _, t := range tokens {
		result = append(result, h.mapEnrollmentTokenToAPI(t, ""))
	}

	res := api.ListWorkerEnrollmentTokensOKApplicationJSON(result)
	return &res, nil
}

// CreateWorkerEnrollmentToken implements createWorkerEnrollmentToken operation.
//
// POST /workers/enrollment-tokens
func (h *Handler) CreateWorkerEnrollmentToken(ctx context.Context, req *api.WorkerEnrollmentToken) (api.CreateWorkerEnrollmentTokenRes, error) {
	// Get user UUID from context
	userUUIDStr, err := getUserUUIDFromContext(ctx)
	if err != nil {
		return nil, ErrWithCode(401, E("%s", err.Error()))
	}
	userUUID, err := uuid.FromString(userUUIDStr)
	if err != nil {
		return nil, ErrWithCode(400, E("invalid user UUID"))
	}

	// Generate secure token
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return nil, err
	}
	rawToken := base64.URLEncoding.EncodeToString(tokenBytes)

	// Hash the token for storage using SHA256 (must match gRPC service validation)
	hasher := sha256.New()
	hasher.Write([]byte(rawToken))
	tokenHash := hex.EncodeToString(hasher.Sum(nil))

	// Set expiration (default 24 hours)
	expiresAt := time.Now().Add(24 * time.Hour)
	if ea, ok := req.ExpiresAt.Get(); ok {
		expiresAt = ea
	}

	// Get is_global
	isGlobal := false
	if g, ok := req.IsGlobal.Get(); ok {
		isGlobal = g
	}

	// Convert workspace UUIDs to pgtype.UUID array
	var workspaceUUIDs []pgtype.UUID
	for _, wsUUIDStr := range req.WorkspaceUuids {
		wsUUID, err := uuid.FromString(wsUUIDStr)
		if err != nil {
			continue
		}
		workspaceUUIDs = append(workspaceUUIDs, pgtype.UUID{Bytes: wsUUID, Valid: true})
	}

	tokenUUID := uuid.Must(uuid.NewV7())
	q := query.New(h.dbp)

	token, err := q.CreateEnrollmentToken(ctx, query.CreateEnrollmentTokenParams{
		UUID:              tokenUUID,
		TokenHash:         string(tokenHash),
		Name:              req.Name,
		IsGlobal:          isGlobal,
		WorkspaceUuids:    workspaceUUIDs,
		ExpiresAt:         pgtype.Timestamptz{Time: expiresAt, Valid: true},
		CreatedByUserUuid: &userUUID,
	})
	if err != nil {
		return nil, err
	}

	// Return the token with the raw token value (only time it's visible)
	apiToken := h.mapEnrollmentTokenToAPI(token, rawToken)
	return &apiToken, nil
}

// GetWorkerEnrollmentToken implements getWorkerEnrollmentToken operation.
//
// GET /workers/enrollment-tokens/{uuid}
func (h *Handler) GetWorkerEnrollmentToken(ctx context.Context, params api.GetWorkerEnrollmentTokenParams) (api.GetWorkerEnrollmentTokenRes, error) {
	tokenUUID, err := uuid.FromString(params.UUID)
	if err != nil {
		return nil, ErrWithCode(400, E("invalid UUID format"))
	}

	q := query.New(h.dbp)
	token, err := q.GetEnrollmentToken(ctx, tokenUUID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrWithCode(404, E("token not found"))
		}
		return nil, err
	}

	apiToken := h.mapEnrollmentTokenToAPI(token, "")
	return &apiToken, nil
}

// DeleteWorkerEnrollmentToken implements deleteWorkerEnrollmentToken operation.
//
// DELETE /workers/enrollment-tokens/{uuid}
func (h *Handler) DeleteWorkerEnrollmentToken(ctx context.Context, params api.DeleteWorkerEnrollmentTokenParams) (api.DeleteWorkerEnrollmentTokenRes, error) {
	tokenUUID, err := uuid.FromString(params.UUID)
	if err != nil {
		return nil, ErrWithCode(400, E("invalid UUID format"))
	}

	q := query.New(h.dbp)

	// Check if token exists
	_, err = q.GetEnrollmentToken(ctx, tokenUUID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrWithCode(404, E("token not found"))
		}
		return nil, err
	}

	if err := q.DeleteEnrollmentToken(ctx, tokenUUID); err != nil {
		return nil, err
	}
	return &api.DeleteWorkerEnrollmentTokenOK{}, nil
}

// mapWorkerToAPI converts a database worker to API worker
func (h *Handler) mapWorkerToAPI(ctx context.Context, q *query.Queries, w query.RegisteredWorker) (api.RegisteredWorker, error) {
	apiWorker := api.RegisteredWorker{
		UUID:     api.NewOptString(w.UUID.String()),
		Name:     w.Name,
		IsGlobal: api.NewOptBool(w.IsGlobal),
	}

	// Map status
	switch w.Status {
	case "online":
		apiWorker.Status = api.NewOptRegisteredWorkerStatus(api.RegisteredWorkerStatusOnline)
	case "offline":
		apiWorker.Status = api.NewOptRegisteredWorkerStatus(api.RegisteredWorkerStatusOffline)
	case "draining":
		apiWorker.Status = api.NewOptRegisteredWorkerStatus(api.RegisteredWorkerStatusDraining)
	}

	// Get active jobs and capacity from KV store (if worker is online)
	if w.Status == "online" && h.workerStore != nil {
		state, err := h.workerStore.Get(ctx, w.UUID.String())
		if err == nil && state != nil {
			apiWorker.ActiveJobs = api.NewOptInt(int(state.ActiveJobs))
			apiWorker.Capacity = api.NewOptInt(int(state.Capacity))
		}
	}

	// Map version
	if w.Version.Valid {
		apiWorker.Version = api.NewOptString(w.Version.String)
	}

	// Map labels
	if len(w.Labels) > 0 {
		var labels api.RegisteredWorkerLabels
		if err := json.Unmarshal(w.Labels, &labels); err == nil {
			apiWorker.Labels = api.NewOptRegisteredWorkerLabels(labels)
		}
	}

	// Map timestamps
	if w.LastHeartbeat.Valid {
		apiWorker.LastHeartbeat = api.NewOptDateTime(w.LastHeartbeat.Time)
	}
	if w.LastConnectedAt.Valid {
		apiWorker.LastConnectedAt = api.NewOptDateTime(w.LastConnectedAt.Time)
	}
	if w.ConnectedFrom.Valid {
		apiWorker.ConnectedFrom = api.NewOptString(w.ConnectedFrom.String)
	}
	if w.CreatedAt.Valid {
		apiWorker.CreatedAt = api.NewOptDateTime(w.CreatedAt.Time)
	}
	if w.UpdatedAt.Valid {
		apiWorker.UpdatedAt = api.NewOptDateTime(w.UpdatedAt.Time)
	}

	// Get workspace UUIDs if not global
	if !w.IsGlobal {
		links, err := q.ListWorkerWorkspaceLinks(ctx, &w.UUID)
		if err == nil {
			wsUUIDs := make([]string, 0, len(links))
			for _, link := range links {
				if link.WorkspaceUUID != nil {
					wsUUIDs = append(wsUUIDs, link.WorkspaceUUID.String())
				}
			}
			apiWorker.WorkspaceUuids = wsUUIDs
		}
	}

	return apiWorker, nil
}

// mapEnrollmentTokenToAPI converts a database enrollment token to API token
func (h *Handler) mapEnrollmentTokenToAPI(t query.WorkerEnrollmentToken, rawToken string) api.WorkerEnrollmentToken {
	apiToken := api.WorkerEnrollmentToken{
		UUID:     api.NewOptString(t.UUID.String()),
		Name:     t.Name,
		IsGlobal: api.NewOptBool(t.IsGlobal),
	}

	// Only include raw token if provided (on creation)
	if rawToken != "" {
		apiToken.Token = api.NewOptString(rawToken)
	}

	// Map workspace UUIDs
	wsUUIDs := make([]string, 0, len(t.WorkspaceUuids))
	for _, wsUUID := range t.WorkspaceUuids {
		if wsUUID.Valid {
			wsUUIDs = append(wsUUIDs, uuid.UUID(wsUUID.Bytes).String())
		}
	}
	apiToken.WorkspaceUuids = wsUUIDs

	// Map timestamps
	if t.ExpiresAt.Valid {
		apiToken.ExpiresAt = api.NewOptDateTime(t.ExpiresAt.Time)
	}
	if t.UsedAt.Valid {
		apiToken.UsedAt = api.NewOptDateTime(t.UsedAt.Time)
	}
	if t.UsedByWorkerUuid != nil {
		apiToken.UsedByWorkerUUID = api.NewOptString(t.UsedByWorkerUuid.String())
	}
	if t.CreatedByUserUuid != nil {
		apiToken.CreatedByUserUUID = api.NewOptString(t.CreatedByUserUuid.String())
	}
	if t.CreatedAt.Valid {
		apiToken.CreatedAt = api.NewOptDateTime(t.CreatedAt.Time)
	}

	return apiToken
}
