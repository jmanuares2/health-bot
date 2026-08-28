-- name: CreateStrengthSession :one
INSERT INTO strength_sessions (user_id, date, exercise_name, sets, reps, weight_kg, calories_burned)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: GetStrengthSessionsByDate :many
SELECT * FROM strength_sessions
WHERE user_id = $1 AND date = $2
ORDER BY created_at ASC;

-- name: GetExerciseHistory :many
SELECT date, MAX(weight_kg) AS max_weight_kg, sets, reps
FROM strength_sessions
WHERE user_id = $1 AND exercise_name = $2
GROUP BY date, sets, reps
ORDER BY date ASC;

-- name: ListExercises :many
SELECT DISTINCT exercise_name FROM strength_sessions
WHERE user_id = $1
ORDER BY exercise_name ASC;

-- name: GetMonthlyPRs :many
SELECT exercise_name,
       DATE_TRUNC('month', date)::date AS month,
       MAX(weight_kg) AS pr_weight_kg
FROM strength_sessions
WHERE user_id = $1
GROUP BY exercise_name, DATE_TRUNC('month', date)
ORDER BY month DESC, exercise_name ASC;
