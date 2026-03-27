package dto

// CreateIncomeProfileRequest represents request to create a new income profile
type CreateIncomeProfileRequest struct {
	CategoryID  string        `json:"category_id" binding:"required,uuid"`
	Source      string        `json:"source" binding:"required,max=100"`
	Amount      float64       `json:"amount" binding:"required,gte=0"`
	Currency    string        `json:"currency,omitempty" binding:"omitempty,len=3"`
	Frequency   string        `json:"frequency" binding:"required,oneof=monthly weekly bi-weekly quarterly yearly one-time"`
	StartDate   FlexibleTime  `json:"start_date" binding:"required"`
	EndDate     *FlexibleTime `json:"end_date,omitempty" binding:"omitempty"`
	IsRecurring *bool         `json:"is_recurring,omitempty"`

	Description *string `json:"description,omitempty" binding:"omitempty,max=1000"`
}

// UpdateIncomeProfileRequest represents request to update an income profile
// NOTE: This creates a NEW version and archives the old one
type UpdateIncomeProfileRequest struct {
	CategoryID  *string       `json:"category_id,omitempty" binding:"omitempty,uuid"`
	Source      *string       `json:"source,omitempty" binding:"omitempty,max=100"`
	Amount      *float64      `json:"amount,omitempty" binding:"omitempty,gte=0"`
	Currency    *string       `json:"currency,omitempty" binding:"omitempty,len=3"`
	Frequency   *string       `json:"frequency,omitempty" binding:"omitempty,oneof=monthly weekly bi-weekly quarterly yearly one-time"`
	EndDate     *FlexibleTime `json:"end_date,omitempty"`
	IsRecurring *bool         `json:"is_recurring,omitempty"`

	Description *string `json:"description,omitempty" binding:"omitempty,max=1000"`
}

// ListIncomeProfilesQuery represents query parameters for listing income profiles
type ListIncomeProfilesQuery struct {
	Status          *string `form:"status" binding:"omitempty,oneof=active ended archived paused"`
	IsRecurring     *bool   `form:"is_recurring"`
	Source          *string `form:"source"`
	IncludeArchived bool    `form:"include_archived"`
}
