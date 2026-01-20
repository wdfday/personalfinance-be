package dto

import (
	budgetDto "personalfinancedss/internal/module/analytics/budget_allocation/dto"

	"github.com/google/uuid"
)

// ==================== Step 4: Budget Allocation ====================
// Import types from analytics budget_allocation module - NO DUPLICATION

// BudgetConstraintInput represents budget constraints from frontend
type BudgetConstraintInput struct {
	CategoryID   uuid.UUID `json:"category_id"`
	CategoryName string    `json:"category_name" binding:"required"`
	MinAmount    float64   `json:"min_amount" binding:"gte=0"`
	MaxAmount    float64   `json:"max_amount,omitempty"`
	Flexibility  string    `json:"flexibility,omitempty"` // "fixed" | "flexible" | "discretionary"
	Priority     int       `json:"priority,omitempty"`
	IsAdHoc      bool      `json:"is_ad_hoc"`
	Description  string    `json:"description,omitempty"`
}

// PreviewBudgetAllocationRequest requests budget allocation preview
type PreviewBudgetAllocationRequest struct {
	MonthID           uuid.UUID               `json:"month_id" binding:"required"`
	Constraints       []BudgetConstraintInput `json:"constraints" binding:"required,min=1"`
	TotalIncome       float64                 `json:"total_income" binding:"required,gt=0"`
	GoalAllocationPct float64                 `json:"goal_allocation_pct" binding:"gte=0,lte=100"`
	DebtAllocationPct float64                 `json:"debt_allocation_pct" binding:"gte=0,lte=100"`
}

// PreviewBudgetAllocationResponse type alias to analytics output
type PreviewBudgetAllocationResponse = budgetDto.BudgetAllocationModelOutput

// ApplyBudgetAllocationRequest applies selected scenario
// Frontend must send allocations from the selected scenario (not re-computed by backend)
type ApplyBudgetAllocationRequest struct {
	MonthID          uuid.UUID             `json:"month_id" binding:"required"`
	SelectedScenario string                `json:"selected_scenario" binding:"required"`
	Allocations      map[uuid.UUID]float64 `json:"allocations" binding:"required"` // CategoryID -> Amount
}

// Re-export analytics types for convenience
type BudgetAllocationModelInput = budgetDto.BudgetAllocationModelInput
type BudgetAllocationModelOutput = budgetDto.BudgetAllocationModelOutput
type MandatoryExpense = budgetDto.MandatoryExpense
type FlexibleExpense = budgetDto.FlexibleExpense
