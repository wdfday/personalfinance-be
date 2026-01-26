package domain

import "time"

// Strategy type cho debt payoff
type Strategy string

const (
	StrategyAvalanche Strategy = "avalanche" // Highest interest first - mathematically optimal
	StrategySnowball  Strategy = "snowball"  // Smallest balance first - psychologically optimal
	StrategyCashFlow  Strategy = "cash_flow" // Highest min_payment/balance ratio - free up budget
	StrategyStress    Strategy = "stress"    // Highest stress score first - mental peace
	StrategyHybrid    Strategy = "hybrid"    // Weighted combination of factors
)

// DebtInfo thông tin chi tiết về một khoản nợ
type DebtInfo struct {
	ID             string  `json:"id"`
	Name           string  `json:"name"`
	Type           string  `json:"type"` // "credit_card", "student_loan", "car_loan", "personal", "mortgage"
	Balance        float64 `json:"balance"`
	InterestRate   float64 `json:"interest_rate"`      // Annual rate (decimal, e.g., 0.18 = 18%)
	MinimumPayment float64 `json:"minimum_payment"`    // Monthly minimum
	IsVariableRate bool    `json:"is_variable_rate"`   // For sensitivity analysis
	Behavior       string  `json:"behavior,omitempty"` // "revolving", "installment", "interest_only" - payment behavior type

	// Psychological factors
	IsEmbarrassing bool `json:"is_embarrassing"` // Nợ gia đình, bạn bè
	StressScore    int  `json:"stress_score"`    // 1-10 scale

	// For credit score optimization
	IsCreditCard       bool    `json:"is_credit_card"`
	CreditLimit        float64 `json:"credit_limit"` // For utilization calculation
	AffectsCreditScore bool    `json:"affects_credit_score"`
}

// HybridWeights weights cho hybrid strategy
type HybridWeights struct {
	InterestRateWeight float64 `json:"interest_rate_weight"` // α - default 0.4
	BalanceWeight      float64 `json:"balance_weight"`       // β - default 0.3
	StressWeight       float64 `json:"stress_weight"`        // γ - default 0.2
	CashFlowWeight     float64 `json:"cash_flow_weight"`     // δ - default 0.1
}

// DefaultHybridWeights returns default weights
func DefaultHybridWeights() HybridWeights {
	return HybridWeights{
		InterestRateWeight: 0.4,
		BalanceWeight:      0.3,
		StressWeight:       0.2,
		CashFlowWeight:     0.1,
	}
}

// PaymentPlan kế hoạch thanh toán cho một khoản nợ
type PaymentPlan struct {
	DebtID         string            `json:"debt_id"`
	DebtName       string            `json:"debt_name"`
	MonthlyPayment float64           `json:"monthly_payment"`
	ExtraPayment   float64           `json:"extra_payment"`
	PayoffMonth    int               `json:"payoff_month"`
	TotalInterest  float64           `json:"total_interest"`
	Timeline       []MonthlySnapshot `json:"timeline,omitempty"`
}

// MonthlySnapshot snapshot trạng thái nợ mỗi tháng
type MonthlySnapshot struct {
	Month        int     `json:"month"`
	StartBalance float64 `json:"start_balance"`
	Interest     float64 `json:"interest"`
	Payment      float64 `json:"payment"`
	EndBalance   float64 `json:"end_balance"`
}

// StrategyComparison so sánh giữa các strategies
type StrategyComparison struct {
	Strategy          Strategy      `json:"strategy"`
	TotalInterest     float64       `json:"total_interest"`
	Months            int           `json:"months"`
	InterestSaved     float64       `json:"interest_saved"`
	FirstDebtCleared  int           `json:"first_debt_cleared"` // Month when first debt is paid off
	Description       string        `json:"description"`
	Pros              []string      `json:"pros"`
	Cons              []string      `json:"cons"`
	PaymentPlans      []PaymentPlan `json:"payment_plans,omitempty"` // Payment plans for this strategy
	MonthlyAllocation float64       `json:"monthly_allocation"`      // Total monthly payment allocation for this strategy
}

// MonthlyAggregate aggregated monthly schedule
type MonthlyAggregate struct {
	Month            int     `json:"month"`
	TotalPayment     float64 `json:"total_payment"`
	TotalTowardsPrin float64 `json:"total_towards_principal"`
	TotalInterest    float64 `json:"total_interest"`
	RemainingDebts   int     `json:"remaining_debts"`
	TotalBalance     float64 `json:"total_balance"`
	DebtsCleared     int     `json:"debts_cleared"` // Cumulative
}

// Milestone các mốc quan trọng trong quá trình trả nợ
type Milestone struct {
	Month       int       `json:"month"`
	Date        time.Time `json:"date"`
	Description string    `json:"description"`
	Type        string    `json:"type"` // "debt_cleared", "halfway", "debt_free", "quick_win"
	DebtID      string    `json:"debt_id,omitempty"`
	Celebration string    `json:"celebration,omitempty"` // Gamification message
}

