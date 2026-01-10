package usagelimits

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/samber/do/v2"

	"github.com/shadowapi/shadowapi/backend/pkg/query"
)

// Manager handles usage limit checking and tracking.
type Manager struct {
	dbp *pgxpool.Pool
	log *slog.Logger
}

// Provide creates a new Manager for the dependency injector.
func Provide(i do.Injector) (*Manager, error) {
	log := do.MustInvoke[*slog.Logger](i)
	dbp := do.MustInvoke[*pgxpool.Pool](i)

	log.Info("initializing usage limits manager")

	return &Manager{
		dbp: dbp,
		log: log,
	}, nil
}

// googleUUIDToPg converts a google/uuid.UUID to pgtype.UUID
func googleUUIDToPg(u uuid.UUID) pgtype.UUID {
	return pgtype.UUID{Bytes: u, Valid: true}
}

// gofrsUUIDToPg converts a gofrs/uuid.UUID (returned by SQLC) to pgtype.UUID
func gofrsUUIDToPg(u [16]byte) pgtype.UUID {
	return pgtype.UUID{Bytes: u, Valid: true}
}

// CheckAndReserve checks both user and worker quotas and reserves the minimum.
// Returns the allowed count (0 if both exceeded) or error.
func (m *Manager) CheckAndReserve(ctx context.Context, params CheckParams) (allowed int64, err error) {
	now := time.Now().UTC()

	// Get effective limits for user and worker
	userLimit, err := m.GetEffectiveUserLimit(ctx, params.UserUUID, params.WorkspaceSlug, params.LimitType)
	if err != nil {
		return 0, err
	}

	workerLimit, err := m.GetWorkerLimit(ctx, params.WorkerUUID, params.WorkspaceSlug, params.LimitType)
	if err != nil {
		return 0, err
	}

	// Calculate remaining for each
	var userRemaining, workerRemaining *int64

	if userLimit != nil && userLimit.IsEnabled && userLimit.LimitValue != nil {
		usage, err := m.getCurrentUserUsage(ctx, params.UserUUID, params.WorkspaceSlug, params.LimitType, userLimit.ResetPeriod, now)
		if err != nil {
			return 0, err
		}
		remaining := *userLimit.LimitValue - usage
		if remaining < 0 {
			remaining = 0
		}
		userRemaining = &remaining
	}

	if workerLimit != nil && workerLimit.IsEnabled && workerLimit.LimitValue != nil {
		usage, err := m.getCurrentWorkerUsage(ctx, params.WorkerUUID, params.WorkspaceSlug, params.LimitType, workerLimit.ResetPeriod, now)
		if err != nil {
			return 0, err
		}
		remaining := *workerLimit.LimitValue - usage
		if remaining < 0 {
			remaining = 0
		}
		workerRemaining = &remaining
	}

	// Determine the effective allowed amount
	allowed = params.Requested

	if userRemaining != nil && *userRemaining < allowed {
		allowed = *userRemaining
	}
	if workerRemaining != nil && *workerRemaining < allowed {
		allowed = *workerRemaining
	}

	if allowed <= 0 {
		return 0, nil
	}

	// Reserve the quota (increment usage now, will be adjusted after job completion)
	if userLimit != nil && userLimit.IsEnabled && userLimit.LimitValue != nil {
		if err := m.incrementUserUsage(ctx, params.UserUUID, params.WorkspaceSlug, params.LimitType, userLimit.ResetPeriod, now, allowed); err != nil {
			return 0, err
		}
	}

	if workerLimit != nil && workerLimit.IsEnabled && workerLimit.LimitValue != nil {
		if err := m.incrementWorkerUsage(ctx, params.WorkerUUID, params.WorkspaceSlug, params.LimitType, workerLimit.ResetPeriod, now, allowed); err != nil {
			// Try to rollback user usage
			_ = m.decrementUserUsage(ctx, params.UserUUID, params.WorkspaceSlug, params.LimitType, userLimit.ResetPeriod, now, allowed)
			return 0, err
		}
	}

	return allowed, nil
}

