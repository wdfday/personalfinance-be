package domain

import "time"

// ========== LARGE PURCHASE DOMAIN MODELS ==========

// FinancialState comprehensive financial state for large purchase analysis
type FinancialState struct {
	// Cash & Savings
	CashAvailable       float64 `json:"cash_available"`
	EmergencyFund       float64 `json:"emergency_fund"`
	EmergencyFundTarget float64 `json:"emergency_fund_target"`
	OtherSavings        float64 `json:"other_savings"`

	// Income & Budget
	MonthlyIncome       float64            `json:"monthly_income"`
	MonthlyBudgetAlloc  map[string]float64 `json:"monthly_budget"`
	MonthlySpent        map[string]float64 `json:"monthly_spent"`
	DiscretionaryBudget float64            `json:"discretionary_budget"`

	// Goals & Debt
	ActiveGoals        []GoalInfo `json:"active_goals"`
	CurrentDebtPayment float64    `json:"current_debt_payment"`
	TotalDebtBalance   float64    `json:"total_debt_balance"`

	// Projections
	DebtFreeDate         time.Time `json:"debt_free_date"`
	RetirementProjection float64   `json:"retirement_projection"`
}

// GoalInfo detailed goal information
type GoalInfo struct {
	ID                  string    `json:"id"`
	Name                string    `json:"name"`
	TargetAmount        float64   `json:"target_amount"`
	CurrentAmount       float64   `json:"current_amount"`
	MonthlyContribution float64   `json:"monthly_contribution"`
	Deadline            time.Time `json:"deadline"`
	Priority            float64   `json:"priority"`
}

// GoalImpact shows impact on financial goals
type GoalImpact struct {
	GoalName          string    `json:"goal_name"`
	CurrentProgress   float64   `json:"current_progress"`
	DelayMonths       int       `json:"delay_months"`
	NewCompletionDate time.Time `json:"new_completion_date"`
}

// Alternative represents an alternative option
type Alternative struct {
	Description     string  `json:"description"`
	PotentialSaving float64 `json:"potential_saving"`
	Link            string  `json:"link"`
}

// FundingOption represents a funding source analysis
type FundingOption struct {
	Source          string  `json:"source"` // "savings", "financing", "budget_realloc"
	Feasible        bool    `json:"feasible"`
	Risk            string  `json:"risk"` // "low", "medium", "high"
	AmountNeeded    float64 `json:"amount_needed"`
	AmountAvailable float64 `json:"amount_available"`

	// Impact
	ImpactOnEF    float64      `json:"impact_on_ef"`
	ImpactOnGoals []GoalImpact `json:"impact_on_goals"`

	// For financing
	InterestRate   float64 `json:"interest_rate,omitempty"`
	TotalInterest  float64 `json:"total_interest,omitempty"`
	MonthlyPayment float64 `json:"monthly_payment,omitempty"`

	Pros []string `json:"pros"`
	Cons []string `json:"cons"`
}

// TrueCostAnalysis comprehensive cost breakdown
type TrueCostAnalysis struct {
	PurchasePrice   float64    `json:"purchase_price"`
	OpportunityCost float64    `json:"opportunity_cost"`
	RecurringCosts  float64    `json:"recurring_costs"`
	TotalCost       float64    `json:"total_cost"`
	Breakdown       []CostItem `json:"breakdown"`
}

// CostItem individual cost component
type CostItem struct {
	Category    string  `json:"category"`
	Amount      float64 `json:"amount"`
	Description string  `json:"description"`
}

// BudgetReallocation plan for budget adjustment
type BudgetReallocation struct {
	Strategy         string               `json:"strategy"`
	Duration         int                  `json:"duration"`
	Adjustments      []CategoryAdjustment `json:"adjustments"`
	NewMonthlyBudget map[string]float64   `json:"new_monthly_budget"`
	IsFeasible       bool                 `json:"is_feasible"`
	Difficulty       string               `json:"difficulty"`
}

// CategoryAdjustment budget adjustment for a category
type CategoryAdjustment struct {
	Category      string  `json:"category"`
	CurrentAmount float64 `json:"current_amount"`
	NewAmount     float64 `json:"new_amount"`
	Reduction     float64 `json:"reduction"`
	ReductionPct  float64 `json:"reduction_pct"`
}

// LongTermImpact projection of long-term effects
type LongTermImpact struct {
	RetirementAt65Before  float64   `json:"retirement_at_65_before"`
	RetirementAt65After   float64   `json:"retirement_at_65_after"`
	RetirementLoss        float64   `json:"retirement_loss"`
	DebtFreeDateBefore    time.Time `json:"debt_free_date_before"`
	DebtFreeDateAfter     time.Time `json:"debt_free_date_after"`
	DebtFreeDelayMonths   int       `json:"debt_free_delay_months"`
	NetWorth5YearsBefore  float64   `json:"net_worth_5_years_before"`
	NetWorth5YearsAfter   float64   `json:"net_worth_5_years_after"`
	NetWorth10YearsBefore float64   `json:"net_worth_10_years_before"`
	NetWorth10YearsAfter  float64   `json:"net_worth_10_years_after"`
}
