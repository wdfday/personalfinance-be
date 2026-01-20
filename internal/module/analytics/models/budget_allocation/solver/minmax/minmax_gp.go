package minmax

import "personalfinancedss/internal/module/analytics/budget_allocation/domain"

import (
	"fmt"
	"math"

	"github.com/google/uuid"
)

// MinmaxGPSolver implements Minmax (Chebyshev) Goal Programming
// Minimizes the maximum weighted deviation across all goals
// This ensures no single goal is disproportionately under-achieved
type MinmaxGPSolver struct {
	goals       []MinmaxGoal
	variables   []MinmaxVariable
	totalIncome float64
}

// MinmaxVariable represents a decision variable
type MinmaxVariable struct {
	ID       uuid.UUID
	Name     string
	Type     string // "category", "goal", "debt"
	MinValue float64
	MaxValue float64
}

// MinmaxGoal represents a goal with target
type MinmaxGoal struct {
	ID          string
	Description string
	TargetValue float64
	VariableIdx int
	Weight      float64 // Weight for normalization (higher = more important)
	GoalType    string  // "at_least", "at_most", "exactly"
}

// MinmaxResult contains the solution
type MinmaxResult struct {
	VariableValues   map[uuid.UUID]float64
	GoalDeviations   map[string]float64
	GoalAchievements map[string]float64 // Percentage achieved (0-100)
	MaxDeviation     float64            // The minimized maximum deviation
	MinAchievement   float64            // Minimum achievement percentage
	AchievedGoals    []string
	PartialGoals     []string
	UnachievedGoals  []string
	Iterations       int
	IsBalanced       bool // True if all goals have similar achievement levels
}

// NewMinmaxGPSolver creates a new solver
func NewMinmaxGPSolver(totalIncome float64) *MinmaxGPSolver {
	return &MinmaxGPSolver{
		goals:       make([]MinmaxGoal, 0),
		variables:   make([]MinmaxVariable, 0),
		totalIncome: totalIncome,
	}
}

// AddVariable adds a decision variable
func (s *MinmaxGPSolver) AddVariable(v MinmaxVariable) int {
	idx := len(s.variables)
	s.variables = append(s.variables, v)
	return idx
}

// AddGoal adds a goal
func (s *MinmaxGPSolver) AddGoal(g MinmaxGoal) {
	s.goals = append(s.goals, g)
}

