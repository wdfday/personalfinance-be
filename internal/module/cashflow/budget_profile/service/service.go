package service

import (
	"context"

	"personalfinancedss/internal/module/cashflow/budget_profile/domain"
	"personalfinancedss/internal/module/cashflow/budget_profile/dto"
	"personalfinancedss/internal/module/cashflow/budget_profile/repository"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// BudgetConstraintCreator defines budget constraint creation operations
type BudgetConstraintCreator interface {
	CreateBudgetConstraint(ctx context.Context, userID string, req dto.CreateBudgetConstraintRequest) (*domain.BudgetConstraint, error)
}

// BudgetConstraintReader defines budget constraint read operations
type BudgetConstraintReader interface {
	GetBudgetConstraint(ctx context.Context, userID string, constraintID string) (*domain.BudgetConstraint, error)
	GetBudgetConstraintWithHistory(ctx context.Context, userID string, constraintID string) (*domain.BudgetConstraint, domain.BudgetConstraints, error)
	GetBudgetConstraintByCategory(ctx context.Context, userID string, categoryID string) (*domain.BudgetConstraint, error)
	ListBudgetConstraints(ctx context.Context, userID string, query dto.ListBudgetConstraintsQuery) (domain.BudgetConstraints, error)
	GetActiveConstraints(ctx context.Context, userID string) (domain.BudgetConstraints, error)
	GetArchivedConstraints(ctx context.Context, userID string) (domain.BudgetConstraints, error)
	GetBudgetConstraintSummary(ctx context.Context, userID string) (*dto.BudgetConstraintSummaryResponse, error)
}

// BudgetConstraintUpdater defines budget constraint update operations (creates new version)
type BudgetConstraintUpdater interface {
	// UpdateBudgetConstraint creates a NEW version and archives the old one
	UpdateBudgetConstraint(ctx context.Context, userID string, constraintID string, req dto.UpdateBudgetConstraintRequest) (*domain.BudgetConstraint, error)

	// ArchiveBudgetConstraint manually archives a constraint
	ArchiveBudgetConstraint(ctx context.Context, userID string, constraintID string) error

	// CheckAndArchiveEnded checks and archives ended constraints
	CheckAndArchiveEnded(ctx context.Context, userID string) (int, error)
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

// parseUserID parses user ID string to UUID
func parseUserID(userID string) (uuid.UUID, error) {
	return uuid.Parse(userID)
}

// parseConstraintID parses constraint ID string to UUID
func parseConstraintID(constraintID string) (uuid.UUID, error) {
	return uuid.Parse(constraintID)
}
