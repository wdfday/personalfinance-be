package service

import (
	"context"
	"personalfinancedss/internal/module/cashflow/budget/domain"
	"personalfinancedss/internal/module/cashflow/budget/dto"
	"personalfinancedss/internal/module/cashflow/budget/repository"
	"time"

	"github.com/google/uuid"
)

// BudgetCreator defines budget creation operations
type BudgetCreator interface {
	CreateBudget(ctx context.Context, userID uuid.UUID, req dto.CreateBudgetRequest) (*domain.Budget, error)
	CreateBudgetFromDomain(ctx context.Context, budget *domain.Budget) error // for DSS, rollover, etc.
}

// BudgetReader defines budget read operations
type BudgetReader interface {
	GetBudgetByIDForUser(ctx context.Context, budgetID, userID uuid.UUID) (*domain.Budget, error)
	GetUserBudgets(ctx context.Context, userID uuid.UUID) ([]domain.Budget, error)
	GetUserBudgetsPaginated(ctx context.Context, userID uuid.UUID, page, pageSize int) (*repository.PaginatedResult, error)
	GetActiveBudgets(ctx context.Context, userID uuid.UUID) ([]domain.Budget, error)
	GetBudgetsByCategory(ctx context.Context, userID, categoryID uuid.UUID) ([]domain.Budget, error)
	GetBudgetsByConstraint(ctx context.Context, userID, constraintID uuid.UUID) ([]domain.Budget, error)
	GetBudgetsByPeriod(ctx context.Context, userID uuid.UUID, period domain.BudgetPeriod) ([]domain.Budget, error)
	GetBudgetSummary(ctx context.Context, userID uuid.UUID, period time.Time) (*dto.BudgetSummary, error)
	GetBudgetVsActual(ctx context.Context, userID uuid.UUID, period domain.BudgetPeriod, startDate, endDate time.Time) ([]*dto.BudgetVsActual, error)
	GetBudgetProgress(ctx context.Context, budgetID, userID uuid.UUID) (*dto.BudgetProgress, error)
	GetBudgetAnalytics(ctx context.Context, budgetID, userID uuid.UUID) (*dto.BudgetAnalytics, error)
}

// BudgetUpdater defines budget update operations
type BudgetUpdater interface {
	UpdateBudgetForUser(ctx context.Context, budget *domain.Budget, userID uuid.UUID) error
	CheckBudgetAlerts(ctx context.Context, budgetID uuid.UUID) ([]domain.AlertThreshold, error)
	MarkExpiredBudgets(ctx context.Context) error
}

// BudgetDeleter defines budget delete operations
type BudgetDeleter interface {
	DeleteBudgetForUser(ctx context.Context, budgetID, userID uuid.UUID) error
}

// BudgetCalculator defines budget calculation operations
type BudgetCalculator interface {
	RecalculateBudgetSpendingForUser(ctx context.Context, budgetID, userID uuid.UUID) error
	RecalculateAllBudgets(ctx context.Context, userID uuid.UUID) error
	RolloverBudgets(ctx context.Context, userID uuid.UUID) error
}

// Service is the composite interface for all budget operations
type Service interface {
	BudgetCreator
	BudgetReader
	BudgetUpdater
	BudgetDeleter
	BudgetCalculator
}
