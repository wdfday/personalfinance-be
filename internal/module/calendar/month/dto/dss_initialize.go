package dto

import "github.com/google/uuid"

// ==================== Initialize DSS Workflow ====================

// InitializeDSSRequest initializes DSS workflow with input snapshot
type InitializeDSSRequest struct {
	// MonthID được lấy từ path param, KHÔNG validate từ body
	MonthID       uuid.UUID `json:"month_id"`
	MonthlyIncome float64   `json:"monthly_income" binding:"required,gt=0"`

	// Input snapshot - user's selected items for this DSS session
	Goals       []InitGoalInput       `json:"goals"`
	Debts       []InitDebtInput       `json:"debts"`
	Constraints []InitConstraintInput `json:"constraints"`
}

// InitGoalInput is simplified goal input for DSS
type InitGoalInput struct {
	ID            string  `json:"id" binding:"required"`
	Name          string  `json:"name" binding:"required"`
	TargetAmount  float64 `json:"target_amount" binding:"required,gt=0"`
	CurrentAmount float64 `json:"current_amount" binding:"gte=0"`
	TargetDate    string  `json:"target_date,omitempty"`
	Type          string  `json:"type" binding:"required"`
	Priority      string  `json:"priority,omitempty"`
}

// InitDebtInput is simplified debt input for DSS
type InitDebtInput struct {
	ID             string  `json:"id" binding:"required"`
	Name           string  `json:"name" binding:"required"`
	CurrentBalance float64 `json:"current_balance" binding:"required,gt=0"`
	InterestRate   float64 `json:"interest_rate" binding:"gte=0"`
	MinimumPayment float64 `json:"minimum_payment" binding:"gte=0"`
}

// InitConstraintInput is simplified constraint input for DSS
type InitConstraintInput struct {
	ID            string    `json:"id" binding:"required"`
	Name          string    `json:"name" binding:"required"`
	CategoryID    uuid.UUID `json:"category_id" binding:"required"`
	MinimumAmount float64   `json:"minimum_amount" binding:"gte=0"`
	MaximumAmount float64   `json:"maximum_amount,omitempty"`
	IsFlexible    bool      `json:"is_flexible"`
	Priority      int       `json:"priority,omitempty"`
}

// InitializeDSSResponse confirms DSS workflow initialization
type InitializeDSSResponse struct {
	MonthID   uuid.UUID `json:"month_id"`
	Status    string    `json:"status"`
	ExpiresIn string    `json:"expires_in"` // e.g. "3h"
	Message   string    `json:"message"`

	// Summary of captured inputs
	GoalCount       int     `json:"goal_count"`
	DebtCount       int     `json:"debt_count"`
	ConstraintCount int     `json:"constraint_count"`
	MonthlyIncome   float64 `json:"monthly_income"`
}
