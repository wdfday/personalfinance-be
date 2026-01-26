package debt_strategy

import (
	"context"
	"errors"
	"fmt"
	"personalfinancedss/internal/module/analytics/debt_strategy/domain"
	"personalfinancedss/internal/module/analytics/debt_strategy/dto"
	"personalfinancedss/internal/module/analytics/models/debt_strategy/engine"
	"sort"
	"time"
)

type DebtStrategyModel struct {
	name        string
	description string
	simulator   *engine.DebtSimulator
	analyzer    *engine.DebtAnalyzer
}

func NewDebtStrategyModel() *DebtStrategyModel {
	return &DebtStrategyModel{
		name:        "debt_payoff_strategy",
		description: "Optimal debt payment allocation with multiple strategies, what-if analysis, and psychological scoring",
		simulator:   engine.NewDebtSimulator(),
		analyzer:    engine.NewDebtAnalyzer(),
	}
}

func (m *DebtStrategyModel) Name() string           { return m.name }
func (m *DebtStrategyModel) Description() string    { return m.description }
func (m *DebtStrategyModel) Dependencies() []string { return []string{"savings_debt_tradeoff"} }

func (m *DebtStrategyModel) Validate(ctx context.Context, input interface{}) error {
	di, ok := input.(*dto.DebtStrategyInput)
	if !ok {
		return errors.New("input must be *dto.DebtStrategyInput type")
	}

	// If no debts, return early (no validation needed)
	if len(di.Debts) == 0 {
		return nil
	}

	if di.TotalDebtBudget <= 0 {
		return errors.New("total debt budget must be positive")
	}

	// Validate budget covers minimum payments
	totalMin := di.GetTotalMinPayments()
	if di.TotalDebtBudget < totalMin {
		return fmt.Errorf("debt budget ($%.2f) insufficient for minimum payments ($%.2f)",
			di.TotalDebtBudget, totalMin)
	}

	// Validate debts
	for i, debt := range di.Debts {
		if debt.Balance < 0 {
			return fmt.Errorf("debt %d: balance cannot be negative", i)
		}
		if debt.InterestRate < 0 || debt.InterestRate > 1 {
			return fmt.Errorf("debt %d: interest rate must be between 0 and 1", i)
		}
	}

	return nil
}

