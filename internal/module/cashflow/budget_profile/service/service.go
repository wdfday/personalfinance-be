package service

import (
	"context"

	"personalfinancedss/internal/module/cashflow/budget_profile/domain"
	"personalfinancedss/internal/module/cashflow/budget_profile/dto"
	"personalfinancedss/internal/module/cashflow/budget_profile/repository"

	"go.uber.org/zap"
)

// BudgetConstraintCreator defines budget constraint creation operations
type BudgetConstraintCreator interface {
	CreateBudgetConstraint(ctx context.Context, userID string, req dto.CreateBudgetConstraintRequest) (*domain.BudgetConstraint, error)
}

// BudgetConstraintReader defines budget constraint read operations
type BudgetConstraintReader interface {
	GetBudgetConstraint(ctx context.Context, userID string, constraintID string) (*domain.BudgetConstraint, error)
	GetBudgetConstraintByCategory(ctx context.Context, userID string, categoryID string) (*domain.BudgetConstraint, error)
	ListBudgetConstraints(ctx context.Context, userID string, query dto.ListBudgetConstraintsQuery) (domain.BudgetConstraints, error)
	GetBudgetConstraintSummary(ctx context.Context, userID string) (*dto.BudgetConstraintSummaryResponse, error)
}

// BudgetConstraintUpdater defines budget constraint update operations
type BudgetConstraintUpdater interface {
	UpdateBudgetConstraint(ctx context.Context, userID string, constraintID string, req dto.UpdateBudgetConstraintRequest) (*domain.BudgetConstraint, error)
}

// BudgetConstraintDeleter defines budget constraint delete operations
type BudgetConstraintDeleter interface {
	DeleteBudgetConstraint(ctx context.Context, userID string, constraintID string) error
}

// Service is the composite interface for all budget constraint operations
type Service interface {
	BudgetConstraintCreator
	BudgetConstraintReader
	BudgetConstraintUpdater
	BudgetConstraintDeleter
}

// budgetConstraintService implements all budget constraint use cases
type budgetConstraintService struct {
	repo   repository.Repository
	logger *zap.Logger
}

// NewService creates a new budget constraint service
func NewService(repo repository.Repository, logger *zap.Logger) Service {
	return &budgetConstraintService{
		repo:   repo,
		logger: logger,
	}
}
