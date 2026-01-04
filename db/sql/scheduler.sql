-- name: CreateScheduler :one
INSERT INTO scheduler (
    uuid,
    pipeline_uuid,
    schedule_type,
    cron_expression,
    run_at,
    timezone,
    next_run,
    last_run,
    last_uid,
    is_enabled,
    is_paused,
    batch_size,
    sync_state,
    last_sync_timestamp,
    oldest_sync_timestamp,
    cutoff_date,
    created_at,
    updated_at
) VALUES (
             sqlc.arg('uuid')::uuid,
             sqlc.arg('pipeline_uuid')::uuid,
             sqlc.arg('schedule_type'),
             sqlc.arg('cron_expression'),
             sqlc.arg('run_at'),
             sqlc.arg('timezone'),
             sqlc.arg('next_run'),
             sqlc.arg('last_run'),
             COALESCE(sqlc.arg('last_uid')::bigint, 0),
             sqlc.arg('is_enabled')::boolean,
             sqlc.arg('is_paused')::boolean,
             sqlc.arg('batch_size')::int,
             COALESCE(sqlc.arg('sync_state'), 'initial'),
             sqlc.arg('last_sync_timestamp'),
             sqlc.arg('oldest_sync_timestamp'),
             sqlc.arg('cutoff_date'),
             NOW(),
             NOW()
         ) RETURNING *;

-- name: GetScheduler :one
SELECT
    sqlc.embed(scheduler)
FROM scheduler
WHERE uuid = sqlc.arg('uuid')::uuid;

-- name: ListSchedulers :many
SELECT
    sqlc.embed(scheduler)
FROM scheduler
ORDER BY created_at DESC
LIMIT NULLIF(sqlc.arg('limit')::int, 0)
    OFFSET sqlc.arg('offset');

-- name: GetSchedulers :many
WITH filtered_schedulers AS (
    SELECT s.*
    FROM scheduler s
    WHERE
        (NULLIF(sqlc.arg('pipeline_uuid'), '') IS NULL OR s.pipeline_uuid = sqlc.arg('pipeline_uuid')::uuid) AND
        (NULLIF(sqlc.arg('is_enabled')::int, -1) IS NULL OR s.is_enabled = sqlc.arg('is_enabled')::boolean) AND
        (NULLIF(sqlc.arg('is_paused')::int, -1) IS NULL OR s.is_paused = sqlc.arg('is_paused')::boolean)
)
SELECT
    *,
    (SELECT count(*) FROM filtered_schedulers) as total_count
FROM filtered_schedulers
ORDER BY
    CASE WHEN sqlc.arg('order_by') = 'created_at' AND sqlc.arg('order_direction') = 'asc' THEN created_at END ASC,
    CASE WHEN sqlc.arg('order_by') = 'created_at' AND sqlc.arg('order_direction') = 'desc' THEN created_at END DESC,
    CASE WHEN sqlc.arg('order_by') = 'updated_at' AND sqlc.arg('order_direction') = 'asc' THEN updated_at END ASC,
    CASE WHEN sqlc.arg('order_by') = 'updated_at' AND sqlc.arg('order_direction') = 'desc' THEN updated_at END DESC,
    created_at DESC
LIMIT NULLIF(sqlc.arg('limit')::int, 0)
    OFFSET sqlc.arg('offset')::int;

-- name: UpdateScheduler :exec
UPDATE scheduler SET
                     cron_expression = sqlc.arg('cron_expression'),
                     run_at = sqlc.arg('run_at'),
                     timezone = sqlc.arg('timezone'),
                     next_run = sqlc.arg('next_run'),
                     last_run = sqlc.arg('last_run'),
                     last_uid = COALESCE(sqlc.arg('last_uid')::bigint, last_uid),
                     is_enabled = sqlc.arg('is_enabled')::boolean,
                     is_paused = sqlc.arg('is_paused')::boolean,
                     batch_size = sqlc.arg('batch_size')::int,
                     cutoff_date = sqlc.arg('cutoff_date'),
                     updated_at = NOW()
WHERE uuid = sqlc.arg('uuid')::uuid;

-- name: DeleteScheduler :exec
DELETE FROM scheduler WHERE uuid = sqlc.arg('uuid')::uuid;

-- name: UpdateSchedulerLastRun :exec
-- Update the scheduler's last successful fetch timestamp
UPDATE scheduler SET
    last_run = sqlc.arg('last_run'),
    updated_at = NOW()
WHERE uuid = sqlc.arg('uuid')::uuid;

-- name: UpdateSchedulerFetchProgress :exec
-- Update the scheduler's fetch progress (last_uid for incremental IMAP fetch)
-- DEPRECATED: Use UpdateSchedulerSyncProgress for new timestamp-based tracking
UPDATE scheduler SET
    last_uid = sqlc.arg('last_uid')::bigint,
    last_run = sqlc.arg('last_run'),
    updated_at = NOW()
