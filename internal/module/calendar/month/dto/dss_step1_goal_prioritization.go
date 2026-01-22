package dto

import (
	goalDto "personalfinancedss/internal/module/analytics/goal_prioritization/dto"

	"github.com/google/uuid"
)

// ==================== Step 1: Goal Prioritization ====================
// Goals are READ FROM REDIS CACHE (set during Initialize)

// PreviewGoalPrioritizationRequest requests AHP goal prioritization
// Goals are read from cached DSS state (initialized via POST /dss/initialize)
type PreviewGoalPrioritizationRequest struct {
	MonthID         uuid.UUID      `json:"month_id" binding:"required"`
	CriteriaRatings map[string]int `json:"criteria_ratings,omitempty"` // Custom weights from Step 0 (1-10 scale)
	// No goals needed - read from Redis cache
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
