package preemptive

import (
	"fmt"
	"personalfinancedss/internal/module/analytics/budget_allocation/domain"
	"sort"

	"github.com/google/uuid"
)

// PreemptiveGPSolver implements Preemptive Goal Programming
// It solves goals in priority order, fixing achieved goals before moving to next priority
type PreemptiveGPSolver struct {
	goals       []GPGoal
	constraints []GPConstraint
	variables   []GPVariable
	totalIncome float64
}

// GPVariable represents a decision variable in the GP model
type GPVariable struct {
	ID       uuid.UUID
	Name     string
	Type     string  // "category", "goal", "debt"
	MinValue float64 // Lower bound
	MaxValue float64 // Upper bound (0 = no upper bound)
}

// GPGoal represents a goal in the GP model
type GPGoal struct {
	ID          string
	Description string
	Priority    int // Lower = higher priority (1 is highest)
	TargetValue float64
	VariableIdx int     // Index of the variable this goal relates to
	GoalType    string  // "min_deviation", "at_least", "at_most", "exactly"
	Weight      float64 // Weight within same priority level
}

// GPConstraint represents a hard constraint
type GPConstraint struct {
	Coefficients []float64 // Coefficients for each variable
	RHS          float64   // Right-hand side
	Type         string    // "<=", ">=", "="
}

// GPResult contains the solution from the GP solver
type GPResult struct {
	VariableValues   map[uuid.UUID]float64
	GoalDeviations   map[string]float64 // Deviation from each goal
	AchievedGoals    []string
	UnachievedGoals  []string
	TotalDeviation   float64
	IsFeasible       bool
	SolverIterations int
}

// NewPreemptiveGPSolver creates a new solver
func NewPreemptiveGPSolver(totalIncome float64) *PreemptiveGPSolver {
	return &PreemptiveGPSolver{
		goals:       make([]GPGoal, 0),
		constraints: make([]GPConstraint, 0),
		variables:   make([]GPVariable, 0),
		totalIncome: totalIncome,
	}
}

// AddVariable adds a decision variable
func (s *PreemptiveGPSolver) AddVariable(v GPVariable) int {
	idx := len(s.variables)
	s.variables = append(s.variables, v)
	return idx
}

// AddGoal adds a goal to optimize
func (s *PreemptiveGPSolver) AddGoal(g GPGoal) {
	s.goals = append(s.goals, g)
}

// AddConstraint adds a hard constraint
func (s *PreemptiveGPSolver) AddConstraint(c GPConstraint) {
	s.constraints = append(s.constraints, c)
}

// Solve executes the preemptive goal programming algorithm
func (s *PreemptiveGPSolver) Solve() (*GPResult, error) {
	result := &GPResult{
		VariableValues:  make(map[uuid.UUID]float64),
		GoalDeviations:  make(map[string]float64),
		AchievedGoals:   make([]string, 0),
		UnachievedGoals: make([]string, 0),
		IsFeasible:      true,
	}

	if len(s.variables) == 0 {
		return result, nil
	}

	// Initialize solution with minimum values
	currentSolution := make([]float64, len(s.variables))
	for i, v := range s.variables {
		currentSolution[i] = v.MinValue
	}

	// Calculate initial remaining budget after minimums
	remainingBudget := s.totalIncome
	for _, v := range s.variables {
		remainingBudget -= v.MinValue
	}

	// Check if minimums exceed budget
	if remainingBudget < 0 {
		result.IsFeasible = false
		// Still allocate what we can
		remainingBudget = s.totalIncome
		for i := range currentSolution {
			currentSolution[i] = 0
		}
	}

	// Group goals by priority
	priorityGroups := s.groupGoalsByPriority()
	priorities := s.getSortedPriorities(priorityGroups)

	// Solve for each priority level (preemptive approach)
	for _, priority := range priorities {
		goals := priorityGroups[priority]

		// Sort goals within priority by weight (higher weight first)
		sort.Slice(goals, func(i, j int) bool {
			return goals[i].Weight > goals[j].Weight
		})

		// Allocate to each goal in this priority level
		for _, goal := range goals {
			if goal.VariableIdx >= len(currentSolution) {
				continue
			}

			v := s.variables[goal.VariableIdx]
			currentValue := currentSolution[goal.VariableIdx]
			targetValue := goal.TargetValue

			var allocation float64

			switch goal.GoalType {
			case "at_least", "exactly", "min_deviation":
				if currentValue < targetValue {
					needed := targetValue - currentValue

					// Respect max bound
					if v.MaxValue > 0 {
						maxAllowable := v.MaxValue - currentValue
						if needed > maxAllowable {
							needed = maxAllowable
						}
					}

					// Respect remaining budget
					if needed > remainingBudget {
						needed = remainingBudget
					}

					if needed > 0 {
						allocation = needed
						currentSolution[goal.VariableIdx] += allocation
						remainingBudget -= allocation
					}
				}
			case "at_most":
				// For at_most, we don't need to allocate more
				// Just ensure we don't exceed
				if currentValue > targetValue {
					excess := currentValue - targetValue
					currentSolution[goal.VariableIdx] -= excess
					remainingBudget += excess
				}
			}

			result.SolverIterations++
		}

		// Check which goals were achieved at this priority level
		for _, goal := range goals {
			deviation := s.calculateDeviation(goal, currentSolution)
			result.GoalDeviations[goal.ID] = deviation

			if deviation < 0.01 { // Small tolerance
				result.AchievedGoals = append(result.AchievedGoals, goal.ID)
			} else {
				result.UnachievedGoals = append(result.UnachievedGoals, goal.ID)
				result.TotalDeviation += deviation
			}
		}
	}

	// Store final solution
	for i, v := range s.variables {
		result.VariableValues[v.ID] = currentSolution[i]
	}

	return result, nil
}