WHERE uuid = sqlc.arg('uuid')::uuid;

-- name: UpdateSchedulerSyncProgress :exec
-- Update the scheduler's sync progress with timestamp-based tracking
UPDATE scheduler SET
    sync_state = COALESCE(sqlc.arg('sync_state'), sync_state),
    last_sync_timestamp = COALESCE(sqlc.arg('last_sync_timestamp'), last_sync_timestamp),
    oldest_sync_timestamp = COALESCE(sqlc.arg('oldest_sync_timestamp'), oldest_sync_timestamp),
    last_run = sqlc.arg('last_run'),
    updated_at = NOW()
WHERE uuid = sqlc.arg('uuid')::uuid;

-- ============================================================================
-- Workspace-filtered scheduler queries
-- ============================================================================

-- name: GetSchedulerByWorkspace :one
-- Get a scheduler by UUID, filtered by workspace (via pipeline join)
SELECT
    sqlc.embed(scheduler)
FROM scheduler
INNER JOIN pipeline p ON scheduler.pipeline_uuid = p.uuid
WHERE scheduler.uuid = sqlc.arg('uuid')::uuid
  AND p.workspace_uuid = sqlc.arg('workspace_uuid')::uuid;

-- name: ListSchedulersByWorkspace :many
-- List all schedulers for a workspace
SELECT
    sqlc.embed(scheduler)
FROM scheduler
INNER JOIN pipeline p ON scheduler.pipeline_uuid = p.uuid
WHERE p.workspace_uuid = sqlc.arg('workspace_uuid')::uuid
ORDER BY scheduler.created_at DESC
LIMIT NULLIF(sqlc.arg('limit')::int, 0)
    OFFSET sqlc.arg('offset');

-- name: GetSchedulersWithWorkspace :many
-- Get schedulers with filters, scoped to a workspace
WITH filtered_schedulers AS (
    SELECT s.*
    FROM scheduler s
    INNER JOIN pipeline p ON s.pipeline_uuid = p.uuid
    WHERE
        p.workspace_uuid = sqlc.arg('workspace_uuid')::uuid AND
        (NULLIF(sqlc.arg('pipeline_uuid'), '') IS NULL OR s.pipeline_uuid = sqlc.arg('pipeline_uuid')::uuid) AND
        (NULLIF(sqlc.arg('is_enabled')::int, -1) IS NULL OR s.is_enabled = sqlc.arg('is_enabled')::boolean) AND
        (NULLIF(sqlc.arg('is_paused')::int, -1) IS NULL OR s.is_paused = sqlc.arg('is_paused')::boolean)
)
SELECT
    *,
    (SELECT count(*) FROM filtered_schedulers) as total_count
FROM filtered_schedulers
ORDER BY
    CASE WHEN sqlc.arg('order_by') = 'created_at' AND sqlc.arg('order_direction') = 'asc' THEN created_at END ASC,
    CASE WHEN sqlc.arg('order_by') = 'created_at' AND sqlc.arg('order_direction') = 'desc' THEN created_at END DESC,
    CASE WHEN sqlc.arg('order_by') = 'updated_at' AND sqlc.arg('order_direction') = 'asc' THEN updated_at END ASC,
    CASE WHEN sqlc.arg('order_by') = 'updated_at' AND sqlc.arg('order_direction') = 'desc' THEN updated_at END DESC,
    created_at DESC
LIMIT NULLIF(sqlc.arg('limit')::int, 0)
    OFFSET sqlc.arg('offset')::int;

-- name: DeleteSchedulerByWorkspace :exec
-- Delete a scheduler, ensuring it belongs to the workspace (via pipeline)
DELETE FROM scheduler
WHERE uuid = sqlc.arg('uuid')::uuid
  AND pipeline_uuid IN (
    SELECT p.uuid FROM pipeline p WHERE p.workspace_uuid = sqlc.arg('workspace_uuid')::uuid
  );

-- name: UpdateSchedulerByWorkspace :exec
-- Update a scheduler, ensuring it belongs to the workspace (via pipeline)
UPDATE scheduler SET
    cron_expression = sqlc.arg('cron_expression'),
    run_at = sqlc.arg('run_at'),
    timezone = sqlc.arg('timezone'),
    next_run = sqlc.arg('next_run'),
    last_run = sqlc.arg('last_run'),
    last_uid = COALESCE(sqlc.arg('last_uid')::bigint, last_uid),
    is_enabled = sqlc.arg('is_enabled')::boolean,
    is_paused = sqlc.arg('is_paused')::boolean,
    batch_size = sqlc.arg('batch_size')::int,
    cutoff_date = sqlc.arg('cutoff_date'),
    updated_at = NOW()
WHERE uuid = sqlc.arg('uuid')::uuid
  AND pipeline_uuid IN (
    SELECT p.uuid FROM pipeline p WHERE p.workspace_uuid = sqlc.arg('workspace_uuid')::uuid
  );
