package solver

import (
	"personalfinancedss/internal/module/analytics/budget_allocation/domain"
	"personalfinancedss/internal/module/analytics/models/budget_allocation/solver/meta"
	"personalfinancedss/internal/module/analytics/models/budget_allocation/solver/minmax"
	"personalfinancedss/internal/module/analytics/models/budget_allocation/solver/preemptive"
	"personalfinancedss/internal/module/analytics/models/budget_allocation/solver/weighted"

	"github.com/google/uuid"
)

// GoalProgrammingSolver implements the budget allocation algorithm
// using Preemptive Goal Programming with Simplex method
type GoalProgrammingSolver struct {
	constraintModel *domain.ConstraintModel
	params          domain.ScenarioParameters
}

// NewGoalProgrammingSolver creates a new solver
func NewGoalProgrammingSolver(model *domain.ConstraintModel, params domain.ScenarioParameters) *GoalProgrammingSolver {
	return &GoalProgrammingSolver{
		constraintModel: model,
		params:          params,
	}
}

// Solve executes the Preemptive Goal Programming algorithm (default)
func (gp *GoalProgrammingSolver) Solve() (*domain.AllocationResult, error) {
	return gp.SolvePreemptive()
}

// SolvePreemptive executes Preemptive GP
func (gp *GoalProgrammingSolver) SolvePreemptive() (*domain.AllocationResult, error) {
	result := &domain.AllocationResult{
		CategoryAllocations: make(map[uuid.UUID]float64),
		GoalAllocations:     make(map[uuid.UUID]float64),
		DebtAllocations:     make(map[uuid.UUID]domain.DebtPayment),
		FeasibilityScore:    100.0,
		AchievedGoals:       make([]string, 0),
		UnachievedGoals:     make([]string, 0),
		SolverType:          "preemptive",
	}

	gpSolver := preemptive.BuildPreemptiveGPFromConstraintModel(gp.constraintModel, gp.params)
	gpResult, err := gpSolver.Solve()

	if err != nil {
		return gp.solveHeuristic()
	}

	gp.mapPreemptiveResult(result, gpResult)
	return result, nil
}

// SolveWeighted executes Weighted GP
func (gp *GoalProgrammingSolver) SolveWeighted() (*domain.AllocationResult, error) {
	result := &domain.AllocationResult{
		CategoryAllocations: make(map[uuid.UUID]float64),
		GoalAllocations:     make(map[uuid.UUID]float64),
		DebtAllocations:     make(map[uuid.UUID]domain.DebtPayment),
		FeasibilityScore:    100.0,
		AchievedGoals:       make([]string, 0),
		UnachievedGoals:     make([]string, 0),
		SolverType:          "weighted",
	}

	wgpSolver := weighted.BuildWeightedGPFromConstraintModel(gp.constraintModel, gp.params)
	wgpResult, err := wgpSolver.Solve()

	if err != nil {
		return gp.solveHeuristic()
	}

	gp.mapWeightedResult(result, wgpResult)
	return result, nil
}

// SolveMinmax executes Minmax (Chebyshev) GP
func (gp *GoalProgrammingSolver) SolveMinmax() (*domain.AllocationResult, error) {
	result := &domain.AllocationResult{
		CategoryAllocations: make(map[uuid.UUID]float64),
		GoalAllocations:     make(map[uuid.UUID]float64),
		DebtAllocations:     make(map[uuid.UUID]domain.DebtPayment),
		FeasibilityScore:    100.0,
		AchievedGoals:       make([]string, 0),
		UnachievedGoals:     make([]string, 0),
		SolverType:          "minmax",
	}

	mmSolver := minmax.BuildMinmaxGPFromConstraintModel(gp.constraintModel, gp.params)
	mmResult, err := mmSolver.Solve()

	if err != nil {
		return gp.solveHeuristic()
	}

	gp.mapMinmaxResult(result, mmResult)
	return result, nil
}

