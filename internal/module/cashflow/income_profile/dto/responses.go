package dto

import "time"

// IncomeProfileResponse represents an income profile in API responses
type IncomeProfileResponse struct {
	ID     string `json:"id"`
	UserID string `json:"user_id"`
	Year   int    `json:"year"`
	Month  int    `json:"month"`

	// Income components
	BaseSalary      float64 `json:"base_salary"`
	Bonus           float64 `json:"bonus"`
	FreelanceIncome float64 `json:"freelance_income"`
	OtherIncome     float64 `json:"other_income"`

	// Computed fields
	TotalIncome     float64            `json:"total_income"`
	IncomeBreakdown map[string]float64 `json:"income_breakdown,omitempty"`

	// Status
	IsActual bool   `json:"is_actual"`
	Notes    string `json:"notes,omitempty"`

	// Timestamps
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// IncomeProfileListResponse represents a list of income profiles
type IncomeProfileListResponse struct {
	IncomeProfiles []IncomeProfileResponse `json:"income_profiles"`
	Count          int                     `json:"count"`
}

// MessageResponse represents a simple message response
type MessageResponse struct {
	Message string `json:"message"`
}
