package service

import (
	"context"
	"personalfinancedss/internal/module/cashflow/goal/domain"

	"github.com/google/uuid"
)

// Service defines the interface for goal business logic
type Service interface {
	// CreateGoal creates a new financial goal
	CreateGoal(ctx context.Context, goal *domain.Goal) error

	// GetGoalByID retrieves a goal by ID
	GetGoalByID(ctx context.Context, goalID uuid.UUID) (*domain.Goal, error)

	// GetUserGoals retrieves all goals for a user
	GetUserGoals(ctx context.Context, userID uuid.UUID) ([]domain.Goal, error)

	// GetActiveGoals retrieves all active goals for a user
	GetActiveGoals(ctx context.Context, userID uuid.UUID) ([]domain.Goal, error)

	// GetGoalsByType retrieves goals of a specific type
	GetGoalsByType(ctx context.Context, userID uuid.UUID, goalType domain.GoalType) ([]domain.Goal, error)

	// GetCompletedGoals retrieves completed goals
	GetCompletedGoals(ctx context.Context, userID uuid.UUID) ([]domain.Goal, error)

	// UpdateGoal updates an existing goal
	UpdateGoal(ctx context.Context, goal *domain.Goal) error

	// DeleteGoal deletes a goal
	DeleteGoal(ctx context.Context, goalID uuid.UUID) error

	// AddContribution adds a contribution to a goal
	AddContribution(ctx context.Context, goalID uuid.UUID, amount float64) (*domain.Goal, error)

	// CalculateProgress recalculates progress for a goal
	CalculateProgress(ctx context.Context, goalID uuid.UUID) error

	// MarkAsCompleted marks a goal as completed
	MarkAsCompleted(ctx context.Context, goalID uuid.UUID) error

	// CheckOverdueGoals checks and marks overdue goals
	CheckOverdueGoals(ctx context.Context, userID uuid.UUID) error

	// GetGoalSummary gets a summary of all goals
	GetGoalSummary(ctx context.Context, userID uuid.UUID) (*GoalSummary, error)
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
