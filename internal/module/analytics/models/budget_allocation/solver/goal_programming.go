package solver

import (
	"fmt"
	"math"
	"personalfinancedss/internal/module/analytics/budget_allocation/domain"
	"personalfinancedss/internal/module/analytics/models/budget_allocation/solver/fuzzy"

	"github.com/google/uuid"
)

// GoalProgrammingSolver implements the budget allocation algorithm
// using Fuzzy Goal Programming
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

// Solve executes the Fuzzy Goal Programming algorithm (default)
func (gp *GoalProgrammingSolver) Solve() (*domain.AllocationResult, error) {
	return gp.SolveFuzzy()
}

// SolveFuzzy executes Fuzzy GP
func (gp *GoalProgrammingSolver) SolveFuzzy() (*domain.AllocationResult, error) {
	result := &domain.AllocationResult{
		CategoryAllocations: make(map[uuid.UUID]float64),
		GoalAllocations:     make(map[uuid.UUID]float64),
		DebtAllocations:     make(map[uuid.UUID]domain.DebtPayment),
		FeasibilityScore:    100.0,
		AchievedGoals:       make([]string, 0),
		UnachievedGoals:     make([]string, 0),
		SolverType:          "fuzzy",
	}

	// CRITICAL: Fixed constraints (mandatory expenses và debts) phải dùng heuristic, KHÔNG được đưa vào LP
	// 1. Tính mandatory expenses bằng heuristic
	totalMandatoryAllocated := 0.0
	for _, constraint := range gp.constraintModel.MandatoryExpenses {
		result.CategoryAllocations[constraint.CategoryID] = constraint.Minimum
		totalMandatoryAllocated += constraint.Minimum
	}

	// 2. Tính debts bằng heuristic
	debtHeuristicResult, _ := gp.solveDebtHeuristic()
	if debtHeuristicResult != nil {
		result.DebtAllocations = debtHeuristicResult.DebtAllocations
	}

	totalDebtAllocated := 0.0
	for _, payment := range result.DebtAllocations {
		totalDebtAllocated += payment.TotalPayment
	}

	// 3. Tạo constraint model mới KHÔNG có mandatory expenses và debts (chỉ flexible expenses và goals)
	modelForLP := &domain.ConstraintModel{
		TotalIncome:       gp.constraintModel.TotalIncome - totalMandatoryAllocated - totalDebtAllocated,
		MandatoryExpenses: make(map[uuid.UUID]domain.CategoryConstraint), // Empty - không đưa vào LP
		FlexibleExpenses:  gp.constraintModel.FlexibleExpenses,
		DebtPayments:      make(map[uuid.UUID]domain.DebtConstraint), // Empty - không đưa vào LP
		GoalTargets:       gp.constraintModel.GoalTargets,
	}

	fuzzySolver := fuzzy.BuildFuzzyGPFromConstraintModel(modelForLP, gp.params)
	fuzzyResult, err := fuzzySolver.Solve()

	if err != nil {
		// DISABLED: Heuristic fallback - return error instead
		return nil, fmt.Errorf("fuzzy GP solver failed: %w", err)
	}

	gp.mapFuzzyResultWithoutDebts(result, fuzzyResult)

	// Làm tròn các allocations
	gp.roundAllocations(result)

	// Tính TotalAllocated và Surplus sau khi làm tròn
	totalAllocated := 0.0
	for _, amount := range result.CategoryAllocations {
		totalAllocated += amount
	}
	for _, amount := range result.GoalAllocations {
		totalAllocated += amount
	}
	for _, payment := range result.DebtAllocations {
		totalAllocated += payment.TotalPayment
	}
	result.TotalAllocated = totalAllocated
	result.Surplus = gp.constraintModel.TotalIncome - totalAllocated

	// Xử lý surplus ngay tại đây (nếu có)
	if result.Surplus > 100 {
		gp.reduceSurplusByDeducting(result, result.Surplus)
		// Tính lại sau khi trừ surplus
		totalAllocated = 0.0
		for _, amount := range result.CategoryAllocations {
			totalAllocated += amount
		}
		for _, amount := range result.GoalAllocations {
			totalAllocated += amount
		}
		for _, payment := range result.DebtAllocations {
			totalAllocated += payment.TotalPayment
		}
		result.TotalAllocated = totalAllocated
		result.Surplus = gp.constraintModel.TotalIncome - totalAllocated
	}

	return result, nil
}

