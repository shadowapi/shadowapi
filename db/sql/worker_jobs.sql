-- name: CreateWorkerJob :one
INSERT INTO worker_jobs (
    uuid,
    scheduler_uuid,
    subject,
    status,
    data,
    started_at,
    finished_at
) VALUES (
             sqlc.arg('uuid')::uuid,
             sqlc.arg('scheduler_uuid')::uuid,
             sqlc.arg('subject'),
             sqlc.arg('status'),
             sqlc.arg('data'),
             NOW(),
             sqlc.arg('finished_at')
         )
RETURNING *;

-- name: GetWorkerJob :one
SELECT
    sqlc.embed(worker_jobs)
FROM worker_jobs
WHERE uuid = sqlc.arg('uuid')::uuid;

-- name: ListWorkerJobs :many
SELECT
    sqlc.embed(worker_jobs)
FROM worker_jobs
ORDER BY started_at DESC
LIMIT NULLIF(sqlc.arg('limit')::int, 0)
    OFFSET sqlc.arg('offset')::int;

-- name: GetWorkerJobs :many
WITH filtered_worker_jobs AS (
    SELECT w.*
    FROM worker_jobs w
    WHERE
        (sqlc.arg('scheduler_uuid')::uuid IS NULL
            OR w.scheduler_uuid = sqlc.arg('scheduler_uuid')::uuid)
      AND (NULLIF(sqlc.arg('subject'), '') IS NULL
        OR w.subject = sqlc.arg('subject'))
      AND (NULLIF(sqlc.arg('status'), '') IS NULL
        OR w.status = sqlc.arg('status'))
)
SELECT
    *,
    (SELECT COUNT(*) FROM filtered_worker_jobs) AS total_count
FROM filtered_worker_jobs
ORDER BY
    CASE WHEN sqlc.arg('order_by') = 'started_at' AND sqlc.arg('order_direction') = 'asc'  THEN started_at END ASC,
    CASE WHEN sqlc.arg('order_by') = 'started_at' AND sqlc.arg('order_direction') = 'desc' THEN started_at END DESC,

    CASE WHEN sqlc.arg('order_by') = 'finished_at' AND sqlc.arg('order_direction') = 'asc'  THEN finished_at END ASC,
    CASE WHEN sqlc.arg('order_by') = 'finished_at' AND sqlc.arg('order_direction') = 'desc' THEN finished_at END DESC,

    started_at DESC
LIMIT NULLIF(sqlc.arg('limit')::int, 0)
    OFFSET sqlc.arg('offset')::int;

-- name: UpdateWorkerJob :exec
UPDATE worker_jobs
SET
    subject     = sqlc.arg('subject'),
    status      = sqlc.arg('status'),
    data        = sqlc.arg('data'),
    finished_at = sqlc.arg('finished_at')
WHERE uuid = sqlc.arg('uuid')::uuid;

-- name: DeleteWorkerJob :exec
DELETE FROM worker_jobs
WHERE uuid = sqlc.arg('uuid')::uuid;
