package lp

import (
	"fmt"
	"personalfinancedss/internal/module/analytics/budget_allocation/domain"

	"github.com/google/uuid"
)

// =============================================================================
// Goal Programming Solver using LP (Linear Programming)
// Chuyển đổi bài toán GP thành LP và giải bằng Simplex hoặc golp
// =============================================================================

// GPVariable represents a decision variable in the LP-based GP model
type GPVariable struct {
	ID       uuid.UUID
	Name     string
	Type     string  // "category", "goal", "debt"
	MinValue float64 // Lower bound
	MaxValue float64 // Upper bound (0 = no upper bound)
}

// GPGoal represents a goal in the LP-based GP model
type GPGoal struct {
	ID          string
	Description string
	Priority    int // Lower = higher priority (1 is highest)
	TargetValue float64
	VariableIdx int     // Index of the variable this goal relates to
	GoalType    string  // "min_deviation", "at_least", "at_most", "exactly"
	Weight      float64 // Weight for the objective function
}

// GPResult contains the solution from the LP-based GP solver
type GPResult struct {
	VariableValues   map[uuid.UUID]float64
	GoalDeviations   map[string]float64 // Deviation from each goal
	AchievedGoals    []string
	UnachievedGoals  []string
	TotalDeviation   float64
	IsFeasible       bool
	SolverIterations int
}

// LPBasedGPSolver solves Goal Programming using Linear Programming
type LPBasedGPSolver struct {
	variables   []GPVariable
	goals       []GPGoal
	totalIncome float64
	solverType  string // "simplex" or "golp"
}

// NewLPBasedGPSolver creates a new LP-based GP solver
// solverType: "simplex" (Pure Go) or "golp" (CGO + lp_solve)
func NewLPBasedGPSolver(totalIncome float64, solverType string) *LPBasedGPSolver {
	return &LPBasedGPSolver{
		variables:   make([]GPVariable, 0),
		goals:       make([]GPGoal, 0),
		totalIncome: totalIncome,
		solverType:  solverType,
	}
}

// AddVariable adds a decision variable
func (s *LPBasedGPSolver) AddVariable(v GPVariable) int {
	idx := len(s.variables)
	s.variables = append(s.variables, v)
	return idx
}

// AddGoal adds a goal
func (s *LPBasedGPSolver) AddGoal(g GPGoal) {
	s.goals = append(s.goals, g)
}

// Solve converts GP to LP and solves
// GP Model:
//
//	minimize: Σ (w_i * d_i^- + w_i * d_i^+)
//	subject to: f_i(x) + d_i^- - d_i^+ = g_i  (goal constraints)
//	            Σ x_i <= totalIncome          (budget constraint)
//	            x_i >= min_i, x_i <= max_i    (bounds)
//	            d_i^-, d_i^+ >= 0             (deviations non-negative)
func (s *LPBasedGPSolver) Solve() (*GPResult, error) {
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

	numVars := len(s.variables)
	numGoals := len(s.goals)

	// LP variables:
	// [0..numVars-1]: original decision variables x_i
	// [numVars..numVars+numGoals-1]: negative deviations d_i^-
	// [numVars+numGoals..numVars+2*numGoals-1]: positive deviations d_i^+
	totalLPVars := numVars + 2*numGoals

	// Create LP solver
	var lpSolver LPSolver
	var err error

	switch s.solverType {
	case "golp":
		lpSolver, err = CreateGolpSolver(totalLPVars)
		if err != nil {
			// Fallback to simplex
			lpSolver = NewPureGoSimplexSolver(totalLPVars)
		}
	default:
		lpSolver = NewPureGoSimplexSolver(totalLPVars)
	}

	// Set objective: minimize Σ (w_i * d_i^-)
	// For "at_least" goals, we only penalize negative deviation
	// For "at_most" goals, we only penalize positive deviation
	// For "exactly" goals, we penalize both
	objective := make([]float64, totalLPVars)
	for i, goal := range s.goals {
		dMinusIdx := numVars + i
		dPlusIdx := numVars + numGoals + i

		switch goal.GoalType {
		case "at_least", "min_deviation":
			objective[dMinusIdx] = goal.Weight // Penalize under-achievement
		case "at_most":
			objective[dPlusIdx] = goal.Weight // Penalize over-achievement
		case "exactly":
			objective[dMinusIdx] = goal.Weight
			objective[dPlusIdx] = goal.Weight
		}
	}
	lpSolver.SetObjective(objective, false) // minimize

	// Set bounds for decision variables
	for i, v := range s.variables {
		lower := v.MinValue
		upper := v.MaxValue
		if upper <= 0 {
			upper = s.totalIncome // No explicit upper bound
		}
		lpSolver.SetBounds(i, lower, upper)
	}

	// Set bounds for deviation variables (>= 0)
	for i := 0; i < numGoals; i++ {
		lpSolver.SetBounds(numVars+i, 0, s.totalIncome)          // d_i^-
		lpSolver.SetBounds(numVars+numGoals+i, 0, s.totalIncome) // d_i^+
	}

	// Add goal constraints: x[varIdx] + d^- - d^+ = target
	for i, goal := range s.goals {
		if goal.VariableIdx >= numVars {
			continue
		}

		coeffs := make([]float64, totalLPVars)
		coeffs[goal.VariableIdx] = 1.0    // x_i
		coeffs[numVars+i] = 1.0           // + d_i^-
		coeffs[numVars+numGoals+i] = -1.0 // - d_i^+

		lpSolver.AddConstraint(coeffs, "=", goal.TargetValue)
	}

	// Add budget constraint: Σ x_i <= totalIncome
	budgetCoeffs := make([]float64, totalLPVars)
	for i := 0; i < numVars; i++ {
		budgetCoeffs[i] = 1.0
	}
	lpSolver.AddConstraint(budgetCoeffs, "<=", s.totalIncome)

	// Solve
	lpResult, err := lpSolver.Solve()
	if err != nil {
		return result, err
	}

	result.SolverIterations = lpResult.Iterations

	if lpResult.Status != LPOptimal {
		result.IsFeasible = false
		// Try to extract partial solution
	}

	// Extract solution
	for i, v := range s.variables {
		if i < len(lpResult.Solution) {
			result.VariableValues[v.ID] = lpResult.Solution[i]
		}
	}

	// Calculate deviations
	for i, goal := range s.goals {
		dMinus := 0.0
		dPlus := 0.0
		if numVars+i < len(lpResult.Solution) {
			dMinus = lpResult.Solution[numVars+i]
		}
		if numVars+numGoals+i < len(lpResult.Solution) {
			dPlus = lpResult.Solution[numVars+numGoals+i]
		}

		deviation := dMinus + dPlus
		result.GoalDeviations[goal.ID] = deviation
		result.TotalDeviation += deviation * goal.Weight

		if deviation < 0.01 {
			result.AchievedGoals = append(result.AchievedGoals, goal.ID)
		} else {
			result.UnachievedGoals = append(result.UnachievedGoals, goal.ID)
		}
	}

	return result, nil
}

