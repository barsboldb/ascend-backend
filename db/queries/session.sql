-- name: CreateSession :one
INSERT INTO sessions (
  program_day_id,
  week_number,
  started_at,
  ended_at,
  notes
)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetSession :one
SELECT * FROM sessions
WHERE id = $1;

-- name: ListSessions :many
SELECT * FROM sessions
ORDER BY started_at DESC;

-- name: DeleteSession :exec
DELETE FROM sessions
WHERE id = $1;
