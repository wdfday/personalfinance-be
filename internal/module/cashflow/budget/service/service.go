package service

import (
	"context"
	"personalfinancedss/internal/module/cashflow/budget/domain"
	"time"

	"github.com/google/uuid"
)

// Service defines the interface for budget business logic
type Service interface {
	// CreateBudget creates a new budget
	CreateBudget(ctx context.Context, budget *domain.Budget) error

	// GetBudgetByID retrieves a budget by ID
	GetBudgetByID(ctx context.Context, budgetID uuid.UUID) (*domain.Budget, error)

	// GetUserBudgets retrieves all budgets for a user
	GetUserBudgets(ctx context.Context, userID uuid.UUID) ([]domain.Budget, error)

	// GetActiveBudgets retrieves all active budgets for a user
	GetActiveBudgets(ctx context.Context, userID uuid.UUID) ([]domain.Budget, error)

	// GetBudgetsByCategory retrieves budgets for a specific category
	GetBudgetsByCategory(ctx context.Context, userID, categoryID uuid.UUID) ([]domain.Budget, error)

	// GetBudgetsByAccount retrieves budgets for a specific account
	GetBudgetsByAccount(ctx context.Context, userID, accountID uuid.UUID) ([]domain.Budget, error)

	// UpdateBudget updates an existing budget
	UpdateBudget(ctx context.Context, budget *domain.Budget) error

	// DeleteBudget deletes a budget
	DeleteBudget(ctx context.Context, budgetID uuid.UUID) error

	// RecalculateBudgetSpending recalculates the spent amount for a budget
	RecalculateBudgetSpending(ctx context.Context, budgetID uuid.UUID) error

	// RecalculateAllBudgets recalculates spending for all active budgets
	RecalculateAllBudgets(ctx context.Context, userID uuid.UUID) error

	// CheckBudgetAlerts checks if any budget alerts should be triggered
	CheckBudgetAlerts(ctx context.Context, budgetID uuid.UUID) ([]domain.AlertThreshold, error)

	// MarkExpiredBudgets marks expired budgets as expired
	MarkExpiredBudgets(ctx context.Context) error

	// RolloverBudgets processes budget rollovers for the new period
	RolloverBudgets(ctx context.Context, userID uuid.UUID) error

	// GetBudgetSummary gets a summary of budget performance
	GetBudgetSummary(ctx context.Context, userID uuid.UUID, period time.Time) (*BudgetSummary, error)
}

// BudgetSummary represents a summary of budget performance
type BudgetSummary struct {
	TotalBudgets      int                           `json:"total_budgets"`
	ActiveBudgets     int                           `json:"active_budgets"`
	ExceededBudgets   int                           `json:"exceeded_budgets"`
	WarningBudgets    int                           `json:"warning_budgets"`
	TotalAmount       float64                       `json:"total_amount"`
	TotalSpent        float64                       `json:"total_spent"`
	TotalRemaining    float64                       `json:"total_remaining"`
	AveragePercentage float64                       `json:"average_percentage"`
	BudgetsByCategory map[string]*CategoryBudgetSum `json:"budgets_by_category"`
}

// CategoryBudgetSum represents budget summary for a category
type CategoryBudgetSum struct {
	CategoryID   uuid.UUID `json:"category_id"`
	CategoryName string    `json:"category_name"`
	Amount       float64   `json:"amount"`
	Spent        float64   `json:"spent"`
	Remaining    float64   `json:"remaining"`
	Percentage   float64   `json:"percentage"`
}
