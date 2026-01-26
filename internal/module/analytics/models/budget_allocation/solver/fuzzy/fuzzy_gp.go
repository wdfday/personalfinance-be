package fuzzy

import (
	"fmt"
	"math"
	"personalfinancedss/internal/module/analytics/budget_allocation/domain"
	"sort"

	"github.com/google/uuid"
)

// FuzzyGPSolver implements Fuzzy Goal Programming
// Uses membership functions to evaluate goal satisfaction with continuous degrees
// Each goal has a fuzzy membership function (triangular or trapezoidal)
type FuzzyGPSolver struct {
	goals       []FuzzyGoal
	variables   []FuzzyVariable
	totalIncome float64
}

// FuzzyVariable represents a decision variable
type FuzzyVariable struct {
	ID       uuid.UUID
	Name     string
	Type     string // "category", "goal", "debt"
	MinValue float64
	MaxValue float64
}

// MembershipFunction defines a fuzzy membership function
// Supports triangular, trapezoidal, and piecewise linear shapes
type MembershipFunction struct {
	Type      string  // "triangular", "trapezoidal", "piecewise", "linear", "s-curve" (s-curve is piecewise-linear approximation)
	Lower     float64 // Left boundary (membership = 0)
	PeakLeft  float64 // Left peak (membership = 1 for triangular, start of plateau for trapezoidal)
	PeakRight float64 // Right peak (membership = 1 for triangular, end of plateau for trapezoidal)
	Upper     float64 // Right boundary (membership = 0)

	// For piecewise linear functions (concave, S-curve, etc.)
	Segments []PiecewiseSegment // Segments with different slopes/weights
}

// PiecewiseSegment represents a linear segment in piecewise membership function
type PiecewiseSegment struct {
	Lower         float64 // Start of segment
	Upper         float64 // End of segment
	Slope         float64 // Slope of membership in this segment
	Intercept     float64 // Intercept at Lower
	MaxMembership float64 // Maximum membership in this segment (for normalization)
}

// FuzzyGoal represents a goal with fuzzy membership function
type FuzzyGoal struct {
	ID             string
	Description    string
	VariableIdx    int
	MembershipFunc MembershipFunction
	Priority       int     // Lower = higher priority
	Weight         float64 // Importance weight
	TargetValue    float64 // Nominal target (center of membership function)
}

// FuzzyResult contains the solution
type FuzzyResult struct {
	VariableValues     map[uuid.UUID]float64
	GoalMemberships    map[string]float64 // Goal ID -> membership degree [0,1]
	GoalValues         map[string]float64 // Goal ID -> actual value
	TotalMembership    float64            // Sum of all membership degrees
	AverageMembership  float64            // Average membership across all goals
	WeightedMembership float64            // Weighted average membership
	AchievedGoals      []string           // Goals with membership >= 0.8
	PartialGoals       []string           // Goals with 0.3 <= membership < 0.8
	UnachievedGoals    []string           // Goals with membership < 0.3
	Iterations         int
}

// NewFuzzyGPSolver creates a new solver
func NewFuzzyGPSolver(totalIncome float64) *FuzzyGPSolver {
	return &FuzzyGPSolver{
		goals:       make([]FuzzyGoal, 0),
		variables:   make([]FuzzyVariable, 0),
		totalIncome: totalIncome,
	}
}

// AddVariable adds a decision variable
func (s *FuzzyGPSolver) AddVariable(v FuzzyVariable) int {
	idx := len(s.variables)
	s.variables = append(s.variables, v)
	return idx
}

// AddGoal adds a goal with fuzzy membership function
func (s *FuzzyGPSolver) AddGoal(g FuzzyGoal) {
	s.goals = append(s.goals, g)
}

// EvaluateMembership calculates the membership degree for a given value
func (mf MembershipFunction) EvaluateMembership(value float64) float64 {
	switch mf.Type {
	case "triangular":
		return mf.evaluateTriangular(value)
	case "trapezoidal":
		return mf.evaluateTrapezoidal(value)
	case "piecewise":
		return mf.evaluatePiecewise(value)
	case "linear":
		return mf.evaluateLinear(value)
	case "s-curve":
		// NOTE: We keep a dedicated type for readability, but the implementation is
		// piecewise-linear to remain LP/MILP-friendly (no sigmoid/exp).
		return mf.evaluateSCurve(value)
	default:
		return 0.0
	}
}

// evaluateTriangular calculates triangular membership
func (mf MembershipFunction) evaluateTriangular(value float64) float64 {
	if value < mf.Lower || value > mf.Upper {
		return 0.0
	}
	if value < mf.PeakLeft {
		// Rising edge
		if mf.PeakLeft == mf.Lower {
			return 1.0
		}
		return (value - mf.Lower) / (mf.PeakLeft - mf.Lower)
	}
	if value > mf.PeakRight {
		// Falling edge
		if mf.PeakRight == mf.Upper {
			return 1.0
		}
		return (mf.Upper - value) / (mf.Upper - mf.PeakRight)
	}
	// Plateau (peak)
	return 1.0
}

// evaluateTrapezoidal calculates trapezoidal membership
func (mf MembershipFunction) evaluateTrapezoidal(value float64) float64 {
	if value < mf.Lower || value > mf.Upper {
		return 0.0
	}
	if value < mf.PeakLeft {
		// Rising edge
		if mf.PeakLeft == mf.Lower {
			return 1.0
		}
		return (value - mf.Lower) / (mf.PeakLeft - mf.Lower)
	}
	if value > mf.PeakRight {
		// Falling edge
		if mf.PeakRight == mf.Upper {
			return 1.0
		}
		return (mf.Upper - value) / (mf.Upper - mf.PeakRight)
	}
	// Plateau
	return 1.0
}