// SolveMeta executes Meta (Multi-Choice) GP
func (gp *GoalProgrammingSolver) SolveMeta() (*domain.AllocationResult, error) {
	result := &domain.AllocationResult{
		CategoryAllocations: make(map[uuid.UUID]float64),
		GoalAllocations:     make(map[uuid.UUID]float64),
		DebtAllocations:     make(map[uuid.UUID]domain.DebtPayment),
		FeasibilityScore:    100.0,
		AchievedGoals:       make([]string, 0),
		UnachievedGoals:     make([]string, 0),
		SolverType:          "meta",
	}

	metaSolver := meta.BuildMetaGPFromConstraintModel(gp.constraintModel, gp.params)
	metaResult, err := metaSolver.Solve()

	if err != nil {
		return gp.solveHeuristic()
	}

	gp.mapMetaResult(result, metaResult)
	return result, nil
}

// SolveDual runs both GP solvers and compares results
func (gp *GoalProgrammingSolver) SolveDual() (*domain.DualGPResult, error) {
	preemptiveResult, err1 := gp.SolvePreemptive()
	weightedResult, err2 := gp.SolveWeighted()

	if err1 != nil && err2 != nil {
		return nil, err1
	}

	dual := &domain.DualGPResult{
		PreemptiveResult: preemptiveResult,
		WeightedResult:   weightedResult,
	}

	// Compare results
	dual.Comparison = gp.compareResults(preemptiveResult, weightedResult)

	return dual, nil
}

// SolveTriple runs all three GP solvers and compares results
func (gp *GoalProgrammingSolver) SolveTriple() (*domain.TripleGPResult, error) {
	preemptiveResult, err1 := gp.SolvePreemptive()
	weightedResult, err2 := gp.SolveWeighted()
	minmaxResult, err3 := gp.SolveMinmax()

	if err1 != nil && err2 != nil && err3 != nil {
		return nil, err1
	}

	triple := &domain.TripleGPResult{
		PreemptiveResult: preemptiveResult,
		WeightedResult:   weightedResult,
		MinmaxResult:     minmaxResult,
	}

	// Compare results
	triple.Comparison = gp.compareTripleResults(preemptiveResult, weightedResult, minmaxResult)

	return triple, nil
}

// compareTripleResults compares all three GP results
func (gp *GoalProgrammingSolver) compareTripleResults(preemptiveResult, weightedResult, minmaxResult *domain.AllocationResult) domain.TripleGPComparison {
	comp := domain.TripleGPComparison{
		PreemptiveAchievedCount: len(preemptiveResult.AchievedGoals),
		WeightedAchievedCount:   len(weightedResult.AchievedGoals),
		MinmaxAchievedCount:     len(minmaxResult.AchievedGoals),
	}

	// Get minmax-specific metrics from the solver
	mmSolver := minmax.BuildMinmaxGPFromConstraintModel(gp.constraintModel, gp.params)
	mmResult, _ := mmSolver.Solve()
	if mmResult != nil {
		comp.MinmaxMinAchievement = mmResult.MinAchievement
		comp.MinmaxIsBalanced = mmResult.IsBalanced
	}

	// Find the best solver
	maxAchieved := comp.PreemptiveAchievedCount
	comp.RecommendedSolver = "preemptive"
	comp.Reason = "Preemptive GP guarantees high-priority goals are fully satisfied"

	if comp.WeightedAchievedCount > maxAchieved {
		maxAchieved = comp.WeightedAchievedCount
		comp.RecommendedSolver = "weighted"
		comp.Reason = "Weighted GP achieved more goals completely"
	}

	if comp.MinmaxAchievedCount > maxAchieved {
		comp.RecommendedSolver = "minmax"
		comp.Reason = "Minmax GP achieved more goals completely"
	} else if comp.MinmaxIsBalanced && comp.MinmaxMinAchievement >= 50 {
		// Minmax is good when it provides balanced allocation
		if comp.MinmaxAchievedCount >= maxAchieved-1 {
			comp.RecommendedSolver = "minmax"
			comp.Reason = "Minmax GP provides balanced allocation across all goals"
		}
	}

	return comp
}

