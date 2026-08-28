package bot

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jmanuares2/health-bot/internal/db"
	"github.com/jmanuares2/health-bot/internal/groq"
)

// timezone for date calculations
const timezone = "America/Argentina/Buenos_Aires"

// Handler routes incoming messages to the correct handler.
type Handler struct {
	parser  *groq.Parser
	queries *db.Queries
	pool    *pgxpool.Pool
	userID  int32
}

// NewHandler creates a Handler. userID is the DB id of the single user.
func NewHandler(parser *groq.Parser, pool *pgxpool.Pool, userID int32) *Handler {
	return &Handler{
		parser:  parser,
		queries: db.New(pool),
		pool:    pool,
		userID:  userID,
	}
}

// Handle parses the message and persists data, returning the reply string.
func (h *Handler) Handle(ctx context.Context, message string) string {
	parsed, err := h.parser.Parse(ctx, message)
	if err != nil {
		log.Printf("groq parse error: %v", err)
		return "No pude procesar el mensaje. Intenta de nuevo."
	}

	today := h.today()

	switch parsed.Type {
	case "food":
		return h.handleFood(ctx, parsed, today)
	case "strength":
		return h.handleStrength(ctx, parsed, today)
	case "cardio":
		return h.handleCardio(ctx, parsed, today)
	case "body_weight":
		return h.handleBodyWeight(ctx, parsed, today)
	case "water":
		return h.handleWater(ctx, parsed, today)
	case "unclear":
		if parsed.ClarificationQuestion != "" {
			return parsed.ClarificationQuestion
		}
		return "No entendi el mensaje. Podes ser mas especifico?"
	default:
		return "Tipo de mensaje desconocido."
	}
}

func (h *Handler) today() pgtype.Date {
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		loc = time.UTC
	}
	now := time.Now().In(loc)
	t := pgtype.Date{}
	_ = t.Scan(now.Format("2006-01-02"))
	return t
}

func (h *Handler) handleFood(ctx context.Context, parsed *groq.ParsedMessage, today pgtype.Date) string {
	if parsed.Food == nil {
		return "No encontre datos de comida en el mensaje."
	}
	f := parsed.Food

	_, err := h.queries.CreateFoodEntry(ctx, db.CreateFoodEntryParams{
		UserID:      pgtype.Int4{Int32: h.userID, Valid: true},
		Date:        today,
		MealType:    pgtype.Text{String: f.MealType, Valid: f.MealType != ""},
		Description: f.Description,
		Calories:    int32(f.Calories),
		ProteinG:    toPgNumeric(f.ProteinG),
		CarbsG:      toPgNumeric(f.CarbsG),
		FatG:        toPgNumeric(f.FatG),
		FiberG:      toPgNumeric(f.FiberG),
	})
	if err != nil {
		log.Printf("createFoodEntry error: %v", err)
		return "Error guardando comida."
	}

	// Get daily totals for remaining calories
	totals, err := h.queries.GetDailyCalories(ctx, db.GetDailyCaloriesParams{
		UserID: pgtype.Int4{Int32: h.userID, Valid: true},
		Date:   today,
	})
	if err != nil {
		log.Printf("getDailyCalories error: %v", err)
	}

	user, _ := h.queries.GetUserByPhone(ctx, "") // we use userID directly
	calorieGoal := int32(2000)
	if user.CalorieGoal != 0 {
		calorieGoal = user.CalorieGoal
	}

	remaining := calorieGoal - totals.TotalCalories
	mealLabel := mealTypeLabel(f.MealType)

	return fmt.Sprintf(
		"✅ %s · +%d kcal (P:%.0fg C:%.0fg G:%.0fg F:%.0fg) · Quedan %d kcal",
		mealLabel, f.Calories,
		f.ProteinG, f.CarbsG, f.FatG, f.FiberG,
		remaining,
	)
}

