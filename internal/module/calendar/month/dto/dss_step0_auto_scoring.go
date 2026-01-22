package dto

import (
	goalDto "personalfinancedss/internal/module/analytics/goal_prioritization/dto"

	"github.com/google/uuid"
)

// ==================== Step 0: Auto-Scoring Preview ====================
// This is a preview-only step - no results are saved to DSSWorkflow
// Goals and Income are READ FROM REDIS CACHE (set during Initialize)

// PreviewAutoScoringRequest requests auto-scoring for goals
// Goals and income are read from cached DSS state (initialized via POST /dss/initialize)
type PreviewAutoScoringRequest struct {
	MonthID uuid.UUID `json:"month_id" binding:"required"`
	// No goals or income needed - read from Redis cache
}

// PreviewAutoScoringResponse type alias to analytics response
type PreviewAutoScoringResponse = goalDto.AutoScoresResponse

// Re-export analytics types for convenience
type GoalForRating = goalDto.GoalForRating
type GoalAutoScores = goalDto.GoalAutoScores
type ScoreWithReason = goalDto.ScoreWithReason
