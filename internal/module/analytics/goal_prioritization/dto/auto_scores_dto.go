package dto

// ================================
// Auto-Scores Request/Response DTOs
// For "auto-score with user review" flow
// ================================

// ScoreWithReason holds a score value and its explanation
type ScoreWithReason struct {
	Score  float64 `json:"score"`
	Reason string  `json:"reason"`
}

// AutoScoresRequest is the input for getting auto-calculated scores
type AutoScoresRequest struct {
	UserID        string          `json:"user_id" binding:"required"`
	MonthlyIncome float64         `json:"monthly_income" binding:"required,gt=0"`
	Goals         []GoalForRating `json:"goals" binding:"required,min=1,dive"`
}

// GoalAutoScores contains auto-calculated scores for a single goal
type GoalAutoScores struct {
	GoalID   string                     `json:"goal_id"`
	GoalName string                     `json:"goal_name"`
	GoalType string                     `json:"goal_type"`
	Scores   map[string]ScoreWithReason `json:"scores"`
}

// AutoScoresResponse contains auto-calculated scores for all goals
// This is returned to frontend for user review before finalizing
type AutoScoresResponse struct {
	Goals []GoalAutoScores `json:"goals"`

	// Default criteria weights (user can adjust these too)
	DefaultCriteriaWeights map[string]float64 `json:"default_criteria_weights"`

	// Suggested criteria ratings (1-10 scale, for easier user input)
	SuggestedCriteriaRatings map[string]int `json:"suggested_criteria_ratings"`
}

// DirectRatingWithOverridesInput extends DirectRatingInput with user overrides
type DirectRatingWithOverridesInput struct {
	DirectRatingInput

	// Optional: User can override auto-calculated scores per goal
	// Map structure: goalID -> criterionName -> overridden score (0.0-1.0)
	GoalScoreOverrides map[string]map[string]float64 `json:"goal_score_overrides,omitempty"`
}

// Validate performs validation on the input including overrides
func (input *DirectRatingWithOverridesInput) Validate() error {
	// First validate base input
	if err := input.DirectRatingInput.Validate(); err != nil {
		return err
	}

	// Validate score overrides if provided
	for goalID, scores := range input.GoalScoreOverrides {
		for criterion, score := range scores {
			if score < 0.0 || score > 1.0 {
				return &ValidationError{
					Field:   "goal_score_overrides." + goalID + "." + criterion,
					Message: "score must be between 0.0 and 1.0",
				}
			}
		}
	}

	return nil
}

// ValidationError represents a field validation error
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return e.Field + ": " + e.Message
}
