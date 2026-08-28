-- name: CreateCardioSession :one
INSERT INTO cardio_sessions (user_id, date, activity, duration_min, distance_km, calories_burned)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetCardioSessionsByDate :many
SELECT * FROM cardio_sessions
WHERE user_id = $1 AND date = $2
ORDER BY created_at ASC;
