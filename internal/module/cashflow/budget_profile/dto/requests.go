package dto

// CreateBudgetConstraintRequest represents request to create a new budget constraint
type CreateBudgetConstraintRequest struct {
	CategoryID string `json:"category_id" binding:"required,uuid"`

	// Minimum required amount
	MinimumAmount float64 `json:"minimum_amount" binding:"required,gte=0"`

	// Flexibility
	IsFlexible    *bool    `json:"is_flexible,omitempty"`
	MaximumAmount *float64 `json:"maximum_amount,omitempty" binding:"omitempty,gte=0"`

	// Priority
	Priority *int `json:"priority,omitempty" binding:"omitempty,gte=1"`
}

// UpdateBudgetConstraintRequest represents request to update a budget constraint
type UpdateBudgetConstraintRequest struct {
	// Minimum required amount
	MinimumAmount *float64 `json:"minimum_amount,omitempty" binding:"omitempty,gte=0"`

	// Flexibility
	IsFlexible    *bool    `json:"is_flexible,omitempty"`
	MaximumAmount *float64 `json:"maximum_amount,omitempty" binding:"omitempty,gte=0"`

	// Priority
	Priority *int `json:"priority,omitempty" binding:"omitempty,gte=1"`
}

// ListBudgetConstraintsQuery represents query parameters for listing budget constraints
type ListBudgetConstraintsQuery struct {
	CategoryID *string `form:"category_id" binding:"omitempty,uuid"`
	IsFlexible *bool   `form:"is_flexible"`
}
