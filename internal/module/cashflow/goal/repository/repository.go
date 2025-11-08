package repository

import (
	"context"
	"personalfinancedss/internal/module/cashflow/goal/domain"

	"github.com/google/uuid"
)

// Repository defines the interface for goal data access
type Repository interface {
	// Create creates a new goal
	Create(ctx context.Context, goal *domain.Goal) error

	// FindByID retrieves a goal by its ID
	FindByID(ctx context.Context, id uuid.UUID) (*domain.Goal, error)

	// FindByUserID retrieves all goals for a user
	FindByUserID(ctx context.Context, userID uuid.UUID) ([]domain.Goal, error)

	// FindActiveByUserID retrieves all active goals for a user
	FindActiveByUserID(ctx context.Context, userID uuid.UUID) ([]domain.Goal, error)

	// FindByType retrieves goals of a specific type
	FindByType(ctx context.Context, userID uuid.UUID, goalType domain.GoalType) ([]domain.Goal, error)

	// FindByStatus retrieves goals with a specific status
	FindByStatus(ctx context.Context, userID uuid.UUID, status domain.GoalStatus) ([]domain.Goal, error)

	// FindCompletedGoals retrieves completed goals for a user
	FindCompletedGoals(ctx context.Context, userID uuid.UUID) ([]domain.Goal, error)

	// FindOverdueGoals retrieves overdue goals
	FindOverdueGoals(ctx context.Context, userID uuid.UUID) ([]domain.Goal, error)

	// Update updates an existing goal
	Update(ctx context.Context, goal *domain.Goal) error

	// Delete soft deletes a goal
	Delete(ctx context.Context, id uuid.UUID) error

	// AddContribution adds a contribution amount to a goal
	AddContribution(ctx context.Context, id uuid.UUID, amount float64) error
}
