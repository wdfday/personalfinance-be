package meta

import (
	"fmt"
	"math"
	"personalfinancedss/internal/module/analytics/models/budget_allocation/solver/lp"

	"github.com/google/uuid"
)

// SolveWithGolp solves the Meta GP problem using Golp (lp_solve) via MILP
func (s *MetaGPSolver) SolveWithGolp() (*MetaResult, error) {
	// 1. Calculate number of variables
	// Continuous vars: s.variables (allocations)
	// Binary vars: for each goal, one for each target level
	numContinuous := len(s.variables)
	numBinary := 0

	// Map to track binary variable indices
	// goalIdx -> levelIdx -> binaryVarIdx within solver
	type levelVarKey struct {
		goalIdx  int
		levelIdx int
	}
	binaryVarMap := make(map[levelVarKey]int)

	for i, goal := range s.goals {
		numBinary += len(goal.TargetLevels)
		for j := range goal.TargetLevels {
			binaryVarMap[levelVarKey{i, j}] = numContinuous + len(binaryVarMap)
		}
	}

	totalVars := numContinuous + numBinary

	// 2. Create Solver
	lpSolver, err := lp.CreateGolpSolver(totalVars)
	if err != nil {
		return nil, fmt.Errorf("failed to create golp solver: %w", err)
	}
	defer lpSolver.Close()

	// 3. Set Objective Function: Maximize Rewards
	// Maximize sum(y_ij * Reward_ij * PriorityWeight)
	// To match Greedy behavior involves strict priority: P1 >>> P2.
	// We apply a large scaling factor based on priority.
	// Assuming Max Priority isn't huge (e.g. 1-10).
	const baseWeight = 1000.0 // Factor between priority levels

	objCoeffs := make([]float64, totalVars)
	// Continuous variables (allocations) have 0 coefficient (we only care about rewards)
	// Let's add a very small negative coeff to allocations to prefer saving if rewards are equal
	epsilon := 1e-6
	for i := 0; i < numContinuous; i++ {
		objCoeffs[i] = -epsilon
	}

	for i, goal := range s.goals {
		// Priority 1 is highest. We want P1 to have weight higher than sum of all P2...
		// Weight = baseWeight ^ (MaxPriority - Priority) might be needed.
		// For simplicity, let's use a simple distinct large mapping if priorities are small.
		// Or: PriorityWeight = 1 / Priority * HugeNumber.
		// Better: Weight = 1000^(10 - Priority). Assuming Priority <= 10.

		// Let's use a safe large multiplier sequence.
		// Priority 1: 1e9, Priority 2: 1e6, Priority 3: 1e3...
		// If Priority is user defined int, we need to be careful.
		// Let's assume standard 1-5 scale.
		priority := goal.Priority
		if priority < 1 {
			priority = 1
		}
		// Determine weight.
		// If we use 1000 as base, then higher priority dominates lower.

		// Note: Use float power
		priorityPower := 10.0 - float64(priority) // e.g. P1->9, P2->8
		if priorityPower < 0 {
			priorityPower = 0
		}
		priorityWeight := math.Pow(baseWeight, priorityPower)

		for j, level := range goal.TargetLevels {
			binIdx := binaryVarMap[levelVarKey{i, j}]
			// Objective = Reward * PriorityWeight
			objCoeffs[binIdx] = level.Reward * priorityWeight
		}
	}

	if err := lpSolver.SetObjective(objCoeffs, true); err != nil {
		return nil, fmt.Errorf("failed to set objective: %w", err)
	}

	// 4. Set Variable Properties (Unbounded positive for allocations, Binary for levels)
	for i := 0; i < numContinuous; i++ {
		// Set bounds based on variable limits
		v := s.variables[i]
		upper := math.Inf(1)
		if v.MaxValue > 0 {
			upper = v.MaxValue
		}
		if err := lpSolver.SetBounds(i, v.MinValue, upper); err != nil {
			return nil, fmt.Errorf("failed to set bounds for var %d: %w", i, err)
		}
	}

	for i := numContinuous; i < totalVars; i++ {
		if err := lpSolver.SetBinary(i); err != nil {
			return nil, fmt.Errorf("failed to set binary for var %d: %w", i, err)
		}
	}

	// 5. Constraints

	// 5.1 Global Budget Constraint: sum(x_k) <= TotalIncome
	budgetCoeffs := make([]float64, totalVars)
	for i := 0; i < numContinuous; i++ {
		budgetCoeffs[i] = 1.0
	}
	if err := lpSolver.AddConstraint(budgetCoeffs, "<=", s.totalIncome); err != nil {
		return nil, fmt.Errorf("failed to add budget constraint: %w", err)
	}

	// 5.2 Link Constraints: x_i >= LevelVal * y_ij
	// Reformulated: x_i - LevelVal * y_ij >= 0
	// For each goal i and level j:
	for i, goal := range s.goals {
		if goal.VariableIdx >= numContinuous {
			continue // Should not happen
		}

		for j, level := range goal.TargetLevels {
			binIdx := binaryVarMap[levelVarKey{i, j}]

			// x_i - V_ij * y_ij >= 0
			// x_i coeff = 1.0
			// y_ij coeff = -V_ij
			coeffs := make([]float64, totalVars)
			coeffs[goal.VariableIdx] = 1.0
			coeffs[binIdx] = -level.Value

			if err := lpSolver.AddConstraint(coeffs, ">=", 0); err != nil {
				return nil, fmt.Errorf("failed to add link constraint for goal %s level %d: %w", goal.ID, j, err)
			}
		}
	}

	// 5.3 Mutually Exclusive Levels Constraint: sum_j y_ij <= 1
	// For each goal, select at most one level (the "best" achieved level)
	for i, goal := range s.goals {
		if len(goal.TargetLevels) == 0 {
			continue
		}

		coeffs := make([]float64, totalVars)
		for j := range goal.TargetLevels {
			binIdx := binaryVarMap[levelVarKey{i, j}]
			coeffs[binIdx] = 1.0
		}

		if err := lpSolver.AddConstraint(coeffs, "<=", 1.0); err != nil {
			return nil, fmt.Errorf("failed to add exclusive constraint for goal %s: %w", goal.ID, err)
		}
	}

	// 6. Solve
	result, err := lpSolver.Solve()
	if err != nil {
		return nil, fmt.Errorf("failed to solve: %w", err)
	}

	if result.Status != lp.LPOptimal {
		return nil, fmt.Errorf("solver failed to find optimal solution, status: %v", result.Status)
	}

	// 7. Map Results
	metaResult := &MetaResult{
		VariableValues: make(map[uuid.UUID]float64),
		GoalLevels:     make(map[string]string),
		GoalValues:     make(map[string]float64),
		AchievedGoals:  make([]string, 0),
		IdealGoals:     make([]string, 0),
		TotalReward:    result.ObjectiveValue,
		// Note: Objective contains negative epsilon terms, might slightly deviate from pure reward sum.
		// We can recalculate exact reward from solution.
	}

	// Store allocation values
	for i, v := range s.variables {
		val := result.Solution[i]
		metaResult.VariableValues[v.ID] = val
	}

	// Determine achieved levels
	recalcReward := 0.0
	for i, goal := range s.goals {
		if goal.VariableIdx >= numContinuous {
			continue
		}

		// Get actual allocation value
		val := result.Solution[goal.VariableIdx]
		metaResult.GoalValues[goal.ID] = val

		// Check which binary var is set (or just check value against levels)
		// Using binary vars is safer to see what the solver "chose"
		chosenLevelIdx := -1
		for j := range goal.TargetLevels {
			binIdx := binaryVarMap[levelVarKey{i, j}]
			if result.Solution[binIdx] > 0.5 { // Binary 1
				chosenLevelIdx = j
				break
			}
		}

		levelName := "none"
		if chosenLevelIdx >= 0 {
			level := goal.TargetLevels[chosenLevelIdx]
			levelName = level.Level
			recalcReward += level.Reward
			metaResult.AchievedGoals = append(metaResult.AchievedGoals, goal.ID)

			// Check if ideal
			if chosenLevelIdx == len(goal.TargetLevels)-1 {
				metaResult.IdealGoals = append(metaResult.IdealGoals, goal.ID)
			}
		}

		metaResult.GoalLevels[goal.ID] = levelName
	}

	metaResult.TotalReward = recalcReward

	// Calculate MaxPossibleReward same as in original code
	for _, goal := range s.goals {
		if len(goal.TargetLevels) > 0 {
			metaResult.MaxPossibleReward += goal.TargetLevels[len(goal.TargetLevels)-1].Reward
		}
	}
	if metaResult.MaxPossibleReward > 0 {
		metaResult.RewardRatio = metaResult.TotalReward / metaResult.MaxPossibleReward
	}

	metaResult.Iterations = result.Iterations

	return metaResult, nil
}
