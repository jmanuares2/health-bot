-- name: CreateWaterLog :one
INSERT INTO water_logs (user_id, date, liters)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetDailyWater :one
SELECT COALESCE(SUM(liters), 0)::decimal AS total_liters
FROM water_logs
WHERE user_id = $1 AND date = $2;

-- name: GetWaterLogsByDate :many
SELECT * FROM water_logs
WHERE user_id = $1 AND date = $2
ORDER BY created_at ASC;
