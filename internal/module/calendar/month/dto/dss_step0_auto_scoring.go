package dto

import (
	goalDto "personalfinancedss/internal/module/analytics/goal_prioritization/dto"

	"github.com/google/uuid"
)

// ==================== Step 0: Auto-Scoring Preview ====================
// This is a preview-only step - no results are saved to DSSWorkflow
// All types imported from analytics module - NO DUPLICATION

// PreviewAutoScoringRequest requests auto-scoring for goals
type PreviewAutoScoringRequest struct {
	MonthID       uuid.UUID               `json:"month_id"`
	MonthlyIncome float64                 `json:"monthly_income" binding:"required,gt=0"`
	Goals         []goalDto.GoalForRating `json:"goals" binding:"required,min=1"`
}

// PreviewAutoScoringResponse type alias to analytics response
type PreviewAutoScoringResponse = goalDto.AutoScoresResponse

// Re-export analytics types for convenience
type GoalForRating = goalDto.GoalForRating
type GoalAutoScores = goalDto.GoalAutoScores
type ScoreWithReason = goalDto.ScoreWithReason
