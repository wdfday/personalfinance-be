package engine

import (
	"fmt"
	"math"
	"personalfinancedss/internal/module/analytics/debt_strategy/domain"
	"time"
)

// DebtAnalyzer handles analysis functions
type DebtAnalyzer struct {
	simulator *DebtSimulator
}

// NewDebtAnalyzer creates a new analyzer
func NewDebtAnalyzer() *DebtAnalyzer {
	return &DebtAnalyzer{
		simulator: NewDebtSimulator(),
	}
}

// AnalyzeRefinancing analyzes if refinancing makes sense
func (a *DebtAnalyzer) AnalyzeRefinancing(
	debts []domain.DebtInfo,
	option domain.RefinanceOption,
	currentExtraPayment float64,
) *domain.RefinanceAnalysis {
	// Calculate current weighted average rate
	totalBalance := 0.0
	weightedRateSum := 0.0
	debtsToConsolidate := make([]domain.DebtInfo, 0)

	for _, d := range debts {
		include := len(option.IncludeDebtIDs) == 0 // Include all if not specified
		for _, id := range option.IncludeDebtIDs {
			if d.ID == id {
				include = true
				break
			}
		}
		if include {
			debtsToConsolidate = append(debtsToConsolidate, d)
			totalBalance += d.Balance
			weightedRateSum += d.Balance * d.InterestRate
		}
	}

	if totalBalance == 0 {
		return &domain.RefinanceAnalysis{
			ShouldRefinance: false,
			Recommendation:  "No debts selected for refinancing",
		}
	}

	currentWeightedRate := weightedRateSum / totalBalance

	// Simulate current scenario
	currentResult := a.simulator.SimulateStrategy(
		domain.StrategyAvalanche,
		debtsToConsolidate,
		currentExtraPayment,
		nil,
	)

	// Calculate refinanced scenario
	consolidatedDebt := domain.DebtInfo{
		ID:             "consolidated",
		Name:           "Consolidated Loan",
		Balance:        totalBalance + option.OriginationFee,
		InterestRate:   option.NewRate,
		MinimumPayment: (totalBalance + option.OriginationFee) / float64(option.NewTerm),
	}

	refinanceResult := a.simulator.SimulateStrategy(
		domain.StrategyAvalanche,
		[]domain.DebtInfo{consolidatedDebt},
		currentExtraPayment,
		nil,
	)

	// Calculate total fees
	totalFees := option.OriginationFee + (option.MonthlyFee * float64(refinanceResult.Months))

	// Net savings
	netSavings := currentResult.TotalInterest - refinanceResult.TotalInterest - totalFees

	// Break-even calculation
	monthlySavings := (currentResult.TotalInterest - refinanceResult.TotalInterest) / float64(currentResult.Months)
	breakEvenMonths := 0
	if monthlySavings > 0 {
		breakEvenMonths = int(math.Ceil(option.OriginationFee / monthlySavings))
	}

	// Warnings
	warnings := make([]string, 0)
	if option.NewRate >= currentWeightedRate {
		warnings = append(warnings, "New rate is not lower than current weighted average")
	}
	if breakEvenMonths > 24 {
		warnings = append(warnings, "Break-even period is over 2 years")
	}
	if refinanceResult.Months > currentResult.Months {
		warnings = append(warnings, "Refinancing extends your payoff timeline")
	}

	// Recommendation
	recommendation := ""
	shouldRefinance := false
	if netSavings > 500 && breakEvenMonths < 24 && option.NewRate < currentWeightedRate {
		shouldRefinance = true
		recommendation = fmt.Sprintf("Refinancing recommended. You'll save $%.2f over the life of the loan with break-even at month %d.", netSavings, breakEvenMonths)
	} else if netSavings > 0 {
		recommendation = fmt.Sprintf("Refinancing provides marginal benefit ($%.2f savings). Consider if the hassle is worth it.", netSavings)
	} else {
		recommendation = "Refinancing is not recommended. Keep your current payment strategy."
	}

	return &domain.RefinanceAnalysis{
		ShouldRefinance:        shouldRefinance,
		CurrentWeightedRate:    currentWeightedRate,
		NewEffectiveRate:       option.NewRate,
		CurrentTotalInterest:   currentResult.TotalInterest,
		RefinanceTotalInterest: refinanceResult.TotalInterest,
		TotalFees:              totalFees,
		NetSavings:             netSavings,
		BreakEvenMonths:        breakEvenMonths,
		CurrentMonthsToPayoff:  currentResult.Months,
		NewMonthsToPayoff:      refinanceResult.Months,
		Recommendation:         recommendation,
		Warnings:               warnings,
	}
}

