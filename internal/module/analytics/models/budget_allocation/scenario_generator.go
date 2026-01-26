package budget_allocation

import (
	"personalfinancedss/internal/module/analytics/budget_allocation/domain"
	"personalfinancedss/internal/module/analytics/models/budget_allocation/constraint"
	"personalfinancedss/internal/module/analytics/models/budget_allocation/solver"

	"github.com/google/uuid"
)

// ScenarioGenerator generates multiple allocation scenarios
type ScenarioGenerator struct {
	constraintBuilder *constraint.ConstraintBuilder
}

// NewScenarioGenerator creates a new scenario generator
func NewScenarioGenerator() *ScenarioGenerator {
	return &ScenarioGenerator{
		constraintBuilder: constraint.NewConstraintBuilder(),
	}
}

// GenerateScenarios generates safe and balanced scenarios (only 2 scenarios)
func (sg *ScenarioGenerator) GenerateScenarios(
	model *domain.ConstraintModel,
	categoryNames map[uuid.UUID]string, // CategoryID -> name
	customParams map[string]*domain.ScenarioParameters, // Optional: custom parameters per scenario type
) ([]domain.AllocationScenario, error) {

	scenarios := make([]domain.AllocationScenario, 0, 2)

	// Check feasibility first
	isFeasible, deficit := sg.constraintBuilder.CheckFeasibility(model)

	// Generate safe scenario
	var safeParams *domain.ScenarioParameters
	if cp, ok := customParams["safe"]; ok {
		safeParams = cp
	}
	safe := sg.GenerateSafeScenarioWithParams(model, categoryNames, safeParams)
	if !isFeasible {
		sg.addDeficitWarnings(&safe, deficit, model)
	}
	scenarios = append(scenarios, safe)

	// Generate balanced scenario
	var balancedParams *domain.ScenarioParameters
	if cp, ok := customParams["balanced"]; ok {
		balancedParams = cp
	}
	balanced := sg.GenerateBalancedScenarioWithParams(model, categoryNames, balancedParams)
	if !isFeasible {
		sg.addDeficitWarnings(&balanced, deficit, model)
	}
	scenarios = append(scenarios, balanced)

	return scenarios, nil
}

// GenerateSafeScenario generates a safe allocation (default parameters)
func (sg *ScenarioGenerator) GenerateSafeScenario(
	model *domain.ConstraintModel,
	categoryNames map[uuid.UUID]string,
) domain.AllocationScenario {
	return sg.GenerateSafeScenarioWithParams(model, categoryNames, nil)
}

// GenerateSafeScenarioWithParams generates a safe allocation with optional custom parameters
func (sg *ScenarioGenerator) GenerateSafeScenarioWithParams(
	model *domain.ConstraintModel,
	categoryNames map[uuid.UUID]string,
	customParams *domain.ScenarioParameters,
) domain.AllocationScenario {

	// Default parameters for Safe scenario
	// Safe: Ưu tiên goals hơn, ít flexible spending (an toàn = tập trung vào goals)
	params := domain.ScenarioParameters{
		ScenarioType:           domain.ScenarioSafe,
		GoalContributionFactor: 1.2, // 120% of suggested - nhiều hơn cho goals
		FlexibleSpendingLevel:  0.3, // Low spending - ít flexible hơn Balanced nhưng không phải 0
		SurplusAllocation: domain.SurplusAllocation{
			EmergencyFundPercent: 0.0,  // 0% - Emergency fund được ưu tiên tự động qua weight trong Fuzzy GP
			DebtExtraPercent:     0.0,  // Not used (debt payments from heuristic)
			GoalsPercent:         0.30, // 30% to goals (chỉ dùng trong heuristic fallback)
			FlexiblePercent:      0.05, // 5% to flexible spending (chỉ dùng trong heuristic fallback)
		},
	}

	// Override with custom parameters if provided
	if customParams != nil {
		if customParams.GoalContributionFactor > 0 {
			params.GoalContributionFactor = customParams.GoalContributionFactor
		}
		if customParams.FlexibleSpendingLevel >= 0 {
			params.FlexibleSpendingLevel = customParams.FlexibleSpendingLevel
		}
		if customParams.SurplusAllocation.EmergencyFundPercent > 0 {
			params.SurplusAllocation.EmergencyFundPercent = customParams.SurplusAllocation.EmergencyFundPercent
		}
		if customParams.SurplusAllocation.GoalsPercent > 0 {
			params.SurplusAllocation.GoalsPercent = customParams.SurplusAllocation.GoalsPercent
		}
		if customParams.SurplusAllocation.FlexiblePercent > 0 {
			params.SurplusAllocation.FlexiblePercent = customParams.SurplusAllocation.FlexiblePercent
		}
	}

	solver := solver.NewGoalProgrammingSolver(model, params)
	result, _ := solver.SolveFuzzy()

	scenario := sg.buildScenario(domain.ScenarioSafe, result, model, categoryNames)
	sg.addSafeWarnings(&scenario, model)

	return scenario
}

