-- ============================================================================
-- Usage Limit Queries (Policy Set Defaults)
-- ============================================================================

-- name: ListUsageLimits :many
SELECT * FROM usage_limit
ORDER BY policy_set_name, limit_type;

-- name: GetUsageLimitsByPolicySet :many
SELECT * FROM usage_limit
WHERE policy_set_name = sqlc.arg('policy_set_name')
ORDER BY limit_type;

-- name: GetUsageLimit :one
SELECT * FROM usage_limit
WHERE uuid = sqlc.arg('uuid')::uuid;

-- name: GetUsageLimitByPolicySetAndType :one
SELECT * FROM usage_limit
WHERE policy_set_name = sqlc.arg('policy_set_name')
  AND limit_type = sqlc.arg('limit_type');

-- name: CreateUsageLimit :one
INSERT INTO usage_limit (
    uuid,
    policy_set_name,
    limit_type,
    limit_value,
    reset_period,
    is_enabled,
    created_at
) VALUES (
    sqlc.arg('uuid')::uuid,
    sqlc.arg('policy_set_name'),
    sqlc.arg('limit_type'),
    sqlc.arg('limit_value'),
    sqlc.arg('reset_period'),
    sqlc.arg('is_enabled')::boolean,
    NOW()
) RETURNING *;

-- name: UpdateUsageLimit :one
UPDATE usage_limit
SET
    limit_value = sqlc.arg('limit_value'),
    reset_period = sqlc.arg('reset_period'),
    is_enabled = sqlc.arg('is_enabled')::boolean,
    updated_at = NOW()
WHERE uuid = sqlc.arg('uuid')::uuid
RETURNING *;

-- name: DeleteUsageLimit :exec
DELETE FROM usage_limit
WHERE uuid = sqlc.arg('uuid')::uuid;

-- name: UsageLimitExists :one
SELECT EXISTS (
    SELECT 1 FROM usage_limit
    WHERE policy_set_name = sqlc.arg('policy_set_name')
      AND limit_type = sqlc.arg('limit_type')
) AS exists;

-- ============================================================================
-- User Usage Limit Override Queries
-- ============================================================================

-- name: ListUserUsageLimitOverrides :many
SELECT * FROM user_usage_limit_override
WHERE user_uuid = sqlc.arg('user_uuid')::uuid
ORDER BY workspace_slug, limit_type;

-- name: ListUserUsageLimitOverridesByWorkspace :many
SELECT * FROM user_usage_limit_override
WHERE user_uuid = sqlc.arg('user_uuid')::uuid
  AND workspace_slug = sqlc.arg('workspace_slug')
ORDER BY limit_type;

-- name: GetUserUsageLimitOverride :one
SELECT * FROM user_usage_limit_override
WHERE uuid = sqlc.arg('uuid')::uuid;

-- name: GetUserUsageLimitOverrideByKey :one
SELECT * FROM user_usage_limit_override
WHERE user_uuid = sqlc.arg('user_uuid')::uuid
  AND workspace_slug = sqlc.arg('workspace_slug')
  AND limit_type = sqlc.arg('limit_type');

-- name: CreateUserUsageLimitOverride :one
INSERT INTO user_usage_limit_override (
    uuid,
    user_uuid,
    workspace_slug,
    limit_type,
    limit_value,
    reset_period,
    is_enabled,
    created_at
) VALUES (
    sqlc.arg('uuid')::uuid,
    sqlc.arg('user_uuid')::uuid,
    sqlc.arg('workspace_slug'),
    sqlc.arg('limit_type'),
    sqlc.arg('limit_value'),
    sqlc.arg('reset_period'),
    sqlc.arg('is_enabled')::boolean,
    NOW()
) RETURNING *;

-- name: UpdateUserUsageLimitOverride :one
UPDATE user_usage_limit_override
SET
    limit_value = sqlc.arg('limit_value'),
    reset_period = sqlc.arg('reset_period'),
    is_enabled = sqlc.arg('is_enabled')::boolean,
    updated_at = NOW()
WHERE uuid = sqlc.arg('uuid')::uuid
RETURNING *;

-- name: DeleteUserUsageLimitOverride :exec
DELETE FROM user_usage_limit_override
WHERE uuid = sqlc.arg('uuid')::uuid;

