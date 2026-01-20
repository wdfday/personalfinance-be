package dto

import (
	"personalfinancedss/internal/module/analytics/debt_strategy/domain"
	"time"
)

// DebtStrategyInput input cho model
type DebtStrategyInput struct {
	UserID string            `json:"user_id"`
	Debts  []domain.DebtInfo `json:"debts"`

	// Budget allocation
	TotalDebtBudget float64 `json:"total_debt_budget"` // Monthly budget cho debt

	// User preferences
	PreferredStrategy domain.Strategy       `json:"preferred_strategy,omitempty"`
	MotivationLevel   string                `json:"motivation_level,omitempty"` // "low", "medium", "high"
	HybridWeights     *domain.HybridWeights `json:"hybrid_weights,omitempty"`   // Custom weights for hybrid

	// What-if scenarios (optional)
	WhatIfScenarios []domain.WhatIfScenario `json:"what_if_scenarios,omitempty"`

	// Refinancing analysis (optional)
	RefinanceOption *domain.RefinanceOption `json:"refinance_option,omitempty"`

	// Sensitivity analysis (optional)
	RunSensitivity bool `json:"run_sensitivity,omitempty"`
}

// CalculateExtraPayment tính extra payment sau khi trả minimum
func (i *DebtStrategyInput) CalculateExtraPayment() float64 {
	totalMin := i.GetTotalMinPayments()
	extra := i.TotalDebtBudget - totalMin
	if extra < 0 {
		return 0
	}
	return extra
}

// GetTotalMinPayments tính tổng minimum payments
func (i *DebtStrategyInput) GetTotalMinPayments() float64 {
	total := 0.0
	for _, d := range i.Debts {
		total += d.MinimumPayment
	}
	return total
}

// GetTotalDebt tính tổng nợ
func (i *DebtStrategyInput) GetTotalDebt() float64 {
	total := 0.0
	for _, d := range i.Debts {
		total += d.Balance
	}
	return total
}

// GetHighestInterestRate lấy lãi suất cao nhất
func (i *DebtStrategyInput) GetHighestInterestRate() float64 {
	max := 0.0
	for _, d := range i.Debts {
		if d.InterestRate > max {
			max = d.InterestRate
		}
	}
	return max
}

// GetWeightedAvgRate tính weighted average interest rate
func (i *DebtStrategyInput) GetWeightedAvgRate() float64 {
	totalDebt := i.GetTotalDebt()
	if totalDebt == 0 {
		return 0
	}
	weightedSum := 0.0
	for _, d := range i.Debts {
		weightedSum += d.Balance * d.InterestRate
	}
	return weightedSum / totalDebt
}

// GetSmallestBalance lấy balance nhỏ nhất
func (i *DebtStrategyInput) GetSmallestBalance() float64 {
	if len(i.Debts) == 0 {
		return 0
	}
	min := i.Debts[0].Balance
	for _, d := range i.Debts {
		if d.Balance < min && d.Balance > 0 {
			min = d.Balance
		}
	}
	return min
}

// GetHighestStressDebt lấy debt có stress cao nhất
func (i *DebtStrategyInput) GetHighestStressDebt() *domain.DebtInfo {
	if len(i.Debts) == 0 {
		return nil
	}
	highest := &i.Debts[0]
	for idx := range i.Debts {
		if i.Debts[idx].StressScore > highest.StressScore {
			highest = &i.Debts[idx]
		}
	}
	return highest
}

// HasHighStressDebt checks if any debt has stress >= 7
func (i *DebtStrategyInput) HasHighStressDebt() bool {
	for _, d := range i.Debts {
		if d.StressScore >= 7 {
			return true
		}
	}
	return false
}

// HasVariableRateDebt checks if any debt has variable rate
func (i *DebtStrategyInput) HasVariableRateDebt() bool {
	for _, d := range i.Debts {
		if d.IsVariableRate {
			return true
		}
	}
	return false
}

// GetHybridWeights returns hybrid weights (default if not set)
func (i *DebtStrategyInput) GetHybridWeights() domain.HybridWeights {
	if i.HybridWeights != nil {
		return *i.HybridWeights
	}
	return domain.DefaultHybridWeights()
}

// DebtStrategyOutput kết quả từ model
type DebtStrategyOutput struct {
	// Recommended strategy và payment plans
	RecommendedStrategy domain.Strategy      `json:"recommended_strategy"`
	PaymentPlans        []domain.PaymentPlan `json:"payment_plans"`

	// Summary metrics
	TotalInterest    float64   `json:"total_interest"`
	MonthsToDebtFree int       `json:"months_to_debt_free"`
	DebtFreeDate     time.Time `json:"debt_free_date"`

	// Comparison với all strategies
	StrategyComparison []domain.StrategyComparison `json:"strategy_comparison"`

	// Monthly payment schedule (aggregated)
	MonthlySchedule []domain.MonthlyAggregate `json:"monthly_schedule,omitempty"`

	// Milestones
	Milestones []domain.Milestone `json:"milestones"`

	// Reasoning
	Reasoning string   `json:"reasoning"`
	KeyFacts  []string `json:"key_facts"`

	// What-if results (if requested)
	WhatIfResults []domain.WhatIfResult `json:"what_if_results,omitempty"`

	// Refinancing analysis (if requested)
	RefinanceAnalysis *domain.RefinanceAnalysis `json:"refinance_analysis,omitempty"`

	// Sensitivity analysis (if requested)
	SensitivityResults []domain.SensitivityResult `json:"sensitivity_results,omitempty"`

	// Psychological scoring
	PsychScore *domain.PsychologicalScore `json:"psychological_score,omitempty"`
}
