-- name: CreateJob :one
INSERT INTO jobs (
  id, type, payload, status, max_retries, idempotency_key, created_at, updated_at
) VALUES (
  $1, $2, $3, $4, $5, $6, now(), now()
)
RETURNING *;

-- name: GetJob :one
SELECT * FROM jobs
WHERE id = $1;

-- name: ListPendingJobs :many
SELECT * FROM jobs
WHERE status = 'PENDING';

-- name: ClaimJob :one
UPDATE jobs SET status = 'IN_PROGRESS', updated_at = NOW()
WHERE id = (
  SELECT id FROM jobs
  WHERE status = 'PENDING' OR (status = 'RETRYING' AND next_retry_at <= NOW())
  ORDER BY created_at ASC
  LIMIT 1
  FOR UPDATE SKIP LOCKED
)
RETURNING *;

-- name: UpdateJobStatus :exec
UPDATE jobs
SET status = $2, retry_count = $3, next_retry_at = $4, updated_at = NOW()
WHERE id = $1;