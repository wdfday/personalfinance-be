package dto

import (
	"errors"
	"fmt"
	"time"
)

// GoalForRating represents a simplified Goal for auto-scoring and rating
type GoalForRating struct {
	ID            string    `json:"id" binding:"required"`
	Name          string    `json:"name" binding:"required"`
	TargetAmount  float64   `json:"target_amount" binding:"required,gt=0"`
	CurrentAmount float64   `json:"current_amount" binding:"gte=0"`
	TargetDate    time.Time `json:"target_date"`
	Type          string    `json:"type" binding:"required,oneof=savings debt investment purchase emergency retirement education other"`
	Priority      string    `json:"priority" binding:"omitempty,oneof=low medium high critical"`
}

// DirectRatingInput represents simplified input using direct ratings instead of pairwise comparisons
// Reduces input from 33 comparisons (3 criteria + 5 goals) to 18 ratings (3 + 15)
type DirectRatingInput struct {
	UserID        string  `json:"user_id" binding:"required"`
	MonthlyIncome float64 `json:"monthly_income" binding:"required,gt=0"`

	// Criteria importance ratings (1-10 scale)
	// These will be normalized to weights summing to 1.0
	CriteriaRatings map[string]int `json:"criteria_ratings" binding:"required"`
	// Expected keys: "urgency", "importance", "feasibility", "impact"
	// Values: 1 (not important) to 10 (very important)

	// Goals to prioritize
	Goals []GoalForRating `json:"goals" binding:"required,min=2,dive"`
}

// Validate performs additional validation on DirectRatingInput
func (dri *DirectRatingInput) Validate() error {
	// Validate criteria ratings
	requiredCriteria := []string{"urgency", "importance", "feasibility", "impact"}
	for _, criterion := range requiredCriteria {
		rating, exists := dri.CriteriaRatings[criterion]
		if !exists {
			return fmt.Errorf("missing criterion rating: %s", criterion)
		}
		if rating < 1 || rating > 10 {
			return fmt.Errorf("%s rating must be between 1 and 10", criterion)
		}
	}

	// Validate goals minimum
	if len(dri.Goals) < 2 {
		return errors.New("at least 2 goals required")
	}

	return nil
}

// BWMInput represents input for Best-Worst Method
// Reduces input from 33 comparisons to ~21 comparisons (36% reduction)
type BWMInput struct {
	UserID        string  `json:"user_id" binding:"required"`
	MonthlyIncome float64 `json:"monthly_income" binding:"required,gt=0"`

	// Step 1: User identifies best and worst criteria
	BestCriterion  string `json:"best_criterion" binding:"required,oneof=urgency importance feasibility impact"`
	WorstCriterion string `json:"worst_criterion" binding:"required,oneof=urgency importance feasibility impact"`

	// Step 2: Compare best criterion with all others (n-1 comparisons)
	// Key:  "{criterion}" (e.g., "importance")
	// Value: How much better is best compared to this? (1-9 scale)
	BestToOthers map[string]float64 `json:"best_to_others" binding:"required"`

	// Step 3: Compare all others with worst criterion (n-1 comparisons)
	// Key:  "{criterion}" (e.g., "urgency")
	// Value: How much better is this compared to worst? (1-9 scale)
	OthersToWorst map[string]float64 `json:"others_to_worst" binding:"required"`

	// Goals (auto-scored)
	Goals []GoalForRating `json:"goals" binding:"required,min=2,dive"`
}

// AutoScoringInput represents input where criteria scores are fully auto-calculated
// Minimal user input - just provide goals and income
type AutoScoringInput struct {
	UserID        string  `json:"user_id" binding:"required"`
	MonthlyIncome float64 `json:"monthly_income" binding:"required,gt=0"`

	// Optional: custom criteria weights (if user wants to override defaults)
	// If nil, uses equal weights of 0.25 each
	CriteriaWeights map[string]float64 `json:"criteria_weights,omitempty"`

	// Goals to prioritize (scores will be auto-calculated)
	Goals []GoalForRating `json:"goals" binding:"required,min=2,dive"`
}

// DirectRatingOutput is the same as AHPOutput but includes method metadata
type DirectRatingOutput struct {
	*AHPOutput

	// Metadata about the method used
	Method             string  `json:"method"`           // "direct_rating", "bwm", "auto_scoring", or "full_ahp"
	InputComplexity    int     `json:"input_complexity"` // Number of user inputs required
	ExecutionTimeMs    float64 `json:"execution_time_ms"`
	UsedAutoScoring    bool    `json:"used_auto_scoring"`
	AutoScoredCriteria int     `json:"auto_scored_criteria"` // How many criteria were auto-scored
}