// groupGoalsByPriority groups goals by their priority level
func (s *PreemptiveGPSolver) groupGoalsByPriority() map[int][]GPGoal {
	groups := make(map[int][]GPGoal)
	for _, goal := range s.goals {
		groups[goal.Priority] = append(groups[goal.Priority], goal)
	}
	return groups
}

// getSortedPriorities returns priorities in ascending order (1 first)
func (s *PreemptiveGPSolver) getSortedPriorities(groups map[int][]GPGoal) []int {
	priorities := make([]int, 0, len(groups))
	for p := range groups {
		priorities = append(priorities, p)
	}
	sort.Ints(priorities)
	return priorities
}

// calculateDeviation calculates how far the solution is from the goal
func (s *PreemptiveGPSolver) calculateDeviation(goal GPGoal, solution []float64) float64 {
	if goal.VariableIdx >= len(solution) {
		return 0
	}

	actual := solution[goal.VariableIdx]
	target := goal.TargetValue

	switch goal.GoalType {
	case "at_least":
		if actual < target {
			return target - actual
		}
		return 0
	case "at_most":
		if actual > target {
			return actual - target
		}
		return 0
	case "exactly", "min_deviation":
		diff := actual - target
		if diff < 0 {
			return -diff
		}
		return diff
	}
	return 0
}