// mapPreemptiveResult maps Preemptive GP result to domain.AllocationResult
func (gp *GoalProgrammingSolver) mapPreemptiveResult(result *domain.AllocationResult, gpResult *preemptive.GPResult) {
	result.SolverIterations = gpResult.SolverIterations
	result.AchievedGoals = gpResult.AchievedGoals
	result.UnachievedGoals = gpResult.UnachievedGoals

	totalAllocated := 0.0

	for id := range gp.constraintModel.MandatoryExpenses {
		if val, exists := gpResult.VariableValues[id]; exists {
			result.CategoryAllocations[id] = val
			totalAllocated += val
		}
	}

	for id := range gp.constraintModel.FlexibleExpenses {
		if val, exists := gpResult.VariableValues[id]; exists {
			result.CategoryAllocations[id] = val
			totalAllocated += val
		}
	}

	for id, constraint := range gp.constraintModel.DebtPayments {
		if val, exists := gpResult.VariableValues[id]; exists {
			extraPayment := val - constraint.MinimumPayment
			if extraPayment < 0 {
				extraPayment = 0
			}
			result.DebtAllocations[id] = domain.DebtPayment{
				TotalPayment:   val,
				MinimumPayment: constraint.MinimumPayment,
				ExtraPayment:   extraPayment,
			}
			totalAllocated += val
		}
	}

	for id := range gp.constraintModel.GoalTargets {
		if val, exists := gpResult.VariableValues[id]; exists {
			result.GoalAllocations[id] = val
			totalAllocated += val
		}
	}

	result.TotalAllocated = totalAllocated
	result.Surplus = gp.constraintModel.TotalIncome - totalAllocated

	if len(gpResult.AchievedGoals)+len(gpResult.UnachievedGoals) > 0 {
		achievedRatio := float64(len(gpResult.AchievedGoals)) /
			float64(len(gpResult.AchievedGoals)+len(gpResult.UnachievedGoals))
		result.FeasibilityScore = achievedRatio * 100
	}

	if !gpResult.IsFeasible {
		result.FeasibilityScore = 0
	}
}

// mapMinmaxResult maps Minmax GP result to domain.AllocationResult
func (gp *GoalProgrammingSolver) mapMinmaxResult(result *domain.AllocationResult, mmResult *minmax.MinmaxResult) {
	result.SolverIterations = mmResult.Iterations
	result.AchievedGoals = mmResult.AchievedGoals
	result.UnachievedGoals = append(mmResult.PartialGoals, mmResult.UnachievedGoals...)

	totalAllocated := 0.0

	for id := range gp.constraintModel.MandatoryExpenses {
		if val, exists := mmResult.VariableValues[id]; exists {
			result.CategoryAllocations[id] = val
			totalAllocated += val
		}
	}

	for id := range gp.constraintModel.FlexibleExpenses {
		if val, exists := mmResult.VariableValues[id]; exists {
			result.CategoryAllocations[id] = val
			totalAllocated += val
		}
	}

	for id, constraint := range gp.constraintModel.DebtPayments {
		if val, exists := mmResult.VariableValues[id]; exists {
			extraPayment := val - constraint.MinimumPayment
			if extraPayment < 0 {
				extraPayment = 0
			}
			result.DebtAllocations[id] = domain.DebtPayment{
				TotalPayment:   val,
				MinimumPayment: constraint.MinimumPayment,
				ExtraPayment:   extraPayment,
			}
			totalAllocated += val
		}
	}

	for id := range gp.constraintModel.GoalTargets {
		if val, exists := mmResult.VariableValues[id]; exists {
			result.GoalAllocations[id] = val
			totalAllocated += val
		}
	}

	result.TotalAllocated = totalAllocated
	result.Surplus = gp.constraintModel.TotalIncome - totalAllocated

	// Use minimum achievement as feasibility score for minmax
	result.FeasibilityScore = mmResult.MinAchievement
}

