package budget_allocation

import (
	"context"
	"errors"
	"fmt"
	"time"

	"personalfinancedss/internal/module/analytics/budget_allocation/domain"
	"personalfinancedss/internal/module/analytics/budget_allocation/dto"
	"personalfinancedss/internal/module/analytics/models/budget_allocation/constraint"

	"github.com/google/uuid"
)

// BudgetAllocationModel implements MBMS Model interface for budget allocation
type BudgetAllocationModel struct {
	name        string
	description string
}

// NewBudgetAllocationModel creates a new budget allocation model
func NewBudgetAllocationModel() *BudgetAllocationModel {
	return &BudgetAllocationModel{
		name:        "budget_allocation_gp",
		description: "Multi-objective budget allocation using Goal Programming with 4 solver variants (Preemptive, Weighted, Minmax, Meta)",
	}
}

func (m *BudgetAllocationModel) Name() string        { return m.name }
func (m *BudgetAllocationModel) Description() string { return m.description }
func (m *BudgetAllocationModel) Dependencies() []string {
	return []string{"goal_prioritization"} // Uses AHP for goal priorities
}

// Validate validates the input before execution
func (m *BudgetAllocationModel) Validate(ctx context.Context, input interface{}) error {
	bi, ok := input.(*dto.BudgetAllocationModelInput)
	if !ok {
		return errors.New("input must be *dto.BudgetAllocationModelInput type")
	}

	if bi.UserID == uuid.Nil {
		return errors.New("user_id is required")
	}

	if bi.Year < 2000 || bi.Year > 2100 {
		return fmt.Errorf("year must be between 2000 and 2100, got %d", bi.Year)
	}

	if bi.Month < 1 || bi.Month > 12 {
		return fmt.Errorf("month must be between 1 and 12, got %d", bi.Month)
	}

	if bi.TotalIncome <= 0 {
		return errors.New("total income must be positive")
	}

	return nil
}

// Execute runs the budget allocation model
func (m *BudgetAllocationModel) Execute(ctx context.Context, input interface{}) (interface{}, error) {
	bi := input.(*dto.BudgetAllocationModelInput)
	startTime := time.Now()

	// Build constraint model from input
	constraintModel := m.buildConstraintModel(bi)

	// Check feasibility
	constraintBuilder := constraint.NewConstraintBuilder()
	isFeasible, deficit := constraintBuilder.CheckFeasibility(constraintModel)

	// Generate scenarios
	scenarios := make([]domain.AllocationScenario, 0)
	globalWarnings := make([]domain.AllocationWarning, 0)

	if !isFeasible {
		globalWarnings = append(globalWarnings, domain.AllocationWarning{
			Severity:    domain.SeverityCritical,
			Category:    "income",
			Message:     fmt.Sprintf("Income insufficient by $%.2f to cover mandatory expenses and minimum debt payments", deficit),
			Suggestions: constraintBuilder.GetSuggestionsForDeficit(constraintModel, deficit),
		})
	}

	// Generate scenarios based on request
	categoryNames := m.buildCategoryNames(bi)
	scenarioGenerator := NewScenarioGenerator()

	if bi.UseAllScenarios {
		// Generate all 3 scenarios with comparison
		generatedScenarios, err := scenarioGenerator.GenerateScenarios(constraintModel, categoryNames)
		if err != nil {
			return nil, fmt.Errorf("failed to generate scenarios: %w", err)
		}
		scenarios = generatedScenarios
	} else {
		// Generate only balanced scenario
		balanced := scenarioGenerator.GenerateBalancedScenario(constraintModel, categoryNames)
		scenarios = append(scenarios, balanced)
	}

	// Run sensitivity analysis if requested
	var sensitivityResults *dto.SensitivityAnalysisResult
	if bi.RunSensitivity {
		sensitivityResults = m.runSensitivityAnalysis(constraintModel, categoryNames)
	}

	// Build output
	output := &dto.BudgetAllocationModelOutput{
		UserID:             bi.UserID,
		Period:             fmt.Sprintf("%d-%02d", bi.Year, bi.Month),
		TotalIncome:        bi.TotalIncome,
		Scenarios:          scenarios,
		IsFeasible:         isFeasible,
		GlobalWarnings:     globalWarnings,
		SensitivityResults: sensitivityResults,
		Metadata: dto.AllocationMetadata{
			GeneratedAt:      time.Now(),
			DataSources:      []string{"income_profile", "budget_constraints", "goals", "debts"},
			ComputationTime:  time.Since(startTime).Milliseconds(),
			ConstraintsCount: len(constraintModel.MandatoryExpenses) + len(constraintModel.FlexibleExpenses),
			GoalsCount:       len(constraintModel.GoalTargets),
			DebtsCount:       len(constraintModel.DebtPayments),
		},
	}

	return output, nil
}

