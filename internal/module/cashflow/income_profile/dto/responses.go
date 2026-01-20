package dto

import "time"

// IncomeProfileResponse represents an income profile in API responses
type IncomeProfileResponse struct {
	ID         string `json:"id"`
	UserID     string `json:"user_id"`
	CategoryID string `json:"category_id"`

	// Period tracking
	StartDate *time.Time `json:"start_date"`
	EndDate   *time.Time `json:"end_date,omitempty"`
	Duration  int        `json:"duration_days,omitempty"` // in days

	// Income details
	Source    string  `json:"source"`
	Amount    float64 `json:"amount"`
	Currency  string  `json:"currency"`
	Frequency string  `json:"frequency"`

	// Income components breakdown
	BaseSalary  float64 `json:"base_salary"`
	Bonus       float64 `json:"bonus"`
	Commission  float64 `json:"commission"`
	Allowance   float64 `json:"allowance"`
	OtherIncome float64 `json:"other_income"`

	// Computed fields
	TotalIncome     float64            `json:"total_income"`
	IncomeBreakdown map[string]float64 `json:"income_breakdown,omitempty"`

	// Status and lifecycle
	Status      string `json:"status"`
	IsRecurring bool   `json:"is_recurring"`
	IsVerified  bool   `json:"is_verified"`
	IsActive    bool   `json:"is_active"`
	IsArchived  bool   `json:"is_archived"`

	// DSS Analysis
	DSSMetadata *DSSMetadataResponse `json:"dss_metadata,omitempty"`
	DSSScore    float64              `json:"dss_score,omitempty"`

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

// DSSMetadataResponse represents DSS analysis metadata
type DSSMetadataResponse struct {
	StabilityScore         float64 `json:"stability_score"`
	RiskLevel              string  `json:"risk_level"`
	Confidence             float64 `json:"confidence"`
	Variance               float64 `json:"variance"`
	Trend                  string  `json:"trend"`
	RecommendedSavingsRate float64 `json:"recommended_savings_rate"`
	LastAnalyzed           string  `json:"last_analyzed"`
	AnalysisVersion        string  `json:"analysis_version"`
}

// IncomeProfileWithHistoryResponse includes version history
type IncomeProfileWithHistoryResponse struct {
	Current        IncomeProfileResponse   `json:"current"`
	VersionHistory []IncomeProfileResponse `json:"version_history,omitempty"`
}

// IncomeProfileListResponse represents a list of income profiles
type IncomeProfileListResponse struct {
	IncomeProfiles []IncomeProfileResponse `json:"income_profiles"`
	Count          int                     `json:"count"`
	Summary        *IncomeSummaryResponse  `json:"summary,omitempty"`
}

// IncomeSummaryResponse provides summary statistics
type IncomeSummaryResponse struct {
	TotalMonthlyIncome   float64 `json:"total_monthly_income"`
	TotalYearlyIncome    float64 `json:"total_yearly_income"`
	ActiveIncomeCount    int     `json:"active_income_count"`
	RecurringIncomeCount int     `json:"recurring_income_count"`
	AverageStability     float64 `json:"average_stability,omitempty"`
}

// MessageResponse represents a simple message response
type MessageResponse struct {
	Message string `json:"message"`
}
