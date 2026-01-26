package dto

import (
	"time"

	"github.com/google/uuid"
)

// MonthViewResponse is the full budget grid for a month
// MonthViewResponse is the full budget grid for a month
type MonthViewResponse struct {
	MonthID      uuid.UUID              `json:"month_id"`
	UserID       uuid.UUID              `json:"user_id"`
	Month        string                 `json:"month"`      // Display: "2024-02"
	StartDate    time.Time              `json:"start_date"` // Actual period start
	EndDate      time.Time              `json:"end_date"`   // Actual period end
	Status       string                 `json:"status"`
	Income       float64                `json:"income"`   // Total income for the month (Month Income + Rollover?) or just Month Income? Simulation implies Income.
	Budgeted     float64                `json:"budgeted"` // Total assigned to categories
	Activity     float64                `json:"activity"` // Total activity across categories
	ToBeBudgeted float64                `json:"to_be_budgeted"`
	Categories   []CategoryLineResponse `json:"categories"`
	CreatedAt    time.Time              `json:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at"`

	// DSS Workflow data (optional - only if DSS has been run)
	DSSWorkflow *DSSWorkflowSummary `json:"dss_workflow,omitempty"`

	// Goal and Debt allocations from DSS (optional - only if DSS has been run)
	GoalAllocations map[string]float64 `json:"goal_allocations,omitempty"` // goal_id -> allocated amount
	DebtAllocations map[string]float64 `json:"debt_allocations,omitempty"` // debt_id -> allocated amount
}

// DSSWorkflowSummary provides a summary of DSS workflow results
type DSSWorkflowSummary struct {
	CurrentStep    int       `json:"current_step"`
	CompletedSteps []int     `json:"completed_steps"`
	IsComplete     bool      `json:"is_complete"`
	LastUpdated    time.Time `json:"last_updated"`

	// Goal Prioritization summary
	GoalPrioritization *GoalPrioritizationSummary `json:"goal_prioritization,omitempty"`

	// Debt Strategy summary
	DebtStrategy *DebtStrategySummary `json:"debt_strategy,omitempty"`

	// Budget Allocation summary
	BudgetAllocation *BudgetAllocationSummary `json:"budget_allocation,omitempty"`
}

// GoalPrioritizationSummary is a simplified view of goal prioritization
type GoalPrioritizationSummary struct {
	Method   string               `json:"method"`
	Rankings []GoalRankingSummary `json:"rankings"`
}

// GoalRankingSummary is a simplified goal ranking
type GoalRankingSummary struct {
	GoalID          string  `json:"goal_id"`
	GoalName        string  `json:"goal_name"`
	Rank            int     `json:"rank"`
	Score           float64 `json:"score"`
	SuggestedAmount float64 `json:"suggested_amount"`
}

// DebtStrategySummary is a simplified view of debt strategy
type DebtStrategySummary struct {
	Strategy    string               `json:"strategy"`
	PaymentPlan []DebtPaymentSummary `json:"payment_plan,omitempty"`
}

// DebtPaymentSummary is a simplified debt payment plan
type DebtPaymentSummary struct {
	DebtID       string  `json:"debt_id"`
	DebtName     string  `json:"debt_name"`
	MinPayment   float64 `json:"min_payment"`
	ExtraPayment float64 `json:"extra_payment,omitempty"`
	TotalPayment float64 `json:"total_payment"`
}

// BudgetAllocationSummary is a simplified view of budget allocation
type BudgetAllocationSummary struct {
	Method              string             `json:"method"`
	OptimalityScore     float64            `json:"optimality_score"`
	CategoryAllocations map[string]float64 `json:"category_allocations"` // category_id -> amount
}

// CategoryLineResponse represents a single category row in the budget grid
type CategoryLineResponse struct {
	CategoryID uuid.UUID `json:"category_id"`
	Name       string    `json:"name"`

	// Budget equation: Available = Rollover + Assigned + Activity
	Rollover  float64 `json:"rollover"`
	Assigned  float64 `json:"assigned"`
	Activity  float64 `json:"activity"`
	Available float64 `json:"available"`

	// Configuration (optional)
	GoalTarget     *float64 `json:"goal_target,omitempty"`
	DebtMinPayment *float64 `json:"debt_min_payment,omitempty"`
	Notes          *string  `json:"notes,omitempty"`
}

// MonthResponse is a summary response for listing months
type MonthResponse struct {
	MonthID      uuid.UUID `json:"month_id"`
	UserID       uuid.UUID `json:"user_id"`
	Month        string    `json:"month"`      // Display: "2024-02"
	StartDate    time.Time `json:"start_date"` // Actual period start
	EndDate      time.Time `json:"end_date"`   // Actual period end
	Status       string    `json:"status"`
	ToBeBudgeted float64   `json:"to_be_budgeted"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// PlanningIterationResponse is the response for a planning iteration
type PlanningIterationResponse struct {
	MonthID  uuid.UUID `json:"month_id"`
	Version  int       `json:"version"`   // Iteration number (1, 2, 3...)
	Total    int       `json:"total"`     // Total iterations in this month
	IsLatest bool      `json:"is_latest"` // Is this the latest iteration?

	// Input snapshot summary
	ProjectedIncome  float64 `json:"projected_income"`
	TotalConstraints float64 `json:"total_constraints"`
	TotalGoalTargets float64 `json:"total_goal_targets"`
	TotalDebtMinimum float64 `json:"total_debt_minimum"`
	TotalAdHoc       float64 `json:"total_adhoc"`
	Disposable       float64 `json:"disposable"` // ProjectedIncome - FixedExpenses

	// State summary
	ToBeBudgeted float64   `json:"to_be_budgeted"`
	CreatedAt    time.Time `json:"created_at"`
}
