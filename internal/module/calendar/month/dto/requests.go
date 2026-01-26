package dto

import "github.com/google/uuid"

// CreateMonthRequest is the request to create a new month
type CreateMonthRequest struct {
	Month string `json:"month" binding:"required"` // YYYY-MM format
	// UserID is taken from auth context, not from request body
}

// AssignCategoryRequest is the request to assign money to a category
type AssignCategoryRequest struct {
	MonthID    uuid.UUID `json:"month_id" binding:"required"`
	CategoryID uuid.UUID `json:"category_id" binding:"required"`
	Amount     float64   `json:"amount" binding:"required"` // Can be positive (assign) or negative (unassign)
}

// MoveMoneyRequest is the request to move money between categories
type MoveMoneyRequest struct {
	MonthID        uuid.UUID `json:"month_id" binding:"required"`
	FromCategoryID uuid.UUID `json:"from_category_id" binding:"required"`
	ToCategoryID   uuid.UUID `json:"to_category_id" binding:"required"`
	Amount         float64   `json:"amount" binding:"required,gt=0"` // Must be positive
}

// IncomeReceivedRequest is the request to add income to TBB
type IncomeReceivedRequest struct {
	MonthID uuid.UUID `json:"month_id" binding:"required"`
	Amount  float64   `json:"amount" binding:"required,gt=0"`
	Source  *string   `json:"source,omitempty"` // Optional description
}

// RecalculatePlanningRequest triggers a new planning iteration
// This appends a NEW state to the States array (does not modify existing states)
type RecalculatePlanningRequest struct {
	MonthID uuid.UUID `json:"month_id" binding:"required"`

	// Optional: Selected entities for this iteration
	SelectedGoalIDs []uuid.UUID `json:"selected_goal_ids,omitempty"`
	SelectedDebtIDs []uuid.UUID `json:"selected_debt_ids,omitempty"`

	// Optional: Ad-hoc items for this iteration
	AdHocExpenses []AdHocExpenseInput `json:"adhoc_expenses,omitempty"`
	AdHocIncome   []AdHocIncomeInput  `json:"adhoc_income,omitempty"`

	// Optional: Override projected income (otherwise calculated from profiles)
	ProjectedIncomeOverride *float64 `json:"projected_income_override,omitempty"`
}

// AdHocExpenseInput is input for one-time expense
type AdHocExpenseInput struct {
	Name         string     `json:"name" binding:"required"`
	Amount       float64    `json:"amount" binding:"required,gt=0"`
	CategoryID   *uuid.UUID `json:"category_id,omitempty"`
	CategoryHint string     `json:"category_hint,omitempty"`
	Notes        string     `json:"notes,omitempty"`
}

// AdHocIncomeInput is input for one-time income
type AdHocIncomeInput struct {
	Name   string  `json:"name" binding:"required"`
	Amount float64 `json:"amount" binding:"required,gt=0"`
	Notes  string  `json:"notes,omitempty"`
}