-- name: UserUsageLimitOverrideExists :one
SELECT EXISTS (
    SELECT 1 FROM user_usage_limit_override
    WHERE user_uuid = sqlc.arg('user_uuid')::uuid
      AND workspace_slug = sqlc.arg('workspace_slug')
      AND limit_type = sqlc.arg('limit_type')
) AS exists;

-- ============================================================================
-- Worker Usage Limit Queries
-- ============================================================================

-- name: ListWorkerUsageLimits :many
SELECT * FROM worker_usage_limit
WHERE worker_uuid = sqlc.arg('worker_uuid')::uuid
ORDER BY workspace_slug, limit_type;

-- name: ListWorkerUsageLimitsByWorkspace :many
SELECT * FROM worker_usage_limit
WHERE worker_uuid = sqlc.arg('worker_uuid')::uuid
  AND workspace_slug = sqlc.arg('workspace_slug')
ORDER BY limit_type;

-- name: GetWorkerUsageLimit :one
SELECT * FROM worker_usage_limit
WHERE uuid = sqlc.arg('uuid')::uuid;

-- name: GetWorkerUsageLimitByKey :one
SELECT * FROM worker_usage_limit
WHERE worker_uuid = sqlc.arg('worker_uuid')::uuid
  AND workspace_slug = sqlc.arg('workspace_slug')
  AND limit_type = sqlc.arg('limit_type');

-- name: CreateWorkerUsageLimit :one
INSERT INTO worker_usage_limit (
    uuid,
    worker_uuid,
    workspace_slug,
    limit_type,
    limit_value,
    reset_period,
    is_enabled,
    created_at
) VALUES (
    sqlc.arg('uuid')::uuid,
    sqlc.arg('worker_uuid')::uuid,
    sqlc.arg('workspace_slug'),
    sqlc.arg('limit_type'),
    sqlc.arg('limit_value'),
    sqlc.arg('reset_period'),
    sqlc.arg('is_enabled')::boolean,
    NOW()
) RETURNING *;

-- name: UpdateWorkerUsageLimit :one
UPDATE worker_usage_limit
SET
    limit_value = sqlc.arg('limit_value'),
    reset_period = sqlc.arg('reset_period'),
    is_enabled = sqlc.arg('is_enabled')::boolean,
    updated_at = NOW()
WHERE uuid = sqlc.arg('uuid')::uuid
RETURNING *;

-- name: DeleteWorkerUsageLimit :exec
DELETE FROM worker_usage_limit
WHERE uuid = sqlc.arg('uuid')::uuid;

-- name: WorkerUsageLimitExists :one
SELECT EXISTS (
    SELECT 1 FROM worker_usage_limit
    WHERE worker_uuid = sqlc.arg('worker_uuid')::uuid
      AND workspace_slug = sqlc.arg('workspace_slug')
      AND limit_type = sqlc.arg('limit_type')
) AS exists;

-- ============================================================================
-- User Usage Tracking Queries
-- ============================================================================

-- name: GetCurrentUserUsage :one
SELECT * FROM user_usage_tracking
WHERE user_uuid = sqlc.arg('user_uuid')::uuid
  AND workspace_slug = sqlc.arg('workspace_slug')
  AND limit_type = sqlc.arg('limit_type')
  AND NOW() >= period_start AND NOW() < period_end;

-- name: GetUserUsageForPeriod :one
SELECT * FROM user_usage_tracking
WHERE user_uuid = sqlc.arg('user_uuid')::uuid
  AND workspace_slug = sqlc.arg('workspace_slug')
  AND limit_type = sqlc.arg('limit_type')
  AND period_start = sqlc.arg('period_start');

-- name: CreateUserUsageTracking :one
INSERT INTO user_usage_tracking (
    uuid,
    user_uuid,
    workspace_slug,
    limit_type,
    period_start,
    period_end,
    current_usage,
    created_at
) VALUES (
    sqlc.arg('uuid')::uuid,
    sqlc.arg('user_uuid')::uuid,
    sqlc.arg('workspace_slug'),
    sqlc.arg('limit_type'),
    sqlc.arg('period_start'),
    sqlc.arg('period_end'),
    sqlc.arg('current_usage'),
    NOW()
) RETURNING *;

-- name: IncrementUserUsage :one
UPDATE user_usage_tracking
SET current_usage = current_usage + sqlc.arg('increment'),
    last_updated = NOW()
WHERE uuid = sqlc.arg('uuid')::uuid
RETURNING *;