func (h *Handler) handleStrength(ctx context.Context, parsed *groq.ParsedMessage, today pgtype.Date) string {
	if len(parsed.Exercises) == 0 {
		return "No encontre ejercicios en el mensaje."
	}

	var replies []string
	for _, ex := range parsed.Exercises {
		_, err := h.queries.CreateStrengthSession(ctx, db.CreateStrengthSessionParams{
			UserID:         pgtype.Int4{Int32: h.userID, Valid: true},
			Date:           today,
			ExerciseName:   ex.Name,
			Sets:           pgtype.Int4{Int32: int32(ex.Sets), Valid: ex.Sets > 0},
			Reps:           pgtype.Int4{Int32: int32(ex.Reps), Valid: ex.Reps > 0},
			WeightKg:       toPgNumeric(ex.WeightKg),
			CaloriesBurned: pgtype.Int4{Int32: int32(ex.CaloriesBurned), Valid: ex.CaloriesBurned > 0},
		})
		if err != nil {
			log.Printf("createStrengthSession error: %v", err)
			replies = append(replies, fmt.Sprintf("Error guardando %s.", ex.Name))
			continue
		}

		reply := fmt.Sprintf("✅ %s %dx%d", capitalize(ex.Name), ex.Sets, ex.Reps)
		if ex.WeightKg > 0 {
			reply += fmt.Sprintf(" @%.1fkg", ex.WeightKg)
		}
		if ex.CaloriesBurned > 0 {
			reply += fmt.Sprintf(" · -%d kcal quemadas", ex.CaloriesBurned)
		}
		replies = append(replies, reply)
	}

	return joinLines(replies)
}

func (h *Handler) handleCardio(ctx context.Context, parsed *groq.ParsedMessage, today pgtype.Date) string {
	if parsed.Cardio == nil {
		return "No encontre datos de cardio en el mensaje."
	}
	c := parsed.Cardio

	_, err := h.queries.CreateCardioSession(ctx, db.CreateCardioSessionParams{
		UserID:         pgtype.Int4{Int32: h.userID, Valid: true},
		Date:           today,
		Activity:       c.Activity,
		DurationMin:    pgtype.Int4{Int32: int32(c.DurationMin), Valid: c.DurationMin > 0},
		DistanceKm:     toPgNumeric(c.DistanceKm),
		CaloriesBurned: pgtype.Int4{Int32: int32(c.CaloriesBurned), Valid: c.CaloriesBurned > 0},
	})
	if err != nil {
		log.Printf("createCardioSession error: %v", err)
		return "Error guardando cardio."
	}

	reply := fmt.Sprintf("✅ %s", capitalize(c.Activity))
	if c.DistanceKm > 0 {
		reply += fmt.Sprintf(" %.1fkm", c.DistanceKm)
	}
	if c.DurationMin > 0 {
		reply += fmt.Sprintf(" / %dmin", c.DurationMin)
	}
	if c.CaloriesBurned > 0 {
		reply += fmt.Sprintf(" · -%d kcal quemadas", c.CaloriesBurned)
	}
	return reply
}

func (h *Handler) handleBodyWeight(ctx context.Context, parsed *groq.ParsedMessage, today pgtype.Date) string {
	if parsed.BodyWeight == nil {
		return "No encontre datos de peso en el mensaje."
	}

	_, err := h.queries.CreateBodyWeight(ctx, db.CreateBodyWeightParams{
		UserID:   pgtype.Int4{Int32: h.userID, Valid: true},
		Date:     today,
		WeightKg: toPgNumeric(parsed.BodyWeight.WeightKg),
	})
	if err != nil {
		log.Printf("createBodyWeight error: %v", err)
		return "Error guardando peso."
	}

	return fmt.Sprintf("✅ Peso registrado: %.1fkg", parsed.BodyWeight.WeightKg)
}

func (h *Handler) handleWater(ctx context.Context, parsed *groq.ParsedMessage, today pgtype.Date) string {
	if parsed.Water == nil {
		return "No encontre datos de agua en el mensaje."
	}

	_, err := h.queries.CreateWaterLog(ctx, db.CreateWaterLogParams{
		UserID: pgtype.Int4{Int32: h.userID, Valid: true},
		Date:   today,
		Liters: toPgNumeric(parsed.Water.Liters),
	})
	if err != nil {
		log.Printf("createWaterLog error: %v", err)
		return "Error guardando agua."
	}

	total, err := h.queries.GetDailyWater(ctx, db.GetDailyWaterParams{
		UserID: pgtype.Int4{Int32: h.userID, Valid: true},
		Date:   today,
	})
	if err != nil {
		log.Printf("getDailyWater error: %v", err)
		return fmt.Sprintf("✅ +%.1fL registrados", parsed.Water.Liters)
	}

	totalF, _ := total.TotalLiters.Float64Value()
	return fmt.Sprintf("✅ +%.1fL · Total hoy: %.1fL", parsed.Water.Liters, totalF.Float64)
}
