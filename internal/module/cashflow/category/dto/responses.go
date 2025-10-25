package dto

import "time"

// CategoryResponse represents a category in API responses
type CategoryResponse struct {
	ID     string  `json:"id"`
	UserID *string `json:"user_id,omitempty"`

	// Category details
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
	Type        string  `json:"type"`

	// Hierarchy
	ParentID *string `json:"parent_id,omitempty"`
	Level    int     `json:"level"`

	// Visual
	Icon  *string `json:"icon,omitempty"`
	Color *string `json:"color,omitempty"`

	// Flags
	IsDefault bool `json:"is_default"`
	IsActive  bool `json:"is_active"`

	// Budget
	MonthlyBudget *float64 `json:"monthly_budget,omitempty"`

	// Statistics
	TransactionCount int     `json:"transaction_count,omitempty"`
	TotalAmount      float64 `json:"total_amount,omitempty"`

	// Display
	DisplayOrder int `json:"display_order"`

	// Timestamps
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`

	// Relationships
	Parent   *CategoryResponse  `json:"parent,omitempty"`
	Children []CategoryResponse `json:"children,omitempty"`
}

// CategoryListResponse represents a list of categories
type CategoryListResponse struct {
	Categories []CategoryResponse `json:"categories"`
	Count      int                `json:"count"`
}

// CategoryStatsResponse represents statistics for a category
type CategoryStatsResponse struct {
	CategoryID       string   `json:"category_id"`
	CategoryName     string   `json:"category_name"`
	TransactionCount int      `json:"transaction_count"`
	TotalAmount      float64  `json:"total_amount"`
	AverageAmount    float64  `json:"average_amount"`
	MonthlyBudget    *float64 `json:"monthly_budget,omitempty"`
	BudgetUsed       *float64 `json:"budget_used,omitempty"`      // Percentage if budget exists
	BudgetRemaining  *float64 `json:"budget_remaining,omitempty"` // Amount remaining
}

// MessageResponse represents a simple message response
type MessageResponse struct {
	Message string `json:"message"`
}
