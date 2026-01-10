package usagelimits

import (
	"time"

	"github.com/google/uuid"
)

// LimitType represents the type of usage limit.
type LimitType string

const (
	LimitTypeMessagesFetch LimitType = "messages_fetch"
	LimitTypeMessagesPush  LimitType = "messages_push"
)

// ResetPeriod represents when the usage counter resets.
type ResetPeriod string

const (
	ResetPeriodDaily      ResetPeriod = "daily"
	ResetPeriodWeekly     ResetPeriod = "weekly"
	ResetPeriodMonthly    ResetPeriod = "monthly"
	ResetPeriodRolling24h ResetPeriod = "rolling_24h"
	ResetPeriodRolling7d  ResetPeriod = "rolling_7d"
	ResetPeriodRolling30d ResetPeriod = "rolling_30d"
)

// EffectiveLimit represents the resolved limit for a user or worker.
type EffectiveLimit struct {
	LimitValue  *int64      // nil means unlimited
	ResetPeriod ResetPeriod
	IsEnabled   bool
}

// LimitStatus represents the current status of a single limit.
type LimitStatus struct {
	LimitValue   *int64      // nil means unlimited
	CurrentUsage int64
	Remaining    *int64 // nil means unlimited
	ResetPeriod  ResetPeriod
	PeriodStart  time.Time
	PeriodEnd    time.Time
	IsLimited    bool // false if unlimited
}

// UsageStatus represents the combined usage status for user and worker.
type UsageStatus struct {
	UserLimit          *LimitStatus
	WorkerLimit        *LimitStatus
	EffectiveRemaining *int64 // min of both, nil if both unlimited
}

// CheckParams contains parameters for checking usage limits.
type CheckParams struct {
	UserUUID      uuid.UUID
	WorkerUUID    uuid.UUID
	WorkspaceSlug string
	LimitType     string
	Requested     int64
}

// IsValidLimitType checks if the given string is a valid limit type.
func IsValidLimitType(t string) bool {
	switch LimitType(t) {
	case LimitTypeMessagesFetch, LimitTypeMessagesPush:
		return true
	}
	return false
}

// IsValidResetPeriod checks if the given string is a valid reset period.
func IsValidResetPeriod(p string) bool {
	switch ResetPeriod(p) {
	case ResetPeriodDaily, ResetPeriodWeekly, ResetPeriodMonthly,
		ResetPeriodRolling24h, ResetPeriodRolling7d, ResetPeriodRolling30d:
		return true
	}
	return false
}