// SimulationResult internal structure cho simulation results
type SimulationResult struct {
	Strategy      Strategy
	Months        int
	TotalInterest float64
	DebtTimelines map[string]*DebtTimeline
	FirstCleared  int // Month when first debt cleared
}

// DebtTimeline timeline cho một khoản nợ cụ thể
type DebtTimeline struct {
	DebtID         string
	DebtName       string
	Snapshots      []MonthlySnapshot
	PayoffMonth    int
	TotalInterest  float64
	TotalPrincipal float64
}

// ========== What-If Scenarios ==========

// WhatIfScenario represents a what-if analysis request
type WhatIfScenario struct {
	Type         string  `json:"type"` // "extra_monthly", "lump_sum", "income_change"
	Amount       float64 `json:"amount"`
	TargetDebtID string  `json:"target_debt_id,omitempty"` // For lump sum - which debt to apply
	Description  string  `json:"description"`
}

// WhatIfResult result of what-if analysis
type WhatIfResult struct {
	Scenario          WhatIfScenario `json:"scenario"`
	OriginalMonths    int            `json:"original_months"`
	NewMonths         int            `json:"new_months"`
	MonthsSaved       int            `json:"months_saved"`
	OriginalInterest  float64        `json:"original_interest"`
	NewInterest       float64        `json:"new_interest"`
	InterestSaved     float64        `json:"interest_saved"`
	RecommendedAction string         `json:"recommended_action"`

	// For lump sum - which debt to pay
	BestDebtForLumpSum string  `json:"best_debt_for_lump_sum,omitempty"`
	LumpSumImpact      float64 `json:"lump_sum_impact,omitempty"`
}

// ========== Refinancing Analysis ==========

// RefinanceOption represents a refinancing option
type RefinanceOption struct {
	NewRate        float64  `json:"new_rate"`         // New interest rate
	NewTerm        int      `json:"new_term"`         // New term in months
	OriginationFee float64  `json:"origination_fee"`  // Upfront fee
	MonthlyFee     float64  `json:"monthly_fee"`      // Monthly service fee
	IncludeDebtIDs []string `json:"include_debt_ids"` // Which debts to consolidate
}

// RefinanceAnalysis result of refinancing analysis
type RefinanceAnalysis struct {
	ShouldRefinance     bool    `json:"should_refinance"`
	CurrentWeightedRate float64 `json:"current_weighted_rate"`
	NewEffectiveRate    float64 `json:"new_effective_rate"`

	CurrentTotalInterest   float64 `json:"current_total_interest"`
	RefinanceTotalInterest float64 `json:"refinance_total_interest"`
	TotalFees              float64 `json:"total_fees"`
	NetSavings             float64 `json:"net_savings"`

	BreakEvenMonths       int `json:"break_even_months"`
	CurrentMonthsToPayoff int `json:"current_months_to_payoff"`
	NewMonthsToPayoff     int `json:"new_months_to_payoff"`

	Recommendation string   `json:"recommendation"`
	Warnings       []string `json:"warnings"`
}

// ========== Sensitivity Analysis ==========

// SensitivityScenario represents a sensitivity test
type SensitivityScenario struct {
	Type        string  `json:"type"`       // "income_decrease", "rate_increase", "expense_increase"
	Percentage  float64 `json:"percentage"` // e.g., 0.10 = 10%
	Description string  `json:"description"`
}

// SensitivityResult result of sensitivity analysis
type SensitivityResult struct {
	Scenario           SensitivityScenario `json:"scenario"`
	BaselineMonths     int                 `json:"baseline_months"`
	AdjustedMonths     int                 `json:"adjusted_months"`
	MonthsImpact       int                 `json:"months_impact"`
	BaselineInterest   float64             `json:"baseline_interest"`
	AdjustedInterest   float64             `json:"adjusted_interest"`
	InterestImpact     float64             `json:"interest_impact"`
	StrategyStillValid bool                `json:"strategy_still_valid"`
	NewRecommendation  Strategy            `json:"new_recommendation,omitempty"`
	RiskLevel          string              `json:"risk_level"` // "low", "medium", "high"
	Advice             string              `json:"advice"`
}

// ========== Psychological Scoring ==========

// PsychologicalScore gamification and motivation tracking
type PsychologicalScore struct {
	QuickWinsCount     int      `json:"quick_wins_count"` // Debts cleared in first 6 months
	FirstWinMonth      int      `json:"first_win_month"`  // When first debt cleared
	MotivationScore    float64  `json:"motivation_score"` // 0-100
	MomentumRating     string   `json:"momentum_rating"`  // "slow_start", "steady", "fast_start"
	Celebrations       []string `json:"celebrations"`     // Gamification messages
	ProgressMilestones []string `json:"progress_milestones"`
}
