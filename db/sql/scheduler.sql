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
    is_enabled,
    is_paused,
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
             sqlc.arg('is_enabled')::boolean,
             sqlc.arg('is_paused')::boolean,
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
                     is_enabled = sqlc.arg('is_enabled')::boolean,
                     is_paused = sqlc.arg('is_paused')::boolean,
                     updated_at = NOW()
WHERE uuid = sqlc.arg('uuid')::uuid;

-- name: DeleteScheduler :exec
DELETE FROM scheduler WHERE uuid = sqlc.arg('uuid')::uuid;