// mapFuzzyResultWithoutDebts maps Fuzzy GP result (KHÔNG có mandatory expenses và debts - dùng heuristic)
func (gp *GoalProgrammingSolver) mapFuzzyResultWithoutDebts(result *domain.AllocationResult, fuzzyResult *fuzzy.FuzzyResult) {
	result.SolverIterations = fuzzyResult.Iterations
	result.AchievedGoals = fuzzyResult.AchievedGoals
	result.UnachievedGoals = append(fuzzyResult.PartialGoals, fuzzyResult.UnachievedGoals...)

	totalAllocated := 0.0

	// Mandatory expenses KHÔNG lấy từ LP - đã được tính bằng heuristic

	// Map flexible expenses từ LP
	for id := range gp.constraintModel.FlexibleExpenses {
		if val, exists := fuzzyResult.VariableValues[id]; exists {
			result.CategoryAllocations[id] = val
			totalAllocated += val
		}
	}

	// Debt KHÔNG lấy từ LP - đã được tính bằng heuristic

	// Map goals từ LP
	for id := range gp.constraintModel.GoalTargets {
		if val, exists := fuzzyResult.VariableValues[id]; exists {
			result.GoalAllocations[id] = val
			totalAllocated += val
		}
	}

	// TotalAllocated sẽ được tính lại trong normalizeSurplus
	// Không set ở đây để tránh tính nhiều lần
}

// mapFuzzyResult maps Fuzzy GP result to domain.AllocationResult
// DEPRECATED: Dùng mapFuzzyResultWithoutDebts thay vì hàm này
func (gp *GoalProgrammingSolver) mapFuzzyResult(result *domain.AllocationResult, fuzzyResult *fuzzy.FuzzyResult) {
	gp.mapFuzzyResultWithoutDebts(result, fuzzyResult)

	// Debt vẫn lấy từ LP (backward compatibility) - nhưng nên dùng heuristic
	for id, constraint := range gp.constraintModel.DebtPayments {
		if val, exists := fuzzyResult.VariableValues[id]; exists {
			extraPayment := val - constraint.MinimumPayment
			if extraPayment < 0 {
				extraPayment = 0
			}
			result.DebtAllocations[id] = domain.DebtPayment{
				TotalPayment:   val,
				MinimumPayment: constraint.MinimumPayment,
				ExtraPayment:   extraPayment,
			}
			result.TotalAllocated += val
		}
	}

	result.Surplus = gp.constraintModel.TotalIncome - result.TotalAllocated

	// Use weighted membership as feasibility score for fuzzy GP (convert from [0,1] to [0,100])
	result.FeasibilityScore = fuzzyResult.WeightedMembership * 100
}

// solveDebtHeuristic calculates debt allocations using heuristic (debt không được đưa vào LP)
func (gp *GoalProgrammingSolver) solveDebtHeuristic() (*domain.AllocationResult, error) {
	result := &domain.AllocationResult{
		DebtAllocations: make(map[uuid.UUID]domain.DebtPayment),
	}

	// Phase 2: Allocate debt payments (Priority 2 - Critical)
	// FORCE FixedPayment if set - this is from DEBT STRATEGY OUTPUT (base + extra from strategy weights)
	// FixedPayment represents the adjustedPayment calculated from debt strategy, which must be respected
	for debtID, constraint := range gp.constraintModel.DebtPayments {
		paymentAmount := constraint.MinimumPayment
		if constraint.FixedPayment > 0 {
			// FORCE: Use debt strategy output (adjustedPayment = base + extra from strategy)
			// This ensures the heuristic respects the debt strategy calculation
			paymentAmount = constraint.FixedPayment
		}

		payment := domain.DebtPayment{
			TotalPayment:   paymentAmount,
			MinimumPayment: constraint.MinimumPayment,
			ExtraPayment:   paymentAmount - constraint.MinimumPayment,
		}
		if payment.ExtraPayment < 0 {
			payment.ExtraPayment = 0
		}
		result.DebtAllocations[debtID] = payment
	}

	return result, nil
}