// evaluatePiecewise calculates piecewise linear membership (for concave, S-curve, etc.)
func (mf MembershipFunction) evaluatePiecewise(value float64) float64 {
	if len(mf.Segments) == 0 {
		return 0.0
	}

	// Find which segment contains the value
	for i, seg := range mf.Segments {
		if value >= seg.Lower && value <= seg.Upper {
			membership := seg.Slope*value + seg.Intercept
			// Ensure membership is non-negative
			if membership < 0 {
				return 0.0
			}
			// Cap at MaxMembership to ensure continuity (MaxMembership is the target at end of segment)
			if membership > seg.MaxMembership {
				return seg.MaxMembership
			}
			return membership
		}
		// If value exceeds the last segment, return the max membership of the last segment
		if i == len(mf.Segments)-1 && value > seg.Upper {
			return seg.MaxMembership
		}
	}
	// Value is below the first segment
	if value < mf.Segments[0].Lower {
		return 0.0
	}
	return 0.0
}

// evaluateLinear calculates linear membership (for debt - simple linear utility)
func (mf MembershipFunction) evaluateLinear(value float64) float64 {
	if value < mf.Lower {
		return 0.0
	}
	if value > mf.Upper {
		return 1.0 // Saturated at max
	}
	// Linear from Lower (0) to Upper (1)
	if mf.Upper == mf.Lower {
		return 1.0
	}
	return (value - mf.Lower) / (mf.Upper - mf.Lower)
}

// evaluateSCurve calculates an S-curve-like membership using a piecewise-linear approximation.
// Shape (in terms of r = value/Upper):
// - slow increase in [0.0, 0.3]
// - fast increase in (0.3, 0.7)
// - saturation in (0.7, 1.0]
// - diminishing returns in (1.0, 1.2] (small extra utility)
func (mf MembershipFunction) evaluateSCurve(value float64) float64 {
	segments := buildSigmoidLikeSegments(mf.Lower, mf.PeakLeft, mf.PeakRight, mf.Upper)
	return evaluatePiecewiseWithSegments(value, segments)
}

// buildSigmoidLikeSegments returns a monotone piecewise-linear S-curve approximation.
// The curve is defined by 5 anchor points (value -> membership):
// - Lower -> 0.00
// - PeakLeft -> 0.15
// - PeakRight -> 0.85
// - Upper -> 1.00
// - Upper*1.2 -> 1.05
//
// To make it "smooth", we split regions into multiple small segments (<= 8 total).
func buildSigmoidLikeSegments(lower, peakLeft, peakRight, upper float64) []PiecewiseSegment {
	// Guard rails
	if upper <= lower {
		return []PiecewiseSegment{
			{Lower: lower, Upper: upper, Slope: 0, Intercept: 1, MaxMembership: 1},
		}
	}

	// If peak markers are not set, fall back to 30% and 70% of target.
	if peakLeft <= lower || peakLeft >= upper {
		peakLeft = lower + 0.3*(upper-lower)
	}
	if peakRight <= peakLeft || peakRight >= upper {
		peakRight = lower + 0.7*(upper-lower)
	}

	// Membership anchor values
	m0 := 0.0
	m1 := 0.15
	m2 := 0.85
	m3 := 1.0
	surplusMax := upper * 1.2
	m4 := 1.05

	// Helper to create a segment from (xA, mA) -> (xB, mB)
	seg := func(xA, mA, xB, mB float64) PiecewiseSegment {
		slope := 0.0
		if xB != xA {
			slope = (mB - mA) / (xB - xA)
		}
		intercept := mA - slope*xA
		return PiecewiseSegment{
			Lower:         xA,
			Upper:         xB,
			Slope:         slope,
			Intercept:     intercept,
			MaxMembership: mB,
		}
	}

	// Split regions for a "smoother" feel (8 segments total)
	// Region [lower, peakLeft] (slow)
	x01 := lower + 0.5*(peakLeft-lower)
	m01 := 0.05

	// Region [peakLeft, peakRight] (fast) split into thirds
	x12a := peakLeft + (peakRight-peakLeft)/3.0
	x12b := peakLeft + 2.0*(peakRight-peakLeft)/3.0
	m12a := 0.35
	m12b := 0.65

	// Region [peakRight, upper] (saturation) split in half
	x23 := peakRight + 0.5*(upper-peakRight)
	m23 := 0.95

	segments := []PiecewiseSegment{
		seg(lower, m0, x01, m01),
		seg(x01, m01, peakLeft, m1),
		seg(peakLeft, m1, x12a, m12a),
		seg(x12a, m12a, x12b, m12b),
		seg(x12b, m12b, peakRight, m2),
		seg(peakRight, m2, x23, m23),
		seg(x23, m23, upper, m3),
	}

	// Surplus tier (diminishing returns)
	if surplusMax > upper {
		segments = append(segments, seg(upper, m3, surplusMax, m4))
	}

	return segments
}

func evaluatePiecewiseWithSegments(value float64, segments []PiecewiseSegment) float64 {
	if len(segments) == 0 {
		return 0.0
	}

	// Below first segment
	if value < segments[0].Lower {
		return 0.0
	}

	for i, seg := range segments {
		if value >= seg.Lower && value <= seg.Upper {
			membership := seg.Slope*value + seg.Intercept
			if membership < 0 {
				return 0.0
			}
			if membership > seg.MaxMembership {
				return seg.MaxMembership
			}
			return membership
		}
		// Beyond last segment: cap at last MaxMembership
		if i == len(segments)-1 && value > seg.Upper {
			return seg.MaxMembership
		}
	}

	return 0.0
}

