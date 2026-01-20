package domain

import (
	"time"

	"github.com/google/uuid"
)

// ===== SNAPSHOT TYPES =====
// These capture the state of external entities at a point in time

// IncomeSnapshot captures income profile state at snapshot time
type IncomeSnapshot struct {
	ProfileID   *uuid.UUID `json:"profile_id,omitempty"` // nil for ad-hoc
	Name        string     `json:"name"`
	Amount      float64    `json:"amount"`
	IsRecurring bool       `json:"is_recurring"`
	Frequency   string     `json:"frequency,omitempty"` // monthly, bi-weekly, etc.
	SourceType  string     `json:"source_type"`         // "profile", "adhoc", "estimate"
	IsConfirmed bool       `json:"is_confirmed"`        // Has actual transaction matched?
}

// ConstraintSnapshot captures budget constraint state at snapshot time
type ConstraintSnapshot struct {
	ConstraintID *uuid.UUID `json:"constraint_id,omitempty"` // nil for ad-hoc
	Name         string     `json:"name"`
	CategoryID   uuid.UUID  `json:"category_id"`
	Type         string     `json:"type"` // "minimum", "maximum", "exact"
	Amount       float64    `json:"amount"`
	IsRecurring  bool       `json:"is_recurring"`
	SourceType   string     `json:"source_type"` // "template", "adhoc"
}

// GoalSnapshot captures goal state at snapshot time
type GoalSnapshot struct {
	GoalID        uuid.UUID `json:"goal_id"`
	Name          string    `json:"name"`
	TargetAmount  float64   `json:"target_amount"`
	CurrentAmount float64   `json:"current_amount"`
	Progress      float64   `json:"progress"` // 0-100
	Priority      int       `json:"priority"`
	Deadline      *string   `json:"deadline,omitempty"`
	MonthlyTarget float64   `json:"monthly_target"` // Calculated: remaining / months left
	IsSelected    bool      `json:"is_selected"`    // User selected for this month
}

// DebtSnapshot captures debt state at snapshot time
type DebtSnapshot struct {
	DebtID         uuid.UUID `json:"debt_id"`
	Name           string    `json:"name"`
	CurrentBalance float64   `json:"current_balance"`
	InterestRate   float64   `json:"interest_rate"`
	MinPayment     float64   `json:"min_payment"`
	Progress       float64   `json:"progress"` // % paid off
	IsSelected     bool      `json:"is_selected"`
}

// AdHocExpense is a one-time expense for this month only (not saved to Constraints table)
type AdHocExpense struct {
	ID           uuid.UUID  `json:"id"`
	Name         string     `json:"name"`
	Amount       float64    `json:"amount"`
	CategoryID   *uuid.UUID `json:"category_id,omitempty"`
	CategoryHint string     `json:"category_hint,omitempty"`
	Notes        string     `json:"notes,omitempty"`
}

// AdHocIncome is a one-time expected income for this month (not saved to Income Profiles)
type AdHocIncome struct {
	ID     uuid.UUID `json:"id"`
	Name   string    `json:"name"`
	Amount float64   `json:"amount"`
	Notes  string    `json:"notes,omitempty"`
}

// ===== DSS RESULT =====

// DSSResult stores the output of a DSS solver execution
type DSSResult struct {
	ExecutedAt   time.Time            `json:"executed_at"`
	SolverType   string               `json:"solver_type"` // "budget_allocation", "goal_prioritization", etc.
	InputSummary map[string]float64   `json:"input_summary"`
	Allocations  []CategoryAllocation `json:"allocations"`
	Feasible     bool                 `json:"feasible"`
	Score        float64              `json:"score,omitempty"`
	Warnings     []string             `json:"warnings,omitempty"`
	Explanations []string             `json:"explanations,omitempty"`
}

// CategoryAllocation represents a single allocation decision from DSS
type CategoryAllocation struct {
	CategoryID uuid.UUID `json:"category_id"`
	Name       string    `json:"name"`
	Amount     float64   `json:"amount"`
	Priority   int       `json:"priority"`
	Source     string    `json:"source"` // "constraint", "goal", "debt", "discretionary"
}

// ===== INPUT SNAPSHOT (Complete snapshot of all inputs) =====

// InputSnapshot aggregates all external entity snapshots
type InputSnapshot struct {
	CapturedAt time.Time `json:"captured_at"`

	// From Income Profiles table
	IncomeProfiles []IncomeSnapshot `json:"income_profiles"`

	// From Budget Constraints table
	Constraints []ConstraintSnapshot `json:"constraints"`

	// From Goals table
	Goals []GoalSnapshot `json:"goals"`

	// From Debts table
	Debts []DebtSnapshot `json:"debts"`

	// User-added one-time items (NOT from tables)
	AdHocExpenses []AdHocExpense `json:"adhoc_expenses,omitempty"`
	AdHocIncome   []AdHocIncome  `json:"adhoc_income,omitempty"`

	// Pre-calculated totals for quick access
	Totals InputTotals `json:"totals"`
}

// InputTotals contains pre-calculated sums
type InputTotals struct {
	ProjectedIncome  float64 `json:"projected_income"`  // Sum of all income (recurring + adhoc)
	RecurringIncome  float64 `json:"recurring_income"`  // Sum of recurring only
	TotalConstraints float64 `json:"total_constraints"` // Sum of fixed expenses
	TotalGoalTargets float64 `json:"total_goal_targets"`
	TotalDebtMinimum float64 `json:"total_debt_minimum"`
	TotalAdHoc       float64 `json:"total_adhoc"`
	Disposable       float64 `json:"disposable"` // ProjectedIncome - TotalConstraints
}

// NewInputSnapshot creates an empty input snapshot
func NewInputSnapshot() *InputSnapshot {
	return &InputSnapshot{
		CapturedAt:     time.Now(),
		IncomeProfiles: []IncomeSnapshot{},
		Constraints:    []ConstraintSnapshot{},
		Goals:          []GoalSnapshot{},
		Debts:          []DebtSnapshot{},
		AdHocExpenses:  []AdHocExpense{},
		AdHocIncome:    []AdHocIncome{},
	}
}

// CalculateTotals computes all totals from the snapshot data
func (s *InputSnapshot) CalculateTotals() {
	var projectedIncome, recurringIncome, totalConstraints, totalGoals, totalDebts, totalAdHoc float64

	for _, inc := range s.IncomeProfiles {
		projectedIncome += inc.Amount
		if inc.IsRecurring {
			recurringIncome += inc.Amount
		}
	}

	for _, c := range s.Constraints {
		totalConstraints += c.Amount
	}

	for _, g := range s.Goals {
		if g.IsSelected {
			totalGoals += g.MonthlyTarget
		}
	}

	for _, d := range s.Debts {
		if d.IsSelected {
			totalDebts += d.MinPayment
		}
	}

	for _, e := range s.AdHocExpenses {
		totalAdHoc += e.Amount
	}

	for _, inc := range s.AdHocIncome {
		projectedIncome += inc.Amount
	}

	s.Totals = InputTotals{
		ProjectedIncome:  projectedIncome,
		RecurringIncome:  recurringIncome,
		TotalConstraints: totalConstraints,
		TotalGoalTargets: totalGoals,
		TotalDebtMinimum: totalDebts,
		TotalAdHoc:       totalAdHoc,
		Disposable:       projectedIncome - totalConstraints - totalAdHoc,
	}
}
