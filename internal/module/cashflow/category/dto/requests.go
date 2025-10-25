package dto

// CreateCategoryRequest represents request to create a new category
type CreateCategoryRequest struct {
	Name        string  `json:"name" binding:"required,min=1,max=100"`
	Description *string `json:"description,omitempty" binding:"omitempty,max=500"`
	Type        string  `json:"type" binding:"required,oneof=income expense both"`

	// Hierarchy
	ParentID *string `json:"parent_id,omitempty" binding:"omitempty,uuid"`

	// Visual
	Icon  *string `json:"icon,omitempty" binding:"omitempty,max=50"`
	Color *string `json:"color,omitempty" binding:"omitempty,max=20"` // Hex color

	// Budget
	MonthlyBudget *float64 `json:"monthly_budget,omitempty" binding:"omitempty,gte=0"`

	// Display
	DisplayOrder *int `json:"display_order,omitempty" binding:"omitempty,gte=0"`
}

// UpdateCategoryRequest represents request to update a category
type UpdateCategoryRequest struct {
	Name        *string `json:"name,omitempty" binding:"omitempty,min=1,max=100"`
	Description *string `json:"description,omitempty" binding:"omitempty,max=500"`
	Type        *string `json:"type,omitempty" binding:"omitempty,oneof=income expense both"`

	// Hierarchy
	ParentID *string `json:"parent_id,omitempty" binding:"omitempty,uuid"`

	// Visual
	Icon  *string `json:"icon,omitempty" binding:"omitempty,max=50"`
	Color *string `json:"color,omitempty" binding:"omitempty,max=20"`

	// Budget
	MonthlyBudget *float64 `json:"monthly_budget,omitempty" binding:"omitempty,gte=0"`

	// Display
	DisplayOrder *int  `json:"display_order,omitempty" binding:"omitempty,gte=0"`
	IsActive     *bool `json:"is_active,omitempty"`
}

// ListCategoriesQuery represents query parameters for listing categories
type ListCategoriesQuery struct {
	Type         *string `form:"type" binding:"omitempty,oneof=income expense both"`
	ParentID     *string `form:"parent_id" binding:"omitempty,uuid"` // Filter by parent
	IsRootOnly   bool    `form:"is_root_only"`                       // Only root categories (no parent)
	IncludeStats bool    `form:"include_stats"`                      // Include transaction statistics
	IsActive     *bool   `form:"is_active"`                          // Filter by active status
}

// InitializeDefaultCategoriesRequest represents request to initialize default categories
type InitializeDefaultCategoriesRequest struct {
	IncludeIncome  bool `json:"include_income" binding:"omitempty"`
	IncludeExpense bool `json:"include_expense" binding:"omitempty"`
}
