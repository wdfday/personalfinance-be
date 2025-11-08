package dto

import "time"

// BudgetConstraintResponse represents a budget constraint in API responses
type BudgetConstraintResponse struct {
	ID         string `json:"id"`
	UserID     string `json:"user_id"`
	CategoryID string `json:"category_id"`

	// Minimum required amount
	MinimumAmount float64 `json:"minimum_amount"`

	// Flexibility
	IsFlexible    bool    `json:"is_flexible"`
	MaximumAmount float64 `json:"maximum_amount"`

	// Priority
	Priority int `json:"priority"`

	// Computed fields
	FlexibilityRange float64 `json:"flexibility_range,omitempty"`
	DisplayString    string  `json:"display_string,omitempty"`

	// Timestamps
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// BudgetConstraintListResponse represents a list of budget constraints
type BudgetConstraintListResponse struct {
	BudgetConstraints []BudgetConstraintResponse `json:"budget_constraints"`
	Count             int                        `json:"count"`
}

// BudgetConstraintSummaryResponse represents summary of budget constraints
type BudgetConstraintSummaryResponse struct {
	TotalMandatoryExpenses float64 `json:"total_mandatory_expenses"`
	TotalFlexible          int     `json:"total_flexible"`
	TotalFixed             int     `json:"total_fixed"`
	Count                  int     `json:"count"`
}

// MessageResponse represents a simple message response
type MessageResponse struct {
	Message string `json:"message"`
}