// AnalyzeWhatIf analyzes what-if scenarios
func (a *DebtAnalyzer) AnalyzeWhatIf(
	scenario domain.WhatIfScenario,
	debts []domain.DebtInfo,
	currentBudget float64,
	strategy domain.Strategy,
	hybridWeights *domain.HybridWeights,
) *domain.WhatIfResult {
	currentExtra := currentBudget - sumMinPayments(debts)
	baseline := a.simulator.SimulateStrategy(strategy, debts, currentExtra, hybridWeights)

	result := &domain.WhatIfResult{
		Scenario:         scenario,
		OriginalMonths:   baseline.Months,
		OriginalInterest: baseline.TotalInterest,
	}

	switch scenario.Type {
	case "extra_monthly":
		// What if I add $X more per month?
		newExtra := currentExtra + scenario.Amount
		newResult := a.simulator.SimulateStrategy(strategy, debts, newExtra, hybridWeights)
		result.NewMonths = newResult.Months
		result.NewInterest = newResult.TotalInterest
		result.MonthsSaved = baseline.Months - newResult.Months
		result.InterestSaved = baseline.TotalInterest - newResult.TotalInterest
		result.RecommendedAction = fmt.Sprintf("Adding $%.2f/month saves $%.2f in interest and %d months",
			scenario.Amount, result.InterestSaved, result.MonthsSaved)

	case "lump_sum":
		// What if I pay $X lump sum?
		if scenario.TargetDebtID != "" {
			// Apply to specific debt
			newResult := a.simulator.SimulateWithLumpSum(strategy, debts, currentExtra, scenario.Amount, scenario.TargetDebtID, hybridWeights)
			result.NewMonths = newResult.Months
			result.NewInterest = newResult.TotalInterest
		} else {
			// Find best debt for lump sum
			bestDebtID, savings := a.simulator.FindBestDebtForLumpSum(strategy, debts, currentExtra, scenario.Amount, hybridWeights)
			result.BestDebtForLumpSum = bestDebtID
			result.LumpSumImpact = savings

			newResult := a.simulator.SimulateWithLumpSum(strategy, debts, currentExtra, scenario.Amount, bestDebtID, hybridWeights)
			result.NewMonths = newResult.Months
			result.NewInterest = newResult.TotalInterest
		}
		result.MonthsSaved = baseline.Months - result.NewMonths
		result.InterestSaved = baseline.TotalInterest - result.NewInterest

		if result.BestDebtForLumpSum != "" {
			result.RecommendedAction = fmt.Sprintf("Apply $%.2f lump sum to debt '%s' to save $%.2f in interest",
				scenario.Amount, result.BestDebtForLumpSum, result.InterestSaved)
		} else {
			result.RecommendedAction = fmt.Sprintf("Lump sum payment saves $%.2f in interest and %d months",
				result.InterestSaved, result.MonthsSaved)
		}

	case "income_change":
		// What if my income changes by X%?
		newBudget := currentBudget * (1 + scenario.Amount)
		newExtra := newBudget - sumMinPayments(debts)
		if newExtra < 0 {
			newExtra = 0
		}
		newResult := a.simulator.SimulateStrategy(strategy, debts, newExtra, hybridWeights)
		result.NewMonths = newResult.Months
		result.NewInterest = newResult.TotalInterest
		result.MonthsSaved = baseline.Months - newResult.Months
		result.InterestSaved = baseline.TotalInterest - newResult.TotalInterest

		if scenario.Amount > 0 {
			result.RecommendedAction = fmt.Sprintf("%.0f%% income increase could save $%.2f and %d months",
				scenario.Amount*100, result.InterestSaved, result.MonthsSaved)
		} else {
			result.RecommendedAction = fmt.Sprintf("%.0f%% income decrease adds $%.2f interest and %d months",
				-scenario.Amount*100, -result.InterestSaved, -result.MonthsSaved)
		}
	}

	return result
}

