package domain

import (
	"time"

	"github.com/google/uuid"
)

// DSSWorkflowResults contains results from all DSS steps
// Current workflow (tradeoff step removed):
// Step 0 (Auto-Scoring, preview only)
// Step 1 (Goal Prioritization)
// Step 2 (Debt Strategy)
// Step 3 (Budget Allocation)
type DSSWorkflowResults struct {
	// Workflow metadata
	CurrentStep    int       `json:"current_step"`    // 0-3 (0=not started, 1-3=steps completed)
	CompletedSteps []int     `json:"completed_steps"` // Which steps have been applied
	StartedAt      time.Time `json:"started_at"`
	LastUpdated    time.Time `json:"last_updated"`

	// Step 0: Auto-Scoring Preview (preview-only, not saved here - just for info)
	// No field needed as it's purely informational

	// Step 1: Goal Prioritization (reuses GoalPriorityResult from dss.go)
	GoalPrioritization *GoalPriorityResult `json:"goal_prioritization,omitempty"`
	AppliedAtStep1     *time.Time          `json:"applied_at_step1,omitempty"`

	// Step 2: Debt Strategy (reuses DebtStrategyResult from dss.go)
	DebtStrategy   *DebtStrategyResult `json:"debt_strategy,omitempty"`
	AppliedAtStep2 *time.Time          `json:"applied_at_step2,omitempty"`

	// Step 3: Budget Allocation (reuses BudgetAllocationResult from dss.go)
	BudgetAllocation *BudgetAllocationResult `json:"budget_allocation,omitempty"`
	AppliedAtStep3   *time.Time              `json:"applied_at_step3,omitempty"`
}

// IsComplete returns true if all 3 apply steps have been applied (steps 1-3)
// Step 0 (Auto-Scoring) is preview-only, so it's optional
func (w *DSSWorkflowResults) IsComplete() bool {
	return w.GoalPrioritization != nil &&
		w.DebtStrategy != nil &&
		w.BudgetAllocation != nil
}

// CanProceedToStep returns true if the given step can be started
func (w *DSSWorkflowResults) CanProceedToStep(step int) bool {
	// Step 1 can always be started
	if step == 1 {
		return true
	}
	// Other steps require previous step to be completed
	return w.CurrentStep >= (step - 1)
}

// MarkStepCompleted marks a step as completed
func (w *DSSWorkflowResults) MarkStepCompleted(step int) {
	now := time.Now()
	w.CurrentStep = step
	w.LastUpdated = now

	// Add to completed steps if not already there
	for _, completedStep := range w.CompletedSteps {
		if completedStep == step {
			return
		}
	}
	w.CompletedSteps = append(w.CompletedSteps, step)
}

// Reset clears all workflow results
func (w *DSSWorkflowResults) Reset() {
	w.CurrentStep = 0
	w.CompletedSteps = []int{}
	w.GoalPrioritization = nil
	w.AppliedAtStep1 = nil
	w.DebtStrategy = nil
	w.AppliedAtStep2 = nil
	w.BudgetAllocation = nil
	w.AppliedAtStep3 = nil
	w.LastUpdated = time.Now()
}

// ==================== Step 1: Goal Prioritization Result ====================

// ==================== Additional helper types for workflow ====================

// CategoryAllocationItem represents allocation to one category (for Step 4: Budget Allocation)
type CategoryAllocationItem struct {
	CategoryID   uuid.UUID `json:"category_id"`
	CategoryName string    `json:"category_name"`
	Amount       float64   `json:"amount"`
	Priority     int       `json:"priority"`
	Source       string    `json:"source"` // "mandatory" | "flexible" | "goal" | "debt"
}

// GoalFundingItem represents funding for one goal (for Step 4: Budget Allocation)
type GoalFundingItem struct {
	GoalID     uuid.UUID `json:"goal_id"`
	GoalName   string    `json:"goal_name"`
	Amount     float64   `json:"amount"`
	Percentage float64   `json:"percentage"` // % of monthly target
}

// DebtPaymentItem represents payment for one debt (for Step 4: Budget Allocation)
type DebtPaymentItem struct {
	DebtID    uuid.UUID `json:"debt_id"`
	DebtName  string    `json:"debt_name"`
	Amount    float64   `json:"amount"`
	IsMinimum bool      `json:"is_minimum"`
}

// ==================== Reused Result Types (originally from dss.go) ====================

// BudgetAllocationResult contains recommendations from budget allocation DSS
type BudgetAllocationResult struct {
	Recommendations map[uuid.UUID]float64 `json:"recommendations"`         // CategoryID -> Amount
	GoalFundings    []GoalFundingItem     `json:"goal_fundings,omitempty"` // Goal allocations from FE
	DebtPayments    []DebtPaymentItem     `json:"debt_payments,omitempty"` // Debt allocations from FE
	OptimalityScore float64               `json:"optimality_score"`        // 0-100
	Method          string                `json:"method"`                  // "linear_programming", "heuristic", "finalize"
	Constraints     []string              `json:"constraints,omitempty"`
}

// GoalPriorityResult contains goal prioritization recommendations
type GoalPriorityResult struct {
	Rankings   []GoalRanking `json:"rankings"`
	TotalScore float64       `json:"total_score"`
	Method     string        `json:"method"` // "meta_gp", "ahp", "weighted"
}

// GoalRanking represents a single goal's priority ranking
type GoalRanking struct {
	GoalID          uuid.UUID `json:"goal_id"`
	GoalName        string    `json:"goal_name"`
	Rank            int       `json:"rank"`
	Score           float64   `json:"score"`
	SuggestedAmount float64   `json:"suggested_amount"`
	Rationale       string    `json:"rationale,omitempty"`
}

// DebtStrategyResult contains debt payoff strategy recommendations
type DebtStrategyResult struct {
	Strategy           string            `json:"strategy"` // "avalanche", "snowball", "hybrid"
	PaymentPlan        []DebtPaymentPlan `json:"payment_plan"`
	TotalInterestSaved float64           `json:"total_interest_saved"`
	PayoffMonths       int               `json:"payoff_months"`
	MonthlyPayment     float64           `json:"monthly_payment"`
}

// DebtPaymentPlan represents payment strategy for a single debt
type DebtPaymentPlan struct {
	DebtID       uuid.UUID `json:"debt_id"`
	DebtName     string    `json:"debt_name"`
	Priority     int       `json:"priority"`
	MinPayment   float64   `json:"min_payment"`
	ExtraPayment float64   `json:"extra_payment"`
	TotalPayment float64   `json:"total_payment"`
}
