package meta

import (
	"fmt"
	"personalfinancedss/internal/module/analytics/budget_allocation/domain"
	"sort"

	"github.com/google/uuid"
)

// MetaGPSolver implements Meta Goal Programming (Multi-Choice GP)
// Each goal can have multiple target levels: minimum, satisfactory, ideal
// The solver tries to achieve the highest possible level for each goal
type MetaGPSolver struct {
	goals       []MetaGoal
	variables   []MetaVariable
	totalIncome float64
}

// MetaVariable represents a decision variable
type MetaVariable struct {
	ID       uuid.UUID
	Name     string
	Type     string // "category", "goal", "debt"
	MinValue float64
	MaxValue float64
}

// TargetLevel represents a target level for a goal
type TargetLevel struct {
	Level       string // "minimum", "satisfactory", "ideal"
	Value       float64
	Reward      float64 // Reward/utility for achieving this level
	Description string
}

// MetaGoal represents a goal with multiple target levels
type MetaGoal struct {
	ID           string
	Description  string
	VariableIdx  int
	TargetLevels []TargetLevel // Ordered from minimum to ideal
	Priority     int           // Lower = higher priority
	Weight       float64
}

// MetaResult contains the solution
type MetaResult struct {
	VariableValues    map[uuid.UUID]float64
	GoalLevels        map[string]string  // Goal ID -> achieved level
	GoalValues        map[string]float64 // Goal ID -> actual value
	TotalReward       float64            // Total utility achieved
	MaxPossibleReward float64            // Maximum possible utility
	RewardRatio       float64            // TotalReward / MaxPossibleReward
	AchievedGoals     []string           // Goals that reached at least minimum
	IdealGoals        []string           // Goals that reached ideal level
	Iterations        int
}

// NewMetaGPSolver creates a new solver
func NewMetaGPSolver(totalIncome float64) *MetaGPSolver {
	return &MetaGPSolver{
		goals:       make([]MetaGoal, 0),
		variables:   make([]MetaVariable, 0),
		totalIncome: totalIncome,
	}
}

// AddVariable adds a decision variable
func (s *MetaGPSolver) AddVariable(v MetaVariable) int {
	idx := len(s.variables)
	s.variables = append(s.variables, v)
	return idx
}

// AddGoal adds a goal with multiple target levels
func (s *MetaGPSolver) AddGoal(g MetaGoal) {
	// Sort target levels by value (ascending)
	sort.Slice(g.TargetLevels, func(i, j int) bool {
		return g.TargetLevels[i].Value < g.TargetLevels[j].Value
	})
	s.goals = append(s.goals, g)
}

