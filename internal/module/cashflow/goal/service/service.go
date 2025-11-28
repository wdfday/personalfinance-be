package service

import (
	"context"
	"personalfinancedss/internal/module/cashflow/goal/domain"
	"time"

	"github.com/google/uuid"
)

// Service defines the interface for goal business logic
type Service interface {
	// Create operations
	CreateGoal(ctx context.Context, goal *domain.Goal) error

	// Read operations
	GetGoalByID(ctx context.Context, goalID uuid.UUID) (*domain.Goal, error)
	GetUserGoals(ctx context.Context, userID uuid.UUID) ([]domain.Goal, error)
	GetActiveGoals(ctx context.Context, userID uuid.UUID) ([]domain.Goal, error)
	GetGoalsByType(ctx context.Context, userID uuid.UUID, goalType domain.GoalType) ([]domain.Goal, error)
	GetCompletedGoals(ctx context.Context, userID uuid.UUID) ([]domain.Goal, error)
	GetGoalSummary(ctx context.Context, userID uuid.UUID) (*GoalSummary, error)
	GetGoalProgress(ctx context.Context, goalID uuid.UUID) (*GoalProgress, error)
	GetGoalAnalytics(ctx context.Context, goalID uuid.UUID) (*GoalAnalytics, error)

	// Update operations
	UpdateGoal(ctx context.Context, goal *domain.Goal) error
	CalculateProgress(ctx context.Context, goalID uuid.UUID) error
	MarkAsCompleted(ctx context.Context, goalID uuid.UUID) error
	CheckOverdueGoals(ctx context.Context, userID uuid.UUID) error

	// Delete operations
	DeleteGoal(ctx context.Context, goalID uuid.UUID) error

	// Contribution operations
	AddContribution(ctx context.Context, goalID uuid.UUID, amount float64) error
}

// GoalSummary represents a summary of user's goals
type GoalSummary struct {
	TotalGoals         int                     `json:"total_goals"`
	ActiveGoals        int                     `json:"active_goals"`
	CompletedGoals     int                     `json:"completed_goals"`
	OverdueGoals       int                     `json:"overdue_goals"`
	TotalTargetAmount  float64                 `json:"total_target_amount"`
	TotalCurrentAmount float64                 `json:"total_current_amount"`
	TotalRemaining     float64                 `json:"total_remaining"`
	AverageProgress    float64                 `json:"average_progress"`
	GoalsByType        map[string]*GoalTypeSum `json:"goals_by_type"`
	GoalsByPriority    map[string]int          `json:"goals_by_priority"`
}

// GoalTypeSum represents summary for a goal type
type GoalTypeSum struct {
	Count         int     `json:"count"`
	TargetAmount  float64 `json:"target_amount"`
	CurrentAmount float64 `json:"current_amount"`
	Progress      float64 `json:"progress"`
}

// GoalProgress represents detailed goal progress
type GoalProgress struct {
	GoalID                  uuid.UUID           `json:"goal_id"`
	Name                    string              `json:"name"`
	Type                    domain.GoalType     `json:"type"`
	Priority                domain.GoalPriority `json:"priority"`
	TargetAmount            float64             `json:"target_amount"`
	CurrentAmount           float64             `json:"current_amount"`
	RemainingAmount         float64             `json:"remaining_amount"`
	PercentageComplete      float64             `json:"percentage_complete"`
	Status                  domain.GoalStatus   `json:"status"`
	StartDate               time.Time           `json:"start_date"`
	TargetDate              *time.Time          `json:"target_date,omitempty"`
	DaysElapsed             int                 `json:"days_elapsed"`
	DaysRemaining           *int                `json:"days_remaining,omitempty"`
	TimeProgress            *float64            `json:"time_progress,omitempty"`
	OnTrack                 *bool               `json:"on_track,omitempty"`
	SuggestedContribution   *float64            `json:"suggested_contribution,omitempty"`
	ProjectedCompletionDate *time.Time          `json:"projected_completion_date,omitempty"`
}

// GoalAnalytics represents goal analytics
type GoalAnalytics struct {
	GoalID                  uuid.UUID       `json:"goal_id"`
	Name                    string          `json:"name"`
	Type                    domain.GoalType `json:"type"`
	TargetAmount            float64         `json:"target_amount"`
	CurrentAmount           float64         `json:"current_amount"`
	PercentageComplete      float64         `json:"percentage_complete"`
	Velocity                float64         `json:"velocity"` // Amount per day
	EstimatedCompletionDate *time.Time      `json:"estimated_completion_date,omitempty"`
	RiskLevel               string          `json:"risk_level"` // low, medium, high, overdue
	RecommendedContribution *float64        `json:"recommended_contribution,omitempty"`
}