// buildConstraintModel builds domain.ConstraintModel from DTO input
func (m *BudgetAllocationModel) buildConstraintModel(input *dto.BudgetAllocationModelInput) *domain.ConstraintModel {
	model := &domain.ConstraintModel{
		TotalIncome:       input.TotalIncome,
		MandatoryExpenses: make(map[uuid.UUID]domain.CategoryConstraint),
		FlexibleExpenses:  make(map[uuid.UUID]domain.CategoryConstraint),
		DebtPayments:      make(map[uuid.UUID]domain.DebtConstraint),
		GoalTargets:       make(map[uuid.UUID]domain.GoalConstraint),
	}

	// Map mandatory expenses
	for _, exp := range input.MandatoryExpenses {
		model.MandatoryExpenses[exp.CategoryID] = domain.CategoryConstraint{
			CategoryID: exp.CategoryID,
			Minimum:    exp.Amount,
			Maximum:    exp.Amount, // Fixed amount
			IsFlexible: false,
			Priority:   exp.Priority,
		}
	}

	// Map flexible expenses
	for _, exp := range input.FlexibleExpenses {
		model.FlexibleExpenses[exp.CategoryID] = domain.CategoryConstraint{
			CategoryID: exp.CategoryID,
			Minimum:    exp.MinAmount,
			Maximum:    exp.MaxAmount,
			IsFlexible: true,
			Priority:   exp.Priority,
		}
	}

	// Map debts
	for _, debt := range input.Debts {
		model.DebtPayments[debt.DebtID] = domain.DebtConstraint{
			DebtID:         debt.DebtID,
			DebtName:       debt.Name,
			MinimumPayment: debt.MinimumPayment,
			CurrentBalance: debt.Balance,
			InterestRate:   debt.InterestRate,
			Priority:       m.calculateDebtPriority(debt.InterestRate),
		}
	}

	// Map goals
	for _, goal := range input.Goals {
		model.GoalTargets[goal.GoalID] = domain.GoalConstraint{
			GoalID:                goal.GoalID,
			GoalName:              goal.Name,
			GoalType:              goal.Type,
			SuggestedContribution: goal.SuggestedContribution,
			Priority:              goal.Priority,
			PriorityWeight:        m.goalPriorityToWeight(goal.Priority),
			RemainingAmount:       goal.RemainingAmount,
		}
	}

	return model
}

// buildCategoryNames builds category name map
func (m *BudgetAllocationModel) buildCategoryNames(input *dto.BudgetAllocationModelInput) map[uuid.UUID]string {
	names := make(map[uuid.UUID]string)

	for _, exp := range input.MandatoryExpenses {
		names[exp.CategoryID] = exp.Name
	}

	for _, exp := range input.FlexibleExpenses {
		names[exp.CategoryID] = exp.Name
	}

	return names
}

// calculateDebtPriority calculates debt priority based on interest rate
func (m *BudgetAllocationModel) calculateDebtPriority(interestRate float64) int {
	switch {
	case interestRate >= 0.20:
		return 1 // Critical
	case interestRate >= 0.10:
		return 10 // High
	case interestRate >= 0.05:
		return 20 // Medium
	default:
		return 30 // Low
	}
}

// goalPriorityToWeight converts priority string to weight
func (m *BudgetAllocationModel) goalPriorityToWeight(priority string) int {
	switch priority {
	case "critical":
		return 1
	case "high":
		return 10
	case "medium":
		return 20
	case "low":
		return 30
	default:
		return 50
	}
}

// runSensitivityAnalysis performs sensitivity analysis on the allocation
func (m *BudgetAllocationModel) runSensitivityAnalysis(
	constraintModel *domain.ConstraintModel,
	categoryNames map[uuid.UUID]string,
) *dto.SensitivityAnalysisResult {
	result := &dto.SensitivityAnalysisResult{
		IncomeImpact:       make([]dto.IncomeImpactResult, 0),
		InterestRateImpact: make([]dto.InterestRateImpactResult, 0),
		GoalPriorityImpact: make([]dto.GoalPriorityImpactResult, 0),
	}

	// Default sensitivity scenarios
	incomeChanges := []float64{-0.20, -0.10, 0.10, 0.20}
	rateChanges := []float64{0.02, 0.05}

	// 1. Income sensitivity analysis
	result.IncomeImpact = m.analyzeIncomeSensitivity(constraintModel, categoryNames, incomeChanges)

	// 2. Interest rate sensitivity analysis
	result.InterestRateImpact = m.analyzeInterestRateSensitivity(constraintModel, rateChanges)

	// 3. Goal priority sensitivity analysis
	result.GoalPriorityImpact = m.analyzeGoalPrioritySensitivity(constraintModel, categoryNames)

	// 4. Generate summary
	result.Summary = m.generateSensitivitySummary(result, constraintModel)

	return result
}