// mapMetaResult maps Meta GP result to domain.AllocationResult
func (gp *GoalProgrammingSolver) mapMetaResult(result *domain.AllocationResult, metaResult *meta.MetaResult) {
	result.SolverIterations = metaResult.Iterations
	result.AchievedGoals = metaResult.AchievedGoals
	// Goals not at ideal level are considered "unachieved" for comparison
	for goalID, level := range metaResult.GoalLevels {
		if level != "ideal" && level != "none" {
			result.UnachievedGoals = append(result.UnachievedGoals, goalID)
		}
	}

	totalAllocated := 0.0

	for id := range gp.constraintModel.MandatoryExpenses {
		if val, exists := metaResult.VariableValues[id]; exists {
			result.CategoryAllocations[id] = val
			totalAllocated += val
		}
	}

	for id := range gp.constraintModel.FlexibleExpenses {
		if val, exists := metaResult.VariableValues[id]; exists {
			result.CategoryAllocations[id] = val
			totalAllocated += val
		}
	}

	for id, constraint := range gp.constraintModel.DebtPayments {
		if val, exists := metaResult.VariableValues[id]; exists {
			extraPayment := val - constraint.MinimumPayment
			if extraPayment < 0 {
				extraPayment = 0
			}
			result.DebtAllocations[id] = domain.DebtPayment{
				TotalPayment:   val,
				MinimumPayment: constraint.MinimumPayment,
				ExtraPayment:   extraPayment,
			}
			totalAllocated += val
		}
	}

	for id := range gp.constraintModel.GoalTargets {
		if val, exists := metaResult.VariableValues[id]; exists {
			result.GoalAllocations[id] = val
			totalAllocated += val
		}
	}

	result.TotalAllocated = totalAllocated
	result.Surplus = gp.constraintModel.TotalIncome - totalAllocated

	// Use reward ratio as feasibility score for meta GP
	result.FeasibilityScore = metaResult.RewardRatio * 100
}

// mapWeightedResult maps Weighted GP result to domain.AllocationResult
func (gp *GoalProgrammingSolver) mapWeightedResult(result *domain.AllocationResult, wgpResult *weighted.WGPResult) {
	result.SolverIterations = wgpResult.Iterations
	result.AchievedGoals = wgpResult.AchievedGoals
	result.UnachievedGoals = append(wgpResult.PartialGoals, wgpResult.UnachievedGoals...)

	totalAllocated := 0.0

	for id := range gp.constraintModel.MandatoryExpenses {
		if val, exists := wgpResult.VariableValues[id]; exists {
			result.CategoryAllocations[id] = val
			totalAllocated += val
		}
	}

	for id := range gp.constraintModel.FlexibleExpenses {
		if val, exists := wgpResult.VariableValues[id]; exists {
			result.CategoryAllocations[id] = val
			totalAllocated += val
		}
	}

	for id, constraint := range gp.constraintModel.DebtPayments {
		if val, exists := wgpResult.VariableValues[id]; exists {
			extraPayment := val - constraint.MinimumPayment
			if extraPayment < 0 {
				extraPayment = 0
			}
			result.DebtAllocations[id] = domain.DebtPayment{
				TotalPayment:   val,
				MinimumPayment: constraint.MinimumPayment,
				ExtraPayment:   extraPayment,
			}
			totalAllocated += val
		}
	}

	for id := range gp.constraintModel.GoalTargets {
		if val, exists := wgpResult.VariableValues[id]; exists {
			result.GoalAllocations[id] = val
			totalAllocated += val
		}
	}

	result.TotalAllocated = totalAllocated
	result.Surplus = gp.constraintModel.TotalIncome - totalAllocated

	totalGoals := len(wgpResult.AchievedGoals) + len(wgpResult.PartialGoals) + len(wgpResult.UnachievedGoals)
	if totalGoals > 0 {
		// Give partial credit for partial goals
		achieved := float64(len(wgpResult.AchievedGoals)) + float64(len(wgpResult.PartialGoals))*0.5
		result.FeasibilityScore = (achieved / float64(totalGoals)) * 100
	}
}

// compareResults compares Preemptive and Weighted GP results
func (gp *GoalProgrammingSolver) compareResults(preemptive, weighted *domain.AllocationResult) domain.GPComparison {
	comp := domain.GPComparison{
		PreemptiveAchievedCount: len(preemptive.AchievedGoals),
		WeightedAchievedCount:   len(weighted.AchievedGoals),
	}

	// Calculate total deviation for each
	// (simplified - in practice would use actual deviation values)
	comp.PreemptiveTotalDev = float64(len(preemptive.UnachievedGoals))
	comp.WeightedTotalDev = float64(len(weighted.UnachievedGoals))

	// Recommend based on results
	if comp.PreemptiveAchievedCount > comp.WeightedAchievedCount {
		comp.RecommendedSolver = "preemptive"
		comp.Reason = "Preemptive GP achieved more goals completely"
	} else if comp.WeightedAchievedCount > comp.PreemptiveAchievedCount {
		comp.RecommendedSolver = "weighted"
		comp.Reason = "Weighted GP achieved more goals completely"
	} else if preemptive.FeasibilityScore > weighted.FeasibilityScore {
		comp.RecommendedSolver = "preemptive"
		comp.Reason = "Preemptive GP has higher feasibility score"
	} else if weighted.FeasibilityScore > preemptive.FeasibilityScore {
		comp.RecommendedSolver = "weighted"
		comp.Reason = "Weighted GP has higher feasibility score"
	} else {
		comp.RecommendedSolver = "preemptive"
		comp.Reason = "Both solvers performed equally; defaulting to preemptive for priority guarantee"
	}

	return comp
}

