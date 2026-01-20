package repository

import (
	"context"
	"personalfinancedss/internal/module/cashflow/goal/domain"
	"time"

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

	// FindByCategory retrieves goals of a specific category
	FindByCategory(ctx context.Context, userID uuid.UUID, category domain.GoalCategory) ([]domain.Goal, error)

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

	// AddContribution adds a contribution amount to a goal (legacy - updates goal amount only)
	AddContribution(ctx context.Context, id uuid.UUID, amount float64) error

	// ============================================================
	// Contribution Methods
	// ============================================================

	// CreateContribution creates a new goal contribution record
	CreateContribution(ctx context.Context, contribution *domain.GoalContribution) error

	// FindContributionsByGoalID retrieves all contributions for a goal
	FindContributionsByGoalID(ctx context.Context, goalID uuid.UUID) ([]domain.GoalContribution, error)

	// FindContributionsByAccountID retrieves all contributions for an account
	FindContributionsByAccountID(ctx context.Context, accountID uuid.UUID) ([]domain.GoalContribution, error)

	// GetNetContributionsByAccountID calculates total net contributions for an account
	// Returns sum of deposits minus sum of withdrawals
	GetNetContributionsByAccountID(ctx context.Context, accountID uuid.UUID) (float64, error)

	// GetNetContributionsByGoalID calculates total net contributions for a goal
	GetNetContributionsByGoalID(ctx context.Context, goalID uuid.UUID) (float64, error)

	// GetContributionsByDateRange retrieves contributions for a goal within a date range
	GetContributionsByDateRange(ctx context.Context, goalID uuid.UUID, startDate, endDate time.Time) ([]domain.GoalContribution, error)
}