// analyzeIncomeSensitivity analyzes how income changes affect allocation
func (m *BudgetAllocationModel) analyzeIncomeSensitivity(
	constraintModel *domain.ConstraintModel,
	categoryNames map[uuid.UUID]string,
	incomeChanges []float64,
) []dto.IncomeImpactResult {
	results := make([]dto.IncomeImpactResult, 0, len(incomeChanges))
	constraintBuilder := constraint.NewConstraintBuilder()
	scenarioGenerator := NewScenarioGenerator()

	// Get baseline allocation
	baselineScenario := scenarioGenerator.GenerateBalancedScenario(constraintModel, categoryNames)

	for _, changePercent := range incomeChanges {
		newIncome := constraintModel.TotalIncome * (1 + changePercent)

		// Create modified constraint model
		modifiedModel := m.cloneConstraintModel(constraintModel)
		modifiedModel.TotalIncome = newIncome

		// Check feasibility
		isFeasible, deficit := constraintBuilder.CheckFeasibility(modifiedModel)

		// Generate scenario with new income
		newScenario := scenarioGenerator.GenerateBalancedScenario(modifiedModel, categoryNames)

		// Calculate deltas
		impact := dto.IncomeImpactResult{
			IncomeChangePercent: changePercent,
			NewIncome:           newIncome,
			IsFeasible:          isFeasible,
			GoalAllocationDelta: newScenario.Summary.TotalGoalContributions - baselineScenario.Summary.TotalGoalContributions,
			FlexibleDelta:       newScenario.Summary.FlexibleExpenses - baselineScenario.Summary.FlexibleExpenses,
			SurplusDelta:        newScenario.Summary.Surplus - baselineScenario.Summary.Surplus,
		}

		// Calculate debt extra delta
		baselineDebtExtra := m.calculateTotalDebtExtra(baselineScenario)
		newDebtExtra := m.calculateTotalDebtExtra(newScenario)
		impact.DebtExtraDelta = newDebtExtra - baselineDebtExtra

		if !isFeasible {
			impact.Deficit = deficit
			impact.Recommendation = fmt.Sprintf("Income decrease of %.0f%% makes allocation infeasible. Need to reduce expenses by $%.2f or increase income.", -changePercent*100, deficit)
		} else if changePercent < 0 {
			// Find affected goals
			impact.AffectedGoals = m.findAffectedGoals(baselineScenario, newScenario)
			if len(impact.AffectedGoals) > 0 {
				impact.Recommendation = fmt.Sprintf("Income decrease would reduce funding for %d goals. Consider adjusting priorities.", len(impact.AffectedGoals))
			} else {
				impact.Recommendation = "Income decrease can be absorbed by reducing flexible spending and surplus."
			}
		} else {
			impact.Recommendation = fmt.Sprintf("Income increase of %.0f%% allows additional $%.2f for goals and debt payoff.", changePercent*100, impact.GoalAllocationDelta+impact.DebtExtraDelta)
		}

		results = append(results, impact)
	}

	return results
}

// analyzeInterestRateSensitivity analyzes how interest rate changes affect debt strategy
func (m *BudgetAllocationModel) analyzeInterestRateSensitivity(
	constraintModel *domain.ConstraintModel,
	rateChanges []float64,
) []dto.InterestRateImpactResult {
	results := make([]dto.InterestRateImpactResult, 0, len(rateChanges))

	for _, rateChange := range rateChanges {
		impact := dto.InterestRateImpactResult{
			RateChangePercent: rateChange,
			AffectedDebts:     make([]dto.DebtRateImpact, 0),
		}

		totalExtraInterest := 0.0
		strategyChangeNeeded := false

		for debtID, debt := range constraintModel.DebtPayments {
			newRate := debt.InterestRate + rateChange
			if newRate > 1.0 {
				newRate = 1.0 // Cap at 100%
			}

			// Calculate extra monthly interest
			monthlyRateOld := debt.InterestRate / 12
			monthlyRateNew := newRate / 12
			extraMonthlyInterest := debt.CurrentBalance * (monthlyRateNew - monthlyRateOld)

			// Calculate new priority
			newPriority := m.calculateDebtPriority(newRate)

			// Check if priority changed significantly
			if newPriority < debt.Priority {
				strategyChangeNeeded = true
			}

			debtImpact := dto.DebtRateImpact{
				DebtID:               debtID,
				DebtName:             debt.DebtName,
				OldRate:              debt.InterestRate,
				NewRate:              newRate,
				ExtraMonthlyInterest: extraMonthlyInterest,
				NewPriority:          newPriority,
			}

			impact.AffectedDebts = append(impact.AffectedDebts, debtImpact)
			totalExtraInterest += extraMonthlyInterest
		}

		impact.TotalExtraInterest = totalExtraInterest
		impact.StrategyChangeNeeded = strategyChangeNeeded

		if strategyChangeNeeded {
			impact.RecommendedAction = fmt.Sprintf("Rate increase of %.0f%% changes debt priorities. Consider reallocating extra payments to higher-rate debts.", rateChange*100)
		} else if totalExtraInterest > 100 {
			impact.RecommendedAction = fmt.Sprintf("Rate increase adds $%.2f/month in interest. Consider increasing debt payments or refinancing.", totalExtraInterest)
		} else {
			impact.RecommendedAction = "Rate increase has minimal impact on current strategy."
		}

		results = append(results, impact)
	}

	return results
}

