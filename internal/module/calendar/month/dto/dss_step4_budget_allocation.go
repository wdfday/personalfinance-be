package dto

import (
	budgetDto "personalfinancedss/internal/module/analytics/budget_allocation/dto"

	"github.com/google/uuid"
)

// ==================== Step 4: Budget Allocation ====================
// Constraints, Goals, Debts are READ FROM REDIS CACHE

// ScenarioParametersOverride allows user to customize scenario parameters
type ScenarioParametersOverride struct {
	ScenarioType           string   `json:"scenario_type" binding:"required,oneof=safe balanced"` // "safe" or "balanced" (matches domain.ScenarioType)
	GoalContributionFactor *float64 `json:"goal_contribution_factor,omitempty"`                   // Multiplier for goal contributions (0.0-2.0)
	FlexibleSpendingLevel  *float64 `json:"flexible_spending_level,omitempty"`                    // 0.0 = minimum, 1.0 = maximum
	EmergencyFundPercent   *float64 `json:"emergency_fund_percent,omitempty"`                     // % of surplus to emergency fund (0.0-1.0)
	GoalsPercent           *float64 `json:"goals_percent,omitempty"`                              // % of surplus to goals (0.0-1.0)
	FlexiblePercent        *float64 `json:"flexible_percent,omitempty"`                           // % of surplus to flexible spending (0.0-1.0)
}

// PreviewBudgetAllocationRequest requests budget allocation preview
// All input data is read from cached DSS state
type PreviewBudgetAllocationRequest struct {
	MonthID           uuid.UUID                    `json:"month_id" binding:"required"`
	GoalAllocationPct float64                      `json:"goal_allocation_pct" binding:"gte=0,lte=100"`                     // From Step 3: adjusts suggested contribution for goals
	DebtAllocationPct *float64                     `json:"debt_allocation_pct,omitempty" binding:"omitempty,gte=0,lte=100"` // Optional: not used in allocation (debts from Step 2), only for default logic
	ScenarioOverrides []ScenarioParametersOverride `json:"scenario_overrides,omitempty"`                                    // Optional: custom scenario parameters
}

// PreviewBudgetAllocationResponse type alias to analytics output
type PreviewBudgetAllocationResponse = budgetDto.BudgetAllocationModelOutput

// ApplyBudgetAllocationRequest applies selected scenario
// IMPORTANT: Frontend must send the exact allocations to apply.
// Backend ONLY saves what frontend sends - no re-calculation or modification.
// Frontend/user is responsible for computing allocations (from scenario preview or custom edits).
type ApplyBudgetAllocationRequest struct {
	MonthID          uuid.UUID             `json:"month_id" binding:"required"`
	SelectedScenario string                `json:"selected_scenario" binding:"required"` // For reference only
	Allocations      map[uuid.UUID]float64 `json:"allocations" binding:"required"`       // CategoryID -> Amount (exactly as frontend computed)
}

// Re-export analytics types for convenience
type BudgetAllocationModelInput = budgetDto.BudgetAllocationModelInput
type BudgetAllocationModelOutput = budgetDto.BudgetAllocationModelOutput
type MandatoryExpense = budgetDto.MandatoryExpense
type FlexibleExpense = budgetDto.FlexibleExpense
