package domain

import "github.com/google/uuid"

// ItemType represents the type of item tracked in CategoryState
type ItemType string

const (
	ItemTypeConstraint ItemType = "CONSTRAINT" // Budget constraint - creates Budget entity when applied
	ItemTypeGoal       ItemType = "GOAL"       // Goal contribution - tracking only
	ItemTypeDebt       ItemType = "DEBT"       // Debt payment - tracking only
)

// CategoryState represents the budget state for a single category/goal/debt within a month
// This is a Value Object that lives inside the Month's JSONB state
type CategoryState struct {
	CategoryID uuid.UUID `json:"category_id"` // For constraints: CategoryID, for goals/debts: GoalID/DebtID
	Type       ItemType  `json:"type"`        // CONSTRAINT, GOAL, or DEBT
	Name       string    `json:"name"`        // Category/Goal/Debt name (denormalized for display)

	// Budget Equation: Available = Rollover + Assigned + Activity
	Rollover  float64 `json:"rollover"`  // Carried over from previous month
	Assigned  float64 `json:"assigned"`  // Money budgeted this month
	Activity  float64 `json:"activity"`  // Money spent/received this month (usually negative for expenses)
	Available float64 `json:"available"` // Calculated field

	// Configuration (deep copied from previous month)
	GoalTarget     *float64 `json:"goal_target,omitempty"`      // Savings goal target
	DebtMinPayment *float64 `json:"debt_min_payment,omitempty"` // Minimum debt payment
	Notes          *string  `json:"notes,omitempty"`            // User notes
}

// NewCategoryState creates a new CategoryState with calculated Available
func NewCategoryState(categoryID uuid.UUID, rollover, assigned, activity float64) *CategoryState {
	cs := &CategoryState{
		CategoryID: categoryID,
		Rollover:   rollover,
		Assigned:   assigned,
		Activity:   activity,
	}
	cs.RecalculateAvailable()
	return cs
}

// RecalculateAvailable recalculates the Available field based on the YNAB equation
// Available = Rollover + Assigned + Activity
func (cs *CategoryState) RecalculateAvailable() {
	cs.Available = cs.Rollover + cs.Assigned + cs.Activity
}

// IsOverspent returns true if Available is negative
func (cs *CategoryState) IsOverspent() bool {
	return cs.Available < 0
}

// AddAssignment adds to the Assigned amount and recalculates Available
func (cs *CategoryState) AddAssignment(amount float64) {
	cs.Assigned += amount
	cs.RecalculateAvailable()
}

// AddActivity adds to the Activity amount (usually negative for expenses) and recalculates Available
func (cs *CategoryState) AddActivity(amount float64) {
	cs.Activity += amount
	cs.RecalculateAvailable()
}

// DeepCopyConfig creates a new CategoryState with the same configuration but reset amounts
// This is used when creating a new month from a previous month
func (cs *CategoryState) DeepCopyConfig() *CategoryState {
	newState := &CategoryState{
		CategoryID: cs.CategoryID,
		Rollover:   0, // Will be set from previous month's Available
		Assigned:   0, // Reset for new month
		Activity:   0, // Reset for new month
		Available:  0, // Will be recalculated
	}

	// Deep copy configuration
	if cs.GoalTarget != nil {
		val := *cs.GoalTarget
		newState.GoalTarget = &val
	}
	if cs.DebtMinPayment != nil {
		val := *cs.DebtMinPayment
		newState.DebtMinPayment = &val
	}
	if cs.Notes != nil {
		val := *cs.Notes
		newState.Notes = &val
	}

	return newState
}

// SetRollover sets the rollover amount (used when creating new month from previous)
func (cs *CategoryState) SetRollover(amount float64) {
	cs.Rollover = amount
	cs.RecalculateAvailable()
}
