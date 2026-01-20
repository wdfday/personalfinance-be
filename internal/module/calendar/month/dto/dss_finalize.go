package dto

import (
	"github.com/google/uuid"
)

// FinalizeDSSRequest contains user's final selections from all DSS steps
// This creates a new MonthState version with complete DSSWorkflowResults
// Note: Keeping simple DTOs here instead of importing analytics DTOs for API stability
type FinalizeDSSRequest struct {
	// Step 0: Auto-scoring (optional, can be skipped)
	UseAutoScoring bool `json:"use_auto_scoring"`

	// Step 1: Goal Prioritization
	GoalPriorities []GoalPrioritySelection `json:"goal_priorities" binding:"required"`

	// Step 2: Debt Strategy (optional if no debts)
	DebtStrategy *string `json:"debt_strategy,omitempty"` // "avalanche", "snowball", "hybrid", or nil

	// Step 3: Goal-Debt Tradeoff (optional if no goals or debts)
	TradeoffChoice *TradeoffChoiceSelection `json:"tradeoff_choice,omitempty"`

	// Step 4: Budget Allocation
	BudgetAllocations map[uuid.UUID]float64  `json:"budget_allocations" binding:"required"` // category_id -> amount
	GoalFundings      []GoalFundingSelection `json:"goal_fundings"`
	DebtPayments      []DebtPaymentSelection `json:"debt_payments"`

	// Metadata
	Notes string `json:"notes,omitempty"` // User notes for this DSS session
}

// GoalPrioritySelection represents user's final choice for a goal's priority
type GoalPrioritySelection struct {
	GoalID   uuid.UUID `json:"goal_id" binding:"required"`
	Priority float64   `json:"priority" binding:"required,min=0,max=1"`
	Method   string    `json:"method"` // "auto" or "manual"
}

// TradeoffChoiceSelection represents user's choice from tradeoff analysis
type TradeoffChoiceSelection struct {
	ScenarioType      string  `json:"scenario_type" binding:"required"` // "conservative", "balanced", "aggressive"
	GoalAllocationPct float64 `json:"goal_allocation_pct" binding:"required,min=0,max=1"`
	DebtAllocationPct float64 `json:"debt_allocation_pct" binding:"required,min=0,max=1"`
	ExpectedOutcome   string  `json:"expected_outcome,omitempty"`
}

// GoalFundingSelection represents user's final funding allocation for a goal
type GoalFundingSelection struct {
	GoalID             uuid.UUID `json:"goal_id" binding:"required"`
	SuggestedAmount    float64   `json:"suggested_amount" binding:"required"`
	UserAdjustedAmount *float64  `json:"user_adjusted_amount,omitempty"`
}

// DebtPaymentSelection represents user's final payment allocation for a debt
type DebtPaymentSelection struct {
	DebtID              uuid.UUID `json:"debt_id" binding:"required"`
	MinimumPayment      float64   `json:"minimum_payment" binding:"required"`
	SuggestedPayment    float64   `json:"suggested_payment" binding:"required"`
	UserAdjustedPayment *float64  `json:"user_adjusted_payment,omitempty"`
}

// FinalizeDSSResponse contains the complete MonthState with DSS results
type FinalizeDSSResponse struct {
	MonthID      uuid.UUID                  `json:"month_id"`
	StateVersion int                        `json:"state_version"` // New version number
	ToBeBudgeted float64                    `json:"to_be_budgeted"`
	Status       string                     `json:"status"` // e.g., "dss_applied"
	DSSWorkflow  *DSSWorkflowStatusResponse `json:"dss_workflow"`
	Message      string                     `json:"message"`
}