// AnalyzeSensitivity runs sensitivity analysis
func (a *DebtAnalyzer) AnalyzeSensitivity(
	debts []domain.DebtInfo,
	currentBudget float64,
	strategy domain.Strategy,
	hybridWeights *domain.HybridWeights,
) []domain.SensitivityResult {
	results := make([]domain.SensitivityResult, 0)
	currentExtra := currentBudget - sumMinPayments(debts)
	baseline := a.simulator.SimulateStrategy(strategy, debts, currentExtra, hybridWeights)

	scenarios := []domain.SensitivityScenario{
		{Type: "income_decrease", Percentage: 0.10, Description: "10% income decrease"},
		{Type: "income_decrease", Percentage: 0.20, Description: "20% income decrease"},
		{Type: "rate_increase", Percentage: 0.02, Description: "2% rate increase (variable debts)"},
		{Type: "rate_increase", Percentage: 0.05, Description: "5% rate increase (variable debts)"},
	}

	for _, scenario := range scenarios {
		var adjustedResult *domain.SimulationResult
		modifiedDebts := make([]domain.DebtInfo, len(debts))
		copy(modifiedDebts, debts)

		switch scenario.Type {
		case "income_decrease":
			newBudget := currentBudget * (1 - scenario.Percentage)
			newExtra := newBudget - sumMinPayments(debts)
			if newExtra < 0 {
				newExtra = 0
			}
			adjustedResult = a.simulator.SimulateStrategy(strategy, modifiedDebts, newExtra, hybridWeights)

		case "rate_increase":
			for i := range modifiedDebts {
				if modifiedDebts[i].IsVariableRate {
					modifiedDebts[i].InterestRate += scenario.Percentage
				}
			}
			adjustedResult = a.simulator.SimulateStrategy(strategy, modifiedDebts, currentExtra, hybridWeights)
		}

		if adjustedResult == nil {
			continue
		}

		monthsImpact := adjustedResult.Months - baseline.Months
		interestImpact := adjustedResult.TotalInterest - baseline.TotalInterest

		// Determine risk level
		riskLevel := "low"
		if monthsImpact > 12 || interestImpact > 1000 {
			riskLevel = "high"
		} else if monthsImpact > 6 || interestImpact > 500 {
			riskLevel = "medium"
		}

		// Check if strategy is still valid
		strategyStillValid := true
		newRecommendation := strategy
		if scenario.Type == "income_decrease" && scenario.Percentage >= 0.20 {
			// With significant income decrease, might need to switch to cash flow strategy
			if strategy != domain.StrategyCashFlow {
				strategyStillValid = false
				newRecommendation = domain.StrategyCashFlow
			}
		}

		advice := ""
		switch riskLevel {
		case "high":
			advice = "Build emergency fund buffer. Consider reducing discretionary spending."
		case "medium":
			advice = "Monitor situation closely. Have contingency plan ready."
		case "low":
			advice = "Current strategy remains robust under this scenario."
		}

		results = append(results, domain.SensitivityResult{
			Scenario:           scenario,
			BaselineMonths:     baseline.Months,
			AdjustedMonths:     adjustedResult.Months,
			MonthsImpact:       monthsImpact,
			BaselineInterest:   baseline.TotalInterest,
			AdjustedInterest:   adjustedResult.TotalInterest,
			InterestImpact:     interestImpact,
			StrategyStillValid: strategyStillValid,
			NewRecommendation:  newRecommendation,
			RiskLevel:          riskLevel,
			Advice:             advice,
		})
	}

	return results
}

