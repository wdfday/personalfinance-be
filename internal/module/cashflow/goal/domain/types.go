package domain

import (
	"time"

	"github.com/google/uuid"
)

// GoalSummary represents a summary of user's goals
type GoalSummary struct {
	TotalGoals         int                         `json:"totalGoals"`
	ActiveGoals        int                         `json:"activeGoals"`
	CompletedGoals     int                         `json:"completedGoals"`
	OverdueGoals       int                         `json:"overdueGoals"`
	TotalTargetAmount  float64                     `json:"totalTargetAmount"`
	TotalCurrentAmount float64                     `json:"totalCurrentAmount"`
	TotalRemaining     float64                     `json:"totalRemaining"`
	AverageProgress    float64                     `json:"averageProgress"`
	GoalsByCategory    map[string]*GoalCategorySum `json:"goalsByCategory"`
	GoalsByPriority    map[string]int              `json:"goalsByPriority"`
}

// GoalCategorySum represents summary for a goal category
type GoalCategorySum struct {
	Count         int     `json:"count"`
	TargetAmount  float64 `json:"targetAmount"`
	CurrentAmount float64 `json:"currentAmount"`
	Progress      float64 `json:"progress"`
}

// GoalProgress represents detailed goal progress
type GoalProgress struct {
	GoalID                  uuid.UUID    `json:"goalId"`
	Name                    string       `json:"name"`
	Behavior                GoalBehavior `json:"behavior"`
	Category                GoalCategory `json:"category"`
	Priority                GoalPriority `json:"priority"`
	TargetAmount            float64      `json:"targetAmount"`
	CurrentAmount           float64      `json:"currentAmount"`
	RemainingAmount         float64      `json:"remainingAmount"`
	PercentageComplete      float64      `json:"percentageComplete"`
	Status                  GoalStatus   `json:"status"`
	StartDate               time.Time    `json:"startDate"`
	TargetDate              *time.Time   `json:"targetDate,omitempty"`
	DaysElapsed             int          `json:"daysElapsed"`
	DaysRemaining           *int         `json:"daysRemaining,omitempty"`
	TimeProgress            *float64     `json:"timeProgress,omitempty"`
	OnTrack                 *bool        `json:"onTrack,omitempty"`
	SuggestedContribution   *float64     `json:"suggestedContribution,omitempty"`
	ProjectedCompletionDate *time.Time   `json:"projectedCompletionDate,omitempty"`
}

// GoalAnalytics represents goal analytics
type GoalAnalytics struct {
	GoalID                  uuid.UUID    `json:"goalId"`
	Name                    string       `json:"name"`
	Behavior                GoalBehavior `json:"behavior"`
	Category                GoalCategory `json:"category"`
	TargetAmount            float64      `json:"targetAmount"`
	CurrentAmount           float64      `json:"currentAmount"`
	PercentageComplete      float64      `json:"percentageComplete"`
	Velocity                float64      `json:"velocity"` // Amount per day
	EstimatedCompletionDate *time.Time   `json:"estimatedCompletionDate,omitempty"`
	RiskLevel               string       `json:"riskLevel"` // low, medium, high, overdue
	RecommendedContribution *float64     `json:"recommendedContribution,omitempty"`
}