// Solve executes the Fuzzy GP algorithm
// Strategy: Tries to use Golp (MILP) first, falls back to Greedy algorithm if Golp is unavailable
func (s *FuzzyGPSolver) Solve() (*FuzzyResult, error) {
	// Try solving with Golp (MILP)
	result, err := s.SolveWithGolp()
	if err == nil {
		return result, nil
	}
	// Fallback to greedy algorithm if Golp is unavailable

	result = &FuzzyResult{
		VariableValues:  make(map[uuid.UUID]float64),
		GoalMemberships: make(map[string]float64),
		GoalValues:      make(map[string]float64),
		Iterations:      0,
	}

	if len(s.variables) == 0 {
		return result, nil
	}

	// Initialize solution with minimum values
	currentSolution := make([]float64, len(s.variables))
	remainingBudget := s.totalIncome

	// Phase 1: Allocate mandatory expenses (hard constraints)
	for i, v := range s.variables {
		if v.Type == "category" {
			// Check if it's a mandatory expense (min == max typically)
			if v.MinValue > 0 {
				currentSolution[i] = v.MinValue
				remainingBudget -= v.MinValue
				if remainingBudget < 0 {
					return result, fmt.Errorf("infeasible: insufficient budget for mandatory expenses")
				}
			}
		}
	}

	// Phase 2: Allocate minimum debt payments
	for i, v := range s.variables {
		if v.Type == "debt" {
			if v.MinValue > 0 {
				currentSolution[i] = v.MinValue
				remainingBudget -= v.MinValue
				if remainingBudget < 0 {
					return result, fmt.Errorf("infeasible: insufficient budget for minimum debt payments")
				}
			}
		}
	}

	// Phase 3: Allocate remaining budget to maximize fuzzy satisfaction
	// Group goals by priority
	priorityGroups := s.groupGoalsByPriority()
	priorities := s.getSortedPriorities(priorityGroups)

	for _, priority := range priorities {
		goals := priorityGroups[priority]
		if remainingBudget <= 0.01 {
			break
		}

		// Allocate within this priority level to maximize weighted membership
		remainingBudget = s.allocateWithinPriority(goals, currentSolution, remainingBudget, result)
		result.Iterations++

		// Second pass: allocate remaining budget within this priority (like mcgp)
		if remainingBudget > 0.01 {
			remainingBudget = s.allocateRemainingWithinPriority(goals, currentSolution, remainingBudget)
		}
	}

	// Allocate remaining budget to ensure total = totalIncome (budget constraint)
	// Distribute remaining budget proportionally to maximize weighted membership
	if remainingBudget > 0.01 {
		// Try to allocate to goals first (higher priority)
		remainingBudget = s.allocateRemainingAcrossAllPriorities(currentSolution, remainingBudget)

		// If still have budget, allocate to flexible expenses
		if remainingBudget > 0.01 {
			remainingBudget = s.allocateToFlexibleOrGoals(currentSolution, remainingBudget)
		}

		// Final fallback: allocate to any variable that can take more
		if remainingBudget > 0.01 {
			for i := range s.variables {
				if remainingBudget <= 0.01 {
					break
				}
				variable := s.variables[i]
				currentValue := currentSolution[i]
				if variable.MaxValue <= 0 || currentValue < variable.MaxValue {
					allocation := remainingBudget
					if variable.MaxValue > 0 {
						allocation = math.Min(remainingBudget, variable.MaxValue-currentValue)
					}
					if allocation > 0.01 {
						currentSolution[i] += allocation
						remainingBudget -= allocation
					}
				}
			}
		}
	}

	// Store solution
	s.storeSolution(result, currentSolution)

	// Calculate final memberships and categorize goals
	s.calculateFinalMemberships(result, currentSolution)

	return result, nil
}