// solveHeuristic is the fallback heuristic solver (original implementation)
func (gp *GoalProgrammingSolver) solveHeuristic() (*domain.AllocationResult, error) {
	result := &domain.AllocationResult{
		CategoryAllocations: make(map[uuid.UUID]float64),
		GoalAllocations:     make(map[uuid.UUID]float64),
		DebtAllocations:     make(map[uuid.UUID]domain.DebtPayment),
		FeasibilityScore:    100.0,
	}

	remainingIncome := gp.constraintModel.TotalIncome

	// Phase 1: Allocate mandatory expenses (Priority 1 - Must Satisfy)
	for categoryID, constraint := range gp.constraintModel.MandatoryExpenses {
		amount := constraint.Minimum
		result.CategoryAllocations[categoryID] = amount
		remainingIncome -= amount
	}

	// Phase 2: Allocate minimum debt payments (Priority 2 - Critical)
	for debtID, constraint := range gp.constraintModel.DebtPayments {
		payment := domain.DebtPayment{
			TotalPayment:   constraint.MinimumPayment,
			MinimumPayment: constraint.MinimumPayment,
			ExtraPayment:   0,
		}
		result.DebtAllocations[debtID] = payment
		remainingIncome -= constraint.MinimumPayment
	}

	// Check if we've gone negative (infeasible)
	if remainingIncome < 0 {
		result.FeasibilityScore = 0
		result.Surplus = remainingIncome
		return result, nil
	}

	// Phase 3: Allocate surplus based on scenario parameters
	gp.allocateSurplusHeuristic(result, remainingIncome)

	result.Surplus = remainingIncome - gp.getTotalSurplusAllocated(result)
	result.TotalAllocated = gp.constraintModel.TotalIncome - result.Surplus

	return result, nil
}

// allocateSurplusHeuristic distributes surplus income according to scenario strategy
func (gp *GoalProgrammingSolver) allocateSurplusHeuristic(result *domain.AllocationResult, surplus float64) {
	if surplus <= 0 {
		return
	}

	sa := gp.params.SurplusAllocation

	// 1. Allocate to emergency fund goals first
	if sa.EmergencyFundPercent > 0 {
		emergencyAmount := surplus * sa.EmergencyFundPercent
		gp.allocateToEmergencyFunds(result, emergencyAmount)
	}

	// 2. Allocate to extra debt payments
	if sa.DebtExtraPercent > 0 {
		debtAmount := surplus * sa.DebtExtraPercent
		gp.allocateToExtraDebtPayments(result, debtAmount)
	}

	// 3. Allocate to high-priority goals
	if sa.GoalsPercent > 0 {
		goalsAmount := surplus * sa.GoalsPercent
		gp.allocateToGoals(result, goalsAmount)
	}

	// 4. Allocate to flexible expense categories
	if sa.FlexiblePercent > 0 {
		flexibleAmount := surplus * sa.FlexiblePercent
		gp.allocateToFlexibleCategories(result, flexibleAmount)
	}
}

// allocateToEmergencyFunds allocates to emergency fund type goals
func (gp *GoalProgrammingSolver) allocateToEmergencyFunds(result *domain.AllocationResult, amount float64) {
	if amount <= 0 {
		return
	}

	var emergencyGoals []domain.GoalConstraint
	for _, goal := range gp.constraintModel.GoalTargets {
		if goal.GoalType == "emergency" {
			emergencyGoals = append(emergencyGoals, goal)
		}
	}

	if len(emergencyGoals) == 0 {
		return
	}

	amountPerGoal := amount / float64(len(emergencyGoals))
	for _, goal := range emergencyGoals {
		allocation := min(amountPerGoal, goal.RemainingAmount)
		result.GoalAllocations[goal.GoalID] += allocation
	}
}