// GetEffectiveUserLimit returns the effective limit for a user in a workspace.
// Priority: user override > policy set default (highest limit from assigned sets)
func (m *Manager) GetEffectiveUserLimit(ctx context.Context, userUUID uuid.UUID, workspaceSlug, limitType string) (*EffectiveLimit, error) {
	q := query.New(m.dbp)

	result, err := q.GetEffectiveUserLimit(ctx, query.GetEffectiveUserLimitParams{
		UserUUID:      googleUUIDToPg(userUUID),
		WorkspaceSlug: pgtype.Text{String: workspaceSlug, Valid: true},
		LimitType:     limitType,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// No limit defined
			return nil, nil
		}
		return nil, err
	}

	// If no limit is defined at all
	if !result.IsEnabled {
		return nil, nil
	}

	var limitValue *int64
	if result.LimitValue.Valid {
		limitValue = &result.LimitValue.Int64
	}

	return &EffectiveLimit{
		LimitValue:  limitValue,
		ResetPeriod: ResetPeriod(result.ResetPeriod),
		IsEnabled:   result.IsEnabled,
	}, nil
}

// GetWorkerLimit returns the worker's limit for a workspace.
func (m *Manager) GetWorkerLimit(ctx context.Context, workerUUID uuid.UUID, workspaceSlug, limitType string) (*EffectiveLimit, error) {
	q := query.New(m.dbp)

	result, err := q.GetWorkerUsageLimitByKey(ctx, query.GetWorkerUsageLimitByKeyParams{
		WorkerUUID:    googleUUIDToPg(workerUUID),
		WorkspaceSlug: workspaceSlug,
		LimitType:     limitType,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	if !result.IsEnabled {
		return nil, nil
	}

	var limitValue *int64
	if result.LimitValue.Valid {
		limitValue = &result.LimitValue.Int64
	}

	return &EffectiveLimit{
		LimitValue:  limitValue,
		ResetPeriod: ResetPeriod(result.ResetPeriod),
		IsEnabled:   result.IsEnabled,
	}, nil
}

// GetUsageStatus returns combined usage status for user and worker in workspace.
func (m *Manager) GetUsageStatus(ctx context.Context, userUUID, workerUUID uuid.UUID, workspaceSlug, limitType string) (*UsageStatus, error) {
	now := time.Now().UTC()
	status := &UsageStatus{}

	// Get user limit status
	userLimit, err := m.GetEffectiveUserLimit(ctx, userUUID, workspaceSlug, limitType)
	if err != nil {
		return nil, err
	}

	if userLimit != nil && userLimit.IsEnabled {
		periodStart, periodEnd := CalculatePeriodBounds(userLimit.ResetPeriod, now)
		usage, err := m.getCurrentUserUsage(ctx, userUUID, workspaceSlug, limitType, userLimit.ResetPeriod, now)
		if err != nil {
			return nil, err
		}

		ls := &LimitStatus{
			LimitValue:   userLimit.LimitValue,
			CurrentUsage: usage,
			ResetPeriod:  userLimit.ResetPeriod,
			PeriodStart:  periodStart,
			PeriodEnd:    periodEnd,
			IsLimited:    userLimit.LimitValue != nil,
		}

		if userLimit.LimitValue != nil {
			remaining := *userLimit.LimitValue - usage
			if remaining < 0 {
				remaining = 0
			}
			ls.Remaining = &remaining
		}

		status.UserLimit = ls
	}

	// Get worker limit status
	workerLimit, err := m.GetWorkerLimit(ctx, workerUUID, workspaceSlug, limitType)
	if err != nil {
		return nil, err
	}

	if workerLimit != nil && workerLimit.IsEnabled {
		periodStart, periodEnd := CalculatePeriodBounds(workerLimit.ResetPeriod, now)
		usage, err := m.getCurrentWorkerUsage(ctx, workerUUID, workspaceSlug, limitType, workerLimit.ResetPeriod, now)
		if err != nil {
			return nil, err
		}

		ls := &LimitStatus{
			LimitValue:   workerLimit.LimitValue,
			CurrentUsage: usage,
			ResetPeriod:  workerLimit.ResetPeriod,
			PeriodStart:  periodStart,
			PeriodEnd:    periodEnd,
			IsLimited:    workerLimit.LimitValue != nil,
		}

		if workerLimit.LimitValue != nil {
			remaining := *workerLimit.LimitValue - usage
			if remaining < 0 {
				remaining = 0
			}
			ls.Remaining = &remaining
		}

		status.WorkerLimit = ls
	}

	// Calculate effective remaining
	if status.UserLimit != nil && status.UserLimit.Remaining != nil {
		if status.EffectiveRemaining == nil || *status.UserLimit.Remaining < *status.EffectiveRemaining {
			status.EffectiveRemaining = status.UserLimit.Remaining
		}
	}
	if status.WorkerLimit != nil && status.WorkerLimit.Remaining != nil {
		if status.EffectiveRemaining == nil || *status.WorkerLimit.Remaining < *status.EffectiveRemaining {
			status.EffectiveRemaining = status.WorkerLimit.Remaining
		}
	}

	return status, nil
}

// RecordUsage records actual usage for both user and worker.
// This should be called after job completion to adjust the reserved quota.
func (m *Manager) RecordUsage(ctx context.Context, userUUID, workerUUID uuid.UUID, workspaceSlug, limitType string, actualCount int64) error {
	// For now, usage is already recorded during CheckAndReserve
	// This method can be used to adjust if actual differs from reserved
	// Implementation can be enhanced later for more accurate tracking
	return nil
}

// ReleaseReservation releases quota that was reserved but not used (e.g., job failed).
func (m *Manager) ReleaseReservation(ctx context.Context, userUUID, workerUUID uuid.UUID, workspaceSlug, limitType string, count int64) error {
	now := time.Now().UTC()

	// Get user limit to know the period
	userLimit, err := m.GetEffectiveUserLimit(ctx, userUUID, workspaceSlug, limitType)
	if err != nil {
		return err
	}
	if userLimit != nil && userLimit.IsEnabled && userLimit.LimitValue != nil {
		if err := m.decrementUserUsage(ctx, userUUID, workspaceSlug, limitType, userLimit.ResetPeriod, now, count); err != nil {
			m.log.Warn("failed to release user reservation", "error", err)
		}
	}

	// Get worker limit to know the period
	workerLimit, err := m.GetWorkerLimit(ctx, workerUUID, workspaceSlug, limitType)
	if err != nil {
		return err
	}
	if workerLimit != nil && workerLimit.IsEnabled && workerLimit.LimitValue != nil {
		if err := m.decrementWorkerUsage(ctx, workerUUID, workspaceSlug, limitType, workerLimit.ResetPeriod, now, count); err != nil {
			m.log.Warn("failed to release worker reservation", "error", err)
		}
	}

	return nil
}

// Internal helper methods

func (m *Manager) getCurrentUserUsage(ctx context.Context, userUUID uuid.UUID, workspaceSlug, limitType string, period ResetPeriod, now time.Time) (int64, error) {
	q := query.New(m.dbp)

	tracking, err := q.GetCurrentUserUsage(ctx, query.GetCurrentUserUsageParams{
		UserUUID:      googleUUIDToPg(userUUID),
		WorkspaceSlug: workspaceSlug,
		LimitType:     limitType,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, nil
		}
		return 0, err
	}

	return tracking.CurrentUsage, nil
}

func (m *Manager) getCurrentWorkerUsage(ctx context.Context, workerUUID uuid.UUID, workspaceSlug, limitType string, period ResetPeriod, now time.Time) (int64, error) {
	q := query.New(m.dbp)

	tracking, err := q.GetCurrentWorkerUsage(ctx, query.GetCurrentWorkerUsageParams{
		WorkerUUID:    googleUUIDToPg(workerUUID),
		WorkspaceSlug: workspaceSlug,
		LimitType:     limitType,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, nil
		}
		return 0, err
	}

	return tracking.CurrentUsage, nil
}

func (m *Manager) incrementUserUsage(ctx context.Context, userUUID uuid.UUID, workspaceSlug, limitType string, period ResetPeriod, now time.Time, count int64) error {
	q := query.New(m.dbp)
	periodStart, periodEnd := CalculatePeriodBounds(period, now)

	// Try to get existing tracking record
	tracking, err := q.GetUserUsageForPeriod(ctx, query.GetUserUsageForPeriodParams{
		UserUUID:      googleUUIDToPg(userUUID),
		WorkspaceSlug: workspaceSlug,
		LimitType:     limitType,
		PeriodStart:   pgtype.Timestamptz{Time: periodStart, Valid: true},
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// Create new tracking record
			_, err = q.CreateUserUsageTracking(ctx, query.CreateUserUsageTrackingParams{
				UUID:          googleUUIDToPg(uuid.New()),
				UserUUID:      googleUUIDToPg(userUUID),
				WorkspaceSlug: workspaceSlug,
				LimitType:     limitType,
				PeriodStart:   pgtype.Timestamptz{Time: periodStart, Valid: true},
				PeriodEnd:     pgtype.Timestamptz{Time: periodEnd, Valid: true},
				CurrentUsage:  count,
			})
			return err
		}
		return err
	}

	// Increment existing record - tracking.UUID is gofrs/uuid.UUID which is [16]byte
	_, err = q.IncrementUserUsage(ctx, query.IncrementUserUsageParams{
		UUID:      gofrsUUIDToPg(tracking.UUID),
		Increment: count,
	})
	return err
}

func (m *Manager) decrementUserUsage(ctx context.Context, userUUID uuid.UUID, workspaceSlug, limitType string, period ResetPeriod, now time.Time, count int64) error {
	q := query.New(m.dbp)
	periodStart, _ := CalculatePeriodBounds(period, now)

	tracking, err := q.GetUserUsageForPeriod(ctx, query.GetUserUsageForPeriodParams{
		UserUUID:      googleUUIDToPg(userUUID),
		WorkspaceSlug: workspaceSlug,
		LimitType:     limitType,
		PeriodStart:   pgtype.Timestamptz{Time: periodStart, Valid: true},
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil // Nothing to decrement
		}
		return err
	}

	_, err = q.DecrementUserUsage(ctx, query.DecrementUserUsageParams{
		UUID:      gofrsUUIDToPg(tracking.UUID),
		Decrement: count,
	})
	return err
}

func (m *Manager) incrementWorkerUsage(ctx context.Context, workerUUID uuid.UUID, workspaceSlug, limitType string, period ResetPeriod, now time.Time, count int64) error {
	q := query.New(m.dbp)
	periodStart, periodEnd := CalculatePeriodBounds(period, now)

	// Try to get existing tracking record
	tracking, err := q.GetWorkerUsageForPeriod(ctx, query.GetWorkerUsageForPeriodParams{
		WorkerUUID:    googleUUIDToPg(workerUUID),
		WorkspaceSlug: workspaceSlug,
		LimitType:     limitType,
		PeriodStart:   pgtype.Timestamptz{Time: periodStart, Valid: true},
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// Create new tracking record
			_, err = q.CreateWorkerUsageTracking(ctx, query.CreateWorkerUsageTrackingParams{
				UUID:          googleUUIDToPg(uuid.New()),
				WorkerUUID:    googleUUIDToPg(workerUUID),
				WorkspaceSlug: workspaceSlug,
				LimitType:     limitType,
				PeriodStart:   pgtype.Timestamptz{Time: periodStart, Valid: true},
				PeriodEnd:     pgtype.Timestamptz{Time: periodEnd, Valid: true},
				CurrentUsage:  count,
			})
			return err
		}
		return err
	}

	// Increment existing record
	_, err = q.IncrementWorkerUsage(ctx, query.IncrementWorkerUsageParams{
		UUID:      gofrsUUIDToPg(tracking.UUID),
		Increment: count,
	})
	return err
}

func (m *Manager) decrementWorkerUsage(ctx context.Context, workerUUID uuid.UUID, workspaceSlug, limitType string, period ResetPeriod, now time.Time, count int64) error {
	q := query.New(m.dbp)
	periodStart, _ := CalculatePeriodBounds(period, now)

	tracking, err := q.GetWorkerUsageForPeriod(ctx, query.GetWorkerUsageForPeriodParams{
		WorkerUUID:    googleUUIDToPg(workerUUID),
		WorkspaceSlug: workspaceSlug,
		LimitType:     limitType,
		PeriodStart:   pgtype.Timestamptz{Time: periodStart, Valid: true},
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil // Nothing to decrement
		}
		return err
	}

	_, err = q.DecrementWorkerUsage(ctx, query.DecrementWorkerUsageParams{
		UUID:      gofrsUUIDToPg(tracking.UUID),
		Decrement: count,
	})
	return err
}
