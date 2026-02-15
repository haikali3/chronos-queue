-- name: GetJobs :one
SELECT * FROM jobs
WHERE id = $1;