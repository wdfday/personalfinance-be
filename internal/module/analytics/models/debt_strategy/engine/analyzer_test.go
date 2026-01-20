package engine

import (
	"personalfinancedss/internal/module/analytics/debt_strategy/domain"
	"testing"
)

func TestDebtAnalyzer_AnalyzeWhatIf_ExtraMonthly(t *testing.T) {
	analyzer := NewDebtAnalyzer()

	debts := []domain.DebtInfo{
		{ID: "cc", Name: "Credit Card", Balance: 5000, InterestRate: 0.18, MinimumPayment: 150},
		{ID: "car", Name: "Car Loan", Balance: 10000, InterestRate: 0.06, MinimumPayment: 200},
	}

	scenario := domain.WhatIfScenario{
		Type:        "extra_monthly",
		Amount:      200,
		Description: "What if I add $200/month?",
	}

	result := analyzer.AnalyzeWhatIf(scenario, debts, 550, domain.StrategyAvalanche, nil)

	if result.MonthsSaved <= 0 {
		t.Errorf("Expected months saved > 0, got %d", result.MonthsSaved)
	}

	if result.InterestSaved <= 0 {
		t.Errorf("Expected interest saved > 0, got %.2f", result.InterestSaved)
	}

	t.Logf("Extra $200/month: saves %d months and $%.2f interest", result.MonthsSaved, result.InterestSaved)
	t.Logf("Recommendation: %s", result.RecommendedAction)
}

func TestDebtAnalyzer_AnalyzeWhatIf_LumpSum(t *testing.T) {
	analyzer := NewDebtAnalyzer()

	debts := []domain.DebtInfo{
		{ID: "cc", Name: "Credit Card", Balance: 5000, InterestRate: 0.18, MinimumPayment: 150},
		{ID: "car", Name: "Car Loan", Balance: 10000, InterestRate: 0.06, MinimumPayment: 200},
	}

	scenario := domain.WhatIfScenario{
		Type:        "lump_sum",
		Amount:      3000,
		Description: "What if I pay $3000 bonus?",
	}

	result := analyzer.AnalyzeWhatIf(scenario, debts, 550, domain.StrategyAvalanche, nil)

	if result.BestDebtForLumpSum == "" {
		t.Error("Expected best debt recommendation")
	}

	if result.InterestSaved <= 0 {
		t.Errorf("Expected interest saved > 0, got %.2f", result.InterestSaved)
	}

	t.Logf("$3000 lump sum: best applied to '%s', saves $%.2f interest",
		result.BestDebtForLumpSum, result.InterestSaved)
}

func TestDebtAnalyzer_AnalyzeWhatIf_IncomeChange(t *testing.T) {
	analyzer := NewDebtAnalyzer()

	debts := []domain.DebtInfo{
		{ID: "cc", Name: "Credit Card", Balance: 5000, InterestRate: 0.18, MinimumPayment: 150},
	}

	// Test income decrease
	scenario := domain.WhatIfScenario{
		Type:        "income_change",
		Amount:      -0.20, // 20% decrease
		Description: "What if income drops 20%?",
	}

	result := analyzer.AnalyzeWhatIf(scenario, debts, 400, domain.StrategyAvalanche, nil)

	if result.NewMonths <= result.OriginalMonths {
		t.Errorf("Expected longer payoff with income decrease")
	}

	t.Logf("20%% income decrease: adds %d months and $%.2f interest",
		-result.MonthsSaved, -result.InterestSaved)
}

func TestDebtAnalyzer_AnalyzeRefinancing(t *testing.T) {
	analyzer := NewDebtAnalyzer()

	debts := []domain.DebtInfo{
		{ID: "cc1", Name: "Credit Card 1", Balance: 5000, InterestRate: 0.22, MinimumPayment: 150},
		{ID: "cc2", Name: "Credit Card 2", Balance: 3000, InterestRate: 0.19, MinimumPayment: 90},
	}

	option := domain.RefinanceOption{
		NewRate:        0.10, // 10% consolidation loan
		NewTerm:        36,
		OriginationFee: 200,
		MonthlyFee:     0,
		IncludeDebtIDs: []string{"cc1", "cc2"},
	}

	result := analyzer.AnalyzeRefinancing(debts, option, 200)

	if result.CurrentWeightedRate <= result.NewEffectiveRate {
		t.Logf("Current rate: %.2f%%, New rate: %.2f%%",
			result.CurrentWeightedRate*100, result.NewEffectiveRate*100)
	}

	t.Logf("Refinancing analysis:")
	t.Logf("  Should refinance: %v", result.ShouldRefinance)
	t.Logf("  Net savings: $%.2f", result.NetSavings)
	t.Logf("  Break-even: %d months", result.BreakEvenMonths)
	t.Logf("  Recommendation: %s", result.Recommendation)
}