// GenerateBalancedScenario generates a balanced allocation (default parameters)
func (sg *ScenarioGenerator) GenerateBalancedScenario(
	model *domain.ConstraintModel,
	categoryNames map[uuid.UUID]string,
) domain.AllocationScenario {
	return sg.GenerateBalancedScenarioWithParams(model, categoryNames, nil)
}

// GenerateBalancedScenarioWithParams generates a balanced allocation with optional custom parameters
func (sg *ScenarioGenerator) GenerateBalancedScenarioWithParams(
	model *domain.ConstraintModel,
	categoryNames map[uuid.UUID]string,
	customParams *domain.ScenarioParameters,
) domain.AllocationScenario {

	// Default parameters for Balanced scenario
	// Balanced: Ưu tiên flexible spending hơn, ít goals (cân bằng = linh hoạt chi tiêu)
	params := domain.ScenarioParameters{
		ScenarioType:           domain.ScenarioBalanced,
		GoalContributionFactor: 0.7, // 70% of suggested - ít hơn cho goals
		FlexibleSpendingLevel:  0.7, // Higher spending - nhiều flexible hơn
		SurplusAllocation: domain.SurplusAllocation{
			EmergencyFundPercent: 0.0,  // 0% - Emergency fund được ưu tiên tự động qua weight trong Fuzzy GP
			DebtExtraPercent:     0.0,  // Not used (debt payments from heuristic)
			GoalsPercent:         0.10, // 10% to goals (chỉ dùng trong heuristic fallback)
			FlexiblePercent:      0.20, // 20% to flexible spending (chỉ dùng trong heuristic fallback)
		},
	}

	// Override with custom parameters if provided
	if customParams != nil {
		if customParams.GoalContributionFactor > 0 {
			params.GoalContributionFactor = customParams.GoalContributionFactor
		}
		if customParams.FlexibleSpendingLevel >= 0 {
			params.FlexibleSpendingLevel = customParams.FlexibleSpendingLevel
		}
		if customParams.SurplusAllocation.EmergencyFundPercent > 0 {
			params.SurplusAllocation.EmergencyFundPercent = customParams.SurplusAllocation.EmergencyFundPercent
		}
		if customParams.SurplusAllocation.GoalsPercent > 0 {
			params.SurplusAllocation.GoalsPercent = customParams.SurplusAllocation.GoalsPercent
		}
		if customParams.SurplusAllocation.FlexiblePercent > 0 {
			params.SurplusAllocation.FlexiblePercent = customParams.SurplusAllocation.FlexiblePercent
		}
	}

	solver := solver.NewGoalProgrammingSolver(model, params)
	result, _ := solver.SolveFuzzy()

	scenario := sg.buildScenario(domain.ScenarioBalanced, result, model, categoryNames)
	sg.addBalancedWarnings(&scenario, model)

	return scenario
}

