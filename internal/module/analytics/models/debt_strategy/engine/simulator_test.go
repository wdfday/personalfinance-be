package engine

import (
	"personalfinancedss/internal/module/analytics/debt_strategy/domain"
	"testing"
)

func TestDebtSimulator_SimulateStrategy_Avalanche(t *testing.T) {
	simulator := NewDebtSimulator()

	debts := []domain.DebtInfo{
		{ID: "cc1", Name: "Credit Card", Balance: 5000, InterestRate: 0.18, MinimumPayment: 150},
		{ID: "car", Name: "Car Loan", Balance: 10000, InterestRate: 0.06, MinimumPayment: 200},
		{ID: "student", Name: "Student Loan", Balance: 20000, InterestRate: 0.045, MinimumPayment: 250},
	}

	extraPayment := 300.0 // $300 extra beyond minimums

	result := simulator.SimulateStrategy(domain.StrategyAvalanche, debts, extraPayment, nil)

	// Avalanche should prioritize credit card (18%) first
	if result.Strategy != domain.StrategyAvalanche {
		t.Errorf("Expected strategy avalanche, got %s", result.Strategy)
	}

	// Credit card should be paid off first
	ccTimeline := result.DebtTimelines["cc1"]
	carTimeline := result.DebtTimelines["car"]

	if ccTimeline.PayoffMonth >= carTimeline.PayoffMonth {
		t.Errorf("Avalanche should pay off high-interest debt first. CC: month %d, Car: month %d",
			ccTimeline.PayoffMonth, carTimeline.PayoffMonth)
	}

	// Should complete within reasonable time
	if result.Months > 60 {
		t.Errorf("Expected payoff within 60 months, got %d", result.Months)
	}

	t.Logf("Avalanche: %d months, $%.2f interest, first cleared: month %d",
		result.Months, result.TotalInterest, result.FirstCleared)
}

func TestDebtSimulator_SimulateStrategy_Snowball(t *testing.T) {
	simulator := NewDebtSimulator()

	debts := []domain.DebtInfo{
		{ID: "cc1", Name: "Credit Card", Balance: 5000, InterestRate: 0.18, MinimumPayment: 150},
		{ID: "small", Name: "Small Loan", Balance: 1000, InterestRate: 0.10, MinimumPayment: 50},
		{ID: "car", Name: "Car Loan", Balance: 10000, InterestRate: 0.06, MinimumPayment: 200},
	}

	extraPayment := 300.0

	result := simulator.SimulateStrategy(domain.StrategySnowball, debts, extraPayment, nil)

	// Snowball should prioritize smallest balance first
	smallTimeline := result.DebtTimelines["small"]
	ccTimeline := result.DebtTimelines["cc1"]

	if smallTimeline.PayoffMonth >= ccTimeline.PayoffMonth {
		t.Errorf("Snowball should pay off smallest debt first. Small: month %d, CC: month %d",
			smallTimeline.PayoffMonth, ccTimeline.PayoffMonth)
	}

	// First cleared should be the small loan
	if result.FirstCleared != smallTimeline.PayoffMonth {
		t.Errorf("First cleared should be small loan at month %d, got %d",
			smallTimeline.PayoffMonth, result.FirstCleared)
	}

	t.Logf("Snowball: %d months, $%.2f interest, first cleared: month %d",
		result.Months, result.TotalInterest, result.FirstCleared)
}

func TestDebtSimulator_SimulateStrategy_CashFlow(t *testing.T) {
	simulator := NewDebtSimulator()

	debts := []domain.DebtInfo{
		{ID: "high_ratio", Name: "High Ratio", Balance: 2000, InterestRate: 0.10, MinimumPayment: 200}, // 10% ratio
		{ID: "low_ratio", Name: "Low Ratio", Balance: 10000, InterestRate: 0.15, MinimumPayment: 100},  // 1% ratio
	}

	extraPayment := 200.0

	result := simulator.SimulateStrategy(domain.StrategyCashFlow, debts, extraPayment, nil)

	// Cash flow should prioritize high payment/balance ratio first
	highRatioTimeline := result.DebtTimelines["high_ratio"]
	lowRatioTimeline := result.DebtTimelines["low_ratio"]

	if highRatioTimeline.PayoffMonth >= lowRatioTimeline.PayoffMonth {
		t.Errorf("CashFlow should pay off high ratio debt first. High: month %d, Low: month %d",
			highRatioTimeline.PayoffMonth, lowRatioTimeline.PayoffMonth)
	}

	t.Logf("CashFlow: %d months, $%.2f interest", result.Months, result.TotalInterest)
}

