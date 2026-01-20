package domain

import "time"

// DebtInfo thông tin về một khoản nợ
type DebtInfo struct {
	ID             string  `json:"id"`
	Name           string  `json:"name"`
	Balance        float64 `json:"balance"`
	InterestRate   float64 `json:"interest_rate"`   // Annual rate (decimal, e.g., 0.18 = 18%)
	MinimumPayment float64 `json:"minimum_payment"` // Monthly minimum
	Type           string  `json:"type"`            // "credit_card", "student_loan", "car_loan", "personal", "mortgage"
}

// GoalInfo thông tin về mục tiêu tài chính
type GoalInfo struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	TargetAmount  float64   `json:"target_amount"`
	CurrentAmount float64   `json:"current_amount"`
	Deadline      time.Time `json:"deadline"`
	Priority      float64   `json:"priority"` // Từ AHP (0-1, optional)
}

// InvestmentProfile profile đầu tư của user
type InvestmentProfile struct {
	RiskTolerance      string  `json:"risk_tolerance"`      // "conservative", "moderate", "aggressive"
	ExpectedReturn     float64 `json:"expected_return"`     // Annual return (decimal, e.g., 0.07 = 7%)
	TimeHorizon        int     `json:"time_horizon"`        // Years
	CurrentInvestments float64 `json:"current_investments"` // Current investment balance
}

// EmergencyFundStatus trạng thái quỹ khẩn cấp
type EmergencyFundStatus struct {
	TargetAmount    float64 `json:"target_amount"`
	CurrentAmount   float64 `json:"current_amount"`
	MonthlyExpenses float64 `json:"monthly_expenses"`
	TargetMonths    int     `json:"target_months"` // Typically 3-6 months
}

// Strategy type for debt vs savings allocation
type Strategy string

const (
	StrategyAggressiveDebt    Strategy = "aggressive_debt"    // 75% debt, 25% savings
	StrategyBalanced          Strategy = "balanced"           // 50% debt, 50% savings
	StrategyAggressiveSavings Strategy = "aggressive_savings" // 25% debt, 75% savings
	StrategyCustom            Strategy = "custom"             // User-defined ratio
)

// AllocationRatio represents debt vs savings split
type AllocationRatio struct {
	DebtPercent    float64 `json:"debt_percent"`    // 0-1
	SavingsPercent float64 `json:"savings_percent"` // 0-1
}

// StrategyResult analysis của một strategy
type StrategyResult struct {
	Strategy          Strategy        `json:"strategy"`
	Ratio             AllocationRatio `json:"ratio"`
	NPV               float64         `json:"npv"`                 // Net Present Value
	TotalInterestPaid float64         `json:"total_interest_paid"` // Total interest over projection period
	InterestSaved     float64         `json:"interest_saved"`      // Compared to minimum payments only
	InvestmentValue   float64         `json:"investment_value"`    // Future value of investments
	MonthsToDebtFree  int             `json:"months_to_debt_free"`
	TimeToGoals       map[string]int  `json:"time_to_goals"` // GoalID -> months to achieve
	RiskScore         float64         `json:"risk_score"`    // 0-10, higher = riskier
	Score             float64         `json:"score"`         // Weighted composite score
	Pros              []string        `json:"pros"`
	Cons              []string        `json:"cons"`
}

// NetWorthPoint timeline point for projection
type NetWorthPoint struct {
	Month     int     `json:"month"`
	NetWorth  float64 `json:"net_worth"`
	DebtTotal float64 `json:"debt_total"`
	Assets    float64 `json:"assets"`
	Savings   float64 `json:"savings"`
}

// MonteCarloResult results from Monte Carlo simulation
type MonteCarloResult struct {
	NumSimulations       int        `json:"num_simulations"`
	SuccessProbability   float64    `json:"success_probability"`    // P(achieving goals within timeframe)
	DebtFreeP50          int        `json:"debt_free_p50"`          // 50th percentile months to debt-free
	DebtFreeP75          int        `json:"debt_free_p75"`          // 75th percentile
	DebtFreeP90          int        `json:"debt_free_p90"`          // 90th percentile (worst case)
	NPVMean              float64    `json:"npv_mean"`               // Mean NPV across simulations
	NPVStdDev            float64    `json:"npv_std_dev"`            // Standard deviation
	NPVP5                float64    `json:"npv_p5"`                 // 5th percentile (worst case)
	NPVP95               float64    `json:"npv_p95"`                // 95th percentile (best case)
	ConfidenceInterval95 [2]float64 `json:"confidence_interval_95"` // 95% CI for NPV
}

// SimulationConfig configuration for Monte Carlo
type SimulationConfig struct {
	NumSimulations   int     `json:"num_simulations"`   // Default 500
	IncomeVariance   float64 `json:"income_variance"`   // Default 0.10 (±10%)
	ExpenseVariance  float64 `json:"expense_variance"`  // Default 0.15 (±15%)
	ReturnVariance   float64 `json:"return_variance"`   // Default 0.20 (±20% of expected return)
	ProjectionMonths int     `json:"projection_months"` // Default 60 (5 years)
	DiscountRate     float64 `json:"discount_rate"`     // For NPV calculation, default 0.05
}

// DefaultSimulationConfig returns default config
func DefaultSimulationConfig() SimulationConfig {
	return SimulationConfig{
		NumSimulations:   500,
		IncomeVariance:   0.10,
		ExpenseVariance:  0.15,
		ReturnVariance:   0.20,
		ProjectionMonths: 60,
		DiscountRate:     0.05,
	}
}
