package weighted

import "personalfinancedss/internal/module/analytics/budget_allocation/domain"

import (
	"fmt"
	"math"

	"github.com/google/uuid"
)

// WeightedGPSolver implements Weighted Goal Programming
// All goals are combined into a single objective function with weights
// Allows trade-offs between goals based on their weights
type WeightedGPSolver struct {
	goals       []WGPGoal
	variables   []WGPVariable
	totalIncome float64
}

// WGPVariable represents a decision variable
type WGPVariable struct {
	ID       uuid.UUID
	Name     string
	Type     string // "category", "goal", "debt"
	MinValue float64
	MaxValue float64
	Weight   float64 // Importance weight for this variable
}

// WGPGoal represents a goal with weight
type WGPGoal struct {
	ID          string
	Description string
	TargetValue float64
	VariableIdx int
	Weight      float64 // Weight in objective function (higher = more important)
	GoalType    string  // "at_least", "at_most", "exactly"
}

// WGPResult contains the solution
type WGPResult struct {
	VariableValues    map[uuid.UUID]float64
	GoalDeviations    map[string]float64
	WeightedDeviation float64 // Total weighted deviation
	AchievedGoals     []string
	PartialGoals      []string // Goals partially achieved
	UnachievedGoals   []string
	Iterations        int
}

// NewWeightedGPSolver creates a new solver
func NewWeightedGPSolver(totalIncome float64) *WeightedGPSolver {
	return &WeightedGPSolver{
		goals:       make([]WGPGoal, 0),
		variables:   make([]WGPVariable, 0),
		totalIncome: totalIncome,
	}
}

// AddVariable adds a decision variable
func (s *WeightedGPSolver) AddVariable(v WGPVariable) int {
	idx := len(s.variables)
	s.variables = append(s.variables, v)
	return idx
}

// AddGoal adds a goal
func (s *WeightedGPSolver) AddGoal(g WGPGoal) {
	s.goals = append(s.goals, g)
}

// Solve executes the Weighted GP algorithm
// Uses gradient-based allocation: allocate proportionally to weights
func (s *WeightedGPSolver) Solve() (*WGPResult, error) {
	result := &WGPResult{
		VariableValues:  make(map[uuid.UUID]float64),
		GoalDeviations:  make(map[string]float64),
		AchievedGoals:   make([]string, 0),
		PartialGoals:    make([]string, 0),
		UnachievedGoals: make([]string, 0),
	}

	if len(s.variables) == 0 {
		return result, nil
	}

	// Initialize solution
	solution := make([]float64, len(s.variables))
	for i, v := range s.variables {
		solution[i] = v.MinValue
	}

	// Calculate remaining budget after minimums
	remainingBudget := s.totalIncome
	for _, v := range s.variables {
		remainingBudget -= v.MinValue
	}

	if remainingBudget <= 0 {
		// No surplus to allocate
		s.storeSolution(result, solution)
		return result, nil
	}

	// Calculate total weight and needed amounts for each goal
	type goalNeed struct {
		goalIdx int
		varIdx  int
		needed  float64
		weight  float64
	}

	needs := make([]goalNeed, 0)
	totalWeightedNeed := 0.0

	for i, goal := range s.goals {
		if goal.VariableIdx >= len(solution) {
			continue
		}

		v := s.variables[goal.VariableIdx]
		currentValue := solution[goal.VariableIdx]
		needed := goal.TargetValue - currentValue

		if needed > 0 {
			// Respect max bound
			if v.MaxValue > 0 {
				maxAllowable := v.MaxValue - currentValue
				if needed > maxAllowable {
					needed = maxAllowable
				}
			}

			if needed > 0 {
				needs = append(needs, goalNeed{
					goalIdx: i,
					varIdx:  goal.VariableIdx,
					needed:  needed,
					weight:  goal.Weight,
				})
				totalWeightedNeed += needed * goal.Weight
			}
		}
	}

	// Iteratively allocate budget proportionally to weights
	maxIterations := 100
	for iter := 0; iter < maxIterations && remainingBudget > 0.01 && len(needs) > 0; iter++ {
		result.Iterations++

		// Recalculate total weighted need
		totalWeightedNeed = 0
		for _, n := range needs {
			totalWeightedNeed += n.needed * n.weight
		}

		if totalWeightedNeed <= 0 {
			break
		}

		// Allocate proportionally
		allocated := false
		newNeeds := make([]goalNeed, 0)

		for _, n := range needs {
			if n.needed <= 0 {
				continue
			}

			// Proportion based on weighted need
			proportion := (n.needed * n.weight) / totalWeightedNeed
			allocation := remainingBudget * proportion

			// Don't allocate more than needed
			if allocation > n.needed {
				allocation = n.needed
			}

			if allocation > 0.01 {
				solution[n.varIdx] += allocation
				remainingBudget -= allocation
				n.needed -= allocation
				allocated = true
			}

			if n.needed > 0.01 {
				newNeeds = append(newNeeds, n)
			}
		}

		needs = newNeeds

		if !allocated {
			break
		}
	}

	// Store solution and calculate deviations
	s.storeSolution(result, solution)

	return result, nil
}