// Solve executes the Minmax GP algorithm
// Uses iterative leveling: allocate to bring all goals to same achievement %
func (s *MinmaxGPSolver) Solve() (*MinmaxResult, error) {
	result := &MinmaxResult{
		VariableValues:   make(map[uuid.UUID]float64),
		GoalDeviations:   make(map[string]float64),
		GoalAchievements: make(map[string]float64),
		AchievedGoals:    make([]string, 0),
		PartialGoals:     make([]string, 0),
		UnachievedGoals:  make([]string, 0),
	}

	if len(s.variables) == 0 {
		return result, nil
	}

	// Initialize solution with minimum values
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
		s.storeSolution(result, solution)
		return result, nil
	}

	// Build goal info for leveling algorithm
	type goalInfo struct {
		goalIdx     int
		varIdx      int
		target      float64
		current     float64
		needed      float64
		weight      float64
		achievement float64
	}

	// Iterative leveling: bring all goals to same achievement level
	maxIterations := 200
	for iter := 0; iter < maxIterations && remainingBudget > 0.01; iter++ {
		result.Iterations++

		// Calculate current achievement for each goal
		goals := make([]goalInfo, 0)
		for i, goal := range s.goals {
			if goal.VariableIdx >= len(solution) {
				continue
			}

			current := solution[goal.VariableIdx]
			target := goal.TargetValue
			needed := target - current

			// Respect max bound
			v := s.variables[goal.VariableIdx]
			if v.MaxValue > 0 {
				maxAllowable := v.MaxValue - current
				if needed > maxAllowable {
					needed = maxAllowable
				}
			}

			if needed <= 0 {
				continue // Goal already achieved
			}

			achievement := 0.0
			if target > 0 {
				achievement = (current / target) * 100
			}

			goals = append(goals, goalInfo{
				goalIdx:     i,
				varIdx:      goal.VariableIdx,
				target:      target,
				current:     current,
				needed:      needed,
				weight:      goal.Weight,
				achievement: achievement,
			})
		}

		if len(goals) == 0 {
			break // All goals achieved
		}

		// Find the goal with minimum achievement (most behind)
		minAchievement := math.MaxFloat64
		for _, g := range goals {
			if g.achievement < minAchievement {
				minAchievement = g.achievement
			}
		}

		// Find target achievement level (next level up)
		// We want to bring all goals at minAchievement up to the next level
		nextLevel := minAchievement + 5.0 // Increase by 5% increments
		if nextLevel > 100 {
			nextLevel = 100
		}

		// Calculate how much budget needed to bring all min-achievement goals to next level
		totalNeeded := 0.0
		goalsToLevel := make([]goalInfo, 0)

		for _, g := range goals {
			if g.achievement <= minAchievement+0.01 { // Goals at minimum level
				// How much to reach next level
				targetValue := (nextLevel / 100) * g.target
				amountNeeded := targetValue - g.current
				if amountNeeded > g.needed {
					amountNeeded = g.needed
				}
				if amountNeeded > 0 {
					g.needed = amountNeeded
					goalsToLevel = append(goalsToLevel, g)
					totalNeeded += amountNeeded
				}
			}
		}

		if len(goalsToLevel) == 0 || totalNeeded <= 0 {
			break
		}

		// Allocate proportionally if not enough budget
		allocRatio := 1.0
		if totalNeeded > remainingBudget {
			allocRatio = remainingBudget / totalNeeded
		}

		for _, g := range goalsToLevel {
			allocation := g.needed * allocRatio
			if allocation > 0.01 {
				solution[g.varIdx] += allocation
				remainingBudget -= allocation
			}
		}
	}

	s.storeSolution(result, solution)
	return result, nil
}

// storeSolution stores the solution and calculates metrics
func (s *MinmaxGPSolver) storeSolution(result *MinmaxResult, solution []float64) {
	// Store variable values
	for i, v := range s.variables {
		result.VariableValues[v.ID] = solution[i]
	}

	// Calculate deviations and achievements
	minAchievement := 100.0
	maxDeviation := 0.0

	for _, goal := range s.goals {
		if goal.VariableIdx >= len(solution) {
			continue
		}

		actual := solution[goal.VariableIdx]
		target := goal.TargetValue

		// Calculate deviation
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

		// Normalize deviation by weight
		normalizedDev := deviation
		if goal.Weight > 0 {
			normalizedDev = deviation / goal.Weight
		}

		result.GoalDeviations[goal.ID] = deviation

		if normalizedDev > maxDeviation {
			maxDeviation = normalizedDev
		}

		// Calculate achievement percentage
		achievement := 100.0
		if target > 0 {
			achievement = (actual / target) * 100
			if achievement > 100 {
				achievement = 100
			}
		}
		result.GoalAchievements[goal.ID] = achievement

		if achievement < minAchievement {
			minAchievement = achievement
		}

		// Classify goal
		if deviation < 0.01 {
			result.AchievedGoals = append(result.AchievedGoals, goal.ID)
		} else if achievement >= 50 {
			result.PartialGoals = append(result.PartialGoals, goal.ID)
		} else {
			result.UnachievedGoals = append(result.UnachievedGoals, goal.ID)
		}
	}

	result.MaxDeviation = maxDeviation
	result.MinAchievement = minAchievement

	// Check if solution is balanced (all achievements within 20% of each other)
	if len(result.GoalAchievements) > 1 {
		maxAch := 0.0
		for _, ach := range result.GoalAchievements {
			if ach > maxAch {
				maxAch = ach
			}
		}
		result.IsBalanced = (maxAch - minAchievement) <= 20
	} else {
		result.IsBalanced = true
	}
}

