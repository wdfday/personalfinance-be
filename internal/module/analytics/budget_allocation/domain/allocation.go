package domain

import (
	"github.com/google/uuid"
)

// ScenarioType represents different allocation strategies
type ScenarioType string

const (
	ScenarioConservative ScenarioType = "conservative"
	ScenarioBalanced     ScenarioType = "balanced"
	ScenarioAggressive   ScenarioType = "aggressive"
)

// WarningSeverity represents the severity level of a warning
type WarningSeverity string

const (
	SeverityCritical WarningSeverity = "critical"
	SeverityWarning  WarningSeverity = "warning"
	SeverityInfo     WarningSeverity = "info"
)

// AllocationScenario represents one complete allocation strategy
type AllocationScenario struct {
	ScenarioType        ScenarioType         `json:"scenario_type"`
	CategoryAllocations []CategoryAllocation `json:"category_allocations"`
	GoalAllocations     []GoalAllocation     `json:"goal_allocations"`
	DebtAllocations     []DebtAllocation     `json:"debt_allocations"`
	Summary             AllocationSummary    `json:"summary"`
	FeasibilityScore    float64              `json:"feasibility_score"` // 0-100
	Warnings            []AllocationWarning  `json:"warnings"`
}

// CategoryAllocation represents amount allocated to an expense category
type CategoryAllocation struct {
	CategoryID   uuid.UUID `json:"category_id"`
	CategoryName string    `json:"category_name"`
	Amount       float64   `json:"amount"`
	Minimum      float64   `json:"minimum"`
	Maximum      float64   `json:"maximum"`
	IsFlexible   bool      `json:"is_flexible"`
	Priority     int       `json:"priority"`
}

// GoalAllocation represents amount allocated to a financial goal
type GoalAllocation struct {
	GoalID                uuid.UUID `json:"goal_id"`
	GoalName              string    `json:"goal_name"`
	Amount                float64   `json:"amount"`
	SuggestedContribution float64   `json:"suggested_contribution"`
	Priority              string    `json:"priority"` // low, medium, high, critical
	PercentageOfTarget    float64   `json:"percentage_of_target"`
}

// DebtAllocation represents amount allocated to debt payment
type DebtAllocation struct {
	DebtID          uuid.UUID `json:"debt_id"`
	DebtName        string    `json:"debt_name"`
	Amount          float64   `json:"amount"`
	MinimumPayment  float64   `json:"minimum_payment"`
	ExtraPayment    float64   `json:"extra_payment"` // Amount above minimum
	InterestRate    float64   `json:"interest_rate"`
	InterestSavings float64   `json:"interest_savings,omitempty"` // Estimated savings from extra payment
}

// AllocationSummary provides high-level summary of the allocation
type AllocationSummary struct {
	TotalIncome            float64 `json:"total_income"`
	TotalAllocated         float64 `json:"total_allocated"`
	Surplus                float64 `json:"surplus"`
	MandatoryExpenses      float64 `json:"mandatory_expenses"`
	FlexibleExpenses       float64 `json:"flexible_expenses"`
	TotalDebtPayments      float64 `json:"total_debt_payments"`
	TotalGoalContributions float64 `json:"total_goal_contributions"`
	SavingsRate            float64 `json:"savings_rate"` // Percentage: (goals + extra debt) / income
}

// AllocationWarning represents an issue or recommendation
type AllocationWarning struct {
	Severity    WarningSeverity `json:"severity"`
	Message     string          `json:"message"`
	Category    string          `json:"category"` // "income", "expense", "debt", "goal"
	Suggestions []string        `json:"suggestions,omitempty"`
}

// NewAllocationScenario creates a new allocation scenario
func NewAllocationScenario(scenarioType ScenarioType) *AllocationScenario {
	return &AllocationScenario{
		ScenarioType:        scenarioType,
		CategoryAllocations: make([]CategoryAllocation, 0),
		GoalAllocations:     make([]GoalAllocation, 0),
		DebtAllocations:     make([]DebtAllocation, 0),
		Warnings:            make([]AllocationWarning, 0),
		FeasibilityScore:    100.0,
	}
}

// CalculateSummary calculates the allocation summary from allocated amounts
func (s *AllocationScenario) CalculateSummary(totalIncome float64) {
	var mandatoryExpenses float64
	var flexibleExpenses float64
	var totalDebtPayments float64
	var totalGoalContributions float64

	// Sum category allocations
	for _, cat := range s.CategoryAllocations {
		if cat.IsFlexible {
			flexibleExpenses += cat.Amount
		} else {
			mandatoryExpenses += cat.Amount
		}
	}

	// Sum debt allocations
	for _, debt := range s.DebtAllocations {
		totalDebtPayments += debt.Amount
	}

	// Sum goal allocations
	for _, goal := range s.GoalAllocations {
		totalGoalContributions += goal.Amount
	}

	totalAllocated := mandatoryExpenses + flexibleExpenses + totalDebtPayments + totalGoalContributions
	surplus := totalIncome - totalAllocated

	// Calculate savings rate
	var savingsRate float64
	if totalIncome > 0 {
		// Savings rate = (goal contributions + extra debt payments) / income
		extraDebtPayments := 0.0
		for _, debt := range s.DebtAllocations {
			extraDebtPayments += debt.ExtraPayment
		}
		savingsRate = ((totalGoalContributions + extraDebtPayments) / totalIncome) * 100
	}

	s.Summary = AllocationSummary{
		TotalIncome:            totalIncome,
		TotalAllocated:         totalAllocated,
		Surplus:                surplus,
		MandatoryExpenses:      mandatoryExpenses,
		FlexibleExpenses:       flexibleExpenses,
		TotalDebtPayments:      totalDebtPayments,
		TotalGoalContributions: totalGoalContributions,
		SavingsRate:            savingsRate,
	}
}

// AddWarning adds a warning to the scenario
func (s *AllocationScenario) AddWarning(severity WarningSeverity, category, message string, suggestions ...string) {
	warning := AllocationWarning{
		Severity:    severity,
		Category:    category,
		Message:     message,
		Suggestions: suggestions,
	}
	s.Warnings = append(s.Warnings, warning)
}

// IsFeasible returns true if the scenario is feasible (score >= 50)
func (s *AllocationScenario) IsFeasible() bool {
	return s.FeasibilityScore >= 50.0
}