// allocateToExtraDebtPayments allocates extra payments to debts (prioritize high interest)
func (gp *GoalProgrammingSolver) allocateToExtraDebtPayments(result *domain.AllocationResult, amount float64) {
	if amount <= 0 {
		return
	}

	var highestPriorityDebtID uuid.UUID
	highestPriority := 999

	for debtID, constraint := range gp.constraintModel.DebtPayments {
		if constraint.Priority < highestPriority {
			highestPriority = constraint.Priority
			highestPriorityDebtID = debtID
		}
	}

	if highestPriorityDebtID != uuid.Nil {
		debt := gp.constraintModel.DebtPayments[highestPriorityDebtID]
		extraPayment := min(amount, debt.CurrentBalance-debt.MinimumPayment)
		if extraPayment > 0 {
			payment := result.DebtAllocations[highestPriorityDebtID]
			payment.ExtraPayment += extraPayment
			payment.TotalPayment += extraPayment
			result.DebtAllocations[highestPriorityDebtID] = payment
		}
	}
}

// allocateToGoals allocates to goals based on priority
func (gp *GoalProgrammingSolver) allocateToGoals(result *domain.AllocationResult, amount float64) {
	if amount <= 0 {
		return
	}

	remainingAmount := amount

	type goalWithPriority struct {
		goalID uuid.UUID
		goal   domain.GoalConstraint
	}

	var sortedGoals []goalWithPriority
	for goalID, goal := range gp.constraintModel.GoalTargets {
		if goal.GoalType == "emergency" {
			continue
		}
		sortedGoals = append(sortedGoals, goalWithPriority{goalID, goal})
	}

	// Sort by priority weight
	for i := 0; i < len(sortedGoals); i++ {
		for j := i + 1; j < len(sortedGoals); j++ {
			if sortedGoals[i].goal.PriorityWeight > sortedGoals[j].goal.PriorityWeight {
				sortedGoals[i], sortedGoals[j] = sortedGoals[j], sortedGoals[i]
			}
		}
	}

	for _, item := range sortedGoals {
		if remainingAmount <= 0 {
			break
		}

		goal := item.goal
		suggestedAmount := goal.SuggestedContribution * gp.params.GoalContributionFactor
		allocation := min(suggestedAmount, goal.RemainingAmount)
		allocation = min(allocation, remainingAmount)

		if allocation > 0 {
			result.GoalAllocations[item.goalID] += allocation
			remainingAmount -= allocation
		}
	}
}

// allocateToFlexibleCategories allocates to flexible expense categories
func (gp *GoalProgrammingSolver) allocateToFlexibleCategories(result *domain.AllocationResult, amount float64) {
	if amount <= 0 {
		return
	}

	level := gp.params.FlexibleSpendingLevel
	remainingAmount := amount

	for categoryID, constraint := range gp.constraintModel.FlexibleExpenses {
		if remainingAmount <= 0 {
			break
		}

		var target float64
		if constraint.Maximum > 0 {
			rangeAmount := constraint.Maximum - constraint.Minimum
			target = constraint.Minimum + (rangeAmount * level)
		} else {
			target = constraint.Minimum + (remainingAmount * level)
		}

		allocation := min(target-constraint.Minimum, remainingAmount)
		if allocation > 0 {
			currentAllocation := result.CategoryAllocations[categoryID]
			if currentAllocation == 0 {
				currentAllocation = constraint.Minimum
			}
			result.CategoryAllocations[categoryID] = currentAllocation + allocation
			remainingAmount -= allocation
		}
	}
}

// getTotalSurplusAllocated calculates total amount allocated from surplus
func (gp *GoalProgrammingSolver) getTotalSurplusAllocated(result *domain.AllocationResult) float64 {
	var total float64

	for _, amount := range result.GoalAllocations {
		total += amount
	}

	for _, payment := range result.DebtAllocations {
		total += payment.ExtraPayment
	}

	return total
}

// min returns the minimum of two float64 values
func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