// CalculatePsychologicalScore calculates gamification metrics
func (a *DebtAnalyzer) CalculatePsychologicalScore(
	result *domain.SimulationResult,
	debts []domain.DebtInfo,
) *domain.PsychologicalScore {
	score := &domain.PsychologicalScore{
		Celebrations:       make([]string, 0),
		ProgressMilestones: make([]string, 0),
	}

	// Count quick wins (debts cleared in first 6 months)
	quickWins := 0
	firstWin := 0
	for _, timeline := range result.DebtTimelines {
		if timeline.PayoffMonth > 0 && timeline.PayoffMonth <= 6 {
			quickWins++
		}
		if timeline.PayoffMonth > 0 && (firstWin == 0 || timeline.PayoffMonth < firstWin) {
			firstWin = timeline.PayoffMonth
		}
	}

	score.QuickWinsCount = quickWins
	score.FirstWinMonth = firstWin

	// Momentum rating
	if firstWin <= 3 {
		score.MomentumRating = "fast_start"
		score.Celebrations = append(score.Celebrations, "🚀 Fast start! You'll clear your first debt in just "+fmt.Sprintf("%d", firstWin)+" months!")
	} else if firstWin <= 6 {
		score.MomentumRating = "steady"
		score.Celebrations = append(score.Celebrations, "💪 Steady progress! First debt cleared in "+fmt.Sprintf("%d", firstWin)+" months.")
	} else {
		score.MomentumRating = "slow_start"
		score.Celebrations = append(score.Celebrations, "🎯 Stay focused! Your first win comes in month "+fmt.Sprintf("%d", firstWin)+".")
	}

	// Motivation score (0-100)
	// Based on: quick wins, total time, number of debts
	motivationScore := 50.0 // Base score

	// Quick wins boost
	motivationScore += float64(quickWins) * 10

	// Time factor (shorter = better)
	if result.Months <= 12 {
		motivationScore += 20
	} else if result.Months <= 24 {
		motivationScore += 10
	} else if result.Months > 60 {
		motivationScore -= 10
	}

	// First win timing
	if firstWin <= 3 {
		motivationScore += 15
	} else if firstWin <= 6 {
		motivationScore += 10
	}

	if motivationScore > 100 {
		motivationScore = 100
	}
	if motivationScore < 0 {
		motivationScore = 0
	}
	score.MotivationScore = motivationScore

	// Progress milestones
	totalDebts := len(debts)
	if totalDebts > 1 {
		score.ProgressMilestones = append(score.ProgressMilestones,
			fmt.Sprintf("📊 You have %d debts to conquer", totalDebts))
	}

	if quickWins > 0 {
		score.ProgressMilestones = append(score.ProgressMilestones,
			fmt.Sprintf("⚡ %d quick win(s) in first 6 months!", quickWins))
	}

	// Debt-free celebration
	debtFreeDate := time.Now().AddDate(0, result.Months, 0)
	score.Celebrations = append(score.Celebrations,
		fmt.Sprintf("🎉 Debt-free by %s!", debtFreeDate.Format("January 2006")))

	return score
}

// GenerateMilestones creates milestone events
func (a *DebtAnalyzer) GenerateMilestones(result *domain.SimulationResult) []domain.Milestone {
	milestones := make([]domain.Milestone, 0)
	now := time.Now()

	// Debt cleared milestones
	for debtID, timeline := range result.DebtTimelines {
		if timeline.PayoffMonth > 0 {
			milestones = append(milestones, domain.Milestone{
				Month:       timeline.PayoffMonth,
				Date:        now.AddDate(0, timeline.PayoffMonth, 0),
				Description: fmt.Sprintf("🎯 %s paid off!", timeline.DebtName),
				Type:        "debt_cleared",
				DebtID:      debtID,
				Celebration: "One down! Keep the momentum going!",
			})
		}
	}

	// Halfway milestone
	halfwayMonth := result.Months / 2
	milestones = append(milestones, domain.Milestone{
		Month:       halfwayMonth,
		Date:        now.AddDate(0, halfwayMonth, 0),
		Description: "🏃 Halfway to debt-free!",
		Type:        "halfway",
		Celebration: "You're crushing it! The finish line is in sight!",
	})

	// Debt-free milestone
	milestones = append(milestones, domain.Milestone{
		Month:       result.Months,
		Date:        now.AddDate(0, result.Months, 0),
		Description: "🎉 DEBT-FREE!",
		Type:        "debt_free",
		Celebration: "Congratulations! You've achieved financial freedom!",
	})

	return milestones
}

// GenerateMonthlySchedule creates aggregated monthly schedule
func (a *DebtAnalyzer) GenerateMonthlySchedule(result *domain.SimulationResult) []domain.MonthlyAggregate {
	schedule := make([]domain.MonthlyAggregate, result.Months)
	debtsCleared := 0

	for month := 1; month <= result.Months; month++ {
		aggregate := domain.MonthlyAggregate{
			Month: month,
		}

		remainingDebts := 0
		totalBalance := 0.0

		for _, timeline := range result.DebtTimelines {
			if month-1 < len(timeline.Snapshots) {
				snapshot := timeline.Snapshots[month-1]
				aggregate.TotalPayment += snapshot.Payment
				aggregate.TotalInterest += snapshot.Interest
				aggregate.TotalTowardsPrin += (snapshot.Payment - snapshot.Interest)

				if snapshot.EndBalance > 0.01 {
					remainingDebts++
					totalBalance += snapshot.EndBalance
				}
			}

			// Check if debt was cleared this month
			if timeline.PayoffMonth == month {
				debtsCleared++
			}
		}

		aggregate.RemainingDebts = remainingDebts
		aggregate.TotalBalance = totalBalance
		aggregate.DebtsCleared = debtsCleared

		schedule[month-1] = aggregate
	}

	return schedule
}

// Helper function
func sumMinPayments(debts []domain.DebtInfo) float64 {
	total := 0.0
	for _, d := range debts {
		total += d.MinimumPayment
	}
	return total
}
