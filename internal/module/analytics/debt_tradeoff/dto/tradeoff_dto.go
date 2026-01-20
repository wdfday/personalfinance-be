package dto

import (
	"personalfinancedss/internal/module/analytics/debt_tradeoff/domain"
	"time"
)

// TradeoffInput input cho model
type TradeoffInput struct {
	UserID            string                     `json:"user_id" binding:"required"`
	MonthlyIncome     float64                    `json:"monthly_income" binding:"required,gt=0"`
	EssentialExpenses float64                    `json:"essential_expenses" binding:"required,gte=0"`
	Debts             []domain.DebtInfo          `json:"debts" binding:"required,min=1"`
	TotalMinPayments  float64                    `json:"total_min_payments"`
	Goals             []domain.GoalInfo          `json:"goals"`
	InvestmentProfile domain.InvestmentProfile   `json:"investment_profile"`
	EmergencyFund     domain.EmergencyFundStatus `json:"emergency_fund"`
	Preferences       TradeoffPreferences        `json:"preferences"`
	SimulationConfig  *domain.SimulationConfig   `json:"simulation_config,omitempty"` // Optional, uses defaults if nil
}

// TradeoffPreferences user preferences for decision making
type TradeoffPreferences struct {
	PsychologicalWeight  float64 `json:"psychological_weight"`   // 0-1, weight for psychological factors
	Priority             string  `json:"priority"`               // "debt_free", "wealth_building", "balanced"
	AcceptInvestmentRisk bool    `json:"accept_investment_risk"` // Willing to invest while in debt
	RiskTolerance        string  `json:"risk_tolerance"`         // "conservative", "moderate", "aggressive"
}

// TradeoffOutput kết quả từ model
type TradeoffOutput struct {
	RecommendedStrategy domain.Strategy          `json:"recommended_strategy"`
	RecommendedRatio    domain.AllocationRatio   `json:"recommended_ratio"`
	StrategyAnalysis    []domain.StrategyResult  `json:"strategy_analysis"`
	Reasoning           string                   `json:"reasoning"`
	KeyFactors          []string                 `json:"key_factors"`
	ProjectedTimelines  ProjectionResult         `json:"projected_timelines"`
	MonteCarloResults   *domain.MonteCarloResult `json:"monte_carlo_results,omitempty"`
	Recommendations     []string                 `json:"recommendations"`
}

// ProjectionResult timeline projections
type ProjectionResult struct {
	DebtFreeDate      time.Time              `json:"debt_free_date"`
	EmergencyFundDate time.Time              `json:"emergency_fund_date"`
	GoalDates         map[string]time.Time   `json:"goal_dates"` // GoalID -> projected completion date
	NetWorthGrowth    []domain.NetWorthPoint `json:"net_worth_growth"`
}

// CalculateExtraMoney returns available money after essentials and min payments
func (i *TradeoffInput) CalculateExtraMoney() float64 {
	return i.MonthlyIncome - i.EssentialExpenses - i.TotalMinPayments
}

// GetSimulationConfig returns config or defaults
func (i *TradeoffInput) GetSimulationConfig() domain.SimulationConfig {
	if i.SimulationConfig != nil {
		return *i.SimulationConfig
	}
	return domain.DefaultSimulationConfig()
}

// GetTotalDebt returns sum of all debt balances
func (i *TradeoffInput) GetTotalDebt() float64 {
	total := 0.0
	for _, debt := range i.Debts {
		total += debt.Balance
	}
	return total
}

// GetWeightedAvgInterestRate returns weighted average interest rate
func (i *TradeoffInput) GetWeightedAvgInterestRate() float64 {
	totalDebt := i.GetTotalDebt()
	if totalDebt == 0 {
		return 0
	}

	weightedSum := 0.0
	for _, debt := range i.Debts {
		weightedSum += debt.Balance * debt.InterestRate
	}
	return weightedSum / totalDebt
}

// GetHighestInterestRate returns the highest interest rate among debts
func (i *TradeoffInput) GetHighestInterestRate() float64 {
	highest := 0.0
	for _, debt := range i.Debts {
		if debt.InterestRate > highest {
			highest = debt.InterestRate
		}
	}
	return highest
}

// GetEmergencyFundGap returns how much more is needed for emergency fund
func (i *TradeoffInput) GetEmergencyFundGap() float64 {
	gap := i.EmergencyFund.TargetAmount - i.EmergencyFund.CurrentAmount
	if gap < 0 {
		return 0
	}
	return gap
}

// GetEmergencyFundProgress returns progress as percentage (0-1)
func (i *TradeoffInput) GetEmergencyFundProgress() float64 {
	if i.EmergencyFund.TargetAmount == 0 {
		return 1.0
	}
	progress := i.EmergencyFund.CurrentAmount / i.EmergencyFund.TargetAmount
	if progress > 1 {
		return 1.0
	}
	return progress
}