// solveHeuristic is the main heuristic solver (switched from Fuzzy GP fallback)
func (gp *GoalProgrammingSolver) solveHeuristic() (*domain.AllocationResult, error) {
	result := &domain.AllocationResult{
		CategoryAllocations: make(map[uuid.UUID]float64),
		GoalAllocations:     make(map[uuid.UUID]float64),
		DebtAllocations:     make(map[uuid.UUID]domain.DebtPayment),
		FeasibilityScore:    100.0,
		AchievedGoals:       make([]string, 0),
		UnachievedGoals:     make([]string, 0),
		SolverType:          "heuristic",
	}

	remainingIncome := gp.constraintModel.TotalIncome

	// Phase 1: Allocate mandatory expenses (Priority 1 - Must Satisfy)
	for categoryID, constraint := range gp.constraintModel.MandatoryExpenses {
		amount := constraint.Minimum
		result.CategoryAllocations[categoryID] = amount
		remainingIncome -= amount
	}

	// Phase 2: Allocate debt payments (Priority 2 - Critical)
	// FORCE FixedPayment if set - this is from DEBT STRATEGY OUTPUT (base + extra from strategy weights)
	// FixedPayment represents the adjustedPayment calculated from debt strategy, which must be respected
	for debtID, constraint := range gp.constraintModel.DebtPayments {
		paymentAmount := constraint.MinimumPayment
		if constraint.FixedPayment > 0 {
			// FORCE: Use debt strategy output (adjustedPayment = base + extra from strategy)
			// This ensures the heuristic respects the debt strategy calculation
			paymentAmount = constraint.FixedPayment
		}

		payment := domain.DebtPayment{
			TotalPayment:   paymentAmount,
			MinimumPayment: constraint.MinimumPayment,
			ExtraPayment:   paymentAmount - constraint.MinimumPayment,
		}
		if payment.ExtraPayment < 0 {
			payment.ExtraPayment = 0
		}
		result.DebtAllocations[debtID] = payment
		remainingIncome -= paymentAmount
	}

	// Check if we've gone negative (infeasible)
	if remainingIncome < 0 {
		result.FeasibilityScore = 0
		result.Surplus = remainingIncome
		return result, nil
	}

	// Phase 3: Allocate surplus based on scenario parameters
	gp.allocateSurplusHeuristic(result, remainingIncome)

	// Bước cuối: Đảm bảo Surplus = 0 bằng cách điều chỉnh allocations
	gp.normalizeSurplus(result)

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

// normalizeSurplus đảm bảo Surplus = 0 bằng cách điều chỉnh allocations
func (gp *GoalProgrammingSolver) normalizeSurplus(result *domain.AllocationResult) {
	// Tính lại total allocated chính xác
	totalAllocated := 0.0

	// Sum category allocations (mandatory + flexible)
	for _, amount := range result.CategoryAllocations {
		totalAllocated += amount
	}

	// Sum goal allocations
	for _, amount := range result.GoalAllocations {
		totalAllocated += amount
	}

	// Sum debt allocations
	for _, payment := range result.DebtAllocations {
		totalAllocated += payment.TotalPayment
	}

	// Tính surplus hiện tại
	currentSurplus := gp.constraintModel.TotalIncome - totalAllocated

	// Luôn đưa Surplus về 0 bằng cách điều chỉnh allocations
	// Nếu surplus < 0: cần giảm allocations (ưu tiên goal extra, flexible extra)
	if currentSurplus < 0 {
		deficit := -currentSurplus
		gp.reduceAllocations(result, deficit)
	}

	// Nếu surplus > 0: tăng goals trước (theo tỷ lệ), sau đó mới tăng flexible expenses
	if currentSurplus > 0 {
		// Bước 1: Tăng goals theo tỷ lệ hiện tại (proportional increase)
		totalGoalAllocation := 0.0
		for _, amount := range result.GoalAllocations {
			totalGoalAllocation += amount
		}

		if totalGoalAllocation > 0 {
			// Tăng goals theo tỷ lệ hiện tại
			for goalID, amount := range result.GoalAllocations {
				if currentSurplus <= 0.01 {
					break
				}
				if amount > 0 {
					goalConstraint := gp.constraintModel.GoalTargets[goalID]
					// Tính tỷ lệ của goal này trong tổng goals
					ratio := amount / totalGoalAllocation
					// Tăng theo tỷ lệ, nhưng không vượt quá RemainingAmount
					increase := min(currentSurplus*ratio, goalConstraint.RemainingAmount-amount)
					if increase > 0 {
						result.GoalAllocations[goalID] = amount + increase
						currentSurplus -= increase
					}
				}
			}
		}

		// Bước 2: Nếu vẫn còn surplus, tăng flexible expenses
		if currentSurplus > 0.01 {
			gp.increaseFlexibleExpenses(result, currentSurplus)
		}
	}

	// Tính lại sau khi điều chỉnh và đảm bảo = 0
	totalAllocated = 0.0
	for _, amount := range result.CategoryAllocations {
		totalAllocated += amount
	}
	for _, amount := range result.GoalAllocations {
		totalAllocated += amount
	}
	for _, payment := range result.DebtAllocations {
		totalAllocated += payment.TotalPayment
	}

	// Nếu vẫn còn lệch, điều chỉnh lần cuối cho đến khi = 0
	finalSurplus := gp.constraintModel.TotalIncome - totalAllocated
	for abs(finalSurplus) > 0.01 {
		if finalSurplus < 0 {
			// Còn deficit: giảm tiếp
			deficit := -finalSurplus

			// Ưu tiên 1: Giảm flexible expenses (phần > minimum)
			reduced := false
			for categoryID, amount := range result.CategoryAllocations {
				if constraint, ok := gp.constraintModel.FlexibleExpenses[categoryID]; ok && amount > constraint.Minimum {
					reduction := min(amount-constraint.Minimum, deficit)
					if reduction > 0 {
						result.CategoryAllocations[categoryID] = amount - reduction
						totalAllocated -= reduction
						finalSurplus += reduction
						deficit -= reduction
						reduced = true
						if abs(finalSurplus) <= 0.01 {
							break
						}
					}
				}
			}

			// Ưu tiên 2: Giảm goals (nếu flexible không đủ)
			if !reduced || abs(finalSurplus) > 0.01 {
				for goalID, amount := range result.GoalAllocations {
					if abs(finalSurplus) <= 0.01 {
						break
					}
					if amount > 0 {
						reduction := min(amount, -finalSurplus)
						if reduction > 0 {
							result.GoalAllocations[goalID] = amount - reduction
							totalAllocated -= reduction
							finalSurplus += reduction
						}
					}
				}
			}

			// Nếu vẫn còn, giảm flexible xuống dưới minimum (nếu có thể)
			if abs(finalSurplus) > 0.01 {
				for categoryID, amount := range result.CategoryAllocations {
					if abs(finalSurplus) <= 0.01 {
						break
					}
					if _, ok := gp.constraintModel.FlexibleExpenses[categoryID]; ok && amount > 0 {
						reduction := min(amount, -finalSurplus)
						if reduction > 0 {
							result.CategoryAllocations[categoryID] = amount - reduction
							totalAllocated -= reduction
							finalSurplus += reduction
						}
					}
				}
			}
		} else if finalSurplus > 0 {
			// Còn surplus: tăng goals trước (theo tỷ lệ), sau đó mới tăng flexible expenses
			// Bước 1: Tăng goals theo tỷ lệ hiện tại (proportional increase)
			totalGoalAllocation := 0.0
			for _, amount := range result.GoalAllocations {
				totalGoalAllocation += amount
			}

			if totalGoalAllocation > 0 {
				// Tăng goals theo tỷ lệ hiện tại
				for goalID, amount := range result.GoalAllocations {
					if abs(finalSurplus) <= 0.01 {
						break
					}
					if amount > 0 {
						goalConstraint := gp.constraintModel.GoalTargets[goalID]
						// Tính tỷ lệ của goal này trong tổng goals
						ratio := amount / totalGoalAllocation
						// Tăng theo tỷ lệ, nhưng không vượt quá RemainingAmount
						increase := min(finalSurplus*ratio, goalConstraint.RemainingAmount-amount)
						if increase > 0 {
							result.GoalAllocations[goalID] = amount + increase
							totalAllocated += increase
							finalSurplus -= increase
						}
					}
				}
			}

			// Bước 2: Nếu vẫn còn surplus, tăng flexible expenses (tôn trọng FlexibleSpendingLevel)
			if abs(finalSurplus) > 0.01 {
				level := gp.params.FlexibleSpendingLevel
				for categoryID, constraint := range gp.constraintModel.FlexibleExpenses {
					if abs(finalSurplus) <= 0.01 {
						break
					}
					current := result.CategoryAllocations[categoryID]
					if current == 0 {
						current = constraint.Minimum
					}

					// Calculate target based on FlexibleSpendingLevel
					var target float64
					if constraint.Maximum > 0 {
						rangeAmount := constraint.Maximum - constraint.Minimum
						target = constraint.Minimum + (rangeAmount * level)
					} else {
						// No maximum: target = minimum + (surplus * level)
						target = constraint.Minimum + (finalSurplus * level)
					}

					// Only increase up to target (respect FlexibleSpendingLevel)
					maxAmount := target
					if constraint.Maximum > 0 && target > constraint.Maximum {
						maxAmount = constraint.Maximum
					}

					if current < maxAmount {
						increase := min(finalSurplus, maxAmount-current)
						if increase > 0 {
							result.CategoryAllocations[categoryID] = current + increase
							totalAllocated += increase
							finalSurplus -= increase
						}
					}
				}
			}
		}

		// Tính lại để kiểm tra
		totalAllocated = 0.0
		for _, amount := range result.CategoryAllocations {
			totalAllocated += amount
		}
		for _, amount := range result.GoalAllocations {
			totalAllocated += amount
		}
		for _, payment := range result.DebtAllocations {
			totalAllocated += payment.TotalPayment
		}
		finalSurplus = gp.constraintModel.TotalIncome - totalAllocated
	}

	// REMOVED: TEST code that force reduced goal allocation
	// Goals should be increased proportionally when there's surplus

	result.TotalAllocated = totalAllocated
	result.Surplus = 0 // Luôn = 0 sau khi normalize
}

// reduceAllocations giảm allocations để bù deficit.
// KHÔNG giảm debt. Ưu tiên: 1) goal extra (nhiều nhất trước), 2) flexible extra (nhiều nhất trước), 3) goals còn lại.
func (gp *GoalProgrammingSolver) reduceAllocations(result *domain.AllocationResult, deficit float64) {
	remainingDeficit := deficit

	// Bước 1: Giảm goal extra (phần vượt SuggestedContribution), ưu tiên extra nhiều nhất trước, rồi thứ 2
	type goalExtra struct {
		goalID uuid.UUID
		amount float64
		extra  float64
	}
	goalsWithExtra := make([]goalExtra, 0)
	for goalID, amount := range result.GoalAllocations {
		if constraint, ok := gp.constraintModel.GoalTargets[goalID]; ok && amount > 0 {
			extra := amount - constraint.SuggestedContribution
			if extra < 0 {
				extra = 0
			}
			if extra > 0 {
				goalsWithExtra = append(goalsWithExtra, goalExtra{goalID, amount, extra})
			}
		}
	}
	// Sort giảm dần theo extra (nhiều nhất trước, thứ 2, ...)
	for i := 0; i < len(goalsWithExtra)-1; i++ {
		for j := i + 1; j < len(goalsWithExtra); j++ {
			if goalsWithExtra[j].extra > goalsWithExtra[i].extra {
				goalsWithExtra[i], goalsWithExtra[j] = goalsWithExtra[j], goalsWithExtra[i]
			}
		}
	}
	for _, g := range goalsWithExtra {
		if remainingDeficit <= 0 {
			break
		}
		reduction := min(g.extra, remainingDeficit)
		if reduction <= 0 {
			continue
		}
		constraint := gp.constraintModel.GoalTargets[g.goalID]
		newAmount := result.GoalAllocations[g.goalID] - reduction
		if newAmount < constraint.SuggestedContribution {
			newAmount = constraint.SuggestedContribution
			reduction = result.GoalAllocations[g.goalID] - newAmount
		}
		result.GoalAllocations[g.goalID] = newAmount
		remainingDeficit -= reduction
	}

	// Bước 2: Giảm flexible expense extra, ưu tiên extra nhiều nhất trước, rồi thứ 2
	type flexExtra struct {
		categoryID uuid.UUID
		amount     float64
		extra      float64
	}
	flexWithExtra := make([]flexExtra, 0)
	for categoryID, amount := range result.CategoryAllocations {
		if constraint, ok := gp.constraintModel.FlexibleExpenses[categoryID]; ok && amount > constraint.Minimum {
			flexWithExtra = append(flexWithExtra, flexExtra{
				categoryID, amount, amount - constraint.Minimum,
			})
		}
	}
	for i := 0; i < len(flexWithExtra)-1; i++ {
		for j := i + 1; j < len(flexWithExtra); j++ {
			if flexWithExtra[j].extra > flexWithExtra[i].extra {
				flexWithExtra[i], flexWithExtra[j] = flexWithExtra[j], flexWithExtra[i]
			}
		}
	}
	for _, f := range flexWithExtra {
		if remainingDeficit <= 0 {
			break
		}
		reduction := min(f.extra, remainingDeficit)
		if reduction <= 0 {
			continue
		}
		result.CategoryAllocations[f.categoryID] = f.amount - reduction
		remainingDeficit -= reduction
	}

	// Bước 3: Giảm goals còn lại (xuống dưới suggested), ưu tiên amount nhiều nhất trước
	if remainingDeficit <= 0 {
		return
	}
	goalsRemain := make([]goalExtra, 0)
	for goalID, amount := range result.GoalAllocations {
		if _, ok := gp.constraintModel.GoalTargets[goalID]; ok && amount > 0 {
			goalsRemain = append(goalsRemain, goalExtra{goalID, amount, amount})
		}
	}
	for i := 0; i < len(goalsRemain)-1; i++ {
		for j := i + 1; j < len(goalsRemain); j++ {
			if goalsRemain[j].amount > goalsRemain[i].amount {
				goalsRemain[i], goalsRemain[j] = goalsRemain[j], goalsRemain[i]
			}
		}
	}
	for _, g := range goalsRemain {
		if remainingDeficit <= 0 {
			break
		}
		reduction := min(g.amount, remainingDeficit)
		if reduction > 0 {
			result.GoalAllocations[g.goalID] = g.amount - reduction
			remainingDeficit -= reduction
		}
	}

	// Bước 4: Nếu vẫn còn deficit, giảm flexible expenses xuống dưới minimum (nếu có thể)
	if remainingDeficit > 0 {
		for categoryID, amount := range result.CategoryAllocations {
			if _, ok := gp.constraintModel.FlexibleExpenses[categoryID]; ok && amount > 0 {
				reduction := min(amount, remainingDeficit)
				if reduction > 0 {
					result.CategoryAllocations[categoryID] = amount - reduction
					remainingDeficit -= reduction
					if remainingDeficit <= 0 {
						break
					}
				}
			}
		}
	}
}

// increaseFlexibleExpenses tăng flexible expenses để sử dụng hết surplus.
// Tôn trọng FlexibleSpendingLevel: chỉ tăng đến target level, không vượt quá.
func (gp *GoalProgrammingSolver) increaseFlexibleExpenses(result *domain.AllocationResult, surplus float64) {
	remainingSurplus := surplus
	level := gp.params.FlexibleSpendingLevel

	type flexRoom struct {
		categoryID uuid.UUID
		current    float64
		target     float64 // Target based on FlexibleSpendingLevel
		maximum    float64
		extra      float64
	}
	flexibles := make([]flexRoom, 0)
	for categoryID, constraint := range gp.constraintModel.FlexibleExpenses {
		current := result.CategoryAllocations[categoryID]
		if current == 0 {
			current = constraint.Minimum
		}

		// Calculate target based on FlexibleSpendingLevel
		var target float64
		if constraint.Maximum > 0 {
			rangeAmount := constraint.Maximum - constraint.Minimum
			target = constraint.Minimum + (rangeAmount * level)
		} else {
			// If no maximum, target = minimum + (surplus * level)
			target = constraint.Minimum + (remainingSurplus * level)
		}

		maxAmount := constraint.Maximum
		if maxAmount == 0 {
			// No maximum: allow up to target only (respect FlexibleSpendingLevel)
			maxAmount = target
		} else {
			// Has maximum: cap at target (don't exceed FlexibleSpendingLevel)
			if target < maxAmount {
				maxAmount = target
			}
		}

		if current < maxAmount {
			extra := current - constraint.Minimum
			if extra < 0 {
				extra = 0
			}
			flexibles = append(flexibles, flexRoom{
				categoryID: categoryID,
				current:    current,
				target:     target,
				maximum:    maxAmount,
				extra:      extra,
			})
		}
	}

	// Sort giảm dần theo extra (nhiều nhất trước, thứ 2, ...)
	for i := 0; i < len(flexibles)-1; i++ {
		for j := i + 1; j < len(flexibles); j++ {
			if flexibles[j].extra > flexibles[i].extra {
				flexibles[i], flexibles[j] = flexibles[j], flexibles[i]
			}
		}
	}

	for _, f := range flexibles {
		if remainingSurplus <= 0 {
			break
		}
		// Only increase up to target (respect FlexibleSpendingLevel)
		increase := min(f.maximum-f.current, remainingSurplus)
		if increase > 0 {
			result.CategoryAllocations[f.categoryID] = f.current + increase
			remainingSurplus -= increase
		}
	}
}

// abs returns absolute value
func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

// roundAllocations làm tròn các allocations đến hàng 100,000
func (gp *GoalProgrammingSolver) roundAllocations(result *domain.AllocationResult) {
	// Làm tròn category allocations đến hàng 100,000
	for categoryID, amount := range result.CategoryAllocations {
		result.CategoryAllocations[categoryID] = math.Round(amount/100000) * 100000
	}

	// Làm tròn goal allocations đến hàng 100,000
	for goalID, amount := range result.GoalAllocations {
		result.GoalAllocations[goalID] = math.Round(amount/100000) * 100000
	}

	// Làm tròn debt allocations đến hàng 100,000
	for debtID, payment := range result.DebtAllocations {
		result.DebtAllocations[debtID] = domain.DebtPayment{
			TotalPayment:   math.Round(payment.TotalPayment/100000) * 100000,
			MinimumPayment: math.Round(payment.MinimumPayment/100000) * 100000,
			ExtraPayment:   math.Round(payment.ExtraPayment/100000) * 100000,
		}
	}
}

// reduceSurplusByDeducting trừ surplus về 0 bằng cách trừ 1 lần từ các phần được chia nhiều
func (gp *GoalProgrammingSolver) reduceSurplusByDeducting(result *domain.AllocationResult, surplus float64) {
	if surplus <= 100 {
		return // Quá nhỏ, không cần trừ
	}

	remainingSurplus := surplus

	// Tạo danh sách các allocations để sắp xếp (ưu tiên trừ từ phần lớn nhất)
	type allocationItem struct {
		typ      string // "category", "goal", "debt"
		id       uuid.UUID
		amount   float64
		minValue float64 // Minimum để không trừ xuống dưới
	}

	items := make([]allocationItem, 0)

	// Collect category allocations (flexible expenses only - không trừ mandatory)
	for categoryID, amount := range result.CategoryAllocations {
		if constraint, ok := gp.constraintModel.FlexibleExpenses[categoryID]; ok {
			items = append(items, allocationItem{
				typ:      "category",
				id:       categoryID,
				amount:   amount,
				minValue: constraint.Minimum,
			})
		}
	}

	// Collect goal allocations
	for goalID, amount := range result.GoalAllocations {
		if amount > 0 {
			items = append(items, allocationItem{
				typ:      "goal",
				id:       goalID,
				amount:   amount,
				minValue: 0, // Goals có thể về 0
			})
		}
	}

	// Collect debt extra payments (chỉ trừ extra payment, không trừ minimum)
	for debtID, payment := range result.DebtAllocations {
		if payment.ExtraPayment > 0 {
			items = append(items, allocationItem{
				typ:      "debt",
				id:       debtID,
				amount:   payment.ExtraPayment,
				minValue: 0, // Chỉ trừ extra payment
			})
		}
	}

	// Sắp xếp theo amount giảm dần (phần lớn nhất trước)
	for i := 0; i < len(items)-1; i++ {
		for j := i + 1; j < len(items); j++ {
			if items[i].amount < items[j].amount {
				items[i], items[j] = items[j], items[i]
			}
		}
	}

	// Trừ surplus từ các phần lớn nhất (1 vòng lặp duy nhất)
	for i := range items {
		if remainingSurplus <= 0.01 {
			break
		}
		if items[i].amount > items[i].minValue {
			// Trừ tối đa có thể (không vượt quá phần có thể trừ được)
			deduction := math.Min(remainingSurplus, items[i].amount-items[i].minValue)
			if deduction > 0 {
				items[i].amount -= deduction
				remainingSurplus -= deduction

				// Update result ngay lập tức
				switch items[i].typ {
				case "category":
					result.CategoryAllocations[items[i].id] = items[i].amount
				case "goal":
					result.GoalAllocations[items[i].id] = items[i].amount
				case "debt":
					payment := result.DebtAllocations[items[i].id]
					payment.ExtraPayment = items[i].amount
					payment.TotalPayment = payment.MinimumPayment + payment.ExtraPayment
					result.DebtAllocations[items[i].id] = payment
				}
			}
		}
	}
}

// min returns the minimum of two float64 values
func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