// =============================================================================
// Builder function để tạo LP-based GP solver từ domain.ConstraintModel
// =============================================================================

// BuildLPGPFromConstraintModel creates an LP-based GP solver
func BuildLPGPFromConstraintModel(model *domain.ConstraintModel, params domain.ScenarioParameters, solverType string) *LPBasedGPSolver {
	solver := NewLPBasedGPSolver(model.TotalIncome, solverType)

	varIndexMap := make(map[uuid.UUID]int)

	// Add variables and goals (same structure as Preemptive GP)
	// Priority 1: Mandatory expenses
	for id, constraint := range model.MandatoryExpenses {
		idx := solver.AddVariable(GPVariable{
			ID:       id,
			Name:     fmt.Sprintf("mandatory_%s", id.String()[:8]),
			Type:     "category",
			MinValue: constraint.Minimum,
			MaxValue: constraint.Maximum,
		})
		varIndexMap[id] = idx

		solver.AddGoal(GPGoal{
			ID:          fmt.Sprintf("mandatory_%s", id.String()[:8]),
			Description: "Mandatory expense",
			Priority:    1,
			TargetValue: constraint.Minimum,
			VariableIdx: idx,
			GoalType:    "at_least",
			Weight:      100.0, // Very high weight
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
			Description: fmt.Sprintf("Min payment: %s", constraint.DebtName),
			Priority:    2,
			TargetValue: constraint.MinimumPayment,
			VariableIdx: idx,
			GoalType:    "at_least",
			Weight:      80.0,
		})
	}

	// Goals
	for id, constraint := range model.GoalTargets {
		var weight float64
		var priority int

		if constraint.GoalType == "emergency" {
			weight = 50.0
			priority = 3
		} else if constraint.PriorityWeight <= 10 {
			weight = 30.0
			priority = 4
		} else {
			weight = 10.0
			priority = 6
		}

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
			ID:          fmt.Sprintf("goal_%s", id.String()[:8]),
			Description: constraint.GoalName,
			Priority:    priority,
			TargetValue: targetContribution,
			VariableIdx: idx,
			GoalType:    "at_least",
			Weight:      weight,
		})
	}

	// Flexible expenses
	for id, constraint := range model.FlexibleExpenses {
		idx := solver.AddVariable(GPVariable{
			ID:       id,
			Name:     fmt.Sprintf("flexible_%s", id.String()[:8]),
			Type:     "category",
			MinValue: constraint.Minimum,
			MaxValue: constraint.Maximum,
		})
		varIndexMap[id] = idx

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
			Weight:      5.0,
		})
	}

	return solver
}