// Solve executes the Meta GP algorithm
// Strategy: Tries to use Golp (MILP) first, falls back to Greedy algorithm if Golp is unavailable
func (s *MetaGPSolver) Solve() (*MetaResult, error) {
	// Try solving with Golp (MILP)
	// We check if Golp is available by trying to create a dummy solver or just calling the method
	// SolveWithGolp handles the check internally (calling CreateGolpSolver)
	result, err := s.SolveWithGolp()
	if err == nil {
		return result, nil
	}
	// If error is "not available", we fallback. Otherwise we might want to return valid error?
	// For now, let's log/print and fallback only if it's an availability issue.
	// But since we want to support non-CGO builds smoothly, we just fallback on any error from Golp setup.
	// fmt.Printf("Golp solver failed or unavailable: %v. Falling back to Greedy.\n", err)

	result = &MetaResult{
		VariableValues: make(map[uuid.UUID]float64),
		GoalLevels:     make(map[string]string),
		GoalValues:     make(map[string]float64),
		AchievedGoals:  make([]string, 0),
		IdealGoals:     make([]string, 0),
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

	// Calculate max possible reward
	for _, goal := range s.goals {
		if len(goal.TargetLevels) > 0 {
			result.MaxPossibleReward += goal.TargetLevels[len(goal.TargetLevels)-1].Reward
		}
	}

	if remainingBudget <= 0 {
		s.storeSolution(result, solution)
		return result, nil
	}

	// Group goals by priority
	priorityGroups := s.groupGoalsByPriority()
	priorities := s.getSortedPriorities(priorityGroups)

	// Process each priority level
	for _, priority := range priorities {
		goals := priorityGroups[priority]

		// For each goal in this priority, try to achieve highest possible level
		for _, goal := range goals {
			if goal.VariableIdx >= len(solution) {
				continue
			}

			result.Iterations++

			v := s.variables[goal.VariableIdx]
			currentValue := solution[goal.VariableIdx]

			// Try to achieve each level from highest to lowest
			for i := len(goal.TargetLevels) - 1; i >= 0; i-- {
				level := goal.TargetLevels[i]
				needed := level.Value - currentValue

				if needed <= 0 {
					// Already achieved this level
					continue
				}

				// Respect max bound
				if v.MaxValue > 0 {
					maxAllowable := v.MaxValue - currentValue
					if needed > maxAllowable {
						needed = maxAllowable
					}
				}

				// Check if we can afford this level
				if needed <= remainingBudget {
					solution[goal.VariableIdx] = level.Value
					remainingBudget -= needed
					break // Move to next goal
				}
			}
		}

		// Second pass: allocate remaining budget proportionally within priority
		if remainingBudget > 0.01 {
			s.allocateRemainingWithinPriority(goals, solution, &remainingBudget)
		}
	}

	s.storeSolution(result, solution)
	return result, nil
}

// allocateRemainingWithinPriority allocates remaining budget to partially achieve higher levels
func (s *MetaGPSolver) allocateRemainingWithinPriority(goals []MetaGoal, solution []float64, remainingBudget *float64) {
	type upgradeOption struct {
		goalIdx     int
		varIdx      int
		levelIdx    int
		needed      float64
		reward      float64
		rewardRatio float64 // reward per dollar
	}

	options := make([]upgradeOption, 0)

	for _, goal := range goals {
		if goal.VariableIdx >= len(solution) {
			continue
		}

		currentValue := solution[goal.VariableIdx]
		v := s.variables[goal.VariableIdx]

		// Find current level and next level
		currentLevelIdx := -1
		for i, level := range goal.TargetLevels {
			if currentValue >= level.Value-0.01 {
				currentLevelIdx = i
			}
		}

		// Check if there's a higher level to achieve
		nextLevelIdx := currentLevelIdx + 1
		if nextLevelIdx < len(goal.TargetLevels) {
			nextLevel := goal.TargetLevels[nextLevelIdx]
			needed := nextLevel.Value - currentValue

			// Respect max bound
			if v.MaxValue > 0 {
				maxAllowable := v.MaxValue - currentValue
				if needed > maxAllowable {
					continue
				}
			}

			currentReward := 0.0
			if currentLevelIdx >= 0 {
				currentReward = goal.TargetLevels[currentLevelIdx].Reward
			}
			additionalReward := nextLevel.Reward - currentReward

			if needed > 0 && additionalReward > 0 {
				options = append(options, upgradeOption{
					goalIdx:     0, // Not used
					varIdx:      goal.VariableIdx,
					levelIdx:    nextLevelIdx,
					needed:      needed,
					reward:      additionalReward,
					rewardRatio: additionalReward / needed,
				})
			}
		}
	}

	// Sort by reward ratio (highest first)
	sort.Slice(options, func(i, j int) bool {
		return options[i].rewardRatio > options[j].rewardRatio
	})

	// Allocate to best options
	for _, opt := range options {
		if *remainingBudget < 0.01 {
			break
		}

		if opt.needed <= *remainingBudget {
			solution[opt.varIdx] += opt.needed
			*remainingBudget -= opt.needed
		}
	}
}

// groupGoalsByPriority groups goals by their priority level
func (s *MetaGPSolver) groupGoalsByPriority() map[int][]MetaGoal {
	groups := make(map[int][]MetaGoal)
	for _, goal := range s.goals {
		groups[goal.Priority] = append(groups[goal.Priority], goal)
	}
	return groups
}

// getSortedPriorities returns priorities in ascending order
func (s *MetaGPSolver) getSortedPriorities(groups map[int][]MetaGoal) []int {
	priorities := make([]int, 0, len(groups))
	for p := range groups {
		priorities = append(priorities, p)
	}
	sort.Ints(priorities)
	return priorities
}

// storeSolution stores the solution and calculates metrics
func (s *MetaGPSolver) storeSolution(result *MetaResult, solution []float64) {
	// Store variable values
	for i, v := range s.variables {
		result.VariableValues[v.ID] = solution[i]
	}

	// Determine achieved level for each goal
	for _, goal := range s.goals {
		if goal.VariableIdx >= len(solution) {
			continue
		}

		actual := solution[goal.VariableIdx]
		result.GoalValues[goal.ID] = actual

		achievedLevel := "none"
		achievedReward := 0.0

		for _, level := range goal.TargetLevels {
			if actual >= level.Value-0.01 {
				achievedLevel = level.Level
				achievedReward = level.Reward
			}
		}

		result.GoalLevels[goal.ID] = achievedLevel
		result.TotalReward += achievedReward

		if achievedLevel != "none" {
			result.AchievedGoals = append(result.AchievedGoals, goal.ID)
		}

		// Check if ideal level achieved
		if len(goal.TargetLevels) > 0 {
			idealLevel := goal.TargetLevels[len(goal.TargetLevels)-1]
			if actual >= idealLevel.Value-0.01 {
				result.IdealGoals = append(result.IdealGoals, goal.ID)
			}
		}
	}

	// Calculate reward ratio
	if result.MaxPossibleReward > 0 {
		result.RewardRatio = result.TotalReward / result.MaxPossibleReward
	}
}

// BuildMetaGPFromConstraintModel creates a Meta GP solver from constraint model
func BuildMetaGPFromConstraintModel(model *domain.ConstraintModel, params domain.ScenarioParameters) *MetaGPSolver {
	solver := NewMetaGPSolver(model.TotalIncome)

	// Priority 1: Mandatory expenses (single level - must achieve)
	for id, constraint := range model.MandatoryExpenses {
		idx := solver.AddVariable(MetaVariable{
			ID:       id,
			Name:     fmt.Sprintf("mandatory_%s", id.String()[:8]),
			Type:     "category",
			MinValue: constraint.Minimum,
			MaxValue: constraint.Maximum,
		})

		solver.AddGoal(MetaGoal{
			ID:          fmt.Sprintf("mandatory_%s", id.String()[:8]),
			Description: "Mandatory expense",
			VariableIdx: idx,
			Priority:    1,
			Weight:      1.0,
			TargetLevels: []TargetLevel{
				{Level: "minimum", Value: constraint.Minimum, Reward: 100, Description: "Required amount"},
			},
		})
	}

	// Priority 2: Debt payments (multiple levels)
	for id, constraint := range model.DebtPayments {
		idx := solver.AddVariable(MetaVariable{
			ID:       id,
			Name:     constraint.DebtName,
			Type:     "debt",
			MinValue: constraint.MinimumPayment,
			MaxValue: constraint.CurrentBalance,
		})

		// Calculate extra payment targets
		extraPercent := params.SurplusAllocation.DebtExtraPercent
		satisfactoryPayment := constraint.MinimumPayment * (1 + extraPercent*0.5)
		idealPayment := constraint.MinimumPayment * (1 + extraPercent)

		solver.AddGoal(MetaGoal{
			ID:          fmt.Sprintf("debt_%s", id.String()[:8]),
			Description: fmt.Sprintf("Debt: %s", constraint.DebtName),
			VariableIdx: idx,
			Priority:    2,
			Weight:      constraint.InterestRate / 10, // Higher interest = higher weight
			TargetLevels: []TargetLevel{
				{Level: "minimum", Value: constraint.MinimumPayment, Reward: 50, Description: "Minimum payment"},
				{Level: "satisfactory", Value: satisfactoryPayment, Reward: 75, Description: "Some extra payment"},
				{Level: "ideal", Value: idealPayment, Reward: 100, Description: "Full extra payment"},
			},
		})
	}

	// Unified Goal Handling (combining all priorities)
	// User request: Satisfaction levels 0%, 40%, 80%, 120%
	// Penalty/Reward = AHP * Level Score (1, 2, 3, 4)
	for id, constraint := range model.GoalTargets {
		if constraint.GoalType == "emergency" {
			// Keep emergency slightly separate ortreat as Critical goal?
			// Let's treat emergency as high priority within the unified logic
		}

		// Calculate Base Weight derived from Priority (simulating AHP weight)
		// PriorityWeight: 1 (Critical) -> 30 (Low)
		// We invert this to get a utility weight:
		var ahpProxyWeight float64
		if constraint.PriorityWeight <= 1 {
			ahpProxyWeight = 5.0 // Critical
		} else if constraint.PriorityWeight <= 10 {
			ahpProxyWeight = 3.0 // High
		} else if constraint.PriorityWeight <= 20 {
			ahpProxyWeight = 2.0 // Medium
		} else {
			ahpProxyWeight = 1.0 // Low
		}

		// Emergency fund gets a boost
		if constraint.GoalType == "emergency" {
			ahpProxyWeight = 6.0
		}

		idx := solver.AddVariable(MetaVariable{
			ID:       id,
			Name:     constraint.GoalName,
			Type:     "goal",
			MinValue: 0,
			MaxValue: constraint.RemainingAmount * 1.5, // Allow up to 150% to accommodate the 120% target
		})

		baseContribution := constraint.SuggestedContribution * params.GoalContributionFactor

		// Define Levels: 40%, 80%, 100%, 120% (User asked for 120%, let's use 100% as standard completion and 120% as over-achievement)
		// Note: User asked 0 40 80 120. 0 is implicit start.

		solver.AddGoal(MetaGoal{
			ID:          fmt.Sprintf("goal_%s", id.String()[:8]),
			Description: constraint.GoalName,
			VariableIdx: idx,
			Priority:    3, // Goals are Priority 3 now (after Mandatory and Debt)
			Weight:      ahpProxyWeight,
			TargetLevels: []TargetLevel{
				{Level: "level_0_0", Value: 0, Reward: 0, Description: "0% Contribution"},
				{Level: "level_1_40", Value: baseContribution * 0.4, Reward: ahpProxyWeight * 1.0, Description: "40% Satisfaction"},
				{Level: "level_2_80", Value: baseContribution * 0.8, Reward: ahpProxyWeight * 2.0, Description: "80% Satisfaction"},
				{Level: "level_3_100", Value: baseContribution * 1.0, Reward: ahpProxyWeight * 3.0, Description: "100% Satisfaction"},
				{Level: "level_4_120", Value: baseContribution * 1.2, Reward: ahpProxyWeight * 4.0, Description: "120% Satisfaction"},
			},
		})
	}

	// Unified Priority 3: Flexible expenses (Competing with Goals)
	// User comment: "budget cần nới ra" -> Treat lifestyle needs as important as goals
	for id, constraint := range model.FlexibleExpenses {
		idx := solver.AddVariable(MetaVariable{
			ID:       id,
			Name:     fmt.Sprintf("flexible_%s", id.String()[:8]),
			Type:     "category",
			MinValue: constraint.Minimum,
			MaxValue: constraint.Maximum * 1.2, // Allow flexible max
		})

		rangeAmount := constraint.Maximum - constraint.Minimum
		if rangeAmount <= 0 {
			rangeAmount = constraint.Minimum * 0.5
		}

		// Proxy weight for flexible expenses
		// To compete with Goals (Weight 1.0-5.0), Flexible expenses need comparable weights.
		// If Priority is 1 (High Importance), give it Weight 4.0
		flexWeight := 2.0
		if constraint.Priority == 1 {
			flexWeight = 4.0
		}

		// Mốc: 0% surplus, 40% surplus, 80% surplus, 120% surplus range
		solver.AddGoal(MetaGoal{
			ID:          fmt.Sprintf("flexible_%s", id.String()[:8]),
			Description: "Flexible spending",
			VariableIdx: idx,
			Priority:    3, // Same priority as Goals to allow trade-offs
			Weight:      flexWeight,
			TargetLevels: []TargetLevel{
				{Level: "level_1", Value: constraint.Minimum + rangeAmount*0.4, Reward: flexWeight * 1.0, Description: "Basic Comfort"},
				{Level: "level_2", Value: constraint.Minimum + rangeAmount*0.8, Reward: flexWeight * 2.0, Description: "Good Comfort"},
				{Level: "level_3", Value: constraint.Maximum, Reward: flexWeight * 3.0, Description: "Max Comfort"},
				{Level: "level_4", Value: constraint.Maximum * 1.2, Reward: flexWeight * 4.0, Description: "Luxury"},
			},
		})
	}

	return solver
}