func (m *DebtStrategyModel) Execute(ctx context.Context, input interface{}) (interface{}, error) {
	di := input.(*dto.DebtStrategyInput)

	// If no debts, return empty result
	if len(di.Debts) == 0 {
		return &dto.DebtStrategyOutput{
			RecommendedStrategy: domain.StrategyAvalanche,
			MonthsToDebtFree:    0,
			TotalInterest:       0,
			StrategyComparison:  make([]domain.StrategyComparison, 0),
			PaymentPlans:        make([]domain.PaymentPlan, 0),
			Milestones:          make([]domain.Milestone, 0),
			MonthlySchedule:     make([]domain.MonthlyAggregate, 0),
			Reasoning:           "No debts to pay off. All extra income can be allocated to savings and goals.",
			KeyFacts:            []string{"No active debts", "Full budget available for savings"},
		}, nil
	}

	extraPayment := di.CalculateExtraPayment()
	hybridWeights := di.GetHybridWeights()

	// Step 1: Simulate all strategies
	strategies := []domain.Strategy{
		domain.StrategyAvalanche,
		domain.StrategySnowball,
		domain.StrategyCashFlow,
		domain.StrategyStress,
		domain.StrategyHybrid,
	}

	strategyResults := make(map[domain.Strategy]*domain.SimulationResult)
	comparisons := make([]domain.StrategyComparison, 0)

	for _, strategy := range strategies {
		result := m.simulator.SimulateStrategy(strategy, di.Debts, extraPayment, &hybridWeights)
		strategyResults[strategy] = result

		// Generate payment plans for this strategy
		paymentPlans := m.generatePaymentPlans(result, di.Debts)

		// Calculate monthly allocation: sum of MonthlyPayment from all payment plans
		// This ensures consistency: monthly allocation = sum of individual debt allocations
		monthlyAllocation := 0.0
		for _, plan := range paymentPlans {
			monthlyAllocation += plan.MonthlyPayment
		}

		comparisons = append(comparisons, domain.StrategyComparison{
			Strategy:          strategy,
			TotalInterest:     result.TotalInterest,
			Months:            result.Months,
			FirstDebtCleared:  result.FirstCleared,
			Description:       m.getStrategyDescription(strategy),
			Pros:              m.getStrategyPros(strategy),
			Cons:              m.getStrategyCons(strategy),
			PaymentPlans:      paymentPlans,
			MonthlyAllocation: monthlyAllocation,
		})
	}

	// Calculate interest saved vs baseline (proportional)
	baselineInterest := strategyResults[domain.StrategyAvalanche].TotalInterest // Use avalanche as baseline
	for i := range comparisons {
		comparisons[i].InterestSaved = baselineInterest - comparisons[i].TotalInterest
	}

	// Step 2: Select best strategy
	recommended, reasoning, keyFacts := m.selectBestStrategy(di, strategyResults)
	selectedResult := strategyResults[recommended]

	// Step 3: Generate payment plans
	paymentPlans := m.generatePaymentPlans(selectedResult, di.Debts)

	// Step 4: Generate milestones
	milestones := m.analyzer.GenerateMilestones(selectedResult)

	// Step 5: Generate monthly schedule
	monthlySchedule := m.analyzer.GenerateMonthlySchedule(selectedResult)

	// Step 6: Psychological scoring
	psychScore := m.analyzer.CalculatePsychologicalScore(selectedResult, di.Debts)

	// Step 7: Add warning if extra payment is zero or very small
	if extraPayment <= 0.01 {
		keyFacts = append(keyFacts, "⚠️ No extra payment available (budget = minimum payments). All strategies will produce identical results.")
		if reasoning != "" {
			reasoning += " Note: With no extra payment, all strategies behave identically since they only differ in how extra money is allocated."
		}
	}

	// Build output
	output := &dto.DebtStrategyOutput{
		RecommendedStrategy: recommended,
		PaymentPlans:        paymentPlans,
		TotalInterest:       selectedResult.TotalInterest,
		MonthsToDebtFree:    selectedResult.Months,
		DebtFreeDate:        time.Now().AddDate(0, selectedResult.Months, 0),
		StrategyComparison:  comparisons,
		MonthlySchedule:     monthlySchedule,
		Milestones:          milestones,
		Reasoning:           reasoning,
		KeyFacts:            keyFacts,
		PsychScore:          psychScore,
	}

	// Step 7: What-if analysis (if requested)
	if len(di.WhatIfScenarios) > 0 {
		whatIfResults := make([]domain.WhatIfResult, 0)
		for _, scenario := range di.WhatIfScenarios {
			result := m.analyzer.AnalyzeWhatIf(scenario, di.Debts, di.TotalDebtBudget, recommended, &hybridWeights)
			whatIfResults = append(whatIfResults, *result)
		}
		output.WhatIfResults = whatIfResults
	}

	// Step 8: Refinancing analysis (if requested)
	if di.RefinanceOption != nil {
		output.RefinanceAnalysis = m.analyzer.AnalyzeRefinancing(di.Debts, *di.RefinanceOption, extraPayment)
	}

	// Step 9: Sensitivity analysis (if requested)
	if di.RunSensitivity {
		output.SensitivityResults = m.analyzer.AnalyzeSensitivity(di.Debts, di.TotalDebtBudget, recommended, &hybridWeights)
	}

	return output, nil
}

