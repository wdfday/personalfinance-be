package repository

import (
	"context"

	"personalfinancedss/internal/module/cashflow/budget_profile/domain"
	"personalfinancedss/internal/module/cashflow/budget_profile/dto"

	"github.com/google/uuid"
)

// Repository defines budget constraint data access operations
type Repository interface {
	// Create creates a new budget constraint
	Create(ctx context.Context, bc *domain.BudgetConstraint) error

	// GetByID retrieves a budget constraint by ID
	GetByID(ctx context.Context, id uuid.UUID) (*domain.BudgetConstraint, error)

	// GetByUser retrieves all budget constraints for a user
	GetByUser(ctx context.Context, userID uuid.UUID) (domain.BudgetConstraints, error)

	// GetByUserAndCategory retrieves a budget constraint by user and category
	GetByUserAndCategory(ctx context.Context, userID, categoryID uuid.UUID) (*domain.BudgetConstraint, error)

	// List retrieves budget constraints with filters
	List(ctx context.Context, userID uuid.UUID, query dto.ListBudgetConstraintsQuery) (domain.BudgetConstraints, error)

	// Update updates an existing budget constraint
	Update(ctx context.Context, bc *domain.BudgetConstraint) error

	// Delete deletes a budget constraint
	Delete(ctx context.Context, id uuid.UUID) error

	// DeleteByUserAndCategory deletes a budget constraint by user and category
	DeleteByUserAndCategory(ctx context.Context, userID, categoryID uuid.UUID) error

	// Exists checks if a budget constraint exists for user and category
	Exists(ctx context.Context, userID, categoryID uuid.UUID) (bool, error)

	// GetTotalMandatory calculates total mandatory expenses for user
	GetTotalMandatory(ctx context.Context, userID uuid.UUID) (float64, error)
}