// storeSolution stores the solution and calculates metrics
func (s *WeightedGPSolver) storeSolution(result *WGPResult, solution []float64) {
	// Store variable values
	for i, v := range s.variables {
		result.VariableValues[v.ID] = solution[i]
	}

	// Calculate deviations and classify goals
	for _, goal := range s.goals {
		if goal.VariableIdx >= len(solution) {
			continue
		}

		actual := solution[goal.VariableIdx]
		target := goal.TargetValue
		var deviation float64

		switch goal.GoalType {
		case "at_least":
			if actual < target {
				deviation = target - actual
			}
		case "at_most":
			if actual > target {
				deviation = actual - target
			}
		case "exactly":
			deviation = math.Abs(actual - target)
		}

		result.GoalDeviations[goal.ID] = deviation
		result.WeightedDeviation += deviation * goal.Weight

		// Classify goal achievement
		if deviation < 0.01 {
			result.AchievedGoals = append(result.AchievedGoals, goal.ID)
		} else if actual > 0 && deviation < target*0.5 {
			result.PartialGoals = append(result.PartialGoals, goal.ID)
		} else {
			result.UnachievedGoals = append(result.UnachievedGoals, goal.ID)
		}
	}
}

// BuildWeightedGPFromConstraintModel creates a Weighted GP solver from constraint model
func BuildWeightedGPFromConstraintModel(model *domain.ConstraintModel, params domain.ScenarioParameters) *WeightedGPSolver {
	solver := NewWeightedGPSolver(model.TotalIncome)

	// Calculate base weights from scenario parameters
	baseWeights := calculateBaseWeights(params)

	// Priority 1: Mandatory expenses (highest weight)
	for id, constraint := range model.MandatoryExpenses {
		idx := solver.AddVariable(WGPVariable{
			ID:       id,
			Name:     fmt.Sprintf("mandatory_%s", id.String()[:8]),
			Type:     "category",
			MinValue: constraint.Minimum,
			MaxValue: constraint.Maximum,
			Weight:   10.0, // Very high weight
		})

		solver.AddGoal(WGPGoal{
			ID:          fmt.Sprintf("mandatory_%s", id.String()[:8]),
			Description: "Mandatory expense",
			TargetValue: constraint.Minimum,
			VariableIdx: idx,
			Weight:      10.0,
			GoalType:    "at_least",
		})
	}

	// Priority 2: Minimum debt payments
	for id, constraint := range model.DebtPayments {
		idx := solver.AddVariable(WGPVariable{
			ID:       id,
			Name:     constraint.DebtName,
			Type:     "debt",
			MinValue: constraint.MinimumPayment,
			MaxValue: constraint.CurrentBalance,
			Weight:   8.0,
		})

		// Minimum payment goal
		solver.AddGoal(WGPGoal{
			ID:          fmt.Sprintf("debt_min_%s", id.String()[:8]),
			Description: fmt.Sprintf("Min payment: %s", constraint.DebtName),
			TargetValue: constraint.MinimumPayment,
			VariableIdx: idx,
			Weight:      8.0,
			GoalType:    "at_least",
		})

		// Extra payment goal (weighted by interest rate and scenario)
		extraTarget := constraint.MinimumPayment * (1 + baseWeights.DebtExtra)
		solver.AddGoal(WGPGoal{
			ID:          fmt.Sprintf("debt_extra_%s", id.String()[:8]),
			Description: fmt.Sprintf("Extra payment: %s", constraint.DebtName),
			TargetValue: extraTarget,
			VariableIdx: idx,
			Weight:      baseWeights.DebtExtra * (constraint.InterestRate / 10), // Higher interest = higher weight
			GoalType:    "at_least",
		})
	}

	// Goals (emergency fund, savings, etc.)
	for id, constraint := range model.GoalTargets {
		var weight float64
		if constraint.GoalType == "emergency" {
			weight = baseWeights.Emergency
		} else if constraint.PriorityWeight <= 10 {
			weight = baseWeights.HighPriorityGoal
		} else {
			weight = baseWeights.LowPriorityGoal
		}

		idx := solver.AddVariable(WGPVariable{
			ID:       id,
			Name:     constraint.GoalName,
			Type:     "goal",
			MinValue: 0,
			MaxValue: constraint.RemainingAmount,
			Weight:   weight,
		})

		targetContribution := constraint.SuggestedContribution * params.GoalContributionFactor
		solver.AddGoal(WGPGoal{
			ID:          fmt.Sprintf("goal_%s", id.String()[:8]),
			Description: constraint.GoalName,
			TargetValue: targetContribution,
			VariableIdx: idx,
			Weight:      weight,
			GoalType:    "at_least",
		})
	}

	// Flexible expenses
	for id, constraint := range model.FlexibleExpenses {
		idx := solver.AddVariable(WGPVariable{
			ID:       id,
			Name:     fmt.Sprintf("flexible_%s", id.String()[:8]),
			Type:     "category",
			MinValue: constraint.Minimum,
			MaxValue: constraint.Maximum,
			Weight:   baseWeights.Flexible,
		})

		target := constraint.Minimum
		if constraint.Maximum > 0 {
			target = constraint.Minimum + (constraint.Maximum-constraint.Minimum)*params.FlexibleSpendingLevel
		}

		solver.AddGoal(WGPGoal{
			ID:          fmt.Sprintf("flexible_%s", id.String()[:8]),
			Description: "Flexible spending",
			TargetValue: target,
			VariableIdx: idx,
			Weight:      baseWeights.Flexible,
			GoalType:    "at_least",
		})
	}

	return solver
}

// BaseWeights contains weight multipliers for different goal types
type BaseWeights struct {
	Emergency        float64
	DebtExtra        float64
	HighPriorityGoal float64
	LowPriorityGoal  float64
	Flexible         float64
}

// calculateBaseWeights calculates weights based on scenario parameters
func calculateBaseWeights(params domain.ScenarioParameters) BaseWeights {
	sa := params.SurplusAllocation

	return BaseWeights{
		Emergency:        5.0 * sa.EmergencyFundPercent * 10, // Scale to reasonable range
		DebtExtra:        4.0 * sa.DebtExtraPercent * 10,
		HighPriorityGoal: 3.0 * sa.GoalsPercent * 10,
		LowPriorityGoal:  1.5 * sa.GoalsPercent * 10,
		Flexible:         1.0 * sa.FlexiblePercent * 10,
	}
}
