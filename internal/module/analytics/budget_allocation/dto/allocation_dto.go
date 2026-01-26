package dto

import (
	"time"

	"personalfinancedss/internal/module/analytics/budget_allocation/domain"

	"github.com/google/uuid"
)

// GenerateAllocationRequest is the request to generate allocation scenarios
type GenerateAllocationRequest struct {
	UserID         uuid.UUID `json:"user_id" binding:"required"`
	Year           int       `json:"year" binding:"required,min=2000,max=2100"`
	Month          int       `json:"month" binding:"required,min=1,max=12"`
	OverrideIncome *float64  `json:"override_income,omitempty"` // Optional: override income instead of using IncomeProfile
}

// GenerateAllocationResponse is the response containing multiple scenarios
type GenerateAllocationResponse struct {
	UserID         uuid.UUID                   `json:"user_id"`
	Period         string                      `json:"period"` // Format: "2024-12"
	TotalIncome    float64                     `json:"total_income"`
	Scenarios      []domain.AllocationScenario `json:"scenarios"`
	IsFeasible     bool                        `json:"is_feasible"`
	GlobalWarnings []domain.AllocationWarning  `json:"global_warnings,omitempty"`
	Metadata       AllocationMetadata          `json:"metadata"`
}

// AllocationMetadata contains metadata about the allocation computation
type AllocationMetadata struct {
	GeneratedAt      time.Time `json:"generated_at"`
	DataSources      []string  `json:"data_sources"`        // e.g., ["income_profile", "budget_constraints", "goals", "debts"]
	ComputationTime  int64     `json:"computation_time_ms"` // Milliseconds
	ConstraintsCount int       `json:"constraints_count"`
	GoalsCount       int       `json:"goals_count"`
	DebtsCount       int       `json:"debts_count"`
}

// ExecuteAllocationRequest is the request to execute/apply a chosen scenario
type ExecuteAllocationRequest struct {
	UserID       uuid.UUID           `json:"user_id" binding:"required"`
	ScenarioType domain.ScenarioType `json:"scenario_type" binding:"required,oneof=conservative balanced aggressive"`
	Year         int                 `json:"year" binding:"required,min=2000,max=2100"`
	Month        int                 `json:"month" binding:"required,min=1,max=12"`
}

// ExecuteAllocationResponse is the confirmation of execution
type ExecuteAllocationResponse struct {
	Success      bool                      `json:"success"`
	ExecutedAt   time.Time                 `json:"executed_at"`
	Scenario     domain.AllocationScenario `json:"scenario"`
	ActionsTaken []ActionItem              `json:"actions_taken"`
}

// ActionItem describes an action that was taken during execution
type ActionItem struct {
	Type        string    `json:"type"` // "budget_created", "goal_contribution", "debt_payment", etc.
	Description string    `json:"description"`
	EntityID    uuid.UUID `json:"entity_id,omitempty"` // ID of created/updated entity
	Amount      float64   `json:"amount"`
	Status      string    `json:"status"` // "completed", "pending", "failed"
	Error       string    `json:"error,omitempty"`
}

// CustomScenarioParams allows customizing scenario parameters
type CustomScenarioParams struct {
	ScenarioType           string  `json:"scenario_type" binding:"required,oneof=safe balanced"` // "safe" or "balanced" (matches domain.ScenarioType)
	GoalContributionFactor float64 `json:"goal_contribution_factor" binding:"gte=0,lte=2"`
	FlexibleSpendingLevel  float64 `json:"flexible_spending_level" binding:"gte=0,lte=1"`
	EmergencyFundPercent   float64 `json:"emergency_fund_percent" binding:"gte=0,lte=1"`
	GoalsPercent           float64 `json:"goals_percent" binding:"gte=0,lte=1"`
	FlexiblePercent        float64 `json:"flexible_percent" binding:"gte=0,lte=1"`
}