// BuildMinmaxGPFromConstraintModel creates a Minmax GP solver from constraint model
func BuildMinmaxGPFromConstraintModel(model *domain.ConstraintModel, params domain.ScenarioParameters) *MinmaxGPSolver {
	solver := NewMinmaxGPSolver(model.TotalIncome)

	// Priority 1: Mandatory expenses (highest weight for normalization)
	for id, constraint := range model.MandatoryExpenses {
		idx := solver.AddVariable(MinmaxVariable{
			ID:       id,
			Name:     fmt.Sprintf("mandatory_%s", id.String()[:8]),
			Type:     "category",
			MinValue: constraint.Minimum,
			MaxValue: constraint.Maximum,
		})

		solver.AddGoal(MinmaxGoal{
			ID:          fmt.Sprintf("mandatory_%s", id.String()[:8]),
			Description: "Mandatory expense",
			TargetValue: constraint.Minimum,
			VariableIdx: idx,
			Weight:      constraint.Minimum, // Use target as weight for normalization
			GoalType:    "at_least",
		})
	}

	// Priority 2: Minimum debt payments
	for id, constraint := range model.DebtPayments {
		idx := solver.AddVariable(MinmaxVariable{
			ID:       id,
			Name:     constraint.DebtName,
			Type:     "debt",
			MinValue: constraint.MinimumPayment,
			MaxValue: constraint.CurrentBalance,
		})

		// Minimum payment goal
		solver.AddGoal(MinmaxGoal{
			ID:          fmt.Sprintf("debt_min_%s", id.String()[:8]),
			Description: fmt.Sprintf("Min payment: %s", constraint.DebtName),
			TargetValue: constraint.MinimumPayment,
			VariableIdx: idx,
			Weight:      constraint.MinimumPayment,
			GoalType:    "at_least",
		})

		// Extra payment goal
		extraTarget := constraint.MinimumPayment * (1 + params.SurplusAllocation.DebtExtraPercent)
		solver.AddGoal(MinmaxGoal{
			ID:          fmt.Sprintf("debt_extra_%s", id.String()[:8]),
			Description: fmt.Sprintf("Extra payment: %s", constraint.DebtName),
			TargetValue: extraTarget,
			VariableIdx: idx,
			Weight:      extraTarget,
			GoalType:    "at_least",
		})
	}

	// Goals (emergency fund, savings, etc.)
	for id, constraint := range model.GoalTargets {
		idx := solver.AddVariable(MinmaxVariable{
			ID:       id,
			Name:     constraint.GoalName,
			Type:     "goal",
			MinValue: 0,
			MaxValue: constraint.RemainingAmount,
		})

		targetContribution := constraint.SuggestedContribution * params.GoalContributionFactor
		solver.AddGoal(MinmaxGoal{
			ID:          fmt.Sprintf("goal_%s", id.String()[:8]),
			Description: constraint.GoalName,
			TargetValue: targetContribution,
			VariableIdx: idx,
			Weight:      targetContribution,
			GoalType:    "at_least",
		})
	}

	// Flexible expenses
	for id, constraint := range model.FlexibleExpenses {
		idx := solver.AddVariable(MinmaxVariable{
			ID:       id,
			Name:     fmt.Sprintf("flexible_%s", id.String()[:8]),
			Type:     "category",
			MinValue: constraint.Minimum,
			MaxValue: constraint.Maximum,
		})

		target := constraint.Minimum
		if constraint.Maximum > 0 {
			target = constraint.Minimum + (constraint.Maximum-constraint.Minimum)*params.FlexibleSpendingLevel
		}

		solver.AddGoal(MinmaxGoal{
			ID:          fmt.Sprintf("flexible_%s", id.String()[:8]),
			Description: "Flexible spending",
			TargetValue: target,
			VariableIdx: idx,
			Weight:      target,
			GoalType:    "at_least",
		})
	}

	return solver
}
