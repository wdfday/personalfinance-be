package budget_allocation

import (
	"math"
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

// GenerateScenarios generates conservative, balanced, and aggressive scenarios
func (sg *ScenarioGenerator) GenerateScenarios(
	model *domain.ConstraintModel,
	categoryNames map[uuid.UUID]string, // CategoryID -> name
) ([]domain.AllocationScenario, error) {

	scenarios := make([]domain.AllocationScenario, 0, 3)

	// Check feasibility first
	isFeasible, deficit := sg.constraintBuilder.CheckFeasibility(model)

	// Generate conservative scenario
	conservative := sg.GenerateConservativeScenario(model, categoryNames)
	if !isFeasible {
		sg.addDeficitWarnings(&conservative, deficit, model)
	}
	scenarios = append(scenarios, conservative)

	// Generate balanced scenario
	balanced := sg.GenerateBalancedScenario(model, categoryNames)
	if !isFeasible {
		sg.addDeficitWarnings(&balanced, deficit, model)
	}
	scenarios = append(scenarios, balanced)

	// Generate aggressive scenario
	aggressive := sg.GenerateAggressiveScenario(model, categoryNames)
	if !isFeasible {
		sg.addDeficitWarnings(&aggressive, deficit, model)
	}
	scenarios = append(scenarios, aggressive)

	return scenarios, nil
}

// GenerateConservativeScenario generates a conservative allocation
func (sg *ScenarioGenerator) GenerateConservativeScenario(
	model *domain.ConstraintModel,
	categoryNames map[uuid.UUID]string,
) domain.AllocationScenario {

	params := domain.ScenarioParameters{
		ScenarioType:           domain.ScenarioConservative,
		GoalContributionFactor: 0.7, // 70% of suggested
		FlexibleSpendingLevel:  0.0, // Minimum spending
		SurplusAllocation: domain.SurplusAllocation{
			EmergencyFundPercent: 0.60, // 60% to emergency fund
			DebtExtraPercent:     0.30, // 30% to extra debt payments
			GoalsPercent:         0.05, // 5% to goals
			FlexiblePercent:      0.05, // 5% to flexible spending
		},
	}

	solver := solver.NewGoalProgrammingSolver(model, params)
	result, _ := solver.SolveMeta()

	scenario := sg.buildScenario(domain.ScenarioConservative, result, model, categoryNames)
	sg.addConservativeWarnings(&scenario, model)

	return scenario
}

// GenerateBalancedScenario generates a balanced allocation
func (sg *ScenarioGenerator) GenerateBalancedScenario(
	model *domain.ConstraintModel,
	categoryNames map[uuid.UUID]string,
) domain.AllocationScenario {

	params := domain.ScenarioParameters{
		ScenarioType:           domain.ScenarioBalanced,
		GoalContributionFactor: 1.0, // 100% of suggested
		FlexibleSpendingLevel:  0.5, // Mid-range spending
		SurplusAllocation: domain.SurplusAllocation{
			EmergencyFundPercent: 0.40, // 40% to emergency fund
			DebtExtraPercent:     0.30, // 30% to extra debt payments
			GoalsPercent:         0.20, // 20% to goals
			FlexiblePercent:      0.10, // 10% to flexible spending
		},
	}

	solver := solver.NewGoalProgrammingSolver(model, params)
	result, _ := solver.SolveMeta()

	scenario := sg.buildScenario(domain.ScenarioBalanced, result, model, categoryNames)
	sg.addBalancedWarnings(&scenario, model)

	return scenario
}

// GenerateAggressiveScenario generates an aggressive allocation
func (sg *ScenarioGenerator) GenerateAggressiveScenario(
	model *domain.ConstraintModel,
	categoryNames map[uuid.UUID]string,
) domain.AllocationScenario {

	params := domain.ScenarioParameters{
		ScenarioType:           domain.ScenarioAggressive,
		GoalContributionFactor: 1.3, // 130% of suggested
		FlexibleSpendingLevel:  1.0, // Maximum spending
		SurplusAllocation: domain.SurplusAllocation{
			EmergencyFundPercent: 0.25, // 25% to emergency fund
			DebtExtraPercent:     0.25, // 25% to extra debt payments
			GoalsPercent:         0.40, // 40% to goals
			FlexiblePercent:      0.10, // 10% to flexible spending
		},
	}

	solver := solver.NewGoalProgrammingSolver(model, params)
	result, _ := solver.SolveMeta()

	scenario := sg.buildScenario(domain.ScenarioAggressive, result, model, categoryNames)
	sg.addAggressiveWarnings(&scenario, model)

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

		// Smart round amount for cleaner presentation
		roundedAmount := sg.smartRound(amount)

		percentageOfTarget := 0.0
		if goalConstraint.RemainingAmount > 0 {
			percentageOfTarget = (roundedAmount / goalConstraint.RemainingAmount) * 100
		}

		scenario.GoalAllocations = append(scenario.GoalAllocations, domain.GoalAllocation{
			GoalID:                goalID,
			GoalName:              goalConstraint.GoalName,
			Amount:                roundedAmount,
			SuggestedContribution: goalConstraint.SuggestedContribution,
			Priority:              goalConstraint.Priority,
			PercentageOfTarget:    percentageOfTarget,
		})
	}

	// Build debt allocations
	for debtID, payment := range result.DebtAllocations {
		debtConstraint := model.DebtPayments[debtID]

		// Round payments
		roundedTotal := sg.smartRound(payment.TotalPayment)
		roundedExtra := sg.smartRound(payment.ExtraPayment)

		// Ensure rounding consistency
		if roundedExtra > 0 {
			// Re-calculate total based on min + extra to be safe, or just trust the rounded total?
			// Best to respect minimum constraint exactly? No, minimum might not be round.
			// Let's just use the rounded values but ensure Total >= Minimum
			if roundedTotal < payment.MinimumPayment {
				roundedTotal = payment.MinimumPayment // Respect hard constraint
			}
		}

		interestSavings := sg.calculateInterestSavings(
			roundedExtra,
			debtConstraint.InterestRate,
			debtConstraint.CurrentBalance,
		)

		scenario.DebtAllocations = append(scenario.DebtAllocations, domain.DebtAllocation{
			DebtID:          debtID,
			DebtName:        debtConstraint.DebtName,
			Amount:          roundedTotal,
			MinimumPayment:  payment.MinimumPayment,
			ExtraPayment:    roundedExtra,
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

		roundedAmount := sg.smartRound(amount)
		if !isFlexible && roundedAmount < constraint.Minimum {
			// Use RoundUp for mandatory defaults to ensure it's both enough AND round
			roundedAmount = sg.smartRoundUp(constraint.Minimum)
		}

		scenario.CategoryAllocations = append(scenario.CategoryAllocations, domain.CategoryAllocation{
			CategoryID:   categoryID,
			CategoryName: categoryName,
			Amount:       roundedAmount,
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

// smartRound rounds amounts to "nice" numbers (Aggressive Round Up)
func (sg *ScenarioGenerator) smartRound(val float64) float64 {
	if val <= 0 {
		return 0
	}

	// For VND (large numbers), round UP to nearest 100,000 (User request: "làm tròn lên trăm")
	if val > 100000 {
		return math.Ceil(val/100000) * 100000
	}
	if val > 10000 {
		return math.Ceil(val/10000) * 10000 // Round UP to nearest 10k
	}

	// For USD or small numbers, round UP to nearest integer
	return math.Ceil(val)
}

// smartRoundUp rounds amounts UP to "nice" numbers (Ceiling)
// Used for mandatory expenses to ensure coverage
func (sg *ScenarioGenerator) smartRoundUp(val float64) float64 {
	return sg.smartRound(val)
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

// addConservativeWarnings adds scenario-specific warnings
func (sg *ScenarioGenerator) addConservativeWarnings(scenario *domain.AllocationScenario, model *domain.ConstraintModel) {
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
			"Low savings rate in conservative scenario",
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

// addAggressiveWarnings adds warnings for aggressive scenario
func (sg *ScenarioGenerator) addAggressiveWarnings(scenario *domain.AllocationScenario, model *domain.ConstraintModel) {
	// Warn if flexible spending is at maximum
	if scenario.Summary.FlexibleExpenses > 0 {
		scenario.AddWarning(
			domain.SeverityInfo,
			"expense",
			"Flexible spending at maximum levels",
			"This scenario assumes higher lifestyle spending. Ensure you can sustain this level.",
		)
	}

	// Warn if surplus is very low
	if scenario.Summary.Surplus < 100 {
		scenario.AddWarning(
			domain.SeverityWarning,
			"income",
			"Very low surplus remaining",
			"This aggressive allocation leaves little buffer for unexpected expenses",
		)
	}
}

// DualScenarioResult contains results from both GP solvers for a single scenario
type DualScenarioResult struct {
	ScenarioType       domain.ScenarioType
	PreemptiveScenario domain.AllocationScenario
	WeightedScenario   domain.AllocationScenario
	Comparison         domain.GPComparison
}

// GenerateScenariosWithComparison generates scenarios using both GP solvers for comparison
func (sg *ScenarioGenerator) GenerateScenariosWithComparison(
	model *domain.ConstraintModel,
	categoryNames map[uuid.UUID]string,
) ([]DualScenarioResult, error) {
	results := make([]DualScenarioResult, 0, 3)

	// Check feasibility first
	isFeasible, deficit := sg.constraintBuilder.CheckFeasibility(model)

	// Generate all three scenarios with dual GP
	scenarioConfigs := []struct {
		scenarioType domain.ScenarioType
		params       domain.ScenarioParameters
	}{
		{
			scenarioType: domain.ScenarioConservative,
			params: domain.ScenarioParameters{
				ScenarioType:           domain.ScenarioConservative,
				GoalContributionFactor: 0.7,
				FlexibleSpendingLevel:  0.0,
				SurplusAllocation: domain.SurplusAllocation{
					EmergencyFundPercent: 0.60,
					DebtExtraPercent:     0.30,
					GoalsPercent:         0.05,
					FlexiblePercent:      0.05,
				},
			},
		},
		{
			scenarioType: domain.ScenarioBalanced,
			params: domain.ScenarioParameters{
				ScenarioType:           domain.ScenarioBalanced,
				GoalContributionFactor: 1.0,
				FlexibleSpendingLevel:  0.5,
				SurplusAllocation: domain.SurplusAllocation{
					EmergencyFundPercent: 0.40,
					DebtExtraPercent:     0.30,
					GoalsPercent:         0.20,
					FlexiblePercent:      0.10,
				},
			},
		},
		{
			scenarioType: domain.ScenarioAggressive,
			params: domain.ScenarioParameters{
				ScenarioType:           domain.ScenarioAggressive,
				GoalContributionFactor: 1.3,
				FlexibleSpendingLevel:  1.0,
				SurplusAllocation: domain.SurplusAllocation{
					EmergencyFundPercent: 0.25,
					DebtExtraPercent:     0.25,
					GoalsPercent:         0.40,
					FlexiblePercent:      0.10,
				},
			},
		},
	}

	for _, config := range scenarioConfigs {
		solver := solver.NewGoalProgrammingSolver(model, config.params)
		dualResult, err := solver.SolveDual()
		if err != nil {
			continue
		}

		preemptiveScenario := sg.buildScenario(config.scenarioType, dualResult.PreemptiveResult, model, categoryNames)
		weightedScenario := sg.buildScenario(config.scenarioType, dualResult.WeightedResult, model, categoryNames)

		// Add warnings
		if !isFeasible {
			sg.addDeficitWarnings(&preemptiveScenario, deficit, model)
			sg.addDeficitWarnings(&weightedScenario, deficit, model)
		}

		switch config.scenarioType {
		case domain.ScenarioConservative:
			sg.addConservativeWarnings(&preemptiveScenario, model)
			sg.addConservativeWarnings(&weightedScenario, model)
		case domain.ScenarioBalanced:
			sg.addBalancedWarnings(&preemptiveScenario, model)
			sg.addBalancedWarnings(&weightedScenario, model)
		case domain.ScenarioAggressive:
			sg.addAggressiveWarnings(&preemptiveScenario, model)
			sg.addAggressiveWarnings(&weightedScenario, model)
		}

		results = append(results, DualScenarioResult{
			ScenarioType:       config.scenarioType,
			PreemptiveScenario: preemptiveScenario,
			WeightedScenario:   weightedScenario,
			Comparison:         dualResult.Comparison,
		})
	}

	return results, nil
}

// TripleScenarioResult contains results from all three GP solvers for a single scenario
type TripleScenarioResult struct {
	ScenarioType       domain.ScenarioType
	PreemptiveScenario domain.AllocationScenario
	WeightedScenario   domain.AllocationScenario
	MinmaxScenario     domain.AllocationScenario
	Comparison         domain.TripleGPComparison
}

// GenerateScenariosWithTripleComparison generates scenarios using all three GP solvers
func (sg *ScenarioGenerator) GenerateScenariosWithTripleComparison(
	model *domain.ConstraintModel,
	categoryNames map[uuid.UUID]string,
) ([]TripleScenarioResult, error) {
	results := make([]TripleScenarioResult, 0, 3)

	// Check feasibility first
	isFeasible, deficit := sg.constraintBuilder.CheckFeasibility(model)

	// Scenario configurations
	scenarioConfigs := []struct {
		scenarioType domain.ScenarioType
		params       domain.ScenarioParameters
	}{
		{
			scenarioType: domain.ScenarioConservative,
			params: domain.ScenarioParameters{
				ScenarioType:           domain.ScenarioConservative,
				GoalContributionFactor: 0.7,
				FlexibleSpendingLevel:  0.0,
				SurplusAllocation: domain.SurplusAllocation{
					EmergencyFundPercent: 0.60,
					DebtExtraPercent:     0.30,
					GoalsPercent:         0.05,
					FlexiblePercent:      0.05,
				},
			},
		},
		{
			scenarioType: domain.ScenarioBalanced,
			params: domain.ScenarioParameters{
				ScenarioType:           domain.ScenarioBalanced,
				GoalContributionFactor: 1.0,
				FlexibleSpendingLevel:  0.5,
				SurplusAllocation: domain.SurplusAllocation{
					EmergencyFundPercent: 0.40,
					DebtExtraPercent:     0.30,
					GoalsPercent:         0.20,
					FlexiblePercent:      0.10,
				},
			},
		},
		{
			scenarioType: domain.ScenarioAggressive,
			params: domain.ScenarioParameters{
				ScenarioType:           domain.ScenarioAggressive,
				GoalContributionFactor: 1.3,
				FlexibleSpendingLevel:  1.0,
				SurplusAllocation: domain.SurplusAllocation{
					EmergencyFundPercent: 0.25,
					DebtExtraPercent:     0.25,
					GoalsPercent:         0.40,
					FlexiblePercent:      0.10,
				},
			},
		},
	}

	for _, config := range scenarioConfigs {
		solver := solver.NewGoalProgrammingSolver(model, config.params)
		tripleResult, err := solver.SolveTriple()
		if err != nil {
			continue
		}

		preemptiveScenario := sg.buildScenario(config.scenarioType, tripleResult.PreemptiveResult, model, categoryNames)
		weightedScenario := sg.buildScenario(config.scenarioType, tripleResult.WeightedResult, model, categoryNames)
		minmaxScenario := sg.buildScenario(config.scenarioType, tripleResult.MinmaxResult, model, categoryNames)

		// Add warnings
		if !isFeasible {
			sg.addDeficitWarnings(&preemptiveScenario, deficit, model)
			sg.addDeficitWarnings(&weightedScenario, deficit, model)
			sg.addDeficitWarnings(&minmaxScenario, deficit, model)
		}

		switch config.scenarioType {
		case domain.ScenarioConservative:
			sg.addConservativeWarnings(&preemptiveScenario, model)
			sg.addConservativeWarnings(&weightedScenario, model)
			sg.addConservativeWarnings(&minmaxScenario, model)
		case domain.ScenarioBalanced:
			sg.addBalancedWarnings(&preemptiveScenario, model)
			sg.addBalancedWarnings(&weightedScenario, model)
			sg.addBalancedWarnings(&minmaxScenario, model)
		case domain.ScenarioAggressive:
			sg.addAggressiveWarnings(&preemptiveScenario, model)
			sg.addAggressiveWarnings(&weightedScenario, model)
			sg.addAggressiveWarnings(&minmaxScenario, model)
		}

		results = append(results, TripleScenarioResult{
			ScenarioType:       config.scenarioType,
			PreemptiveScenario: preemptiveScenario,
			WeightedScenario:   weightedScenario,
			MinmaxScenario:     minmaxScenario,
			Comparison:         tripleResult.Comparison,
		})
	}

	return results, nil
}