// allocateWithinPriority allocates budget within a priority level to maximize fuzzy satisfaction
// IMPROVED: Distribute proportionally based on target values to ensure balanced allocation
func (s *FuzzyGPSolver) allocateWithinPriority(
	goals []FuzzyGoal,
	solution []float64,
	remainingBudget float64,
	result *FuzzyResult,
) float64 {
	// Calculate total target for proportional distribution
	totalTarget := 0.0
	goalTargets := make(map[int]float64) // idx -> target
	for _, goal := range goals {
		idx := goal.VariableIdx
		variable := s.variables[idx]

		// Check if we can allocate more
		if variable.MaxValue > 0 && solution[idx] >= variable.MaxValue {
			continue
		}

		target := goal.TargetValue
		goalTargets[idx] = target
		totalTarget += target
	}

	// If no targets, fall back to equal distribution
	if totalTarget <= 0 {
		// Equal distribution fallback
		validGoals := 0
		for _, goal := range goals {
			idx := goal.VariableIdx
			variable := s.variables[idx]
			if variable.MaxValue <= 0 || solution[idx] < variable.MaxValue {
				validGoals++
			}
		}
		if validGoals > 0 {
			perGoal := remainingBudget / float64(validGoals)
			for _, goal := range goals {
				idx := goal.VariableIdx
				variable := s.variables[idx]
				if variable.MaxValue <= 0 || solution[idx] < variable.MaxValue {
					allocation := perGoal
					if variable.MaxValue > 0 {
						allocation = math.Min(allocation, variable.MaxValue-solution[idx])
					}
					if allocation > 0 {
						solution[idx] += allocation
						remainingBudget -= allocation
					}
				}
			}
		}
		return remainingBudget
	}

	// Use iterative improvement to maximize weighted membership sum
	// Objective: maximize sum(membership(value) * weight)
	// Membership function maps value to satisfaction level based on segments
	iterationLimit := 500
	iteration := 0
	increment := math.Max(remainingBudget/200, 0.5)

	for remainingBudget > 0.01 && iteration < iterationLimit {
		// Find best allocation that maximizes weighted membership improvement
		type allocationOption struct {
			idx         int
			allocation  float64
			improvement float64 // weighted membership improvement
			ratio       float64 // improvement per dollar
		}
		options := make([]allocationOption, 0)

		for idx := range goalTargets {
			goal := goals[0] // Find goal by idx
			for _, g := range goals {
				if g.VariableIdx == idx {
					goal = g
					break
				}
			}
			variable := s.variables[idx]

			// Check if we can allocate more
			if variable.MaxValue > 0 && solution[idx] >= variable.MaxValue {
				continue
			}

			currentValue := solution[idx]
			currentMembership := goal.MembershipFunc.EvaluateMembership(currentValue)

			// Try allocating increment
			testValue := currentValue + increment
			if variable.MaxValue > 0 && testValue > variable.MaxValue {
				testValue = variable.MaxValue
			}
			testMembership := goal.MembershipFunc.EvaluateMembership(testValue)
			improvement := (testMembership - currentMembership) * goal.Weight
			allocationAmount := testValue - currentValue

			if allocationAmount > 0.01 && improvement > 0.001 {
				ratio := improvement / allocationAmount
				options = append(options, allocationOption{
					idx:         idx,
					allocation:  allocationAmount,
					improvement: improvement,
					ratio:       ratio,
				})
			}
		}

		if len(options) == 0 {
			break // No improvement possible
		}

		// Allocate to ALL goals with positive improvement (not just best one)
		// This ensures all goals increase when there's more budget
		// Sort by improvement ratio (best first) for ordering, but allocate to all
		for i := 0; i < len(options)-1; i++ {
			for j := i + 1; j < len(options); j++ {
				if options[i].ratio < options[j].ratio {
					options[i], options[j] = options[j], options[i]
				}
			}
		}

		// Calculate total improvement for normalization
		totalImprovement := 0.0
		for _, opt := range options {
			totalImprovement += opt.improvement
		}

		// Allocate to all goals proportionally based on improvement
		// This ensures all goals increase when there's more budget
		if totalImprovement > 0 {
			for _, opt := range options {
				if remainingBudget <= 0.01 {
					break
				}
				variable := s.variables[opt.idx]

				// Proportional allocation based on improvement ratio
				// Goals with better improvement get more, but all get some
				improvementRatio := opt.improvement / totalImprovement
				proportionalAllocation := increment * improvementRatio

				actualAllocation := math.Min(proportionalAllocation, remainingBudget)
				if variable.MaxValue > 0 {
					actualAllocation = math.Min(actualAllocation, variable.MaxValue-solution[opt.idx])
				}
				actualAllocation = math.Min(actualAllocation, opt.allocation) // Don't exceed what was tested

				if actualAllocation > 0.01 {
					solution[opt.idx] += actualAllocation
					remainingBudget -= actualAllocation
				}
			}
		}

		iteration++
	}

	return remainingBudget
}

// allocateRemainingWithinPriority allocates remaining budget within a priority level
func (s *FuzzyGPSolver) allocateRemainingWithinPriority(goals []FuzzyGoal, solution []float64, remainingBudget float64) float64 {
	// Calculate total target for proportional distribution
	totalTarget := 0.0
	goalTargets := make(map[int]float64) // idx -> target
	validGoals := make([]int, 0)         // List of valid goal indices

	for _, goal := range goals {
		idx := goal.VariableIdx
		if idx >= len(solution) {
			continue
		}

		variable := s.variables[idx]
		currentValue := solution[idx]

		// Check if we can allocate more
		if variable.MaxValue > 0 && currentValue >= variable.MaxValue {
			continue
		}

		target := goal.TargetValue
		goalTargets[idx] = target
		validGoals = append(validGoals, idx)
		totalTarget += target
	}

	if len(validGoals) == 0 {
		return remainingBudget
	}

	// If no targets, fall back to equal distribution
	if totalTarget <= 0 {
		perGoal := remainingBudget / float64(len(validGoals))
		for _, idx := range validGoals {
			variable := s.variables[idx]
			allocation := perGoal
			if variable.MaxValue > 0 {
				allocation = math.Min(allocation, variable.MaxValue-solution[idx])
			}
			if allocation > 0.01 {
				solution[idx] += allocation
				remainingBudget -= allocation
			}
		}
		return remainingBudget
	}

	// Maximize weighted membership improvement
	// Objective: maximize sum(membership(value) * weight)
	// Membership function is monotone increasing, so we allocate to maximize improvement
	iterationLimit := 100
	iteration := 0
	increment := math.Max(remainingBudget/50, 1.0)

	for remainingBudget > 0.01 && iteration < iterationLimit {
		// Find best allocation that maximizes weighted membership improvement
		type allocationOption struct {
			idx         int
			allocation  float64
			improvement float64 // weighted membership improvement
			ratio       float64 // improvement per dollar
		}
		options := make([]allocationOption, 0)

		for _, idx := range validGoals {
			// Find goal by idx
			var goal FuzzyGoal
			found := false
			for _, g := range goals {
				if g.VariableIdx == idx {
					goal = g
					found = true
					break
				}
			}
			if !found {
				continue
			}

			variable := s.variables[idx]
			currentValue := solution[idx]

			// Check if we can allocate more
			if variable.MaxValue > 0 && currentValue >= variable.MaxValue {
				continue
			}

			currentMembership := goal.MembershipFunc.EvaluateMembership(currentValue)

			// Try allocating increment
			testValue := currentValue + increment
			if variable.MaxValue > 0 && testValue > variable.MaxValue {
				testValue = variable.MaxValue
			}
			testMembership := goal.MembershipFunc.EvaluateMembership(testValue)
			improvement := (testMembership - currentMembership) * goal.Weight
			allocationAmount := testValue - currentValue

			if allocationAmount > 0.01 && improvement > 0.001 {
				ratio := improvement / allocationAmount
				options = append(options, allocationOption{
					idx:         idx,
					allocation:  allocationAmount,
					improvement: improvement,
					ratio:       ratio,
				})
			}
		}

		if len(options) == 0 {
			break // No improvement possible
		}

		// Allocate to ALL goals with positive improvement (not just best one)
		// This ensures all goals increase when there's more budget
		// Sort by improvement ratio (best first) for ordering, but allocate to all
		for i := 0; i < len(options)-1; i++ {
			for j := i + 1; j < len(options); j++ {
				if options[i].ratio < options[j].ratio {
					options[i], options[j] = options[j], options[i]
				}
			}
		}

		// Calculate total improvement for normalization
		totalImprovement := 0.0
		for _, opt := range options {
			totalImprovement += opt.improvement
		}

		// Allocate to all goals proportionally based on improvement
		// This ensures all goals increase when there's more budget
		if totalImprovement > 0 {
			for _, opt := range options {
				if remainingBudget <= 0.01 {
					break
				}
				variable := s.variables[opt.idx]

				// Proportional allocation based on improvement ratio
				// Goals with better improvement get more, but all get some
				improvementRatio := opt.improvement / totalImprovement
				proportionalAllocation := increment * improvementRatio

				actualAllocation := math.Min(proportionalAllocation, remainingBudget)
				if variable.MaxValue > 0 {
					actualAllocation = math.Min(actualAllocation, variable.MaxValue-solution[opt.idx])
				}
				actualAllocation = math.Min(actualAllocation, opt.allocation) // Don't exceed what was tested

				if actualAllocation > 0.01 {
					solution[opt.idx] += actualAllocation
					remainingBudget -= actualAllocation
				}
			}
		}

		iteration++
	}

	return remainingBudget
}