-- name: DecrementUserUsage :one
UPDATE user_usage_tracking
SET current_usage = GREATEST(0, current_usage - sqlc.arg('decrement')),
    last_updated = NOW()
WHERE uuid = sqlc.arg('uuid')::uuid
RETURNING *;

-- ============================================================================
-- Worker Usage Tracking Queries
-- ============================================================================

-- name: GetCurrentWorkerUsage :one
SELECT * FROM worker_usage_tracking
WHERE worker_uuid = sqlc.arg('worker_uuid')::uuid
  AND workspace_slug = sqlc.arg('workspace_slug')
  AND limit_type = sqlc.arg('limit_type')
  AND NOW() >= period_start AND NOW() < period_end;

-- name: GetWorkerUsageForPeriod :one
SELECT * FROM worker_usage_tracking
WHERE worker_uuid = sqlc.arg('worker_uuid')::uuid
  AND workspace_slug = sqlc.arg('workspace_slug')
  AND limit_type = sqlc.arg('limit_type')
  AND period_start = sqlc.arg('period_start');

-- name: CreateWorkerUsageTracking :one
INSERT INTO worker_usage_tracking (
    uuid,
    worker_uuid,
    workspace_slug,
    limit_type,
    period_start,
    period_end,
    current_usage,
    created_at
) VALUES (
    sqlc.arg('uuid')::uuid,
    sqlc.arg('worker_uuid')::uuid,
    sqlc.arg('workspace_slug'),
    sqlc.arg('limit_type'),
    sqlc.arg('period_start'),
    sqlc.arg('period_end'),
    sqlc.arg('current_usage'),
    NOW()
) RETURNING *;

-- name: IncrementWorkerUsage :one
UPDATE worker_usage_tracking
SET current_usage = current_usage + sqlc.arg('increment'),
    last_updated = NOW()
WHERE uuid = sqlc.arg('uuid')::uuid
RETURNING *;

-- name: DecrementWorkerUsage :one
UPDATE worker_usage_tracking
SET current_usage = GREATEST(0, current_usage - sqlc.arg('decrement')),
    last_updated = NOW()
WHERE uuid = sqlc.arg('uuid')::uuid
RETURNING *;

-- ============================================================================
-- Effective Limit Resolution Queries
-- ============================================================================

-- name: GetEffectiveUserLimit :one
-- Returns the effective limit for a user in a workspace
-- Priority: user override > policy set default (highest limit from assigned policy sets)
-- NULL limit_value means unlimited
WITH user_policy_sets AS (
    SELECT ups.policy_set_name FROM user_policy_set ups
    WHERE ups.user_uuid = sqlc.arg('user_uuid')::uuid
      AND (ups.workspace_slug = sqlc.arg('workspace_slug') OR ups.workspace_slug IS NULL)
),
policy_limit AS (
    SELECT ul.limit_value, ul.reset_period, ul.is_enabled
    FROM usage_limit ul
    WHERE ul.policy_set_name IN (SELECT policy_set_name FROM user_policy_sets)
      AND ul.limit_type = sqlc.arg('limit_type')
      AND ul.is_enabled = TRUE
    ORDER BY ul.limit_value DESC NULLS FIRST  -- NULL (unlimited) takes priority
    LIMIT 1
),
user_override AS (
    SELECT limit_value, reset_period, is_enabled
    FROM user_usage_limit_override
    WHERE user_uuid = sqlc.arg('user_uuid')::uuid
      AND workspace_slug = sqlc.arg('workspace_slug')
      AND limit_type = sqlc.arg('limit_type')
)
SELECT
    COALESCE(uo.limit_value, pl.limit_value) AS limit_value,
    COALESCE(uo.reset_period, pl.reset_period, 'monthly') AS reset_period,
    COALESCE(uo.is_enabled, pl.is_enabled, FALSE) AS is_enabled,
    (uo.limit_value IS NOT NULL OR uo.is_enabled IS NOT NULL) AS has_override
FROM (SELECT 1) dummy
LEFT JOIN policy_limit pl ON TRUE
LEFT JOIN user_override uo ON TRUE;

-- ============================================================================
-- Cleanup Queries
-- ============================================================================

-- name: CleanupExpiredUserUsageTracking :exec
DELETE FROM user_usage_tracking
WHERE period_end < NOW() - INTERVAL '90 days';

-- name: CleanupExpiredWorkerUsageTracking :exec
DELETE FROM worker_usage_tracking
WHERE period_end < NOW() - INTERVAL '90 days';
