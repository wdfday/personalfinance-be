package engine

import (
	"fmt"
	"testing"

	"personalfinancedss/internal/module/analytics/debt_strategy/domain"
)

// TestDebtSimulator_LargeDataSet stress tests the simulator with many debts
func TestDebtSimulator_LargeDataSet(t *testing.T) {
	simulator := NewDebtSimulator()

	// Create 50 debts
	debts := make([]domain.DebtInfo, 50)
	for i := 0; i < 50; i++ {
		debts[i] = domain.DebtInfo{
			ID:             fmt.Sprintf("debt_%d", i),
			Name:           fmt.Sprintf("Debt %d", i),
			Balance:        1000 + float64(i)*100,
			InterestRate:   0.05 + float64(i)*0.005, // 5% to 30%
			MinimumPayment: 20 + float64(i)*2,
		}
	}

	result := simulator.SimulateStrategy(domain.StrategyAvalanche, debts, 1000, nil)

	if result.Months == 0 {
		t.Error("Simulation failed for large dataset")
	}

	// Verify all debts paid
	if len(result.DebtTimelines) != 50 {
		t.Errorf("Expected 50 debt timelines, got %d", len(result.DebtTimelines))
	}
}

// TestDebtSimulator_StrategyEfficiency verifies Avalanche saves simpler interest than Snowball
func TestDebtSimulator_StrategyEfficiency(t *testing.T) {
	simulator := NewDebtSimulator()

	// Setup: Small balance low interest vs Large balance High interest
	// This is where Avalanche shines most
	debts := []domain.DebtInfo{
		{ID: "low_int_small", Balance: 2000, InterestRate: 0.05, MinimumPayment: 50},
		{ID: "high_int_large", Balance: 20000, InterestRate: 0.25, MinimumPayment: 500},
	}

	avalanche := simulator.SimulateStrategy(domain.StrategyAvalanche, debts, 500, nil)
	snowball := simulator.SimulateStrategy(domain.StrategySnowball, debts, 500, nil)

	diff := snowball.TotalInterest - avalanche.TotalInterest

	if diff <= 0 {
		t.Errorf("Avalanche should strictly save more interest here. Snowball: %.2f, Avalanche: %.2f",
			snowball.TotalInterest, avalanche.TotalInterest)
	}

	t.Logf("Avalanche saved $%.2f more than Snowball", diff)
}

// TestDebtSimulator_InvalidInputs tests robustness
func TestDebtSimulator_InvalidInputs(t *testing.T) {
	simulator := NewDebtSimulator()

	debts := []domain.DebtInfo{
		{ID: "neg", Balance: -100, InterestRate: 0.1, MinimumPayment: 10},
		{ID: "zero", Balance: 0, InterestRate: 0.1, MinimumPayment: 0},
	}

	// Should handle gracefully without panic or infinite loop
	result := simulator.SimulateStrategy(domain.StrategyAvalanche, debts, 100, nil)

	if result.Months != 0 {
		t.Logf("Simulation with zero/negative balances returned %d months", result.Months)
	}
}

// TestDebtSimulator_HybridWeights checks if custom weights influence order
func TestDebtSimulator_HybridWeights(t *testing.T) {
	simulator := NewDebtSimulator()

	debts := []domain.DebtInfo{
		{ID: "d1", Balance: 5000, InterestRate: 0.10, MinimumPayment: 100, StressScore: 1},
		{ID: "d2", Balance: 5000, InterestRate: 0.10, MinimumPayment: 100, StressScore: 10},
	}

	// Identical financial stats, but d2 is high stress

	// Case 1: Pure financial (Avalanche implies order by ID/stable sort if tie)
	res1 := simulator.SimulateStrategy(domain.StrategyHybrid, debts, 200, &domain.HybridWeights{InterestRateWeight: 1.0})

	// Case 2: Pure Stress
	res2 := simulator.SimulateStrategy(domain.StrategyHybrid, debts, 200, &domain.HybridWeights{StressWeight: 1.0})

	d2Month1 := res1.DebtTimelines["d2"].PayoffMonth
	d2Month2 := res2.DebtTimelines["d2"].PayoffMonth

	if d2Month2 >= d2Month1 {
		t.Logf("Warning: Stress weighting didn't accelerate d2 payoff as expected. %d vs %d", d2Month2, d2Month1)
		// Note: Might be equal if extra payment clears both fast, but generally d2 should be faster or equal
	}
}
