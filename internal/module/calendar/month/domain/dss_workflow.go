package domain

import (
	"time"

	"github.com/google/uuid"
)

// DSSWorkflowResults contains results from all 5 sequential DSS steps
// This is embedded in MonthState.DSSWorkflow for the new sequential workflow
type DSSWorkflowResults struct {
	// Workflow metadata
	CurrentStep    int       `json:"current_step"`    // 0-4 (0=not started, 1-4=steps completed)
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

	// Step 3: Goal-Debt Trade-off (new type)
	GoalDebtTradeoff *GoalDebtTradeoffResult `json:"goal_debt_tradeoff,omitempty"`
	AppliedAtStep3   *time.Time              `json:"applied_at_step3,omitempty"`

	// Step 4: Budget Allocation (reuses BudgetAllocationResult from dss.go)
	BudgetAllocation *BudgetAllocationResult `json:"budget_allocation,omitempty"`
	AppliedAtStep4   *time.Time              `json:"applied_at_step4,omitempty"`
}

// IsComplete returns true if all 4 steps have been applied (steps 1-4)
func (w *DSSWorkflowResults) IsComplete() bool {
	return len(w.CompletedSteps) == 4 && w.CurrentStep == 4
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
	w.GoalDebtTradeoff = nil
	w.AppliedAtStep3 = nil
	w.BudgetAllocation = nil
	w.AppliedAtStep4 = nil
	w.LastUpdated = time.Now()
}

// ==================== Step 3: Goal-Debt Trade-off Result (NEW) ====================

// GoalDebtTradeoffResult stores the trade-off analysis result
type GoalDebtTradeoffResult struct {
	GoalAllocationPercent float64            `json:"goal_allocation_percent"`
	DebtAllocationPercent float64            `json:"debt_allocation_percent"`
	Analysis              TradeoffAnalysis   `json:"analysis"`
	SelectedScenario      string             `json:"selected_scenario"` // "conservative" | "balanced" | "aggressive"
	InterestSavings       float64            `json:"interest_savings"`
	OpportunityCost       float64            `json:"opportunity_cost"`
	MonteCarloRuns        int                `json:"monte_carlo_runs,omitempty"`
	Metadata              map[string]float64 `json:"metadata,omitempty"`
}

// TradeoffAnalysis contains detailed trade-off analysis
type TradeoffAnalysis struct {
	GoalPriority   string `json:"goal_priority"` // "high" | "medium" | "low"
	DebtPriority   string `json:"debt_priority"` // "high" | "medium" | "low"
	Recommendation string `json:"recommendation"`
	RiskLevel      string `json:"risk_level"` // "low" | "medium" | "high"
}

// ==================== Additional helper types for new workflow ====================

// CategoryAllocationItem represents allocation to one category (for Step 4)
type CategoryAllocationItem struct {
	CategoryID   uuid.UUID `json:"category_id"`
	CategoryName string    `json:"category_name"`
	Amount       float64   `json:"amount"`
	Priority     int       `json:"priority"`
	Source       string    `json:"source"` // "mandatory" | "flexible" | "goal" | "debt"
}

// GoalFundingItem represents funding for one goal (for Step 4)
type GoalFundingItem struct {
	GoalID     uuid.UUID `json:"goal_id"`
	GoalName   string    `json:"goal_name"`
	Amount     float64   `json:"amount"`
	Percentage float64   `json:"percentage"` // % of monthly target
}

// DebtPaymentItem represents payment for one debt (for Step 4)
type DebtPaymentItem struct {
	DebtID    uuid.UUID `json:"debt_id"`
	DebtName  string    `json:"debt_name"`
	Amount    float64   `json:"amount"`
	IsMinimum bool      `json:"is_minimum"`
}

// ==================== Reused Result Types (originally from dss.go) ====================

// BudgetAllocationResult contains recommendations from budget allocation DSS
type BudgetAllocationResult struct {
	Recommendations map[uuid.UUID]float64 `json:"recommendations"`  // CategoryID -> Amount
	OptimalityScore float64               `json:"optimality_score"` // 0-100
	Method          string                `json:"method"`           // "linear_programming", "heuristic"
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
