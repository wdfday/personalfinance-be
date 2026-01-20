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
