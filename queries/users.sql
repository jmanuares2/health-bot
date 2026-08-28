-- name: GetOrCreateUser :one
INSERT INTO users (phone)
VALUES ($1)
ON CONFLICT (phone) DO UPDATE SET phone = EXCLUDED.phone
RETURNING *;

-- name: GetUserByPhone :one
SELECT * FROM users WHERE phone = $1;

-- name: UpdateUserGoals :one
UPDATE users
SET calorie_goal = $2, protein_goal_g = $3
WHERE id = $1
RETURNING *;