// allocateRemainingAcrossAllPriorities allocates remaining budget across all goals
func (s *FuzzyGPSolver) allocateRemainingAcrossAllPriorities(solution []float64, remainingBudget float64) float64 {
	type upgradeOption struct {
		idx          int
		currentValue float64
		needed       float64
		improvement  float64
		ratio        float64
	}

	options := make([]upgradeOption, 0)

	// Collect all goals (across all priorities)
	for _, goal := range s.goals {
		idx := goal.VariableIdx
		if idx >= len(solution) {
			continue
		}

		currentValue := solution[idx]
		variable := s.variables[idx]

		// Check if we can allocate more
		if variable.MaxValue > 0 && currentValue >= variable.MaxValue {
			continue
		}

		// Calculate current membership
		currentMembership := goal.MembershipFunc.EvaluateMembership(currentValue)

		// Try allocating remaining budget (or up to max)
		testValue := currentValue + remainingBudget
		if variable.MaxValue > 0 && testValue > variable.MaxValue {
			testValue = variable.MaxValue
		}

		testMembership := goal.MembershipFunc.EvaluateMembership(testValue)
		improvement := (testMembership - currentMembership) * goal.Weight
		needed := testValue - currentValue

		if needed > 0.01 && improvement > 0 {
			ratio := improvement / needed
			options = append(options, upgradeOption{
				idx:          idx,
				currentValue: currentValue,
				needed:       needed,
				improvement:  improvement,
				ratio:        ratio,
			})
		}
	}

	// Sort by ratio (highest first)
	for i := 0; i < len(options); i++ {
		for j := i + 1; j < len(options); j++ {
			if options[i].ratio < options[j].ratio {
				options[i], options[j] = options[j], options[i]
			}
		}
	}

	// Allocate to best options until budget is exhausted
	for _, opt := range options {
		if remainingBudget < 0.01 {
			break
		}

		allocation := math.Min(opt.needed, remainingBudget)
		if allocation > 0.01 {
			solution[opt.idx] += allocation
			remainingBudget -= allocation
		}
	}

	return remainingBudget
}

// allocateRemainingBudget allocates any remaining budget to goals with highest potential
func (s *FuzzyGPSolver) allocateRemainingBudget(solution []float64, remainingBudget float64) float64 {
	// Find all goals sorted by weight and current membership
	type goalWithInfo struct {
		goal         FuzzyGoal
		idx          int
		currentValue float64
		membership   float64
		potential    float64
	}

	goalsInfo := make([]goalWithInfo, 0)
	for _, goal := range s.goals {
		idx := goal.VariableIdx
		currentValue := solution[idx]
		variable := s.variables[idx]

		if variable.MaxValue > 0 && currentValue >= variable.MaxValue {
			continue
		}

		membership := goal.MembershipFunc.EvaluateMembership(currentValue)
		// Calculate potential: what membership could we achieve with remaining budget?
		testValue := math.Min(currentValue+remainingBudget, variable.MaxValue)
		if variable.MaxValue > 0 && testValue > variable.MaxValue {
			testValue = variable.MaxValue
		}
		potentialMembership := goal.MembershipFunc.EvaluateMembership(testValue)
		potential := (potentialMembership - membership) * goal.Weight

		goalsInfo = append(goalsInfo, goalWithInfo{
			goal:         goal,
			idx:          idx,
			currentValue: currentValue,
			membership:   membership,
			potential:    potential,
		})
	}

	// Sort by potential (descending)
	for i := 0; i < len(goalsInfo); i++ {
		for j := i + 1; j < len(goalsInfo); j++ {
			if goalsInfo[i].potential < goalsInfo[j].potential {
				goalsInfo[i], goalsInfo[j] = goalsInfo[j], goalsInfo[i]
			}
		}
	}

	// Allocate to best goal
	if len(goalsInfo) > 0 && goalsInfo[0].potential > 0 {
		best := goalsInfo[0]
		variable := s.variables[best.idx]
		allocation := math.Min(remainingBudget, variable.MaxValue-best.currentValue)
		if variable.MaxValue <= 0 {
			allocation = remainingBudget
		}
		if allocation > 0 {
			solution[best.idx] += allocation
			remainingBudget -= allocation
		}
	}

	return remainingBudget
}

