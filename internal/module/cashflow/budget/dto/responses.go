package dto

import (
	"personalfinancedss/internal/module/cashflow/budget/domain"
	"personalfinancedss/internal/module/cashflow/budget/repository"
	"time"

	"github.com/google/uuid"
)

// BudgetResponse represents a budget in API responses
type BudgetResponse struct {
	ID     uuid.UUID `json:"id"`
	UserID uuid.UUID `json:"user_id"`

	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`

	Amount   float64 `json:"amount"`
	Currency string  `json:"currency"`

	Period    domain.BudgetPeriod `json:"period"`
	StartDate time.Time           `json:"start_date"`
	EndDate   *time.Time          `json:"end_date,omitempty"`

	CategoryID   *uuid.UUID `json:"category_id,omitempty"`
	ConstraintID *uuid.UUID `json:"constraint_id,omitempty"` // FK to budget_constraint (if created from DSS)

	SpentAmount      float64             `json:"spent_amount"`
	RemainingAmount  float64             `json:"remaining_amount"`
	PercentageSpent  float64             `json:"percentage_spent"`
	Status           domain.BudgetStatus `json:"status"`
	LastCalculatedAt *time.Time          `json:"last_calculated_at,omitempty"`

	EnableAlerts     bool                    `json:"enable_alerts"`
	AlertThresholds  []domain.AlertThreshold `json:"alert_thresholds"`
	NotificationSent bool                    `json:"notification_sent"`

	AllowRollover    bool    `json:"allow_rollover"`
	RolloverAmount   float64 `json:"rollover_amount"`
	CarryOverPercent *int    `json:"carry_over_percent,omitempty"`

	AutoAdjust           bool    `json:"auto_adjust"`
	AutoAdjustPercentage *int    `json:"auto_adjust_percentage,omitempty"`
	AutoAdjustBasedOn    *string `json:"auto_adjust_based_on,omitempty"`

	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}

// BudgetSummaryResponse represents a budget summary in API responses
type BudgetSummaryResponse struct {
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

// BudgetSummary represents a summary of budget performance (internal service type)
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

// CategoryBudgetSum represents budget summary for a category (internal service type)
type CategoryBudgetSum struct {
	CategoryID   uuid.UUID `json:"category_id"`
	CategoryName string    `json:"category_name"`
	Amount       float64   `json:"amount"`
	Spent        float64   `json:"spent"`
	Remaining    float64   `json:"remaining"`
	Percentage   float64   `json:"percentage"`
}

// BudgetVsActual represents budget vs actual comparison (internal service type)
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

// BudgetProgress represents detailed budget progress (internal service type)
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

// BudgetAnalytics represents budget analytics (internal service type)
type BudgetAnalytics struct {
	BudgetID          uuid.UUID `json:"budget_id"`
	HistoricalAverage float64   `json:"historical_average"`
	Trend             string    `json:"trend"` // increasing, stable, decreasing
	Volatility        float64   `json:"volatility"`
	ComplianceRate    float64   `json:"compliance_rate"`
	RecommendedAmount float64   `json:"recommended_amount"`
	OptimizationScore float64   `json:"optimization_score"`
}

// ToBudgetResponse converts a domain budget to response DTO
func ToBudgetResponse(budget *domain.Budget) *BudgetResponse {
	if budget == nil {
		return nil
	}

	return &BudgetResponse{
		ID:                   budget.ID,
		UserID:               budget.UserID,
		Name:                 budget.Name,
		Description:          budget.Description,
		Amount:               budget.Amount,
		Currency:             budget.Currency,
		Period:               budget.Period,
		StartDate:            budget.StartDate,
		EndDate:              budget.EndDate,
		CategoryID:           budget.CategoryID,
		ConstraintID:         budget.ConstraintID,
		SpentAmount:          budget.SpentAmount,
		RemainingAmount:      budget.RemainingAmount,
		PercentageSpent:      budget.PercentageSpent,
		Status:               budget.Status,
		LastCalculatedAt:     budget.LastCalculatedAt,
		EnableAlerts:         budget.EnableAlerts,
		AlertThresholds:      []domain.AlertThreshold(budget.AlertThresholds),
		NotificationSent:     budget.NotificationSent,
		AllowRollover:        budget.AllowRollover,
		RolloverAmount:       budget.RolloverAmount,
		CarryOverPercent:     budget.CarryOverPercent,
		AutoAdjust:           budget.AutoAdjust,
		AutoAdjustPercentage: budget.AutoAdjustPercentage,
		AutoAdjustBasedOn:    budget.AutoAdjustBasedOn,
		CreatedAt:            budget.CreatedAt,
		UpdatedAt:            budget.UpdatedAt,
	}
}

// ToBudgetResponseList converts a list of domain budgets to response DTOs
func ToBudgetResponseList(budgets []domain.Budget) []*BudgetResponse {
	responses := make([]*BudgetResponse, len(budgets))
	for i := range budgets {
		responses[i] = ToBudgetResponse(&budgets[i])
	}
	return responses
}

// ToBudgetSummaryResponse converts a budget summary to response DTO
func ToBudgetSummaryResponse(summary *BudgetSummary) *BudgetSummaryResponse {
	if summary == nil {
		return nil
	}

	return &BudgetSummaryResponse{
		TotalBudgets:      summary.TotalBudgets,
		ActiveBudgets:     summary.ActiveBudgets,
		ExceededBudgets:   summary.ExceededBudgets,
		WarningBudgets:    summary.WarningBudgets,
		TotalAmount:       summary.TotalAmount,
		TotalSpent:        summary.TotalSpent,
		TotalRemaining:    summary.TotalRemaining,
		AveragePercentage: summary.AveragePercentage,
		BudgetsByCategory: summary.BudgetsByCategory,
	}
}

// PaginatedBudgetResponse represents a paginated list of budgets
type PaginatedBudgetResponse struct {
	Data       []*BudgetResponse `json:"data"`
	Total      int64             `json:"total"`
	Page       int               `json:"page"`
	PageSize   int               `json:"page_size"`
	TotalPages int               `json:"total_pages"`
}

// ToPaginatedBudgetResponse converts a paginated result to response DTO
func ToPaginatedBudgetResponse(result *repository.PaginatedResult) *PaginatedBudgetResponse {
	if result == nil {
		return nil
	}

	return &PaginatedBudgetResponse{
		Data:       ToBudgetResponseList(result.Data),
		Total:      result.Total,
		Page:       result.Page,
		PageSize:   result.PageSize,
		TotalPages: result.TotalPages,
	}
}

// BudgetProgressResponse represents budget progress in API responses
type BudgetProgressResponse struct {
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

// ToBudgetProgressResponse converts BudgetProgress to response DTO
func ToBudgetProgressResponse(progress *BudgetProgress) *BudgetProgressResponse {
	if progress == nil {
		return nil
	}

	return &BudgetProgressResponse{
		BudgetID:         progress.BudgetID,
		Name:             progress.Name,
		Period:           progress.Period,
		StartDate:        progress.StartDate,
		EndDate:          progress.EndDate,
		Amount:           progress.Amount,
		SpentAmount:      progress.SpentAmount,
		RemainingAmount:  progress.RemainingAmount,
		PercentageSpent:  progress.PercentageSpent,
		Status:           progress.Status,
		DaysElapsed:      progress.DaysElapsed,
		DaysRemaining:    progress.DaysRemaining,
		DailyAverage:     progress.DailyAverage,
		ProjectedTotal:   progress.ProjectedTotal,
		OnTrack:          progress.OnTrack,
		TransactionCount: progress.TransactionCount,
		LastTransaction:  progress.LastTransaction,
	}
}

// BudgetAnalyticsResponse represents budget analytics in API responses
type BudgetAnalyticsResponse struct {
	BudgetID          uuid.UUID `json:"budget_id"`
	HistoricalAverage float64   `json:"historical_average"`
	Trend             string    `json:"trend"`
	Volatility        float64   `json:"volatility"`
	ComplianceRate    float64   `json:"compliance_rate"`
	RecommendedAmount float64   `json:"recommended_amount"`
	OptimizationScore float64   `json:"optimization_score"`
}

// ToBudgetAnalyticsResponse converts BudgetAnalytics to response DTO
func ToBudgetAnalyticsResponse(analytics *BudgetAnalytics) *BudgetAnalyticsResponse {
	if analytics == nil {
		return nil
	}

	return &BudgetAnalyticsResponse{
		BudgetID:          analytics.BudgetID,
		HistoricalAverage: analytics.HistoricalAverage,
		Trend:             analytics.Trend,
		Volatility:        analytics.Volatility,
		ComplianceRate:    analytics.ComplianceRate,
		RecommendedAmount: analytics.RecommendedAmount,
		OptimizationScore: analytics.OptimizationScore,
	}
}
