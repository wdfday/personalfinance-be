package spending_decision

import (
	"fmt"
	"math"
	"personalfinancedss/internal/module/analytics/spending_decision/domain"
	"personalfinancedss/internal/module/analytics/spending_decision/dto"
)

// Large Purchase Helper Functions

func (m *LargePurchaseModel) analyzeFundingOptions(input *dto.LargePurchaseInput) []domain.FundingOption {
	options := make([]domain.FundingOption, 0)

	// Option 1: Savings
	savingsOption := m.evaluateSavingsFunding(input)
	options = append(options, savingsOption)

	// Option 2: Financing (if applicable)
	if m.isFinanceable(input.ItemType) {
		financingOption := m.evaluateFinancing(input)
		options = append(options, financingOption)
	}

	// Option 3: Budget Reallocation
	reallocOption := m.evaluateBudgetReallocation(input)
	options = append(options, reallocOption)

	return options
}

func (m *LargePurchaseModel) evaluateSavingsFunding(input *dto.LargePurchaseInput) domain.FundingOption {
	state := input.CurrentState

	option := domain.FundingOption{
		Source:          "savings",
		AmountNeeded:    input.PurchaseAmount,
		AmountAvailable: state.OtherSavings,
		Pros:            make([]string, 0),
		Cons:            make([]string, 0),
	}

	option.Feasible = option.AmountAvailable >= option.AmountNeeded

	if option.Feasible {
		option.Pros = append(option.Pros, "✓ No interest or debt")
		option.Pros = append(option.Pros, "✓ One-time payment, no commitment")

		remainingSavings := option.AmountAvailable - option.AmountNeeded

		if remainingSavings > state.MonthlyIncome*3 {
			option.Risk = "low"
			option.Pros = append(option.Pros, "✓ Still have good savings buffer")
		} else if remainingSavings > state.MonthlyIncome {
			option.Risk = "medium"
			option.Cons = append(option.Cons, "⚠ Reduces savings buffer significantly")
		} else {
			option.Risk = "high"
			option.Cons = append(option.Cons, "❌ Leaves minimal savings buffer")
		}

		// Goal impact
		option.ImpactOnGoals = m.calculateGoalImpactLarge(
			input.PurchaseAmount,
			state.ActiveGoals,
			"savings_reduction",
		)

		if len(option.ImpactOnGoals) > 0 {
			option.Cons = append(option.Cons,
				fmt.Sprintf("⚠ Delays %d goal(s)", len(option.ImpactOnGoals)))
		}

	} else {
		option.Risk = "N/A"
		option.Cons = append(option.Cons,
			fmt.Sprintf("❌ Insufficient savings: need $%.0f, have $%.0f",
				option.AmountNeeded, option.AmountAvailable))
	}

	return option
}

func (m *LargePurchaseModel) evaluateFinancing(input *dto.LargePurchaseInput) domain.FundingOption {
	state := input.CurrentState

	interestRate := m.estimateInterestRate(input.ItemType)
	months := input.RecurringMonths
	if months == 0 {
		months = 36
	}

	monthlyRate := interestRate / 12
	monthlyPayment := input.PurchaseAmount *
		(monthlyRate * math.Pow(1+monthlyRate, float64(months))) /
		(math.Pow(1+monthlyRate, float64(months)) - 1)

	totalPaid := monthlyPayment * float64(months)
	totalInterest := totalPaid - input.PurchaseAmount

	option := domain.FundingOption{
		Source:          "financing",
		AmountNeeded:    input.PurchaseAmount,
		AmountAvailable: math.MaxFloat64,
		InterestRate:    interestRate,
		MonthlyPayment:  monthlyPayment,
		TotalInterest:   totalInterest,
		Pros:            make([]string, 0),
		Cons:            make([]string, 0),
	}

	availableForPayment := state.DiscretionaryBudget * 0.3
	option.Feasible = monthlyPayment <= availableForPayment

	if option.Feasible {
		option.Risk = "medium"

		option.Pros = append(option.Pros, "✓ Preserve cash/savings")
		option.Pros = append(option.Pros,
			fmt.Sprintf("✓ Spread cost over %d months", months))

		option.Cons = append(option.Cons,
			fmt.Sprintf("⚠ Pay $%.0f in interest over term", totalInterest))
		option.Cons = append(option.Cons,
			fmt.Sprintf("⚠ Commit to $%.0f/month for %d months", monthlyPayment, months))

		option.ImpactOnGoals = m.calculateGoalImpactLarge(
			monthlyPayment,
			state.ActiveGoals,
			"monthly_reduction",
		)

	} else {
		option.Risk = "high"
		option.Cons = append(option.Cons,
			fmt.Sprintf("❌ Monthly payment ($%.0f) exceeds affordable amount ($%.0f)",
				monthlyPayment, availableForPayment))
	}

	return option
}