// allocateToFlexibleOrGoals allocates remaining budget to flexible expenses or goals beyond max
// IMPROVED: Distribute proportionally based on target values to maintain similar achievement ratios
func (s *FuzzyGPSolver) allocateToFlexibleOrGoals(solution []float64, remainingBudget float64) float64 {
	// Collect flexible expenses (categories) with their targets
	type flexExpense struct {
		idx    int
		target float64
		goal   FuzzyGoal
	}
	flexExpenses := make([]flexExpense, 0)
	totalFlexTarget := 0.0

	// Find flexible expense goals (Priority 4 typically)
	for _, goal := range s.goals {
		idx := goal.VariableIdx
		if idx >= len(s.variables) {
			continue
		}
		variable := s.variables[idx]

		// Check if it's a flexible expense (category type)
		if variable.Type == "category" {
			currentValue := solution[idx]
			// Check if we can allocate more
			if variable.MaxValue <= 0 || currentValue < variable.MaxValue {
				target := goal.TargetValue
				if target > 0 {
					flexExpenses = append(flexExpenses, flexExpense{
						idx:    idx,
						target: target,
						goal:   goal,
					})
					totalFlexTarget += target
				}
			}
		}
	}

	// Allocate to flexible expenses proportionally based on targets
	if len(flexExpenses) > 0 && totalFlexTarget > 0 {
		iterationLimit := 100
		iteration := 0
		increment := math.Max(remainingBudget/50, 1.0)

		for remainingBudget > 0.01 && iteration < iterationLimit {
			allocations := make(map[int]float64)
			totalAllocatable := 0.0

			for _, flex := range flexExpenses {
				variable := s.variables[flex.idx]
				currentValue := solution[flex.idx]

				// Check if we can allocate more
				if variable.MaxValue > 0 && currentValue >= variable.MaxValue {
					continue
				}

				// Proportional allocation based on target
				// This maintains similar achievement ratios (value/target) across flexible expenses
				proportionalAmount := (flex.target / totalFlexTarget) * increment

				// Cap at remaining budget and max value
				allocation := math.Min(proportionalAmount, remainingBudget)
				if variable.MaxValue > 0 {
					allocation = math.Min(allocation, variable.MaxValue-currentValue)
				}

				if allocation > 0.01 {
					allocations[flex.idx] = allocation
					totalAllocatable += allocation
				}
			}

			if len(allocations) == 0 {
				break // No more allocation possible
			}

			// Normalize if total exceeds remaining budget
			if totalAllocatable > remainingBudget {
				scale := remainingBudget / totalAllocatable
				for idx := range allocations {
					allocations[idx] *= scale
				}
			}

			// Allocate to all flexible expenses proportionally
			for idx, allocation := range allocations {
				if remainingBudget <= 0.01 {
					break
				}
				if allocation > 0.01 {
					solution[idx] += allocation
					remainingBudget -= allocation
				}
			}

			iteration++
		}
	}

	// If still have budget, allocate to goals beyond max (surplus tier)
	// Find goal with highest weight
	bestGoalIdx := -1
	bestWeight := 0.0
	for _, goal := range s.goals {
		idx := goal.VariableIdx
		if idx >= len(s.variables) {
			continue
		}
		variable := s.variables[idx]
		currentValue := solution[idx]

		// Check if we can allocate more (beyond target, in surplus tier)
		if variable.MaxValue <= 0 || currentValue < variable.MaxValue {
			if goal.Weight > bestWeight {
				bestWeight = goal.Weight
				bestGoalIdx = idx
			}
		}
	}

	if bestGoalIdx >= 0 {
		// Allocate all remaining to highest priority goal (surplus tier)
		variable := s.variables[bestGoalIdx]
		allocation := remainingBudget
		if variable.MaxValue > 0 {
			allocation = math.Min(allocation, variable.MaxValue-solution[bestGoalIdx])
		}
		if allocation > 0.01 {
			solution[bestGoalIdx] += allocation
			remainingBudget -= allocation
		}
	}

	return remainingBudget
}

// groupGoalsByPriority groups goals by their priority
func (s *FuzzyGPSolver) groupGoalsByPriority() map[int][]FuzzyGoal {
	groups := make(map[int][]FuzzyGoal)
	for _, goal := range s.goals {
		groups[goal.Priority] = append(groups[goal.Priority], goal)
	}
	return groups
}

// getSortedPriorities returns priorities sorted in ascending order (lower = higher priority)
func (s *FuzzyGPSolver) getSortedPriorities(groups map[int][]FuzzyGoal) []int {
	priorities := make([]int, 0, len(groups))
	for p := range groups {
		priorities = append(priorities, p)
	}
	sort.Ints(priorities)
	return priorities
}

// storeSolution stores the solution values
func (s *FuzzyGPSolver) storeSolution(result *FuzzyResult, solution []float64) {
	for i, v := range s.variables {
		if solution[i] > 0 {
			result.VariableValues[v.ID] = solution[i]
		}
	}
}