// analyzeGoalPrioritySensitivity analyzes how goal priority changes affect allocation
func (m *BudgetAllocationModel) analyzeGoalPrioritySensitivity(
	constraintModel *domain.ConstraintModel,
	categoryNames map[uuid.UUID]string,
) []dto.GoalPriorityImpactResult {
	results := make([]dto.GoalPriorityImpactResult, 0)
	scenarioGenerator := NewScenarioGenerator()

	// Get baseline allocation
	baselineScenario := scenarioGenerator.GenerateBalancedScenario(constraintModel, categoryNames)
	baselineAllocations := make(map[uuid.UUID]float64)
	for _, ga := range baselineScenario.GoalAllocations {
		baselineAllocations[ga.GoalID] = ga.Amount
	}

	for goalID, goal := range constraintModel.GoalTargets {
		// Skip emergency funds (usually fixed priority)
		if goal.GoalType == "emergency" {
			continue
		}

		impact := dto.GoalPriorityImpactResult{
			GoalID:            goalID,
			GoalName:          goal.GoalName,
			CurrentPriority:   goal.Priority,
			CurrentAllocation: baselineAllocations[goalID],
		}

		// Simulate higher priority
		higherModel := m.cloneConstraintModel(constraintModel)
		higherGoal := higherModel.GoalTargets[goalID]
		higherGoal.PriorityWeight = max(1, goal.PriorityWeight-10)
		higherModel.GoalTargets[goalID] = higherGoal

		higherScenario := scenarioGenerator.GenerateBalancedScenario(higherModel, categoryNames)
		for _, ga := range higherScenario.GoalAllocations {
			if ga.GoalID == goalID {
				impact.IfHigherPriority = ga.Amount
				break
			}
		}

		// Simulate lower priority
		lowerModel := m.cloneConstraintModel(constraintModel)
		lowerGoal := lowerModel.GoalTargets[goalID]
		lowerGoal.PriorityWeight = min(99, goal.PriorityWeight+10)
		lowerModel.GoalTargets[goalID] = lowerGoal

		lowerScenario := scenarioGenerator.GenerateBalancedScenario(lowerModel, categoryNames)
		for _, ga := range lowerScenario.GoalAllocations {
			if ga.GoalID == goalID {
				impact.IfLowerPriority = ga.Amount
				break
			}
		}

		// Determine sensitivity
		allocationRange := impact.IfHigherPriority - impact.IfLowerPriority
		if impact.CurrentAllocation > 0 {
			sensitivityRatio := allocationRange / impact.CurrentAllocation
			if sensitivityRatio > 0.5 {
				impact.AllocationSensitivity = "high"
			} else if sensitivityRatio > 0.2 {
				impact.AllocationSensitivity = "medium"
			} else {
				impact.AllocationSensitivity = "low"
			}
		} else {
			impact.AllocationSensitivity = "low"
		}

		results = append(results, impact)
	}

	return results
}

