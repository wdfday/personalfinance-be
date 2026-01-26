package domain

import (
	"github.com/google/uuid"
)

// ConstraintModel represents all constraints for budget allocation
type ConstraintModel struct {
	TotalIncome float64

	// Category constraints from BudgetProfile
	MandatoryExpenses map[uuid.UUID]CategoryConstraint // CategoryID -> constraint
	FlexibleExpenses  map[uuid.UUID]CategoryConstraint

	// Debt constraints
	DebtPayments map[uuid.UUID]DebtConstraint // DebtID -> constraint

	// Goal constraints
	GoalTargets map[uuid.UUID]GoalConstraint // GoalID -> constraint
}

// CategoryConstraint represents a budget constraint for a category
type CategoryConstraint struct {
	CategoryID uuid.UUID
	Minimum    float64
	Maximum    float64 // 0 means no maximum
	IsFlexible bool
	Priority   int
}

// DebtConstraint represents a debt payment constraint
type DebtConstraint struct {
	DebtID         uuid.UUID
	DebtName       string
	MinimumPayment float64
	FixedPayment   float64 // If > 0, force this payment amount (user's desired payment). If 0, use MinimumPayment
	CurrentBalance float64
	InterestRate   float64
	Priority       int // Calculated based on interest rate and balance
}

// GoalConstraint represents a goal contribution constraint
type GoalConstraint struct {
	GoalID                uuid.UUID
	GoalName              string
	GoalType              string
	SuggestedContribution float64
	Priority              string // low, medium, high, critical
	PriorityWeight        int    // Numerical weight for sorting
	RemainingAmount       float64
}

// AllocationResult represents the result of the allocation algorithm
type AllocationResult struct {
	CategoryAllocations map[uuid.UUID]float64 // CategoryID -> amount
	GoalAllocations     map[uuid.UUID]float64 // GoalID -> amount
	DebtAllocations     map[uuid.UUID]DebtPayment
	TotalAllocated      float64
	Surplus             float64
	FeasibilityScore    float64
	SolverIterations    int
	AchievedGoals       []string
	UnachievedGoals     []string
	SolverType          string // "preemptive" or "weighted"
}

// DebtPayment represents a debt payment breakdown
type DebtPayment struct {
	TotalPayment   float64
	MinimumPayment float64
	ExtraPayment   float64
}

// ScenarioParameters defines parameters for different allocation scenarios
type ScenarioParameters struct {
	ScenarioType           ScenarioType
	GoalContributionFactor float64 // Multiplier for suggested goal contributions
	FlexibleSpendingLevel  float64 // 0.0 = minimum, 0.5 = mid, 1.0 = maximum
	SurplusAllocation      SurplusAllocation
}

// SurplusAllocation defines how to allocate surplus income
type SurplusAllocation struct {
	EmergencyFundPercent float64 // % to emergency fund goals
	DebtExtraPercent     float64 // % to extra debt payments
	GoalsPercent         float64 // % to high-priority goals
	FlexiblePercent      float64 // % to flexible categories
}

// DualGPResult contains results from both GP solvers for comparison
type DualGPResult struct {
	PreemptiveResult *AllocationResult
	WeightedResult   *AllocationResult
	Comparison       GPComparison
}

// GPComparison compares the two GP approaches
type GPComparison struct {
	PreemptiveAchievedCount int
	WeightedAchievedCount   int
	PreemptiveTotalDev      float64
	WeightedTotalDev        float64
	RecommendedSolver       string
	Reason                  string
}

// TripleGPResult contains results from all three GP solvers
type TripleGPResult struct {
	PreemptiveResult *AllocationResult
	WeightedResult   *AllocationResult
	MinmaxResult     *AllocationResult
	Comparison       TripleGPComparison
}

// TripleGPComparison compares all three GP approaches
type TripleGPComparison struct {
	PreemptiveAchievedCount int
	WeightedAchievedCount   int
	MinmaxAchievedCount     int
	MinmaxMinAchievement    float64 // Minimum achievement % across all goals
	MinmaxIsBalanced        bool
	RecommendedSolver       string
	Reason                  string
}