// calculateFinalMemberships calculates final membership degrees and categorizes goals
func (s *FuzzyGPSolver) calculateFinalMemberships(result *FuzzyResult, solution []float64) {
	totalWeightedMembership := 0.0
	totalWeight := 0.0

	for _, goal := range s.goals {
		value := solution[goal.VariableIdx]
		membership := goal.MembershipFunc.EvaluateMembership(value)

		result.GoalMemberships[goal.ID] = membership
		result.GoalValues[goal.ID] = value
		result.TotalMembership += membership

		weightedMembership := membership * goal.Weight
		totalWeightedMembership += weightedMembership
		totalWeight += goal.Weight

		// Categorize goals
		if membership >= 0.8 {
			result.AchievedGoals = append(result.AchievedGoals, goal.Description)
		} else if membership >= 0.3 {
			result.PartialGoals = append(result.PartialGoals, goal.Description)
		} else {
			result.UnachievedGoals = append(result.UnachievedGoals, goal.Description)
		}
	}

	if len(s.goals) > 0 {
		result.AverageMembership = result.TotalMembership / float64(len(s.goals))
	}

	if totalWeight > 0 {
		result.WeightedMembership = totalWeightedMembership / totalWeight
	}
}

// BuildFuzzyGPFromConstraintModel creates a Fuzzy GP solver from constraint model
func BuildFuzzyGPFromConstraintModel(model *domain.ConstraintModel, params domain.ScenarioParameters) *FuzzyGPSolver {
	solver := NewFuzzyGPSolver(model.TotalIncome)

	// Priority 1: Mandatory expenses (trapezoidal membership - must achieve exactly)
	for id, constraint := range model.MandatoryExpenses {
		idx := solver.AddVariable(FuzzyVariable{
			ID:       id,
			Name:     fmt.Sprintf("mandatory_%s", id.String()[:8]),
			Type:     "category",
			MinValue: constraint.Minimum,
			MaxValue: constraint.Maximum,
		})

		// Trapezoidal membership: full satisfaction at exact amount
		target := constraint.Minimum
		tolerance := target * 0.05 // 5% tolerance
		solver.AddGoal(FuzzyGoal{
			ID:          fmt.Sprintf("mandatory_%s", id.String()[:8]),
			Description: "Mandatory expense",
			VariableIdx: idx,
			Priority:    1,
			Weight:      10.0, // Very high weight
			TargetValue: target,
			MembershipFunc: MembershipFunction{
				Type:      "trapezoidal",
				Lower:     target - tolerance*2,
				PeakLeft:  target - tolerance,
				PeakRight: target + tolerance,
				Upper:     target + tolerance*2,
			},
		})
	}

	// Priority 2: Debt payments (linear membership - simple linear utility)
	for id, constraint := range model.DebtPayments {
		minValue := constraint.MinimumPayment
		maxValue := constraint.CurrentBalance

		if constraint.FixedPayment > 0 {
			// Force the user's desired payment amount
			minValue = constraint.FixedPayment
			maxValue = constraint.FixedPayment
		}

		idx := solver.AddVariable(FuzzyVariable{
			ID:       id,
			Name:     constraint.DebtName,
			Type:     "debt",
			MinValue: minValue,
			MaxValue: maxValue,
		})

		minPayment := minValue
		maxPayment := maxValue

		solver.AddGoal(FuzzyGoal{
			ID:          fmt.Sprintf("debt_%s", id.String()[:8]),
			Description: fmt.Sprintf("Debt: %s", constraint.DebtName),
			VariableIdx: idx,
			Priority:    2,
			Weight:      constraint.InterestRate * 10, // Higher interest = higher weight
			TargetValue: maxPayment,                   // Target is max (pay off completely) or FixedPayment if forced
			MembershipFunc: MembershipFunction{
				Type:      "linear",
				Lower:     minPayment,
				Upper:     maxPayment,
				PeakLeft:  minPayment, // Not used for linear
				PeakRight: maxPayment, // Not used for linear
			},
		})
	}

	// Priority 3: Goals (triangular membership with satisfaction levels)
	for id, constraint := range model.GoalTargets {
		idx := solver.AddVariable(FuzzyVariable{
			ID:       id,
			Name:     constraint.GoalName,
			Type:     "goal",
			MinValue: 0,
			MaxValue: constraint.RemainingAmount * 1.5,
		})

		// Calculate base weight from priority
		var weight float64
		if constraint.PriorityWeight <= 1 {
			weight = 5.0 // Critical
		} else if constraint.PriorityWeight <= 10 {
			weight = 3.0 // High
		} else if constraint.PriorityWeight <= 20 {
			weight = 2.0 // Medium
		} else {
			weight = 1.0 // Low
		}

		// Emergency fund gets boost
		if constraint.GoalType == "emergency" {
			weight = 6.0
		}

		baseContribution := constraint.SuggestedContribution * params.GoalContributionFactor

		// S-curve-like membership (piecewise-linear, LP/MILP-friendly):
		// - slow increase up to ~30% of target
		// - fast increase in ~30%→70%
		// - saturation in ~70%→100%
		// - small extra utility in 100%→120%
		target := baseContribution
		peakLeft := target * 0.3
		peakRight := target * 0.7
		surplusMax := target * 1.2 // Allow up to 120% for surplus tier
		segments := buildSigmoidLikeSegments(0.0, peakLeft, peakRight, target)

		solver.AddGoal(FuzzyGoal{
			ID:          fmt.Sprintf("goal_%s", id.String()[:8]),
			Description: constraint.GoalName,
			VariableIdx: idx,
			Priority:    3,
			Weight:      weight,
			TargetValue: target,
			MembershipFunc: MembershipFunction{
				Type:      "piecewise",
				Lower:     0.0,
				PeakLeft:  peakLeft,
				PeakRight: peakRight,
				Upper:     surplusMax,
				Segments:  segments,
			},
		})

		// Update max value to allow surplus allocation
		if constraint.RemainingAmount*1.5 < surplusMax {
			// Update variable max to allow surplus
			for i := range solver.variables {
				if solver.variables[i].ID == id {
					solver.variables[i].MaxValue = surplusMax
					break
				}
			}
		}
	}

	// Priority 4: Flexible expenses (concave piecewise linear - diminishing marginal utility)
	// 500k đầu tiên: độ sướng 10/10, 500k tiếp: 7/10, 500k tiếp: 3/10
	for id, constraint := range model.FlexibleExpenses {
		minVal := constraint.Minimum
		maxVal := constraint.Maximum
		if maxVal == 0 {
			maxVal = minVal * 3.0 // Default to 3x min for flexible
		}

		// Calculate target (optimal spending level)
		target := minVal + (maxVal-minVal)*params.FlexibleSpendingLevel

		// Surplus tier: allow up to target * 1.2 (respect FlexibleSpendingLevel)
		// For Safe (level=0.0): target = minVal, surplusMax = minVal * 1.2 (very limited)
		// For Balanced (level=0.7): target cao hơn, surplusMax = target * 1.2 (more room)
		surplusMax := target * 1.2
		// But cap at maxVal * 1.2 to avoid exceeding reasonable bounds
		if surplusMax > maxVal*1.2 {
			surplusMax = maxVal * 1.2
		}

		// Calculate surplus max for variable bounds (respect FlexibleSpendingLevel)
		surplusMaxForVar := surplusMax

		idx := solver.AddVariable(FuzzyVariable{
			ID:       id,
			Name:     fmt.Sprintf("flexible_%s", id.String()[:8]),
			Type:     "category",
			MinValue: constraint.Minimum,
			MaxValue: surplusMaxForVar, // Allow surplus tier (up to target * 1.2, capped at maxVal * 1.2)
		})

		// Define 3 tiers with diminishing marginal utility (cumulative increase)
		// Tier 1: 0 → 1.0 (highest utility per unit)
		// Tier 2: 1.0 → 1.7 (moderate utility per unit)
		// Tier 3: 1.7 → 2.0 (low utility per unit)
		// Surplus tier: 2.0 → 2.1 (very low utility - diminishing returns)
		tier1Max := minVal + (target-minVal)*0.33 // First third: highest utility
		tier2Max := minVal + (target-minVal)*0.67 // Second third: medium utility

		// Create piecewise segments with cumulative increasing membership
		segments := []PiecewiseSegment{
			{
				Lower:         minVal,
				Upper:         tier1Max,
				Slope:         1.0 / (tier1Max - minVal), // Steep rise: 0 → 1.0
				Intercept:     0.0,
				MaxMembership: 1.0, // End of tier 1: membership = 1.0
			},
			{
				Lower:         tier1Max,
				Upper:         tier2Max,
				Slope:         0.7 / (tier2Max - tier1Max),              // Moderate rise: 1.0 → 1.7
				Intercept:     1.0 - tier1Max*(0.7/(tier2Max-tier1Max)), // Start from 1.0
				MaxMembership: 1.7,                                      // End of tier 2: membership = 1.7
			},
			{
				Lower:         tier2Max,
				Upper:         target,
				Slope:         0.3 / (target - tier2Max),              // Gentle rise: 1.7 → 2.0
				Intercept:     1.7 - tier2Max*(0.3/(target-tier2Max)), // Start from 1.7
				MaxMembership: 2.0,                                    // End of tier 3: membership = 2.0
			},
		}

		// Surplus tier: from target to max (or max*1.2) with diminishing returns
		// Membership: 2.0 → 2.1 (very small increase for going beyond target)
		if surplusMax > target {
			segments = append(segments, PiecewiseSegment{
				Lower:         target,
				Upper:         surplusMax,
				Slope:         0.1 / (surplusMax - target),            // Very gentle rise: 2.0 → 2.1
				Intercept:     2.0 - target*(0.1/(surplusMax-target)), // Start from 2.0
				MaxMembership: 2.1,                                    // End of surplus tier: membership = 2.1
			})
		}

		// Use constraint.Priority instead of hardcoded 4
		priority := constraint.Priority
		if priority == 0 {
			priority = 4 // Default fallback
		}

		// Calculate weight based on priority (similar to Goals)
		// Lower priority number = higher priority = higher weight
		var weight float64
		if priority <= 2 {
			weight = 3.0 // High priority
		} else if priority <= 4 {
			weight = 2.0 // Medium-high priority
		} else if priority <= 6 {
			weight = 1.5 // Medium priority
		} else {
			weight = 1.0 // Low priority
		}

		solver.AddGoal(FuzzyGoal{
			ID:          fmt.Sprintf("flexible_%s", id.String()[:8]),
			Description: "Flexible expense",
			VariableIdx: idx,
			Priority:    priority,
			Weight:      weight, // Weight based on priority (1.0-3.0)
			TargetValue: target,
			MembershipFunc: MembershipFunction{
				Type:      "piecewise",
				Lower:     minVal,
				Upper:     surplusMax, // Upper bound includes surplus tier
				PeakLeft:  minVal,     // Not used for piecewise
				PeakRight: target,     // Not used for piecewise
				Segments:  segments,
			},
		})
	}

	return solver
}
