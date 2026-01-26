package repository

import (
	"context"
	"personalfinancedss/internal/module/cashflow/budget/domain"
	"time"

	"github.com/google/uuid"
)

// PaginationParams contains pagination parameters
type PaginationParams struct {
	Page     int
	PageSize int
}

// PaginatedResult contains paginated results
type PaginatedResult struct {
	Data       []domain.Budget
	Total      int64
	Page       int
	PageSize   int
	TotalPages int
}

// Repository defines the interface for budget data access
type Repository interface {
	// Create creates a new budget
	Create(ctx context.Context, budget *domain.Budget) error

	// FindByID retrieves a budget by its ID
	FindByID(ctx context.Context, id uuid.UUID) (*domain.Budget, error)

	// FindByIDAndUserID retrieves a budget by ID and verifies user ownership
	FindByIDAndUserID(ctx context.Context, id, userID uuid.UUID) (*domain.Budget, error)

	// FindByUserID retrieves all budgets for a user
	FindByUserID(ctx context.Context, userID uuid.UUID) ([]domain.Budget, error)

	// FindByUserIDPaginated retrieves budgets for a user with pagination
	FindByUserIDPaginated(ctx context.Context, userID uuid.UUID, params PaginationParams) (*PaginatedResult, error)

	// FindActiveByUserID retrieves all active budgets for a user
	FindActiveByUserID(ctx context.Context, userID uuid.UUID) ([]domain.Budget, error)

	// FindByUserIDAndCategory retrieves budgets for a specific category
	FindByUserIDAndCategory(ctx context.Context, userID, categoryID uuid.UUID) ([]domain.Budget, error)

	// FindByConstraintID retrieves budgets for a specific constraint
	FindByConstraintID(ctx context.Context, userID, constraintID uuid.UUID) ([]domain.Budget, error)

	// FindByPeriod retrieves budgets for a specific period
	FindByPeriod(ctx context.Context, userID uuid.UUID, startDate, endDate time.Time) ([]domain.Budget, error)

	// Update updates an existing budget
	Update(ctx context.Context, budget *domain.Budget) error

	// Delete soft deletes a budget
	Delete(ctx context.Context, id uuid.UUID) error

	// DeleteByIDAndUserID deletes a budget with ownership verification
	DeleteByIDAndUserID(ctx context.Context, id, userID uuid.UUID) error

	// UpdateSpentAmount updates the spent amount for a budget
	UpdateSpentAmount(ctx context.Context, id uuid.UUID, spentAmount float64) error

	// FindExpiredBudgets retrieves all expired budgets
	FindExpiredBudgets(ctx context.Context) ([]domain.Budget, error)

	// FindBudgetsNeedingRecalculation retrieves budgets that need spent amount recalculation
	FindBudgetsNeedingRecalculation(ctx context.Context, threshold time.Duration) ([]domain.Budget, error)

	// ExistsByUserIDAndName checks if a budget with the same name exists for user
	ExistsByUserIDAndName(ctx context.Context, userID uuid.UUID, name string) (bool, error)
}
