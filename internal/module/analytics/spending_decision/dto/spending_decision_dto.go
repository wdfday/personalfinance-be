package dto

import (
	"personalfinancedss/internal/module/analytics/spending_decision/domain"
)

// LargePurchaseInput input for large purchase analysis
type LargePurchaseInput struct {
	UserID string `json:"user_id" binding:"required"`

	// Purchase details
	ItemName       string  `json:"item_name" binding:"required"`
	ItemType       string  `json:"item_type" binding:"required"`
	PurchaseAmount float64 `json:"purchase_amount" binding:"required,gt=0"`
	IsRecurring    bool    `json:"is_recurring"`

	// Recurring costs
	MonthlyRecurringCost float64 `json:"monthly_recurring_cost"`
	RecurringMonths      int     `json:"recurring_months"`

	// Funding
	PreferredFundingSource string `json:"preferred_funding_source"` // "savings", "financing", "budget_realloc"

	// Financial state
	CurrentState domain.FinancialState `json:"current_state" binding:"required"`

	// Context
	MotivationLevel string `json:"motivation_level"` // "impulse", "planned", "necessary"
	Justification   string `json:"justification"`
}

// LargePurchaseOutput comprehensive analysis output
type LargePurchaseOutput struct {
	Recommendation   DecisionRecommendation     `json:"recommendation"`
	FundingOptions   []domain.FundingOption     `json:"funding_options"`
	TrueCost         domain.TrueCostAnalysis    `json:"true_cost"`
	ReallocationPlan *domain.BudgetReallocation `json:"reallocation_plan,omitempty"`
	LongTermImpact   domain.LongTermImpact      `json:"long_term_impact"`
	Alternatives     []domain.Alternative       `json:"alternatives"`
	BehavioralNudges []string                   `json:"behavioral_nudges"`
}

// DecisionRecommendation recommendation for large purchase
type DecisionRecommendation struct {
	Decision           string   `json:"decision"` // "approve", "reconsider", "reject"
	Confidence         float64  `json:"confidence"`
	Reasoning          []string `json:"reasoning"`
	Warnings           []string `json:"warnings"`
	RecommendedFunding string   `json:"recommended_funding,omitempty"`
}
