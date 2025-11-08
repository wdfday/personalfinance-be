package dto

// CreateIncomeProfileRequest represents request to create a new income profile
type CreateIncomeProfileRequest struct {
	Year  int `json:"year" binding:"required,gte=2000,lte=2100"`
	Month int `json:"month" binding:"required,gte=1,lte=12"`

	// Income components
	BaseSalary      *float64 `json:"base_salary,omitempty" binding:"omitempty,gte=0"`
	Bonus           *float64 `json:"bonus,omitempty" binding:"omitempty,gte=0"`
	FreelanceIncome *float64 `json:"freelance_income,omitempty" binding:"omitempty,gte=0"`
	OtherIncome     *float64 `json:"other_income,omitempty" binding:"omitempty,gte=0"`

	// Status
	IsActual *bool   `json:"is_actual,omitempty"`
	Notes    *string `json:"notes,omitempty" binding:"omitempty,max=1000"`
}

// UpdateIncomeProfileRequest represents request to update an income profile
type UpdateIncomeProfileRequest struct {
	// Income components
	BaseSalary      *float64 `json:"base_salary,omitempty" binding:"omitempty,gte=0"`
	Bonus           *float64 `json:"bonus,omitempty" binding:"omitempty,gte=0"`
	FreelanceIncome *float64 `json:"freelance_income,omitempty" binding:"omitempty,gte=0"`
	OtherIncome     *float64 `json:"other_income,omitempty" binding:"omitempty,gte=0"`

	// Status
	IsActual *bool   `json:"is_actual,omitempty"`
	Notes    *string `json:"notes,omitempty" binding:"omitempty,max=1000"`
}

// ListIncomeProfilesQuery represents query parameters for listing income profiles
type ListIncomeProfilesQuery struct {
	Year     *int  `form:"year" binding:"omitempty,gte=2000,lte=2100"`
	IsActual *bool `form:"is_actual"`
}