// buildScenario converts allocation result to domain scenario
func (sg *ScenarioGenerator) buildScenario(
	scenarioType domain.ScenarioType,
	result *domain.AllocationResult,
	model *domain.ConstraintModel,
	categoryNames map[uuid.UUID]string,
) domain.AllocationScenario {

	scenario := domain.NewAllocationScenario(scenarioType)
	scenario.FeasibilityScore = result.FeasibilityScore

	// Build goal allocations
	for goalID, amount := range result.GoalAllocations {
		goalConstraint := model.GoalTargets[goalID]

		// Use raw amount from GP (no rounding)
		percentageOfTarget := 0.0
		if goalConstraint.RemainingAmount > 0 {
			percentageOfTarget = (amount / goalConstraint.RemainingAmount) * 100
		}

		scenario.GoalAllocations = append(scenario.GoalAllocations, domain.GoalAllocation{
			GoalID:                goalID,
			GoalName:              goalConstraint.GoalName,
			Amount:                amount, // Raw from GP
			SuggestedContribution: goalConstraint.SuggestedContribution,
			Priority:              goalConstraint.Priority,
			PercentageOfTarget:    percentageOfTarget,
		})
	}

	// Build debt allocations
	for debtID, payment := range result.DebtAllocations {
		debtConstraint := model.DebtPayments[debtID]

		// Use raw amounts from GP (no rounding)
		// Ensure Total >= Minimum
		totalPayment := payment.TotalPayment
		if totalPayment < payment.MinimumPayment {
			totalPayment = payment.MinimumPayment // Respect hard constraint
		}

		interestSavings := sg.calculateInterestSavings(
			payment.ExtraPayment,
			debtConstraint.InterestRate,
			debtConstraint.CurrentBalance,
		)

		scenario.DebtAllocations = append(scenario.DebtAllocations, domain.DebtAllocation{
			DebtID:          debtID,
			DebtName:        debtConstraint.DebtName,
			Amount:          totalPayment, // Raw from GP
			MinimumPayment:  payment.MinimumPayment,
			ExtraPayment:    payment.ExtraPayment, // Raw from GP
			InterestRate:    debtConstraint.InterestRate,
			InterestSavings: interestSavings,
		})
	}

	// Build category allocations (Flexible/Mandatory)
	for categoryID, amount := range result.CategoryAllocations {
		var constraint domain.CategoryConstraint
		var isFlexible bool
		if c, ok := model.MandatoryExpenses[categoryID]; ok {
			constraint = c
			isFlexible = false
		} else if c, ok := model.FlexibleExpenses[categoryID]; ok {
			constraint = c
			isFlexible = true
		}

		categoryName := categoryNames[categoryID]
		if categoryName == "" {
			categoryName = "Unknown Category"
		}

		// Use raw amount from GP (no rounding)
		// Ensure mandatory expenses meet minimum
		finalAmount := amount
		if !isFlexible && finalAmount < constraint.Minimum {
			finalAmount = constraint.Minimum // Respect hard constraint
		}

		scenario.CategoryAllocations = append(scenario.CategoryAllocations, domain.CategoryAllocation{
			CategoryID:   categoryID,
			CategoryName: categoryName,
			Amount:       finalAmount, // Raw from GP
			Minimum:      constraint.Minimum,
			Maximum:      constraint.Maximum,
			IsFlexible:   isFlexible,
			Priority:     constraint.Priority,
		})
	}

	// Calculate summary
	scenario.CalculateSummary(model.TotalIncome)

	return *scenario
}

// calculateInterestSavings estimates interest savings from extra debt payment
func (sg *ScenarioGenerator) calculateInterestSavings(extraPayment, annualRate, _ float64) float64 {
	if extraPayment <= 0 || annualRate <= 0 {
		return 0
	}

	// Simple estimate: (extra payment) * (annual rate / 12)
	// This is a rough approximation of one month's interest savings
	monthlyRate := annualRate / 100.0 / 12.0
	return extraPayment * monthlyRate
}

// addDeficitWarnings adds warnings when income is insufficient
func (sg *ScenarioGenerator) addDeficitWarnings(scenario *domain.AllocationScenario, deficit float64, model *domain.ConstraintModel) {
	suggestions := sg.constraintBuilder.GetSuggestionsForDeficit(model, deficit)

	scenario.AddWarning(
		domain.SeverityCritical,
		"income",
		"Insufficient income to cover mandatory expenses and minimum debt payments",
		suggestions...,
	)

	scenario.FeasibilityScore = 0
}

// addSafeWarnings adds scenario-specific warnings
func (sg *ScenarioGenerator) addSafeWarnings(scenario *domain.AllocationScenario, model *domain.ConstraintModel) {
	// Check if emergency fund goals exist
	hasEmergencyFund := false
	for _, goal := range model.GoalTargets {
		if goal.GoalType == "emergency" {
			hasEmergencyFund = true
			break
		}
	}

	if !hasEmergencyFund {
		scenario.AddWarning(
			domain.SeverityWarning,
			"goal",
			"No emergency fund goal detected",
			"Consider creating an emergency fund goal for 3-6 months of expenses",
		)
	}

	// Check savings rate
	if scenario.Summary.SavingsRate < 10 {
		scenario.AddWarning(
			domain.SeverityInfo,
			"goal",
			"Low savings rate in safe scenario",
			"Your current income may be tight. Consider increasing income or reducing expenses.",
		)
	}
}

// addBalancedWarnings adds warnings for balanced scenario
func (sg *ScenarioGenerator) addBalancedWarnings(scenario *domain.AllocationScenario, model *domain.ConstraintModel) {
	// Check if any high-interest debts exist
	for _, debt := range model.DebtPayments {
		if debt.InterestRate >= 15 {
			scenario.AddWarning(
				domain.SeverityWarning,
				"debt",
				"High-interest debt detected",
				"Consider paying off high-interest debts faster to save on interest",
			)
			break
		}
	}
}
