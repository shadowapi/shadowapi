package handler

import (
	"context"
	"net/http"

	gofrs "github.com/gofrs/uuid"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/shadowapi/shadowapi/backend/pkg/api"
	"github.com/shadowapi/shadowapi/backend/pkg/query"
)

// googleUUIDToPg converts a google/uuid.UUID to pgtype.UUID
func googleUUIDToPg(u uuid.UUID) pgtype.UUID {
	return pgtype.UUID{Bytes: u, Valid: true}
}

// gofrsToGoogleUUID converts a gofrs/uuid.UUID to google/uuid.UUID
func gofrsToGoogleUUID(u gofrs.UUID) uuid.UUID {
	return uuid.UUID(u)
}

// ============================================================================
// Usage Limit (Policy Set Defaults)
// ============================================================================

// ListUsageLimits implements listUsageLimits operation.
func (h *Handler) ListUsageLimits(ctx context.Context, params api.ListUsageLimitsParams) (api.ListUsageLimitsRes, error) {
	q := query.New(h.dbp)

	var limits []query.UsageLimit
	var err error

	if params.PolicySetName.IsSet() {
		limits, err = q.GetUsageLimitsByPolicySet(ctx, params.PolicySetName.Value)
	} else {
		limits, err = q.ListUsageLimits(ctx)
	}

	if err != nil {
		h.log.Error("failed to list usage limits", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to list usage limits"))
	}

	result := make([]api.UsageLimit, 0, len(limits))
	for _, l := range limits {
		result = append(result, qUsageLimitToAPI(l))
	}

	res := api.ListUsageLimitsOKApplicationJSON(result)
	return &res, nil
}

// CreateUsageLimit implements createUsageLimit operation.
func (h *Handler) CreateUsageLimit(ctx context.Context, req *api.UsageLimit) (api.CreateUsageLimitRes, error) {
	q := query.New(h.dbp)

	limitUUID := uuid.New()

	// Set defaults
	resetPeriod := "monthly"
	if req.ResetPeriod.IsSet() {
		resetPeriod = string(req.ResetPeriod.Value)
	}

	isEnabled := true
	if req.IsEnabled.IsSet() {
		isEnabled = req.IsEnabled.Value
	}

	var limitValue pgtype.Int8
	if req.LimitValue.IsSet() && !req.LimitValue.Null {
		limitValue = pgtype.Int8{Int64: req.LimitValue.Value, Valid: true}
	}

	limit, err := q.CreateUsageLimit(ctx, query.CreateUsageLimitParams{
		UUID:          googleUUIDToPg(limitUUID),
		PolicySetName: req.PolicySetName,
		LimitType:     string(req.LimitType),
		LimitValue:    limitValue,
		ResetPeriod:   resetPeriod,
		IsEnabled:     isEnabled,
	})
	if err != nil {
		h.log.Error("failed to create usage limit", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to create usage limit"))
	}

	result := qUsageLimitToAPI(limit)
	return &result, nil
}

// GetUsageLimit implements getUsageLimit operation.
func (h *Handler) GetUsageLimit(ctx context.Context, params api.GetUsageLimitParams) (api.GetUsageLimitRes, error) {
	q := query.New(h.dbp)

	limit, err := q.GetUsageLimit(ctx, googleUUIDToPg(params.UUID))
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrWithCode(http.StatusNotFound, E("usage limit not found"))
		}
		h.log.Error("failed to get usage limit", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to get usage limit"))
	}

	result := qUsageLimitToAPI(limit)
	return &result, nil
}

// UpdateUsageLimit implements updateUsageLimit operation.
func (h *Handler) UpdateUsageLimit(ctx context.Context, req *api.UsageLimit, params api.UpdateUsageLimitParams) (api.UpdateUsageLimitRes, error) {
	q := query.New(h.dbp)

	// Check if exists
	existing, err := q.GetUsageLimit(ctx, googleUUIDToPg(params.UUID))
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrWithCode(http.StatusNotFound, E("usage limit not found"))
		}
		h.log.Error("failed to get usage limit", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to get usage limit"))
	}

	// Use existing values as defaults
	resetPeriod := existing.ResetPeriod
	if req.ResetPeriod.IsSet() {
		resetPeriod = string(req.ResetPeriod.Value)
	}

	isEnabled := existing.IsEnabled
	if req.IsEnabled.IsSet() {
		isEnabled = req.IsEnabled.Value
	}

	var limitValue pgtype.Int8
	if req.LimitValue.IsSet() {
		if !req.LimitValue.Null {
			limitValue = pgtype.Int8{Int64: req.LimitValue.Value, Valid: true}
		}
	} else {
		limitValue = existing.LimitValue
	}

	limit, err := q.UpdateUsageLimit(ctx, query.UpdateUsageLimitParams{
		UUID:        googleUUIDToPg(params.UUID),
		LimitValue:  limitValue,
		ResetPeriod: resetPeriod,
		IsEnabled:   isEnabled,
	})
	if err != nil {
		h.log.Error("failed to update usage limit", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to update usage limit"))
	}

	result := qUsageLimitToAPI(limit)
	return &result, nil
}

// DeleteUsageLimit implements deleteUsageLimit operation.
func (h *Handler) DeleteUsageLimit(ctx context.Context, params api.DeleteUsageLimitParams) (api.DeleteUsageLimitRes, error) {
	q := query.New(h.dbp)

	err := q.DeleteUsageLimit(ctx, googleUUIDToPg(params.UUID))
	if err != nil {
		h.log.Error("failed to delete usage limit", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to delete usage limit"))
	}

	return &api.DeleteUsageLimitNoContent{}, nil
}

// ============================================================================
// User Usage Limit Override
// ============================================================================

// ListUserUsageLimitOverrides implements listUserUsageLimitOverrides operation.
func (h *Handler) ListUserUsageLimitOverrides(ctx context.Context, params api.ListUserUsageLimitOverridesParams) (api.ListUserUsageLimitOverridesRes, error) {
	q := query.New(h.dbp)

	var overrides []query.UserUsageLimitOverride
	var err error

	if params.WorkspaceSlug.IsSet() {
		overrides, err = q.ListUserUsageLimitOverridesByWorkspace(ctx, query.ListUserUsageLimitOverridesByWorkspaceParams{
			UserUUID:      googleUUIDToPg(params.UserUUID),
			WorkspaceSlug: params.WorkspaceSlug.Value,
		})
	} else {
		overrides, err = q.ListUserUsageLimitOverrides(ctx, googleUUIDToPg(params.UserUUID))
	}

	if err != nil {
		h.log.Error("failed to list user usage limit overrides", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to list user usage limit overrides"))
	}

	result := make([]api.UserUsageLimitOverride, 0, len(overrides))
	for _, o := range overrides {
		result = append(result, qUserUsageLimitOverrideToAPI(o))
	}

	res := api.ListUserUsageLimitOverridesOKApplicationJSON(result)
	return &res, nil
}

// CreateUserUsageLimitOverride implements createUserUsageLimitOverride operation.
func (h *Handler) CreateUserUsageLimitOverride(ctx context.Context, req *api.UserUsageLimitOverride, params api.CreateUserUsageLimitOverrideParams) (api.CreateUserUsageLimitOverrideRes, error) {
	q := query.New(h.dbp)

	overrideUUID := uuid.New()

	var resetPeriod pgtype.Text
	if req.ResetPeriod.IsSet() && !req.ResetPeriod.Null {
		resetPeriod = pgtype.Text{String: string(req.ResetPeriod.Value), Valid: true}
	}

	isEnabled := true
	if req.IsEnabled.IsSet() {
		isEnabled = req.IsEnabled.Value
	}

	var limitValue pgtype.Int8
	if req.LimitValue.IsSet() && !req.LimitValue.Null {
		limitValue = pgtype.Int8{Int64: req.LimitValue.Value, Valid: true}
	}

	override, err := q.CreateUserUsageLimitOverride(ctx, query.CreateUserUsageLimitOverrideParams{
		UUID:          googleUUIDToPg(overrideUUID),
		UserUUID:      googleUUIDToPg(params.UserUUID),
		WorkspaceSlug: req.WorkspaceSlug,
		LimitType:     string(req.LimitType),
		LimitValue:    limitValue,
		ResetPeriod:   resetPeriod,
		IsEnabled:     isEnabled,
	})
	if err != nil {
		h.log.Error("failed to create user usage limit override", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to create user usage limit override"))
	}

	result := qUserUsageLimitOverrideToAPI(override)
	return &result, nil
}

// UpdateUserUsageLimitOverride implements updateUserUsageLimitOverride operation.
func (h *Handler) UpdateUserUsageLimitOverride(ctx context.Context, req *api.UserUsageLimitOverride, params api.UpdateUserUsageLimitOverrideParams) (api.UpdateUserUsageLimitOverrideRes, error) {
	q := query.New(h.dbp)

	// Check if exists
	existing, err := q.GetUserUsageLimitOverride(ctx, googleUUIDToPg(params.UUID))
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrWithCode(http.StatusNotFound, E("user usage limit override not found"))
		}
		h.log.Error("failed to get user usage limit override", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to get user usage limit override"))
	}

	resetPeriod := existing.ResetPeriod
	if req.ResetPeriod.IsSet() {
		if req.ResetPeriod.Null {
			resetPeriod = pgtype.Text{Valid: false}
		} else {
			resetPeriod = pgtype.Text{String: string(req.ResetPeriod.Value), Valid: true}
		}
	}

	isEnabled := existing.IsEnabled
	if req.IsEnabled.IsSet() {
		isEnabled = req.IsEnabled.Value
	}

	var limitValue pgtype.Int8
	if req.LimitValue.IsSet() {
		if !req.LimitValue.Null {
			limitValue = pgtype.Int8{Int64: req.LimitValue.Value, Valid: true}
		}
	} else {
		limitValue = existing.LimitValue
	}

	override, err := q.UpdateUserUsageLimitOverride(ctx, query.UpdateUserUsageLimitOverrideParams{
		UUID:        googleUUIDToPg(params.UUID),
		LimitValue:  limitValue,
		ResetPeriod: resetPeriod,
		IsEnabled:   isEnabled,
	})
	if err != nil {
		h.log.Error("failed to update user usage limit override", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to update user usage limit override"))
	}

	result := qUserUsageLimitOverrideToAPI(override)
	return &result, nil
}

// DeleteUserUsageLimitOverride implements deleteUserUsageLimitOverride operation.
func (h *Handler) DeleteUserUsageLimitOverride(ctx context.Context, params api.DeleteUserUsageLimitOverrideParams) (api.DeleteUserUsageLimitOverrideRes, error) {
	q := query.New(h.dbp)

	err := q.DeleteUserUsageLimitOverride(ctx, googleUUIDToPg(params.UUID))
	if err != nil {
		h.log.Error("failed to delete user usage limit override", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to delete user usage limit override"))
	}

	return &api.DeleteUserUsageLimitOverrideNoContent{}, nil
}

// ============================================================================
// Worker Usage Limit
// ============================================================================

// ListWorkerUsageLimits implements listWorkerUsageLimits operation.
func (h *Handler) ListWorkerUsageLimits(ctx context.Context, params api.ListWorkerUsageLimitsParams) (api.ListWorkerUsageLimitsRes, error) {
	q := query.New(h.dbp)

	var limits []query.WorkerUsageLimit
	var err error

	if params.WorkspaceSlug.IsSet() {
		limits, err = q.ListWorkerUsageLimitsByWorkspace(ctx, query.ListWorkerUsageLimitsByWorkspaceParams{
			WorkerUUID:    googleUUIDToPg(params.WorkerUUID),
			WorkspaceSlug: params.WorkspaceSlug.Value,
		})
	} else {
		limits, err = q.ListWorkerUsageLimits(ctx, googleUUIDToPg(params.WorkerUUID))
	}

	if err != nil {
		h.log.Error("failed to list worker usage limits", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to list worker usage limits"))
	}

	result := make([]api.WorkerUsageLimit, 0, len(limits))
	for _, l := range limits {
		result = append(result, qWorkerUsageLimitToAPI(l))
	}

	res := api.ListWorkerUsageLimitsOKApplicationJSON(result)
	return &res, nil
}

// CreateWorkerUsageLimit implements createWorkerUsageLimit operation.
func (h *Handler) CreateWorkerUsageLimit(ctx context.Context, req *api.WorkerUsageLimit, params api.CreateWorkerUsageLimitParams) (api.CreateWorkerUsageLimitRes, error) {
	q := query.New(h.dbp)

	limitUUID := uuid.New()

	resetPeriod := "monthly"
	if req.ResetPeriod.IsSet() {
		resetPeriod = string(req.ResetPeriod.Value)
	}

	isEnabled := true
	if req.IsEnabled.IsSet() {
		isEnabled = req.IsEnabled.Value
	}

	var limitValue pgtype.Int8
	if req.LimitValue.IsSet() && !req.LimitValue.Null {
		limitValue = pgtype.Int8{Int64: req.LimitValue.Value, Valid: true}
	}

	limit, err := q.CreateWorkerUsageLimit(ctx, query.CreateWorkerUsageLimitParams{
		UUID:          googleUUIDToPg(limitUUID),
		WorkerUUID:    googleUUIDToPg(params.WorkerUUID),
		WorkspaceSlug: req.WorkspaceSlug,
		LimitType:     string(req.LimitType),
		LimitValue:    limitValue,
		ResetPeriod:   resetPeriod,
		IsEnabled:     isEnabled,
	})
	if err != nil {
		h.log.Error("failed to create worker usage limit", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to create worker usage limit"))
	}

	result := qWorkerUsageLimitToAPI(limit)
	return &result, nil
}

// UpdateWorkerUsageLimit implements updateWorkerUsageLimit operation.
func (h *Handler) UpdateWorkerUsageLimit(ctx context.Context, req *api.WorkerUsageLimit, params api.UpdateWorkerUsageLimitParams) (api.UpdateWorkerUsageLimitRes, error) {
	q := query.New(h.dbp)

	// Check if exists
	existing, err := q.GetWorkerUsageLimit(ctx, googleUUIDToPg(params.UUID))
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrWithCode(http.StatusNotFound, E("worker usage limit not found"))
		}
		h.log.Error("failed to get worker usage limit", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to get worker usage limit"))
	}

	resetPeriod := existing.ResetPeriod
	if req.ResetPeriod.IsSet() {
		resetPeriod = string(req.ResetPeriod.Value)
	}

	isEnabled := existing.IsEnabled
	if req.IsEnabled.IsSet() {
		isEnabled = req.IsEnabled.Value
	}

	var limitValue pgtype.Int8
	if req.LimitValue.IsSet() {
		if !req.LimitValue.Null {
			limitValue = pgtype.Int8{Int64: req.LimitValue.Value, Valid: true}
		}
	} else {
		limitValue = existing.LimitValue
	}

	limit, err := q.UpdateWorkerUsageLimit(ctx, query.UpdateWorkerUsageLimitParams{
		UUID:        googleUUIDToPg(params.UUID),
		LimitValue:  limitValue,
		ResetPeriod: resetPeriod,
		IsEnabled:   isEnabled,
	})
	if err != nil {
		h.log.Error("failed to update worker usage limit", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to update worker usage limit"))
	}

	result := qWorkerUsageLimitToAPI(limit)
	return &result, nil
}

// DeleteWorkerUsageLimit implements deleteWorkerUsageLimit operation.
func (h *Handler) DeleteWorkerUsageLimit(ctx context.Context, params api.DeleteWorkerUsageLimitParams) (api.DeleteWorkerUsageLimitRes, error) {
	q := query.New(h.dbp)

	err := q.DeleteWorkerUsageLimit(ctx, googleUUIDToPg(params.UUID))
	if err != nil {
		h.log.Error("failed to delete worker usage limit", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to delete worker usage limit"))
	}

	return &api.DeleteWorkerUsageLimitNoContent{}, nil
}

// ============================================================================
// Usage Status
// ============================================================================

// GetUsageStatus implements getUsageStatus operation.
func (h *Handler) GetUsageStatus(ctx context.Context, params api.GetUsageStatusParams) (api.GetUsageStatusRes, error) {
	status, err := h.usageLimits.GetUsageStatus(
		ctx,
		params.UserUUID,
		params.WorkerUUID,
		params.WorkspaceSlug,
		string(params.LimitType),
	)
	if err != nil {
		h.log.Error("failed to get usage status", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to get usage status"))
	}

	result := api.UsageStatus{}

	if status.UserLimit != nil {
		ul := api.UsageStatusUserLimit{
			CurrentUsage: api.NewOptInt64(status.UserLimit.CurrentUsage),
			ResetPeriod:  api.NewOptUsageStatusUserLimitResetPeriod(api.UsageStatusUserLimitResetPeriod(status.UserLimit.ResetPeriod)),
			PeriodStart:  api.NewOptDateTime(status.UserLimit.PeriodStart),
			PeriodEnd:    api.NewOptDateTime(status.UserLimit.PeriodEnd),
			IsLimited:    api.NewOptBool(status.UserLimit.IsLimited),
		}
		if status.UserLimit.LimitValue != nil {
			ul.LimitValue = api.NewOptNilInt64(*status.UserLimit.LimitValue)
		}
		if status.UserLimit.Remaining != nil {
			ul.Remaining = api.NewOptNilInt64(*status.UserLimit.Remaining)
		}
		result.UserLimit = api.NewOptUsageStatusUserLimit(ul)
	}

	if status.WorkerLimit != nil {
		wl := api.UsageStatusWorkerLimit{
			CurrentUsage: api.NewOptInt64(status.WorkerLimit.CurrentUsage),
			ResetPeriod:  api.NewOptUsageStatusWorkerLimitResetPeriod(api.UsageStatusWorkerLimitResetPeriod(status.WorkerLimit.ResetPeriod)),
			PeriodStart:  api.NewOptDateTime(status.WorkerLimit.PeriodStart),
			PeriodEnd:    api.NewOptDateTime(status.WorkerLimit.PeriodEnd),
			IsLimited:    api.NewOptBool(status.WorkerLimit.IsLimited),
		}
		if status.WorkerLimit.LimitValue != nil {
			wl.LimitValue = api.NewOptNilInt64(*status.WorkerLimit.LimitValue)
		}
		if status.WorkerLimit.Remaining != nil {
			wl.Remaining = api.NewOptNilInt64(*status.WorkerLimit.Remaining)
		}
		result.WorkerLimit = api.NewOptUsageStatusWorkerLimit(wl)
	}

	if status.EffectiveRemaining != nil {
		result.EffectiveRemaining = api.NewOptNilInt64(*status.EffectiveRemaining)
	}

	return &result, nil
}

// ============================================================================
// Converters
// ============================================================================

func qUsageLimitToAPI(l query.UsageLimit) api.UsageLimit {
	result := api.UsageLimit{
		UUID:          api.NewOptUUID(gofrsToGoogleUUID(l.UUID)),
		PolicySetName: l.PolicySetName,
		LimitType:     api.UsageLimitLimitType(l.LimitType),
		ResetPeriod:   api.NewOptUsageLimitResetPeriod(api.UsageLimitResetPeriod(l.ResetPeriod)),
		IsEnabled:     api.NewOptBool(l.IsEnabled),
	}

	if l.LimitValue.Valid {
		result.LimitValue = api.NewOptNilInt64(l.LimitValue.Int64)
	}

	if l.CreatedAt.Valid {
		result.CreatedAt = api.NewOptDateTime(l.CreatedAt.Time)
	}

	if l.UpdatedAt.Valid {
		result.UpdatedAt = api.NewOptDateTime(l.UpdatedAt.Time)
	}

	return result
}

func qUserUsageLimitOverrideToAPI(o query.UserUsageLimitOverride) api.UserUsageLimitOverride {
	result := api.UserUsageLimitOverride{
		UUID:          api.NewOptUUID(gofrsToGoogleUUID(o.UUID)),
		WorkspaceSlug: o.WorkspaceSlug,
		LimitType:     api.UserUsageLimitOverrideLimitType(o.LimitType),
		IsEnabled:     api.NewOptBool(o.IsEnabled),
	}

	if o.UserUUID != nil {
		result.UserUUID = gofrsToGoogleUUID(*o.UserUUID)
	}

	if o.LimitValue.Valid {
		result.LimitValue = api.NewOptNilInt64(o.LimitValue.Int64)
	}

	if o.ResetPeriod.Valid {
		result.ResetPeriod = api.NewOptNilUserUsageLimitOverrideResetPeriod(api.UserUsageLimitOverrideResetPeriod(o.ResetPeriod.String))
	}

	if o.CreatedAt.Valid {
		result.CreatedAt = api.NewOptDateTime(o.CreatedAt.Time)
	}

	if o.UpdatedAt.Valid {
		result.UpdatedAt = api.NewOptDateTime(o.UpdatedAt.Time)
	}

	return result
}

func qWorkerUsageLimitToAPI(l query.WorkerUsageLimit) api.WorkerUsageLimit {
	result := api.WorkerUsageLimit{
		UUID:          api.NewOptUUID(gofrsToGoogleUUID(l.UUID)),
		WorkspaceSlug: l.WorkspaceSlug,
		LimitType:     api.WorkerUsageLimitLimitType(l.LimitType),
		ResetPeriod:   api.NewOptWorkerUsageLimitResetPeriod(api.WorkerUsageLimitResetPeriod(l.ResetPeriod)),
		IsEnabled:     api.NewOptBool(l.IsEnabled),
	}

	if l.WorkerUUID != nil {
		result.WorkerUUID = gofrsToGoogleUUID(*l.WorkerUUID)
	}

	if l.LimitValue.Valid {
		result.LimitValue = api.NewOptNilInt64(l.LimitValue.Int64)
	}

	if l.CreatedAt.Valid {
		result.CreatedAt = api.NewOptDateTime(l.CreatedAt.Time)
	}

	if l.UpdatedAt.Valid {
		result.UpdatedAt = api.NewOptDateTime(l.UpdatedAt.Time)
	}

	return result
}