// BudgetAllocationModelInput is the input for the MBMS budget allocation model
// This is the standardized input format for the model interface
type BudgetAllocationModelInput struct {
	UserID               uuid.UUID              `json:"user_id" binding:"required"`
	Year                 int                    `json:"year" binding:"required,min=2000,max=2100"`
	Month                int                    `json:"month" binding:"required,min=1,max=12"`
	OverrideIncome       *float64               `json:"override_income,omitempty"`
	UseAllScenarios      bool                   `json:"use_all_scenarios"`                // If true, return 2 scenarios (Safe and Balanced); false returns only balanced
	RunSensitivity       bool                   `json:"run_sensitivity"`                  // If true, run sensitivity analysis
	CustomScenarioParams []CustomScenarioParams `json:"custom_scenario_params,omitempty"` // Optional: custom parameters for scenarios

	// Financial data for allocation
	TotalIncome         float64             `json:"total_income" binding:"required,gt=0"`
	MandatoryExpenses   []MandatoryExpense  `json:"mandatory_expenses"`
	FlexibleExpenses    []FlexibleExpense   `json:"flexible_expenses"`
	Debts               []DebtInput         `json:"debts"`
	Goals               []GoalInput         `json:"goals"`
	SensitivityOptions  *SensitivityOptions `json:"sensitivity_options,omitempty"`
	DebtStrategyPreview interface{}         `json:"debt_strategy_preview,omitempty"` // Debt strategy output to get PayoffMonth for each debt
}

// MandatoryExpense represents a fixed expense category
type MandatoryExpense struct {
	CategoryID uuid.UUID `json:"category_id" binding:"required"`
	Name       string    `json:"name" binding:"required"`
	Amount     float64   `json:"amount" binding:"required,gte=0"`
	Priority   int       `json:"priority"` // Lower = higher priority
}

// FlexibleExpense represents a variable expense category
type FlexibleExpense struct {
	CategoryID uuid.UUID `json:"category_id" binding:"required"`
	Name       string    `json:"name" binding:"required"`
	MinAmount  float64   `json:"min_amount" binding:"gte=0"`
	MaxAmount  float64   `json:"max_amount" binding:"gte=0"`
	Priority   int       `json:"priority"`
}

// DebtInput represents debt information for allocation
type DebtInput struct {
	DebtID         uuid.UUID `json:"debt_id" binding:"required"`
	Name           string    `json:"name" binding:"required"`
	Balance        float64   `json:"balance" binding:"required,gte=0"`
	InterestRate   float64   `json:"interest_rate" binding:"gte=0,lte=1"` // Annual rate as decimal
	MinimumPayment float64   `json:"minimum_payment" binding:"gte=0"`
}

// GoalInput represents goal information for allocation
type GoalInput struct {
	GoalID                uuid.UUID `json:"goal_id" binding:"required"`
	Name                  string    `json:"name" binding:"required"`
	Type                  string    `json:"type"`     // "emergency", "savings", "investment", etc.
	Priority              string    `json:"priority"` // "critical", "high", "medium", "low"
	RemainingAmount       float64   `json:"remaining_amount" binding:"gte=0"`
	SuggestedContribution float64   `json:"suggested_contribution" binding:"gte=0"`
}

// SensitivityOptions configures sensitivity analysis parameters
type SensitivityOptions struct {
	IncomeChangePercents []float64 `json:"income_change_percents"` // e.g., [-0.20, -0.10, 0.10, 0.20]
	RateChangePercents   []float64 `json:"rate_change_percents"`   // e.g., [0.02, 0.05] for +2%, +5%
	AnalyzeGoalPriority  bool      `json:"analyze_goal_priority"`  // Analyze impact of goal priority changes
}