// generateSensitivitySummary generates overall sensitivity insights
func (m *BudgetAllocationModel) generateSensitivitySummary(
	result *dto.SensitivityAnalysisResult,
	constraintModel *domain.ConstraintModel,
) dto.SensitivitySummary {
	summary := dto.SensitivitySummary{
		HighRiskDebts:      make([]string, 0),
		MostFlexibleGoals:  make([]string, 0),
		KeyRecommendations: make([]string, 0),
	}

	// Find income break-even point (total mandatory expenses + minimum debt payments)
	var totalMandatory float64
	for _, cat := range constraintModel.MandatoryExpenses {
		totalMandatory += cat.Minimum
	}
	for _, debt := range constraintModel.DebtPayments {
		totalMandatory += debt.MinimumPayment
	}
	summary.IncomeBreakEvenPoint = totalMandatory

	// Check if sensitive to income
	for _, impact := range result.IncomeImpact {
		if impact.IncomeChangePercent == -0.10 && !impact.IsFeasible {
			summary.MostSensitiveToIncome = true
			break
		}
	}

	// Find high-risk debts (most affected by rate changes)
	for _, rateImpact := range result.InterestRateImpact {
		for _, debtImpact := range rateImpact.AffectedDebts {
			if debtImpact.ExtraMonthlyInterest > 50 {
				summary.HighRiskDebts = append(summary.HighRiskDebts, debtImpact.DebtName)
			}
		}
		break // Only check first rate scenario
	}

	// Find most flexible goals
	for _, goalImpact := range result.GoalPriorityImpact {
		if goalImpact.AllocationSensitivity == "high" {
			summary.MostFlexibleGoals = append(summary.MostFlexibleGoals, goalImpact.GoalName)
		}
	}

	// Determine overall risk level
	riskScore := 0
	if summary.MostSensitiveToIncome {
		riskScore += 3
	}
	if len(summary.HighRiskDebts) > 0 {
		riskScore += 2
	}
	surplus := constraintModel.TotalIncome - totalMandatory
	if surplus < constraintModel.TotalIncome*0.1 {
		riskScore += 2
	}

	if riskScore >= 5 {
		summary.OverallRiskLevel = "high"
	} else if riskScore >= 3 {
		summary.OverallRiskLevel = "medium"
	} else {
		summary.OverallRiskLevel = "low"
	}

	// Generate recommendations
	if summary.MostSensitiveToIncome {
		summary.KeyRecommendations = append(summary.KeyRecommendations,
			"Build emergency fund to at least 3 months of expenses to buffer income volatility")
	}
	if len(summary.HighRiskDebts) > 0 {
		summary.KeyRecommendations = append(summary.KeyRecommendations,
			fmt.Sprintf("Consider refinancing or paying down high-risk debts: %v", summary.HighRiskDebts))
	}
	if summary.OverallRiskLevel == "high" {
		summary.KeyRecommendations = append(summary.KeyRecommendations,
			"Current allocation has low margin for error. Consider reducing flexible expenses or increasing income")
	}
	if len(summary.MostFlexibleGoals) > 0 {
		summary.KeyRecommendations = append(summary.KeyRecommendations,
			fmt.Sprintf("Goals %v are most sensitive to priority changes - review their importance", summary.MostFlexibleGoals))
	}

	return summary
}

// Helper functions

func (m *BudgetAllocationModel) cloneConstraintModel(original *domain.ConstraintModel) *domain.ConstraintModel {
	clone := &domain.ConstraintModel{
		TotalIncome:       original.TotalIncome,
		MandatoryExpenses: make(map[uuid.UUID]domain.CategoryConstraint),
		FlexibleExpenses:  make(map[uuid.UUID]domain.CategoryConstraint),
		DebtPayments:      make(map[uuid.UUID]domain.DebtConstraint),
		GoalTargets:       make(map[uuid.UUID]domain.GoalConstraint),
	}

	for k, v := range original.MandatoryExpenses {
		clone.MandatoryExpenses[k] = v
	}
	for k, v := range original.FlexibleExpenses {
		clone.FlexibleExpenses[k] = v
	}
	for k, v := range original.DebtPayments {
		clone.DebtPayments[k] = v
	}
	for k, v := range original.GoalTargets {
		clone.GoalTargets[k] = v
	}

	return clone
}

func (m *BudgetAllocationModel) calculateTotalDebtExtra(scenario domain.AllocationScenario) float64 {
	total := 0.0
	for _, da := range scenario.DebtAllocations {
		total += da.ExtraPayment
	}
	return total
}

func (m *BudgetAllocationModel) findAffectedGoals(baseline, newScenario domain.AllocationScenario) []string {
	affected := make([]string, 0)

	baselineMap := make(map[uuid.UUID]float64)
	for _, ga := range baseline.GoalAllocations {
		baselineMap[ga.GoalID] = ga.Amount
	}

	for _, ga := range newScenario.GoalAllocations {
		baselineAmount := baselineMap[ga.GoalID]
		if ga.Amount < baselineAmount*0.8 { // Reduced by more than 20%
			affected = append(affected, ga.GoalName)
		}
	}

	return affected
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
