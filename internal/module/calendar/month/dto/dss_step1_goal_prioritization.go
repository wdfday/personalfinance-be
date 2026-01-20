package dto

import (
	goalDto "personalfinancedss/internal/module/analytics/goal_prioritization/dto"

	"github.com/google/uuid"
)

// ==================== Step 1: Goal Prioritization ====================

// PreviewGoalPrioritizationRequest requests AHP goal prioritization
type PreviewGoalPrioritizationRequest struct {
	MonthID         uuid.UUID                  `json:"month_id" binding:"required"`
	Goals           []goalDto.GoalForRating    `json:"goals" binding:"required,min=1"`
	CriteriaRatings map[string]int             `json:"criteria_ratings,omitempty"` // Custom weights from Step 0 (1-10 scale)
	UseAutoScoring  bool                       `json:"use_auto_scoring"`
	ManualScores    map[string]ManualGoalScore `json:"manual_scores,omitempty"`
}

// ManualGoalScore allows user to override auto-scores
type ManualGoalScore struct {
	GoalID     uuid.UUID `json:"goal_id" binding:"required"`
	Urgency    float64   `json:"urgency" binding:"required,gte=0,lte=10"`
	Importance float64   `json:"importance" binding:"required,gte=0,lte=10"`
	ROI        float64   `json:"roi" binding:"required,gte=0,lte=10"`
	Effort     float64   `json:"effort" binding:"required,gte=0,lte=10"`
}

// PreviewGoalPrioritizationResponse uses AHP output from analytics
type PreviewGoalPrioritizationResponse struct {
	*goalDto.AHPOutput
	AutoScores map[string]goalDto.GoalAutoScores `json:"auto_scores,omitempty"`
}

// ApplyGoalPrioritizationRequest applies user-accepted ranking
type ApplyGoalPrioritizationRequest struct {
	MonthID         uuid.UUID   `json:"month_id" binding:"required"`
	AcceptedRanking []uuid.UUID `json:"accepted_ranking" binding:"required"`
}
