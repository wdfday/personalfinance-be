package dto

import (
	"personalfinancedss/internal/module/cashflow/budget/domain"
	"personalfinancedss/internal/module/cashflow/budget/service"
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

	CategoryID *uuid.UUID `json:"category_id,omitempty"`
	AccountID  *uuid.UUID `json:"account_id,omitempty"`

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
	TotalBudgets      int                                   `json:"total_budgets"`
	ActiveBudgets     int                                   `json:"active_budgets"`
	ExceededBudgets   int                                   `json:"exceeded_budgets"`
	WarningBudgets    int                                   `json:"warning_budgets"`
	TotalAmount       float64                               `json:"total_amount"`
	TotalSpent        float64                               `json:"total_spent"`
	TotalRemaining    float64                               `json:"total_remaining"`
	AveragePercentage float64                               `json:"average_percentage"`
	BudgetsByCategory map[string]*service.CategoryBudgetSum `json:"budgets_by_category"`
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
		AccountID:            budget.AccountID,
		SpentAmount:          budget.SpentAmount,
		RemainingAmount:      budget.RemainingAmount,
		PercentageSpent:      budget.PercentageSpent,
		Status:               budget.Status,
		LastCalculatedAt:     budget.LastCalculatedAt,
		EnableAlerts:         budget.EnableAlerts,
		AlertThresholds:      budget.AlertThresholds,
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

// ToBudgetSummaryResponse converts a service budget summary to response DTO
func ToBudgetSummaryResponse(summary *service.BudgetSummary) *BudgetSummaryResponse {
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