func (m *DebtStrategyModel) selectBestStrategy(
	input *dto.DebtStrategyInput,
	results map[domain.Strategy]*domain.SimulationResult,
) (domain.Strategy, string, []string) {
	keyFacts := make([]string, 0)

	// Rule 1: User explicitly prefers a strategy
	if input.PreferredStrategy != "" {
		return input.PreferredStrategy,
			fmt.Sprintf("Following your preference for %s strategy", input.PreferredStrategy),
			append(keyFacts, "User preference: "+string(input.PreferredStrategy))
	}

	// Rule 2: High stress debt exists
	if input.HasHighStressDebt() {
		highStress := input.GetHighestStressDebt()
		keyFacts = append(keyFacts, fmt.Sprintf("High stress debt detected: %s (stress: %d/10)", highStress.Name, highStress.StressScore))

		// If stress debt is also small, snowball might clear it fast
		if highStress.Balance < input.GetTotalDebt()*0.2 {
			return domain.StrategySnowball,
				"High stress debt is relatively small - Snowball will clear it quickly for peace of mind",
				keyFacts
		}
		return domain.StrategyStress,
			"Prioritizing high-stress debt for mental peace",
			keyFacts
	}

	// Rule 3: Low motivation - need quick wins
	if input.MotivationLevel == "low" {
		smallestBalance := input.GetSmallestBalance()
		keyFacts = append(keyFacts, fmt.Sprintf("Low motivation level, smallest debt: $%.2f", smallestBalance))

		if smallestBalance < 2000 {
			return domain.StrategySnowball,
				"Snowball strategy recommended for quick wins to build momentum",
				keyFacts
		}
	}

	// Rule 4: High interest rate spread - Avalanche is clearly better
	highRate := input.GetHighestInterestRate()
	avgRate := input.GetWeightedAvgRate()
	rateSpread := highRate - avgRate

	if rateSpread > 0.10 { // >10% spread
		keyFacts = append(keyFacts, fmt.Sprintf("High interest rate spread: %.1f%% (max: %.1f%%, avg: %.1f%%)",
			rateSpread*100, highRate*100, avgRate*100))
		return domain.StrategyAvalanche,
			"Avalanche recommended due to significant interest rate spread - will save substantial interest",
			keyFacts
	}

	// Rule 5: Very high interest debt (>18%)
	if highRate > 0.18 {
		keyFacts = append(keyFacts, fmt.Sprintf("High interest debt detected: %.1f%%", highRate*100))
		return domain.StrategyAvalanche,
			"Avalanche recommended to tackle high-interest debt first",
			keyFacts
	}

	// Rule 6: Need cash flow flexibility
	totalMin := input.GetTotalMinPayments()
	budgetRatio := totalMin / input.TotalDebtBudget
	if budgetRatio > 0.7 { // Minimum payments are >70% of budget
		keyFacts = append(keyFacts, fmt.Sprintf("Tight budget: minimum payments are %.0f%% of debt budget", budgetRatio*100))
		return domain.StrategyCashFlow,
			"Cash Flow strategy recommended to free up monthly budget faster",
			keyFacts
	}

	// Rule 7: Compare interest savings
	avalancheResult := results[domain.StrategyAvalanche]
	snowballResult := results[domain.StrategySnowball]
	interestDiff := snowballResult.TotalInterest - avalancheResult.TotalInterest

	if interestDiff > 500 {
		keyFacts = append(keyFacts, fmt.Sprintf("Avalanche saves $%.2f more in interest", interestDiff))
		return domain.StrategyAvalanche,
			"Avalanche recommended - significant interest savings over Snowball",
			keyFacts
	}

	// Rule 8: Similar outcomes - consider psychological factors
	timeDiff := avalancheResult.Months - snowballResult.Months
	if timeDiff < 3 && interestDiff < 200 {
		// Very similar outcomes - go with snowball for motivation
		keyFacts = append(keyFacts, "Strategies have similar outcomes - choosing for psychological benefit")
		return domain.StrategySnowball,
			"Snowball recommended - similar financial outcome but better for motivation",
			keyFacts
	}

	// Default: Avalanche (mathematically optimal)
	keyFacts = append(keyFacts, "Default to mathematically optimal strategy")
	return domain.StrategyAvalanche,
		"Avalanche strategy recommended as the mathematically optimal approach",
		keyFacts
}