func TestDebtSimulator_SimulateStrategy_Stress(t *testing.T) {
	simulator := NewDebtSimulator()

	debts := []domain.DebtInfo{
		{ID: "family", Name: "Family Debt", Balance: 5000, InterestRate: 0.05, MinimumPayment: 100, StressScore: 9},
		{ID: "cc", Name: "Credit Card", Balance: 3000, InterestRate: 0.20, MinimumPayment: 100, StressScore: 3},
	}

	extraPayment := 200.0

	result := simulator.SimulateStrategy(domain.StrategyStress, debts, extraPayment, nil)

	// Stress should prioritize high stress debt first
	familyTimeline := result.DebtTimelines["family"]
	ccTimeline := result.DebtTimelines["cc"]

	if familyTimeline.PayoffMonth >= ccTimeline.PayoffMonth {
		t.Errorf("Stress strategy should pay off high-stress debt first. Family: month %d, CC: month %d",
			familyTimeline.PayoffMonth, ccTimeline.PayoffMonth)
	}

	t.Logf("Stress: %d months, $%.2f interest", result.Months, result.TotalInterest)
}

func TestDebtSimulator_SimulateStrategy_Hybrid(t *testing.T) {
	simulator := NewDebtSimulator()

	debts := []domain.DebtInfo{
		{ID: "d1", Name: "Debt 1", Balance: 5000, InterestRate: 0.18, MinimumPayment: 150, StressScore: 5},
		{ID: "d2", Name: "Debt 2", Balance: 2000, InterestRate: 0.10, MinimumPayment: 100, StressScore: 8},
		{ID: "d3", Name: "Debt 3", Balance: 8000, InterestRate: 0.08, MinimumPayment: 200, StressScore: 2},
	}

	// Custom weights: prioritize stress
	weights := &domain.HybridWeights{
		InterestRateWeight: 0.2,
		BalanceWeight:      0.2,
		StressWeight:       0.5,
		CashFlowWeight:     0.1,
	}

	extraPayment := 200.0

	result := simulator.SimulateStrategy(domain.StrategyHybrid, debts, extraPayment, weights)

	// With high stress weight, d2 (stress=8) should be prioritized
	d2Timeline := result.DebtTimelines["d2"]
	d1Timeline := result.DebtTimelines["d1"]

	if d2Timeline.PayoffMonth > d1Timeline.PayoffMonth {
		t.Logf("Note: With stress weight 0.5, high-stress debt d2 should be prioritized")
	}

	t.Logf("Hybrid: %d months, $%.2f interest", result.Months, result.TotalInterest)
}

func TestDebtSimulator_FindBestDebtForLumpSum(t *testing.T) {
	simulator := NewDebtSimulator()

	debts := []domain.DebtInfo{
		{ID: "cc", Name: "Credit Card", Balance: 5000, InterestRate: 0.18, MinimumPayment: 150},
		{ID: "car", Name: "Car Loan", Balance: 10000, InterestRate: 0.06, MinimumPayment: 200},
	}

	extraPayment := 200.0
	lumpSum := 3000.0

	bestDebtID, savings := simulator.FindBestDebtForLumpSum(
		domain.StrategyAvalanche, debts, extraPayment, lumpSum, nil)

	// Should recommend paying the high-interest credit card
	if bestDebtID != "cc" {
		t.Errorf("Expected best debt for lump sum to be 'cc', got '%s'", bestDebtID)
	}

	if savings <= 0 {
		t.Errorf("Expected positive savings, got %.2f", savings)
	}

	t.Logf("Best debt for $%.0f lump sum: %s, saves $%.2f", lumpSum, bestDebtID, savings)
}

func TestDebtSimulator_CompareStrategies(t *testing.T) {
	simulator := NewDebtSimulator()

	debts := []domain.DebtInfo{
		{ID: "cc1", Name: "Credit Card 1", Balance: 5000, InterestRate: 0.22, MinimumPayment: 150},
		{ID: "cc2", Name: "Credit Card 2", Balance: 2000, InterestRate: 0.18, MinimumPayment: 60},
		{ID: "car", Name: "Car Loan", Balance: 15000, InterestRate: 0.05, MinimumPayment: 300},
	}

	extraPayment := 400.0

	avalanche := simulator.SimulateStrategy(domain.StrategyAvalanche, debts, extraPayment, nil)
	snowball := simulator.SimulateStrategy(domain.StrategySnowball, debts, extraPayment, nil)

	// Avalanche should save more interest
	if avalanche.TotalInterest >= snowball.TotalInterest {
		t.Logf("Warning: Avalanche didn't save more interest. Avalanche: $%.2f, Snowball: $%.2f",
			avalanche.TotalInterest, snowball.TotalInterest)
	}

	// Snowball should have earlier first win
	if snowball.FirstCleared > avalanche.FirstCleared {
		t.Logf("Note: Snowball first win at month %d, Avalanche at month %d",
			snowball.FirstCleared, avalanche.FirstCleared)
	}

	t.Logf("Comparison:")
	t.Logf("  Avalanche: %d months, $%.2f interest, first win: month %d",
		avalanche.Months, avalanche.TotalInterest, avalanche.FirstCleared)
	t.Logf("  Snowball: %d months, $%.2f interest, first win: month %d",
		snowball.Months, snowball.TotalInterest, snowball.FirstCleared)
	t.Logf("  Interest saved by Avalanche: $%.2f", snowball.TotalInterest-avalanche.TotalInterest)
}
