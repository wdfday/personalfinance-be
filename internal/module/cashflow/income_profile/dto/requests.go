package dto

import "time"

// CreateIncomeProfileRequest represents request to create a new income profile
type CreateIncomeProfileRequest struct {
	Source      string     `json:"source" binding:"required,max=100"`
	Amount      float64    `json:"amount" binding:"required,gte=0"`
	Currency    string     `json:"currency,omitempty" binding:"omitempty,len=3"`
	Frequency   string     `json:"frequency" binding:"required,oneof=monthly weekly bi-weekly quarterly yearly one-time"`
	StartDate   time.Time  `json:"start_date" binding:"required"`
	EndDate     *time.Time `json:"end_date,omitempty" binding:"omitempty"`
	IsRecurring *bool      `json:"is_recurring,omitempty"`

	// Income components breakdown (optional)
	BaseSalary  *float64 `json:"base_salary,omitempty" binding:"omitempty,gte=0"`
	Bonus       *float64 `json:"bonus,omitempty" binding:"omitempty,gte=0"`
	Commission  *float64 `json:"commission,omitempty" binding:"omitempty,gte=0"`
	Allowance   *float64 `json:"allowance,omitempty" binding:"omitempty,gte=0"`
	OtherIncome *float64 `json:"other_income,omitempty" binding:"omitempty,gte=0"`

	Description *string `json:"description,omitempty" binding:"omitempty,max=1000"`
}

// UpdateIncomeProfileRequest represents request to update an income profile
// NOTE: This creates a NEW version and archives the old one
type UpdateIncomeProfileRequest struct {
	Source      *string    `json:"source,omitempty" binding:"omitempty,max=100"`
	Amount      *float64   `json:"amount,omitempty" binding:"omitempty,gte=0"`
	Currency    *string    `json:"currency,omitempty" binding:"omitempty,len=3"`
	Frequency   *string    `json:"frequency,omitempty" binding:"omitempty,oneof=monthly weekly bi-weekly quarterly yearly one-time"`
	EndDate     *time.Time `json:"end_date,omitempty"`
	IsRecurring *bool      `json:"is_recurring,omitempty"`

	// Income components breakdown (optional)
	BaseSalary  *float64 `json:"base_salary,omitempty" binding:"omitempty,gte=0"`
	Bonus       *float64 `json:"bonus,omitempty" binding:"omitempty,gte=0"`
	Commission  *float64 `json:"commission,omitempty" binding:"omitempty,gte=0"`
	Allowance   *float64 `json:"allowance,omitempty" binding:"omitempty,gte=0"`
	OtherIncome *float64 `json:"other_income,omitempty" binding:"omitempty,gte=0"`

	Description *string `json:"description,omitempty" binding:"omitempty,max=1000"`
}

// VerifyIncomeRequest marks income as verified by user
type VerifyIncomeRequest struct {
	Verified bool `json:"verified" binding:"required"`
}

// UpdateDSSMetadataRequest updates DSS analysis metadata
type UpdateDSSMetadataRequest struct {
	StabilityScore         *float64 `json:"stability_score,omitempty" binding:"omitempty,gte=0,lte=1"`
	RiskLevel              *string  `json:"risk_level,omitempty" binding:"omitempty,oneof=low medium high"`
	Confidence             *float64 `json:"confidence,omitempty" binding:"omitempty,gte=0,lte=1"`
	Variance               *float64 `json:"variance,omitempty" binding:"omitempty,gte=0,lte=1"`
	Trend                  *string  `json:"trend,omitempty" binding:"omitempty,oneof=increasing stable decreasing"`
	RecommendedSavingsRate *float64 `json:"recommended_savings_rate,omitempty" binding:"omitempty,gte=0,lte=1"`
}

// ListIncomeProfilesQuery represents query parameters for listing income profiles
type ListIncomeProfilesQuery struct {
	Status          *string `form:"status" binding:"omitempty,oneof=active pending ended archived paused"`
	IsRecurring     *bool   `form:"is_recurring"`
	IsVerified      *bool   `form:"is_verified"`
	Source          *string `form:"source"`
	IncludeArchived bool    `form:"include_archived"`
}