func TestDebtAnalyzer_AnalyzeRefinancing_NotWorthIt(t *testing.T) {
	analyzer := NewDebtAnalyzer()

	debts := []domain.DebtInfo{
		{ID: "car", Name: "Car Loan", Balance: 10000, InterestRate: 0.05, MinimumPayment: 200},
	}

	option := domain.RefinanceOption{
		NewRate:        0.08, // Higher rate!
		NewTerm:        48,
		OriginationFee: 500,
		MonthlyFee:     10,
	}

	result := analyzer.AnalyzeRefinancing(debts, option, 100)

	if result.ShouldRefinance {
		t.Error("Should not recommend refinancing to higher rate")
	}

	if len(result.Warnings) == 0 {
		t.Error("Expected warnings about higher rate")
	}

	t.Logf("Correctly rejected refinancing: %s", result.Recommendation)
}

func TestDebtAnalyzer_AnalyzeSensitivity(t *testing.T) {
	analyzer := NewDebtAnalyzer()

	debts := []domain.DebtInfo{
		{ID: "cc", Name: "Credit Card", Balance: 5000, InterestRate: 0.18, MinimumPayment: 150, IsVariableRate: true},
		{ID: "car", Name: "Car Loan", Balance: 10000, InterestRate: 0.06, MinimumPayment: 200},
	}

	results := analyzer.AnalyzeSensitivity(debts, 600, domain.StrategyAvalanche, nil)

	if len(results) == 0 {
		t.Error("Expected sensitivity results")
	}

	for _, r := range results {
		t.Logf("Scenario: %s", r.Scenario.Description)
		t.Logf("  Impact: %d months, $%.2f interest", r.MonthsImpact, r.InterestImpact)
		t.Logf("  Risk level: %s", r.RiskLevel)
		t.Logf("  Strategy still valid: %v", r.StrategyStillValid)
	}
}

func TestDebtAnalyzer_CalculatePsychologicalScore(t *testing.T) {
	analyzer := NewDebtAnalyzer()
	simulator := NewDebtSimulator()

	debts := []domain.DebtInfo{
		{ID: "small", Name: "Small Debt", Balance: 500, InterestRate: 0.10, MinimumPayment: 50},
		{ID: "medium", Name: "Medium Debt", Balance: 3000, InterestRate: 0.15, MinimumPayment: 100},
		{ID: "large", Name: "Large Debt", Balance: 10000, InterestRate: 0.08, MinimumPayment: 200},
	}

	result := simulator.SimulateStrategy(domain.StrategySnowball, debts, 300, nil)
	score := analyzer.CalculatePsychologicalScore(result, debts)

	if score.QuickWinsCount == 0 {
		t.Log("No quick wins in first 6 months")
	}

	if score.MotivationScore < 0 || score.MotivationScore > 100 {
		t.Errorf("Motivation score out of range: %.2f", score.MotivationScore)
	}

	if len(score.Celebrations) == 0 {
		t.Error("Expected celebration messages")
	}

	t.Logf("Psychological Score:")
	t.Logf("  Quick wins: %d", score.QuickWinsCount)
	t.Logf("  First win: month %d", score.FirstWinMonth)
	t.Logf("  Motivation: %.0f/100", score.MotivationScore)
	t.Logf("  Momentum: %s", score.MomentumRating)
	for _, c := range score.Celebrations {
		t.Logf("  %s", c)
	}
}

func TestDebtAnalyzer_GenerateMilestones(t *testing.T) {
	analyzer := NewDebtAnalyzer()
	simulator := NewDebtSimulator()

	debts := []domain.DebtInfo{
		{ID: "d1", Name: "Debt 1", Balance: 2000, InterestRate: 0.15, MinimumPayment: 100},
		{ID: "d2", Name: "Debt 2", Balance: 5000, InterestRate: 0.10, MinimumPayment: 150},
	}

	result := simulator.SimulateStrategy(domain.StrategySnowball, debts, 200, nil)
	milestones := analyzer.GenerateMilestones(result)

	if len(milestones) < 3 {
		t.Errorf("Expected at least 3 milestones (2 debts + debt-free), got %d", len(milestones))
	}

	hasDebtFree := false
	hasHalfway := false
	for _, m := range milestones {
		if m.Type == "debt_free" {
			hasDebtFree = true
		}
		if m.Type == "halfway" {
			hasHalfway = true
		}
		t.Logf("Month %d: %s (%s)", m.Month, m.Description, m.Type)
	}

	if !hasDebtFree {
		t.Error("Missing debt-free milestone")
	}
	if !hasHalfway {
		t.Error("Missing halfway milestone")
	}
}