func (m *LargePurchaseModel) evaluateBudgetReallocation(input *dto.LargePurchaseInput) domain.FundingOption {
	state := input.CurrentState

	monthlyDiscretionary := state.DiscretionaryBudget
	monthsNeeded := int(math.Ceil(input.PurchaseAmount / monthlyDiscretionary))

	option := domain.FundingOption{
		Source:          "budget_reallocation",
		AmountNeeded:    input.PurchaseAmount,
		AmountAvailable: monthlyDiscretionary * float64(monthsNeeded),
		Pros:            make([]string, 0),
		Cons:            make([]string, 0),
	}

	option.Feasible = monthsNeeded <= 12

	if option.Feasible {
		if monthsNeeded <= 3 {
			option.Risk = "low"
		} else if monthsNeeded <= 6 {
			option.Risk = "medium"
		} else {
			option.Risk = "medium"
		}

		option.Pros = append(option.Pros, "✓ No debt, no interest")
		option.Pros = append(option.Pros, "✓ Saves up systematically")
		option.Pros = append(option.Pros,
			fmt.Sprintf("✓ Can purchase in %d months", monthsNeeded))

		option.Cons = append(option.Cons,
			fmt.Sprintf("⚠ Must wait %d months before purchase", monthsNeeded))
		option.Cons = append(option.Cons,
			"⚠ Requires discipline to save")

		option.ImpactOnGoals = m.calculateGoalImpactLarge(
			monthlyDiscretionary,
			state.ActiveGoals,
			"temporary_reduction",
		)

	} else {
		option.Risk = "high"
		option.Cons = append(option.Cons,
			fmt.Sprintf("❌ Would take %d months to save - too long", monthsNeeded))
	}

	return option
}

func (m *LargePurchaseModel) calculateTrueCost(input *dto.LargePurchaseInput) domain.TrueCostAnalysis {
	analysis := domain.TrueCostAnalysis{
		PurchasePrice: input.PurchaseAmount,
		Breakdown:     make([]domain.CostItem, 0),
	}

	// Purchase price
	analysis.Breakdown = append(analysis.Breakdown, domain.CostItem{
		Category:    "Purchase Price",
		Amount:      input.PurchaseAmount,
		Description: "Base cost of item",
	})

	// Recurring costs
	if input.IsRecurring && input.MonthlyRecurringCost > 0 {
		totalRecurring := input.MonthlyRecurringCost * float64(input.RecurringMonths)
		analysis.RecurringCosts = totalRecurring

		analysis.Breakdown = append(analysis.Breakdown, domain.CostItem{
			Category: "Recurring Costs",
			Amount:   totalRecurring,
			Description: fmt.Sprintf("$%.0f/month × %d months",
				input.MonthlyRecurringCost, input.RecurringMonths),
		})
	}

	// Opportunity cost (20 years @ 7%)
	years := 20.0
	annualReturn := 0.07

	futureValue := input.PurchaseAmount * math.Pow(1+annualReturn, years)
	opportunityCost := futureValue - input.PurchaseAmount

	analysis.OpportunityCost = opportunityCost

	analysis.Breakdown = append(analysis.Breakdown, domain.CostItem{
		Category: "Opportunity Cost",
		Amount:   opportunityCost,
		Description: fmt.Sprintf("Lost growth if $%.0f invested for 20 years at 7%% return",
			input.PurchaseAmount),
	})

	// Interest cost if financing
	if input.PreferredFundingSource == "financing" {
		interestRate := m.estimateInterestRate(input.ItemType)
		months := input.RecurringMonths
		if months == 0 {
			months = 36
		}

		monthlyRate := interestRate / 12
		monthlyPayment := input.PurchaseAmount *
			(monthlyRate * math.Pow(1+monthlyRate, float64(months))) /
			(math.Pow(1+monthlyRate, float64(months)) - 1)

		totalInterest := (monthlyPayment * float64(months)) - input.PurchaseAmount

		analysis.Breakdown = append(analysis.Breakdown, domain.CostItem{
			Category: "Interest Cost",
			Amount:   totalInterest,
			Description: fmt.Sprintf("Interest on loan at %.1f%% over %d months",
				interestRate*100, months),
		})
	}

	// Total
	analysis.TotalCost = analysis.PurchasePrice +
		analysis.RecurringCosts +
		analysis.OpportunityCost

	if input.PreferredFundingSource == "financing" {
		for _, item := range analysis.Breakdown {
			if item.Category == "Interest Cost" {
				analysis.TotalCost += item.Amount
			}
		}
	}

	return analysis
}

