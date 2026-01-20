package engine

import (
	"testing"
	"time"

	"personalfinancedss/internal/module/analytics/debt_tradeoff/domain"
)

// TestMonteCarloSimulator_StressScenarios tests resilience to bad market conditions
func TestMonteCarloSimulator_StressScenarios(t *testing.T) {
	simulator := NewMonteCarloSimulator()

	input := SimulationInput{
		Debts: []domain.DebtInfo{
			{Balance: 10000, InterestRate: 0.15, MinimumPayment: 200},
		},
		MonthlyIncome:     5000,
		EssentialExpenses: 4000, // High expenses tight budget
		TotalMinPayments:  200,
		Ratio:             domain.AllocationRatio{DebtPercent: 1.0, SavingsPercent: 0.0},
		Config: domain.SimulationConfig{
			NumSimulations:   500,
			IncomeVariance:   0.20, // High volatility
			ExpenseVariance:  0.20,
			ReturnVariance:   0.30, // Market crash potential
			ProjectionMonths: 36,
			DiscountRate:     0.05,
		},
	}

	result := simulator.RunSimulation(input)

	// With high variance, success probability should be less than 100%
	if result.SuccessProbability > 0.99 {
		t.Logf("Warning: Success probability is very high (%v) despite stress scenario", result.SuccessProbability)
	}

	// Check if there are significant failures (outliers)
	if result.NPVP5 > result.NPVMean {
		t.Errorf("P5 should be lower than Mean, got P5=%v Mean=%v", result.NPVP5, result.NPVMean)
	}
}

// TestMonteCarloSimulator_Convergence checks if results stabilize
func TestMonteCarloSimulator_Convergence(t *testing.T) {
	simulator := NewMonteCarloSimulator()

	baseInput := SimulationInput{
		Debts:             []domain.DebtInfo{{Balance: 5000, InterestRate: 0.1, MinimumPayment: 100}},
		MonthlyIncome:     3000,
		EssentialExpenses: 2000,
		TotalMinPayments:  100,
		Ratio:             domain.AllocationRatio{DebtPercent: 0.5, SavingsPercent: 0.5},
		Config: domain.SimulationConfig{
			NumSimulations:   50,
			ProjectionMonths: 24,
		},
	}

	// Run small batch
	res1 := simulator.RunSimulation(baseInput)

	// Run larger batch
	baseInput.Config.NumSimulations = 200
	res2 := simulator.RunSimulation(baseInput)

	// Means should be relatively close (within 20%)
	diff := (res1.NPVMean - res2.NPVMean) / res2.NPVMean
	if diff > 0.20 || diff < -0.20 {
		t.Logf("Warning: Convergence might be poor. Means: %v vs %v (Diff: %.2f%%)",
			res1.NPVMean, res2.NPVMean, diff*100)
	}
}

// TestMonteCarloSimulator_TimePerformance benchmarks simulation speed
func TestMonteCarloSimulator_TimePerformance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	simulator := NewMonteCarloSimulator()

	input := SimulationInput{
		Debts:             []domain.DebtInfo{{Balance: 10000, InterestRate: 0.1}},
		MonthlyIncome:     5000,
		EssentialExpenses: 3000,
		Ratio:             domain.AllocationRatio{DebtPercent: 0.5, SavingsPercent: 0.5},
		Config: domain.SimulationConfig{
			NumSimulations:   1000, // Heavy load
			ProjectionMonths: 60,
		},
	}

	start := time.Now()
	_ = simulator.RunSimulation(input)
	duration := time.Since(start)

	t.Logf("Run 1000 simulations (60 months): %v", duration)

	if duration > 2*time.Second {
		t.Errorf("Simulation too slow, took %v", duration)
	}
}
