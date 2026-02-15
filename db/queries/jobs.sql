-- name: CreateJob :one
INSERT INTO jobs (
  id, type, payload, status, max_retries, idempotency_key, created_at, updated_at
) VALUES (
  $1, $2, $3, $4, $5, $6, now(), now()
)

RETURNING *;
-- name: GetJobs :one
SELECT * FROM jobs
WHERE id = $1;


-- name: 