func (m *LargePurchaseModel) generateReallocationPlan(input *dto.LargePurchaseInput) *domain.BudgetReallocation {
	state := input.CurrentState

	plan := &domain.BudgetReallocation{
		Strategy:         "spread_months",
		Adjustments:      make([]domain.CategoryAdjustment, 0),
		NewMonthlyBudget: make(map[string]float64),
	}

	monthlyDiscretionary := state.DiscretionaryBudget
	monthsNeeded := int(math.Ceil(input.PurchaseAmount / monthlyDiscretionary))
	plan.Duration = monthsNeeded

	reducibleCategories := []string{
		"entertainment",
		"dining_out",
		"shopping",
		"hobbies",
		"personal",
	}

	totalReducible := 0.0
	for _, cat := range reducibleCategories {
		if amount, exists := state.MonthlyBudgetAlloc[cat]; exists {
			totalReducible += amount
		}
	}

	neededPerMonth := input.PurchaseAmount / float64(monthsNeeded)
	plan.IsFeasible = totalReducible >= neededPerMonth

	if plan.IsFeasible {
		reductionRatio := neededPerMonth / totalReducible

		for _, cat := range reducibleCategories {
			if currentAmount, exists := state.MonthlyBudgetAlloc[cat]; exists && currentAmount > 0 {
				reduction := currentAmount * reductionRatio
				newAmount := currentAmount - reduction

				plan.Adjustments = append(plan.Adjustments, domain.CategoryAdjustment{
					Category:      cat,
					CurrentAmount: currentAmount,
					NewAmount:     newAmount,
					Reduction:     reduction,
					ReductionPct:  reductionRatio * 100,
				})

				plan.NewMonthlyBudget[cat] = newAmount
			}
		}

		if reductionRatio <= 0.30 {
			plan.Difficulty = "easy"
		} else if reductionRatio <= 0.50 {
			plan.Difficulty = "moderate"
		} else {
			plan.Difficulty = "difficult"
		}
	}

	return plan
}

func (m *LargePurchaseModel) projectLongTermImpact(input *dto.LargePurchaseInput) domain.LongTermImpact {
	impact := domain.LongTermImpact{}

	annualReturn := 0.07

	// Retirement impact
	yearsToRetirement := 30.0
	lostRetirement := input.PurchaseAmount *
		math.Pow(1+annualReturn, yearsToRetirement)

	impact.RetirementAt65Before = input.CurrentState.RetirementProjection
	impact.RetirementAt65After = impact.RetirementAt65Before - lostRetirement
	impact.RetirementLoss = lostRetirement

	// Net worth projections
	impact.NetWorth5YearsBefore = input.PurchaseAmount * math.Pow(1+annualReturn, 5)
	impact.NetWorth5YearsAfter = 0

	impact.NetWorth10YearsBefore = input.PurchaseAmount * math.Pow(1+annualReturn, 10)
	impact.NetWorth10YearsAfter = 0

	// Debt impact
	impact.DebtFreeDateBefore = input.CurrentState.DebtFreeDate
	impact.DebtFreeDateAfter = impact.DebtFreeDateBefore // No change if not financing

	return impact
}

func (m *LargePurchaseModel) generateLargePurchaseAlternatives(input *dto.LargePurchaseInput) []domain.Alternative {
	alternatives := make([]domain.Alternative, 0)

	// Generic alternatives based on item type
	switch input.ItemType {
	case "vehicle":
		alternatives = append(alternatives, domain.Alternative{
			Description:     "Consider a certified pre-owned vehicle instead of new",
			PotentialSaving: input.PurchaseAmount * 0.40,
			Link:            "",
		})
	case "tech":
		alternatives = append(alternatives, domain.Alternative{
			Description:     "Wait for next generation or buy refurbished",
			PotentialSaving: input.PurchaseAmount * 0.30,
			Link:            "",
		})
	case "vacation":
		alternatives = append(alternatives, domain.Alternative{
			Description:     "Choose off-season travel or domestic destination",
			PotentialSaving: input.PurchaseAmount * 0.50,
			Link:            "",
		})
	}

	return alternatives
}

