package dto

import "time"

// BudgetConstraintResponse represents a budget constraint in API responses
type BudgetConstraintResponse struct {
	ID         string `json:"id"`
	UserID     string `json:"user_id"`
	CategoryID string `json:"category_id"`

	// Period tracking
	StartDate *time.Time `json:"start_date"`
	EndDate   *time.Time `json:"end_date,omitempty"`
	Duration  int        `json:"duration_days,omitempty"` // in days

	// Minimum required amount
	MinimumAmount float64 `json:"minimum_amount"`

	// Flexibility
	IsFlexible    bool    `json:"is_flexible"`
	MaximumAmount float64 `json:"maximum_amount"`

	// Priority
	Priority int `json:"priority"`

	// Status and lifecycle
	Status      string `json:"status"`
	IsRecurring bool   `json:"is_recurring"`
	IsActive    bool   `json:"is_active"`
	IsArchived  bool   `json:"is_archived"`

	// Computed fields
	FlexibilityRange float64 `json:"flexibility_range,omitempty"`
	DisplayString    string  `json:"display_string,omitempty"`

	// Additional metadata
	Description string   `json:"description,omitempty"`
	Tags        []string `json:"tags,omitempty"`

	// Versioning
	PreviousVersionID *string `json:"previous_version_id,omitempty"`

	// Timestamps
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
	ArchivedAt *time.Time `json:"archived_at,omitempty"`
}

// BudgetConstraintWithHistoryResponse includes version history
type BudgetConstraintWithHistoryResponse struct {
	Current        BudgetConstraintResponse   `json:"current"`
	VersionHistory []BudgetConstraintResponse `json:"version_history,omitempty"`
}

// BudgetConstraintListResponse represents a list of budget constraints
type BudgetConstraintListResponse struct {
	BudgetConstraints []BudgetConstraintResponse       `json:"budget_constraints"`
	Count             int                              `json:"count"`
	Summary           *BudgetConstraintSummaryResponse `json:"summary,omitempty"`
}

// BudgetConstraintSummaryResponse represents summary of budget constraints
type BudgetConstraintSummaryResponse struct {
	TotalMandatoryExpenses float64 `json:"total_mandatory_expenses"`
	TotalFlexible          int     `json:"total_flexible"`
	TotalFixed             int     `json:"total_fixed"`
	Count                  int     `json:"count"`
	ActiveCount            int     `json:"active_count"`
}

// MessageResponse represents a simple message response
type MessageResponse struct {
	Message string `json:"message"`
}
