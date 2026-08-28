package api

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jmanuares2/health-bot/internal/db"
)

const timezone = "America/Argentina/Buenos_Aires"

// Handlers holds all API handler dependencies.
type Handlers struct {
	queries     *db.Queries
	userID      int32
	calorieGoal int32
}

// NewHandlers creates an API handler set for the given user.
func NewHandlers(queries *db.Queries, user db.User) *Handlers {
	goal := user.CalorieGoal
	if goal == 0 {
		goal = 2000
	}
	return &Handlers{queries: queries, userID: user.ID, calorieGoal: goal}
}

func (h *Handlers) today() pgtype.Date {
	loc, _ := time.LoadLocation(timezone)
	now := time.Now().In(loc)
	d := pgtype.Date{}
	_ = d.Scan(now.Format("2006-01-02"))
	return d
}

// GET /api/today
func (h *Handlers) Today(c *gin.Context) {
	ctx := c.Request.Context()
	today := h.today()

	uid := pgtype.Int4{Int32: h.userID, Valid: true}

	calories, err := h.queries.GetDailyCalories(ctx, db.GetDailyCaloriesParams{UserID: uid, Date: today})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	waterTotal, err := h.queries.GetDailyWater(ctx, db.GetDailyWaterParams{UserID: uid, Date: today})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	food, err := h.queries.GetFoodEntriesByDate(ctx, db.GetFoodEntriesByDateParams{UserID: uid, Date: today})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	strength, err := h.queries.GetStrengthSessionsByDate(ctx, db.GetStrengthSessionsByDateParams{UserID: uid, Date: today})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	cardio, err := h.queries.GetCardioSessionsByDate(ctx, db.GetCardioSessionsByDateParams{UserID: uid, Date: today})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	waterF, _ := waterTotal.Float64Value()
	proteinF, _ := calories.TotalProtein.Float64Value()
	carbsF, _ := calories.TotalCarbs.Float64Value()
	fatF, _ := calories.TotalFat.Float64Value()
	fiberF, _ := calories.TotalFiber.Float64Value()

	c.JSON(http.StatusOK, gin.H{
		"date":         today.Time.Format("2006-01-02"),
		"calorie_goal": h.calorieGoal,
		"calories":     calories.TotalCalories,
		"protein_g":    proteinF.Float64,
		"carbs_g":      carbsF.Float64,
		"fat_g":        fatF.Float64,
		"fiber_g":      fiberF.Float64,
		"water_liters": waterF.Float64,
		"food_entries": food,
		"strength":     strength,
		"cardio":       cardio,
	})
}

// GET /api/progress?month=YYYY-MM
func (h *Handlers) Progress(c *gin.Context) {
	ctx := c.Request.Context()
	month := c.Query("month")
	if month == "" {
		loc, _ := time.LoadLocation(timezone)
		month = time.Now().In(loc).Format("2006-01")
	}
	monthDate := pgtype.Date{}
	_ = monthDate.Scan(month + "-01")

	uid := pgtype.Int4{Int32: h.userID, Valid: true}

	calories, err := h.queries.GetMonthlyCalories(ctx, db.GetMonthlyCaloriesParams{
		UserID:  uid,
		Column2: monthDate,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	weights, err := h.queries.GetBodyWeightHistory(ctx, uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Adherence: % of days where calories <= goal
	daysOnTarget := 0
	for _, day := range calories {
		if day.TotalCalories <= h.calorieGoal {
			daysOnTarget++
		}
	}
	adherence := 0.0
	if len(calories) > 0 {
		adherence = float64(daysOnTarget) / float64(len(calories)) * 100
	}

	c.JSON(http.StatusOK, gin.H{
		"month":       month,
		"calories":    calories,
		"body_weight": weights,
		"adherence":   adherence,
	})
}

// GET /api/gym
func (h *Handlers) GymList(c *gin.Context) {
	ctx := c.Request.Context()
	exercises, err := h.queries.ListExercises(ctx, pgtype.Int4{Int32: h.userID, Valid: true})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"exercises": exercises})
}

// GET /api/gym/:exercise
func (h *Handlers) GymExercise(c *gin.Context) {
	ctx := c.Request.Context()
	exercise := c.Param("exercise")
	uid := pgtype.Int4{Int32: h.userID, Valid: true}

	history, err := h.queries.GetExerciseHistory(ctx, db.GetExerciseHistoryParams{
		UserID:       uid,
		ExerciseName: exercise,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	prs, err := h.queries.GetMonthlyPRs(ctx, uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var exercisePRs []db.GetMonthlyPRsRow
	for _, pr := range prs {
		if pr.ExerciseName == exercise {
			exercisePRs = append(exercisePRs, pr)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"exercise": exercise,
		"history":  history,
		"prs":      exercisePRs,
	})
}

// GET /api/body-weight
func (h *Handlers) BodyWeight(c *gin.Context) {
	ctx := c.Request.Context()
	weights, err := h.queries.GetBodyWeightHistory(ctx, pgtype.Int4{Int32: h.userID, Valid: true})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"body_weight": weights})
}
