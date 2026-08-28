-- name: CreateFoodEntry :one
INSERT INTO food_entries (user_id, date, meal_type, description, calories, protein_g, carbs_g, fat_g, fiber_g)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING *;

-- name: GetFoodEntriesByDate :many
SELECT * FROM food_entries
WHERE user_id = $1 AND date = $2
ORDER BY created_at ASC;

-- name: GetDailyCalories :one
SELECT COALESCE(SUM(calories), 0)::int AS total_calories,
       COALESCE(SUM(protein_g), 0)::decimal AS total_protein,
       COALESCE(SUM(carbs_g), 0)::decimal AS total_carbs,
       COALESCE(SUM(fat_g), 0)::decimal AS total_fat,
       COALESCE(SUM(fiber_g), 0)::decimal AS total_fiber
FROM food_entries
WHERE user_id = $1 AND date = $2;

-- name: GetMonthlyCalories :many
SELECT date, SUM(calories)::int AS total_calories
FROM food_entries
WHERE user_id = $1
  AND date >= $2::date
  AND date < ($2::date + INTERVAL '1 month')
GROUP BY date
ORDER BY date ASC;