// BuildPreemptiveGPFromConstraintModel creates a Preemptive GP solver from the constraint model
func BuildPreemptiveGPFromConstraintModel(model *domain.ConstraintModel, params domain.ScenarioParameters) *PreemptiveGPSolver {
	solver := NewPreemptiveGPSolver(model.TotalIncome)

	varIndexMap := make(map[uuid.UUID]int)

	// Priority 1: Mandatory expenses (must be satisfied)
	for id, constraint := range model.MandatoryExpenses {
		idx := solver.AddVariable(GPVariable{
			ID:       id,
			Name:     fmt.Sprintf("mandatory_%s", id.String()[:8]),
			Type:     "category",
			MinValue: constraint.Minimum,
			MaxValue: constraint.Maximum,
		})
		varIndexMap[id] = idx

		// Goal: allocate at least minimum
		solver.AddGoal(GPGoal{
			ID:          fmt.Sprintf("mandatory_%s", id.String()[:8]),
			Description: "Satisfy mandatory expense",
			Priority:    1,
			TargetValue: constraint.Minimum,
			VariableIdx: idx,
			GoalType:    "at_least",
			Weight:      1.0,
		})
	}

	// Priority 2: Minimum debt payments
	for id, constraint := range model.DebtPayments {
		idx := solver.AddVariable(GPVariable{
			ID:       id,
			Name:     constraint.DebtName,
			Type:     "debt",
			MinValue: constraint.MinimumPayment,
			MaxValue: constraint.CurrentBalance,
		})
		varIndexMap[id] = idx

		solver.AddGoal(GPGoal{
			ID:          fmt.Sprintf("debt_min_%s", id.String()[:8]),
			Description: fmt.Sprintf("Minimum payment for %s", constraint.DebtName),
			Priority:    2,
			TargetValue: constraint.MinimumPayment,
			VariableIdx: idx,
			GoalType:    "at_least",
			Weight:      1.0,
		})
	}

	// Priority 3: Emergency fund goals
	for id, constraint := range model.GoalTargets {
		if constraint.GoalType == "emergency" {
			idx := solver.AddVariable(GPVariable{
				ID:       id,
				Name:     constraint.GoalName,
				Type:     "goal",
				MinValue: 0,
				MaxValue: constraint.RemainingAmount,
			})
			varIndexMap[id] = idx

			targetContribution := constraint.SuggestedContribution * params.GoalContributionFactor
			solver.AddGoal(GPGoal{
				ID:          fmt.Sprintf("emergency_%s", id.String()[:8]),
				Description: fmt.Sprintf("Emergency fund: %s", constraint.GoalName),
				Priority:    3,
				TargetValue: targetContribution,
				VariableIdx: idx,
				GoalType:    "at_least",
				Weight:      1.0,
			})
		}
	}

	// Priority 4: High-priority goals
	for id, constraint := range model.GoalTargets {
		if constraint.GoalType != "emergency" && constraint.PriorityWeight <= 10 {
			idx := solver.AddVariable(GPVariable{
				ID:       id,
				Name:     constraint.GoalName,
				Type:     "goal",
				MinValue: 0,
				MaxValue: constraint.RemainingAmount,
			})
			varIndexMap[id] = idx

			targetContribution := constraint.SuggestedContribution * params.GoalContributionFactor
			solver.AddGoal(GPGoal{
				ID:          fmt.Sprintf("goal_high_%s", id.String()[:8]),
				Description: constraint.GoalName,
				Priority:    4,
				TargetValue: targetContribution,
				VariableIdx: idx,
				GoalType:    "at_least",
				Weight:      float64(11 - constraint.PriorityWeight),
			})
		}
	}

	// Priority 5: Extra debt payments (debt avalanche)
	for id, constraint := range model.DebtPayments {
		if idx, exists := varIndexMap[id]; exists {
			extraTarget := constraint.MinimumPayment * (1 + params.SurplusAllocation.DebtExtraPercent)
			solver.AddGoal(GPGoal{
				ID:          fmt.Sprintf("debt_extra_%s", id.String()[:8]),
				Description: fmt.Sprintf("Extra payment for %s", constraint.DebtName),
				Priority:    5,
				TargetValue: extraTarget,
				VariableIdx: idx,
				GoalType:    "at_least",
				Weight:      float64(100 - constraint.Priority),
			})
		}
	}

	// Priority 6: Medium/Low priority goals
	for id, constraint := range model.GoalTargets {
		if constraint.GoalType != "emergency" && constraint.PriorityWeight > 10 {
			idx := solver.AddVariable(GPVariable{
				ID:       id,
				Name:     constraint.GoalName,
				Type:     "goal",
				MinValue: 0,
				MaxValue: constraint.RemainingAmount,
			})
			varIndexMap[id] = idx

			targetContribution := constraint.SuggestedContribution * params.GoalContributionFactor
			solver.AddGoal(GPGoal{
				ID:          fmt.Sprintf("goal_low_%s", id.String()[:8]),
				Description: constraint.GoalName,
				Priority:    6,
				TargetValue: targetContribution,
				VariableIdx: idx,
				GoalType:    "at_least",
				Weight:      float64(100 - constraint.PriorityWeight),
			})
		}
	}

	// Priority 7: Flexible expenses
	for id, constraint := range model.FlexibleExpenses {
		idx := solver.AddVariable(GPVariable{
			ID:       id,
			Name:     fmt.Sprintf("flexible_%s", id.String()[:8]),
			Type:     "category",
			MinValue: constraint.Minimum,
			MaxValue: constraint.Maximum,
		})
		varIndexMap[id] = idx

		// Target based on spending level parameter
		target := constraint.Minimum
		if constraint.Maximum > 0 {
			target = constraint.Minimum + (constraint.Maximum-constraint.Minimum)*params.FlexibleSpendingLevel
		}

		solver.AddGoal(GPGoal{
			ID:          fmt.Sprintf("flexible_%s", id.String()[:8]),
			Description: "Flexible spending",
			Priority:    7,
			TargetValue: target,
			VariableIdx: idx,
			GoalType:    "at_least",
			Weight:      1.0,
		})
	}

	return solver
}
