package dto

import (
	budgetDto "personalfinancedss/internal/module/analytics/budget_allocation/dto"

	"github.com/google/uuid"
)

// ==================== Step 4: Budget Allocation ====================
// Constraints, Goals, Debts are READ FROM REDIS CACHE

// PreviewBudgetAllocationRequest requests budget allocation preview
// All input data is read from cached DSS state
type PreviewBudgetAllocationRequest struct {
	MonthID           uuid.UUID `json:"month_id" binding:"required"`
	GoalAllocationPct float64   `json:"goal_allocation_pct" binding:"gte=0,lte=100"` // From Step 3
	DebtAllocationPct float64   `json:"debt_allocation_pct" binding:"gte=0,lte=100"` // From Step 3
	// No constraints or income needed - read from Redis cache
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
