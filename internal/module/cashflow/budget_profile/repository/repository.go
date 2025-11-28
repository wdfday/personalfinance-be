package repository

import (
	"context"

	"personalfinancedss/internal/module/cashflow/budget_profile/domain"
	"personalfinancedss/internal/module/cashflow/budget_profile/dto"

	"github.com/google/uuid"
)

// Repository defines budget constraint data access operations with versioning support
type Repository interface {
	// Create creates a new budget constraint
	Create(ctx context.Context, bc *domain.BudgetConstraint) error

	// GetByID retrieves a budget constraint by ID
	GetByID(ctx context.Context, id uuid.UUID) (*domain.BudgetConstraint, error)

	// GetByUser retrieves all active budget constraints for a user (not archived)
	GetByUser(ctx context.Context, userID uuid.UUID) (domain.BudgetConstraints, error)

	// GetActiveByUser retrieves all currently active budget constraints
	GetActiveByUser(ctx context.Context, userID uuid.UUID) (domain.BudgetConstraints, error)

	// GetArchivedByUser retrieves all archived budget constraints
	GetArchivedByUser(ctx context.Context, userID uuid.UUID) (domain.BudgetConstraints, error)

	// GetByUserAndCategory retrieves active constraint by user and category
	GetByUserAndCategory(ctx context.Context, userID, categoryID uuid.UUID) (*domain.BudgetConstraint, error)

	// GetByStatus retrieves constraints by user and status
	GetByStatus(ctx context.Context, userID uuid.UUID, status domain.ConstraintStatus) (domain.BudgetConstraints, error)

	// GetVersionHistory retrieves all versions of a constraint
	GetVersionHistory(ctx context.Context, constraintID uuid.UUID) (domain.BudgetConstraints, error)

	// GetLatestVersion retrieves the latest version of a constraint chain
	GetLatestVersion(ctx context.Context, constraintID uuid.UUID) (*domain.BudgetConstraint, error)

	// List retrieves budget constraints with filters
	List(ctx context.Context, userID uuid.UUID, query dto.ListBudgetConstraintsQuery) (domain.BudgetConstraints, error)

	// Update updates an existing budget constraint
	Update(ctx context.Context, bc *domain.BudgetConstraint) error

	// Delete soft deletes a budget constraint
	Delete(ctx context.Context, id uuid.UUID) error

	// DeleteByUserAndCategory deletes a budget constraint by user and category
	DeleteByUserAndCategory(ctx context.Context, userID, categoryID uuid.UUID) error

	// Archive archives a budget constraint
	Archive(ctx context.Context, id uuid.UUID, archivedBy uuid.UUID) error

	// Exists checks if a budget constraint exists for user and category
	Exists(ctx context.Context, userID, categoryID uuid.UUID) (bool, error)

	// GetTotalMandatory calculates total mandatory expenses for user
	GetTotalMandatory(ctx context.Context, userID uuid.UUID) (float64, error)
}
