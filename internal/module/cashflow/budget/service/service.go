package service

import (
	"context"
	"personalfinancedss/internal/module/cashflow/budget/domain"
	"personalfinancedss/internal/module/cashflow/budget/repository"
	"time"

	"github.com/google/uuid"
)

// Service defines the interface for budget business logic
type Service interface {
	// Create operations
	CreateBudget(ctx context.Context, budget *domain.Budget) error

	// Read operations
	GetBudgetByID(ctx context.Context, budgetID uuid.UUID) (*domain.Budget, error)
	GetBudgetByIDForUser(ctx context.Context, budgetID, userID uuid.UUID) (*domain.Budget, error)
	GetUserBudgets(ctx context.Context, userID uuid.UUID) ([]domain.Budget, error)
	GetUserBudgetsPaginated(ctx context.Context, userID uuid.UUID, page, pageSize int) (*repository.PaginatedResult, error)
	GetActiveBudgets(ctx context.Context, userID uuid.UUID) ([]domain.Budget, error)
	GetBudgetsByCategory(ctx context.Context, userID, categoryID uuid.UUID) ([]domain.Budget, error)
	GetBudgetsByAccount(ctx context.Context, userID, accountID uuid.UUID) ([]domain.Budget, error)
	GetBudgetsByPeriod(ctx context.Context, userID uuid.UUID, period domain.BudgetPeriod) ([]domain.Budget, error)
	GetBudgetSummary(ctx context.Context, userID uuid.UUID, period time.Time) (*BudgetSummary, error)
	GetBudgetVsActual(ctx context.Context, userID uuid.UUID, period domain.BudgetPeriod, startDate, endDate time.Time) ([]*BudgetVsActual, error)
	GetBudgetProgress(ctx context.Context, budgetID, userID uuid.UUID) (*BudgetProgress, error)
	GetBudgetAnalytics(ctx context.Context, budgetID, userID uuid.UUID) (*BudgetAnalytics, error)

	// Update operations
	UpdateBudget(ctx context.Context, budget *domain.Budget) error
	UpdateBudgetForUser(ctx context.Context, budget *domain.Budget, userID uuid.UUID) error
	CheckBudgetAlerts(ctx context.Context, budgetID uuid.UUID) ([]domain.AlertThreshold, error)
	MarkExpiredBudgets(ctx context.Context) error

	// Delete operations
	DeleteBudget(ctx context.Context, budgetID uuid.UUID) error
	DeleteBudgetForUser(ctx context.Context, budgetID, userID uuid.UUID) error

	// Calculation operations
	RecalculateBudgetSpending(ctx context.Context, budgetID uuid.UUID) error
	RecalculateBudgetSpendingForUser(ctx context.Context, budgetID, userID uuid.UUID) error
	RecalculateAllBudgets(ctx context.Context, userID uuid.UUID) error
	RolloverBudgets(ctx context.Context, userID uuid.UUID) error
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

// BudgetVsActual represents budget vs actual comparison
type BudgetVsActual struct {
	BudgetID     uuid.UUID  `json:"budget_id"`
	CategoryID   *uuid.UUID `json:"category_id,omitempty"`
	CategoryName string     `json:"category_name,omitempty"`
	BudgetAmount float64    `json:"budget_amount"`
	ActualSpent  float64    `json:"actual_spent"`
	Difference   float64    `json:"difference"`
	Percentage   float64    `json:"percentage"`
	Status       string     `json:"status"` // under, on_track, over
}

// BudgetProgress represents detailed budget progress
type BudgetProgress struct {
	BudgetID         uuid.UUID           `json:"budget_id"`
	Name             string              `json:"name"`
	Period           domain.BudgetPeriod `json:"period"`
	StartDate        time.Time           `json:"start_date"`
	EndDate          *time.Time          `json:"end_date,omitempty"`
	Amount           float64             `json:"amount"`
	SpentAmount      float64             `json:"spent_amount"`
	RemainingAmount  float64             `json:"remaining_amount"`
	PercentageSpent  float64             `json:"percentage_spent"`
	Status           domain.BudgetStatus `json:"status"`
	DaysElapsed      int                 `json:"days_elapsed"`
	DaysRemaining    int                 `json:"days_remaining"`
	DailyAverage     float64             `json:"daily_average"`
	ProjectedTotal   float64             `json:"projected_total"`
	OnTrack          bool                `json:"on_track"`
	TransactionCount int                 `json:"transaction_count"`
	LastTransaction  *time.Time          `json:"last_transaction,omitempty"`
}

// BudgetAnalytics represents budget analytics
type BudgetAnalytics struct {
	BudgetID          uuid.UUID `json:"budget_id"`
	HistoricalAverage float64   `json:"historical_average"`
	Trend             string    `json:"trend"` // increasing, stable, decreasing
	Volatility        float64   `json:"volatility"`
	ComplianceRate    float64   `json:"compliance_rate"`
	RecommendedAmount float64   `json:"recommended_amount"`
	OptimizationScore float64   `json:"optimization_score"`
}