func (m *DebtStrategyModel) generatePaymentPlans(result *domain.SimulationResult, originalDebts []domain.DebtInfo) []domain.PaymentPlan {
	plans := make([]domain.PaymentPlan, 0, len(result.DebtTimelines))

	// Create a map to quickly find original debt info
	debtMap := make(map[string]domain.DebtInfo)
	for _, debt := range originalDebts {
		debtMap[debt.ID] = debt
	}

	for _, timeline := range result.DebtTimelines {
		totalPayment := 0.0
		totalMinPayment := 0.0
		monthsWithPayment := 0

		// Find original debt to get minimum payment
		originalDebt, exists := debtMap[timeline.DebtID]
		minPayment := 0.0
		if exists {
			minPayment = originalDebt.MinimumPayment
		}

		// Get FIRST MONTH payment for MonthlyPayment (actual allocation for current month)
		firstMonthPayment := 0.0
		if len(timeline.Snapshots) > 0 {
			firstMonthPayment = timeline.Snapshots[0].Payment
		}

		for _, snapshot := range timeline.Snapshots {
			if snapshot.Payment > 0 {
				totalPayment += snapshot.Payment
				totalMinPayment += minPayment
				monthsWithPayment++
			}
		}

		// Calculate extra payment for first month (first month payment - minimum payment)
		firstMonthExtra := firstMonthPayment - minPayment
		if firstMonthExtra < 0 {
			firstMonthExtra = 0
		}

		// Calculate total extra payment (total payment - total minimum payments)
		totalExtraPayment := totalPayment - totalMinPayment
		if totalExtraPayment < 0 {
			totalExtraPayment = 0
		}

		plan := domain.PaymentPlan{
			DebtID:         timeline.DebtID,
			DebtName:       timeline.DebtName,
			MonthlyPayment: firstMonthPayment, // Use first month payment (actual allocation for current month)
			ExtraPayment:   firstMonthExtra,   // Extra payment for first month
			PayoffMonth:    timeline.PayoffMonth,
			TotalInterest:  timeline.TotalInterest,
			Timeline:       timeline.Snapshots,
		}
		plans = append(plans, plan)
	}

	// Sort by payoff month
	sort.Slice(plans, func(i, j int) bool {
		return plans[i].PayoffMonth < plans[j].PayoffMonth
	})

	return plans
}

func (m *DebtStrategyModel) getStrategyDescription(strategy domain.Strategy) string {
	switch strategy {
	case domain.StrategyAvalanche:
		return "Pay highest interest rate first - minimizes total interest paid"
	case domain.StrategySnowball:
		return "Pay smallest balance first - quick wins for motivation"
	case domain.StrategyCashFlow:
		return "Pay highest payment-to-balance ratio first - frees up monthly budget"
	case domain.StrategyStress:
		return "Pay highest stress debt first - prioritizes mental peace"
	case domain.StrategyHybrid:
		return "Weighted combination of interest, balance, stress, and cash flow factors"
	default:
		return "Custom strategy"
	}
}

func (m *DebtStrategyModel) getStrategyPros(strategy domain.Strategy) []string {
	switch strategy {
	case domain.StrategyAvalanche:
		return []string{"Minimizes total interest", "Mathematically optimal", "Fastest debt-free (usually)", "Guaranteed savings"}
	case domain.StrategySnowball:
		return []string{"Quick psychological wins", "Builds momentum", "Easier to stay motivated", "Simplifies finances faster"}
	case domain.StrategyCashFlow:
		return []string{"Frees up monthly budget", "Increases flexibility", "Reduces financial stress", "Good for tight budgets"}
	case domain.StrategyStress:
		return []string{"Reduces anxiety", "Improves mental health", "Eliminates embarrassing debts", "Peace of mind"}
	case domain.StrategyHybrid:
		return []string{"Customizable", "Balances multiple factors", "Adapts to your situation", "Flexible approach"}
	default:
		return []string{}
	}
}

func (m *DebtStrategyModel) getStrategyCons(strategy domain.Strategy) []string {
	switch strategy {
	case domain.StrategyAvalanche:
		return []string{"Slower initial wins", "Can feel discouraging", "Requires discipline"}
	case domain.StrategySnowball:
		return []string{"Pays more interest", "Not mathematically optimal", "May take longer"}
	case domain.StrategyCashFlow:
		return []string{"May pay more interest", "Not optimal for high-rate debt"}
	case domain.StrategyStress:
		return []string{"Ignores financial optimization", "May cost more overall"}
	case domain.StrategyHybrid:
		return []string{"Requires tuning", "More complex", "Results vary by weights"}
	default:
		return []string{}
	}
}