// BudgetAllocationModelOutput is the output from the MBMS budget allocation model
type BudgetAllocationModelOutput struct {
	UserID             uuid.UUID                   `json:"user_id"`
	Period             string                      `json:"period"`
	TotalIncome        float64                     `json:"total_income"`
	Scenarios          []domain.AllocationScenario `json:"scenarios"`
	IsFeasible         bool                        `json:"is_feasible"`
	GlobalWarnings     []domain.AllocationWarning  `json:"global_warnings,omitempty"`
	SensitivityResults *SensitivityAnalysisResult  `json:"sensitivity_results,omitempty"`
	Metadata           AllocationMetadata          `json:"metadata"`
}

// SensitivityAnalysisResult contains results from sensitivity analysis
type SensitivityAnalysisResult struct {
	IncomeImpact       []IncomeImpactResult       `json:"income_impact,omitempty"`
	InterestRateImpact []InterestRateImpactResult `json:"interest_rate_impact,omitempty"`
	GoalPriorityImpact []GoalPriorityImpactResult `json:"goal_priority_impact,omitempty"`
	Summary            SensitivitySummary         `json:"summary"`
}

// IncomeImpactResult shows how income changes affect allocation
type IncomeImpactResult struct {
	IncomeChangePercent float64  `json:"income_change_percent"` // e.g., -0.10 for -10%
	NewIncome           float64  `json:"new_income"`
	IsFeasible          bool     `json:"is_feasible"`
	Deficit             float64  `json:"deficit,omitempty"` // If not feasible
	GoalAllocationDelta float64  `json:"goal_allocation_delta"`
	DebtExtraDelta      float64  `json:"debt_extra_delta"`
	FlexibleDelta       float64  `json:"flexible_delta"`
	SurplusDelta        float64  `json:"surplus_delta"`
	AffectedGoals       []string `json:"affected_goals,omitempty"` // Goals that would be reduced/eliminated
	Recommendation      string   `json:"recommendation"`
}

// InterestRateImpactResult shows how interest rate changes affect debt strategy
type InterestRateImpactResult struct {
	RateChangePercent    float64          `json:"rate_change_percent"` // e.g., 0.02 for +2%
	AffectedDebts        []DebtRateImpact `json:"affected_debts"`
	TotalExtraInterest   float64          `json:"total_extra_interest"` // Monthly extra interest
	RecommendedAction    string           `json:"recommended_action"`
	StrategyChangeNeeded bool             `json:"strategy_change_needed"`
}

// DebtRateImpact shows impact on individual debt
type DebtRateImpact struct {
	DebtID               uuid.UUID `json:"debt_id"`
	DebtName             string    `json:"debt_name"`
	OldRate              float64   `json:"old_rate"`
	NewRate              float64   `json:"new_rate"`
	ExtraMonthlyInterest float64   `json:"extra_monthly_interest"`
	NewPriority          int       `json:"new_priority"`
}

// GoalPriorityImpactResult shows how goal priority changes affect allocation
type GoalPriorityImpactResult struct {
	GoalID                uuid.UUID `json:"goal_id"`
	GoalName              string    `json:"goal_name"`
	CurrentPriority       string    `json:"current_priority"`
	CurrentAllocation     float64   `json:"current_allocation"`
	IfHigherPriority      float64   `json:"if_higher_priority"`     // Allocation if priority increased
	IfLowerPriority       float64   `json:"if_lower_priority"`      // Allocation if priority decreased
	AllocationSensitivity string    `json:"allocation_sensitivity"` // "high", "medium", "low"
}

// SensitivitySummary provides overall sensitivity insights
type SensitivitySummary struct {
	MostSensitiveToIncome bool     `json:"most_sensitive_to_income"`
	IncomeBreakEvenPoint  float64  `json:"income_break_even_point"`       // Income at which allocation becomes infeasible
	HighRiskDebts         []string `json:"high_risk_debts,omitempty"`     // Debts most affected by rate changes
	MostFlexibleGoals     []string `json:"most_flexible_goals,omitempty"` // Goals that can absorb changes
	OverallRiskLevel      string   `json:"overall_risk_level"`            // "low", "medium", "high"
	KeyRecommendations    []string `json:"key_recommendations"`
}