func (m *LargePurchaseModel) makeLargePurchaseRecommendation(
	input *dto.LargePurchaseInput,
	output *dto.LargePurchaseOutput,
) dto.DecisionRecommendation {

	recommendation := dto.DecisionRecommendation{
		Reasoning: make([]string, 0),
		Warnings:  make([]string, 0),
	}

	// Find best funding option
	var bestOption *domain.FundingOption
	for i, option := range output.FundingOptions {
		if option.Feasible {
			if bestOption == nil || option.Risk < bestOption.Risk {
				bestOption = &output.FundingOptions[i]
			}
		}
	}

	if bestOption == nil {
		recommendation.Decision = "reject"
		recommendation.Confidence = 0.9
		recommendation.Reasoning = append(recommendation.Reasoning,
			"No feasible funding options available")
		return recommendation
	}

	// Evaluate based on best option risk
	switch bestOption.Risk {
	case "low":
		recommendation.Decision = "approve"
		recommendation.Confidence = 0.85
		recommendation.RecommendedFunding = bestOption.Source
	case "medium":
		recommendation.Decision = "reconsider"
		recommendation.Confidence = 0.60
		recommendation.RecommendedFunding = bestOption.Source
		recommendation.Warnings = append(recommendation.Warnings,
			"This purchase carries moderate financial risk")
	case "high":
		recommendation.Decision = "reject"
		recommendation.Confidence = 0.75
		recommendation.Reasoning = append(recommendation.Reasoning,
			"High financial risk - consider alternatives")
	}

	// Add reasoning based on true cost
	if output.TrueCost.TotalCost > input.PurchaseAmount*2 {
		recommendation.Warnings = append(recommendation.Warnings,
			fmt.Sprintf("True cost ($%.0f) is %.1fx the purchase price",
				output.TrueCost.TotalCost,
				output.TrueCost.TotalCost/input.PurchaseAmount))
	}

	return recommendation
}

func (m *LargePurchaseModel) generateBehavioralNudges(
	input *dto.LargePurchaseInput,
	output *dto.LargePurchaseOutput,
) []string {

	nudges := make([]string, 0)

	// 48-hour rule
	if input.PurchaseAmount > 1000 && input.MotivationLevel == "impulse" {
		nudges = append(nudges,
			"💡 Sleep on it: Wait 48 hours before making this purchase decision")
	}

	// Opportunity cost visualization
	if output.TrueCost.OpportunityCost > 0 {
		nudges = append(nudges,
			fmt.Sprintf("💰 This purchase costs you $%.0f in future wealth (opportunity cost)",
				output.TrueCost.OpportunityCost))
	}

	// Hours worked equivalent
	if input.CurrentState.MonthlyIncome > 0 {
		hourlyRate := input.CurrentState.MonthlyIncome / 160 // Assume 160 hours/month
		hoursWorked := input.PurchaseAmount / hourlyRate
		nudges = append(nudges,
			fmt.Sprintf("⏰ This equals %.0f hours of your work life", hoursWorked))
	}

	// Goal impact nudge
	if len(output.FundingOptions) > 0 && len(output.FundingOptions[0].ImpactOnGoals) > 0 {
		nudges = append(nudges,
			"🎯 This will delay your financial goals - is it worth it?")
	}

	return nudges
}

// Helper functions

func (m *LargePurchaseModel) isFinanceable(itemType string) bool {
	financeable := map[string]bool{
		"vehicle":   true,
		"housing":   true,
		"tech":      true,
		"furniture": true,
	}
	return financeable[itemType]
}

func (m *LargePurchaseModel) estimateInterestRate(itemType string) float64 {
	rates := map[string]float64{
		"vehicle": 0.05,  // 5%
		"housing": 0.035, // 3.5%
		"tech":    0.15,  // 15%
		"other":   0.12,  // 12%
	}

	if rate, exists := rates[itemType]; exists {
		return rate
	}
	return rates["other"]
}

func (m *LargePurchaseModel) calculateGoalImpactLarge(
	amount float64,
	goals []domain.GoalInfo,
	impactType string,
) []domain.GoalImpact {

	impacts := make([]domain.GoalImpact, 0)

	for _, goal := range goals {
		impact := domain.GoalImpact{
			GoalName: goal.Name,
		}

		switch impactType {
		case "savings_reduction":
			if goal.MonthlyContribution > 0 {
				monthsDelay := int(math.Ceil(amount * goal.Priority / goal.MonthlyContribution))
				impact.DelayMonths = monthsDelay
			}

		case "monthly_reduction":
			reduction := amount * goal.Priority / float64(len(goals))
			remaining := goal.TargetAmount - goal.CurrentAmount
			newMonthly := goal.MonthlyContribution - reduction
			if newMonthly > 0 {
				newMonthsToGoal := int(math.Ceil(remaining / newMonthly))
				oldMonthsToGoal := int(math.Ceil(remaining / goal.MonthlyContribution))
				impact.DelayMonths = newMonthsToGoal - oldMonthsToGoal
			}

		case "temporary_reduction":
			// Temporary budget reduction during saving period
			if goal.MonthlyContribution > 0 {
				impact.DelayMonths = int(math.Ceil(amount / goal.MonthlyContribution))
			}
		}

		if impact.DelayMonths > 0 {
			impacts = append(impacts, impact)
		}
	}

	return impacts
}
