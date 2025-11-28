package dto

import "time"

// CreateBudgetConstraintRequest represents request to create a new budget constraint
type CreateBudgetConstraintRequest struct {
	CategoryID string `json:"category_id" binding:"required,uuid"`

	// Period tracking
	StartDate time.Time  `json:"start_date" binding:"required"`
	EndDate   *time.Time `json:"end_date,omitempty"`

	// Minimum required amount
	MinimumAmount float64 `json:"minimum_amount" binding:"required,gte=0"`

	// Flexibility
	IsFlexible    *bool    `json:"is_flexible,omitempty"`
	MaximumAmount *float64 `json:"maximum_amount,omitempty" binding:"omitempty,gte=0"`

	// Priority
	Priority *int `json:"priority,omitempty" binding:"omitempty,gte=1"`

	// Additional metadata
	Description *string `json:"description,omitempty" binding:"omitempty,max=1000"`
	IsRecurring *bool   `json:"is_recurring,omitempty"`
}

// UpdateBudgetConstraintRequest represents request to update a budget constraint
// NOTE: This creates a NEW version and archives the old one
type UpdateBudgetConstraintRequest struct {
	// Minimum required amount
	MinimumAmount *float64 `json:"minimum_amount,omitempty" binding:"omitempty,gte=0"`

	// Flexibility
	IsFlexible    *bool    `json:"is_flexible,omitempty"`
	MaximumAmount *float64 `json:"maximum_amount,omitempty" binding:"omitempty,gte=0"`

	// Priority
	Priority *int `json:"priority,omitempty" binding:"omitempty,gte=1"`

	// Period
	EndDate *time.Time `json:"end_date,omitempty"`

	// Additional metadata
	Description *string `json:"description,omitempty" binding:"omitempty,max=1000"`
}

// ListBudgetConstraintsQuery represents query parameters for listing budget constraints
type ListBudgetConstraintsQuery struct {
	CategoryID      *string `form:"category_id" binding:"omitempty,uuid"`
	IsFlexible      *bool   `form:"is_flexible"`
	Status          *string `form:"status" binding:"omitempty,oneof=active pending ended archived paused"`
	IncludeArchived bool    `form:"include_archived"`
}
