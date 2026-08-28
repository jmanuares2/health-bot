package groq

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	openai "github.com/sashabaranov/go-openai"
)

const (
	modelFast     = "groq/compound-mini"
	modelFallback = "groq/compound"
	groqBaseURL   = "https://api.groq.com/openai/v1"
)

var systemPrompt = `You are a health and fitness tracker assistant. The user sends messages in natural language (Spanish or English) describing food intake, exercise, body weight, or water consumption.

You MUST always respond with a single JSON object following this exact schema. Only include the field that matches the detected type.

{
  "type": "food | strength | cardio | body_weight | water | unclear",
  "confidence": "high | low",

  "food": {
    "meal_type": "breakfast | lunch | dinner | snack",
    "description": "short description",
    "calories": 0,
    "protein_g": 0.0,
    "carbs_g": 0.0,
    "fat_g": 0.0,
    "fiber_g": 0.0
  },

  "exercises": [
    {
      "name": "exercise name in English (e.g. bench press, squat)",
      "sets": 0,
      "reps": 0,
      "weight_kg": 0.0,
      "calories_burned": 0
    }
  ],

  "cardio": {
    "activity": "activity in English (e.g. running, cycling)",
    "duration_min": 0,
    "distance_km": 0.0,
    "calories_burned": 0
  },

  "body_weight": {
    "weight_kg": 0.0
  },

  "water": {
    "liters": 0.0
  },

  "clarification_question": "Only present if type=unclear. Ask the user for missing info in Spanish."
}

Rules:
- Normalize exercise names to English (e.g. "press de banca" → "bench press", "sentadilla" → "squat").
- Estimate calories and macros from nutritional knowledge if not explicitly stated.
- Set confidence to "low" if you are unsure about the classification or nutrient values.
- Always respond with valid JSON only. No markdown, no extra text.`

// FoodData holds parsed food information.
type FoodData struct {
	MealType    string  `json:"meal_type"`
	Description string  `json:"description"`
	Calories    int     `json:"calories"`
	ProteinG    float64 `json:"protein_g"`
	CarbsG      float64 `json:"carbs_g"`
	FatG        float64 `json:"fat_g"`
	FiberG      float64 `json:"fiber_g"`
}

// ExerciseData holds parsed strength exercise information.
type ExerciseData struct {
	Name           string  `json:"name"`
	Sets           int     `json:"sets"`
	Reps           int     `json:"reps"`
	WeightKg       float64 `json:"weight_kg"`
	CaloriesBurned int     `json:"calories_burned"`
}

// CardioData holds parsed cardio information.
type CardioData struct {
	Activity       string  `json:"activity"`
	DurationMin    int     `json:"duration_min"`
	DistanceKm     float64 `json:"distance_km"`
	CaloriesBurned int     `json:"calories_burned"`
}

// BodyWeightData holds parsed body weight information.
type BodyWeightData struct {
	WeightKg float64 `json:"weight_kg"`
}

// WaterData holds parsed water intake information.
type WaterData struct {
	Liters float64 `json:"liters"`
}

// ParsedMessage is the full struct returned by the LLM parser.
type ParsedMessage struct {
	Type                  string          `json:"type"`
	Confidence            string          `json:"confidence"`
	Food                  *FoodData       `json:"food,omitempty"`
	Exercises             []ExerciseData  `json:"exercises,omitempty"`
	Cardio                *CardioData     `json:"cardio,omitempty"`
	BodyWeight            *BodyWeightData `json:"body_weight,omitempty"`
	Water                 *WaterData      `json:"water,omitempty"`
	ClarificationQuestion string          `json:"clarification_question,omitempty"`
}

// Parser calls the Groq LLM to parse a user message.
type Parser struct {
	client *openai.Client
}

// NewParser creates a new Parser using GROQ_API_KEY from environment.
func NewParser() *Parser {
	cfg := openai.DefaultConfig(os.Getenv("GROQ_API_KEY"))
	cfg.BaseURL = groqBaseURL
	cfg.HTTPClient = &http.Client{Timeout: 30 * time.Second}
	return &Parser{client: openai.NewClientWithConfig(cfg)}
}

// Parse sends the message to the LLM and returns a structured ParsedMessage.
// If confidence is "low", it retries with the fallback model.
func (p *Parser) Parse(ctx context.Context, message string) (*ParsedMessage, error) {
	result, err := p.callLLM(ctx, message, modelFast)
	if err != nil {
		return nil, err
	}

	if result.Confidence == "low" {
		fallback, err := p.callLLM(ctx, message, modelFallback)
		if err != nil {
			// Return the fast result if fallback fails
			return result, nil
		}
		return fallback, nil
	}

	return result, nil
}

func (p *Parser) callLLM(ctx context.Context, message, model string) (*ParsedMessage, error) {
	resp, err := p.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: model,
		Messages: []openai.ChatCompletionMessage{
			{Role: openai.ChatMessageRoleSystem, Content: systemPrompt},
			{Role: openai.ChatMessageRoleUser, Content: message},
		},
		Temperature: 0.1,
	})
	if err != nil {
		return nil, fmt.Errorf("groq chat completion (%s): %w", model, err)
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("groq returned no choices")
	}

	content := resp.Choices[0].Message.Content

	var parsed ParsedMessage
	if err := json.Unmarshal([]byte(content), &parsed); err != nil {
		return nil, fmt.Errorf("json unmarshal LLM response: %w\nraw: %s", err, content)
	}

	return &parsed, nil
}
