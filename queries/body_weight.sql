-- name: CreateBodyWeight :one
INSERT INTO body_weight (user_id, date, weight_kg)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetBodyWeightHistory :many
SELECT * FROM body_weight
WHERE user_id = $1
ORDER BY date ASC;

-- name: GetLatestBodyWeight :one
SELECT * FROM body_weight
WHERE user_id = $1
ORDER BY date DESC
LIMIT 1;
