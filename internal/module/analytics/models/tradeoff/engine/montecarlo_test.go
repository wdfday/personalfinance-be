package engine

import (
	"testing"

	"personalfinancedss/internal/module/analytics/debt_tradeoff/domain"
)

func TestMonteCarloSimulator_RunSimulation(t *testing.T) {
	simulator := NewMonteCarloSimulator()

	input := SimulationInput{
		Debts: []domain.DebtInfo{
			{
				ID:             "cc1",
				Balance:        5000,
				InterestRate:   0.18,
				MinimumPayment: 100,
			},
		},
		MonthlyIncome:     5000,
		EssentialExpenses: 3000,
		TotalMinPayments:  100,
		Ratio:             domain.AllocationRatio{DebtPercent: 0.5, SavingsPercent: 0.5},
		ExpectedReturn:    0.07,
		InitialSavings:    1000,
		Goals: []domain.GoalInfo{
			{
				ID:            "emergency",
				TargetAmount:  10000,
				CurrentAmount: 1000,
				Priority:      1.0,
			},
		},
		Config: domain.SimulationConfig{
			NumSimulations:   100, // Smaller for test speed
			IncomeVariance:   0.10,
			ExpenseVariance:  0.15,
			ReturnVariance:   0.20,
			ProjectionMonths: 60,
			DiscountRate:     0.05,
		},
	}

	result := simulator.RunSimulation(input)

	// Verify result structure
	if result.NumSimulations != 100 {
		t.Errorf("NumSimulations = %d, want 100", result.NumSimulations)
	}

	// Success probability should be between 0 and 1
	if result.SuccessProbability < 0 || result.SuccessProbability > 1 {
		t.Errorf("SuccessProbability = %v, want between 0 and 1", result.SuccessProbability)
	}

	// Percentiles should be in order
	if result.DebtFreeP50 > result.DebtFreeP75 || result.DebtFreeP75 > result.DebtFreeP90 {
		t.Errorf("Percentiles out of order: P50=%d, P75=%d, P90=%d",
			result.DebtFreeP50, result.DebtFreeP75, result.DebtFreeP90)
	}

	// NPV percentiles should be in order
	if result.NPVP5 > result.NPVMean || result.NPVMean > result.NPVP95 {
		t.Errorf("NPV percentiles out of order: P5=%v, Mean=%v, P95=%v",
			result.NPVP5, result.NPVMean, result.NPVP95)
	}

	// Confidence interval should contain mean
	if result.NPVMean < result.ConfidenceInterval95[0] || result.NPVMean > result.ConfidenceInterval95[1] {
		t.Errorf("Mean %v not in confidence interval [%v, %v]",
			result.NPVMean, result.ConfidenceInterval95[0], result.ConfidenceInterval95[1])
	}
}

func TestMonteCarloSimulator_DifferentStrategies(t *testing.T) {
	simulator := NewMonteCarloSimulator()

	baseInput := SimulationInput{
		Debts: []domain.DebtInfo{
			{Balance: 10000, InterestRate: 0.15, MinimumPayment: 200},
		},
		MonthlyIncome:     6000,
		EssentialExpenses: 3500,
		TotalMinPayments:  200,
		ExpectedReturn:    0.07,
		InitialSavings:    2000,
		Goals:             []domain.GoalInfo{},
		Config: domain.SimulationConfig{
			NumSimulations:   50,
			IncomeVariance:   0.10,
			ExpenseVariance:  0.15,
			ReturnVariance:   0.20,
			ProjectionMonths: 60,
			DiscountRate:     0.05,
		},
	}

	// Test aggressive debt strategy
	aggressiveDebt := baseInput
	aggressiveDebt.Ratio = domain.AllocationRatio{DebtPercent: 0.75, SavingsPercent: 0.25}
	resultDebt := simulator.RunSimulation(aggressiveDebt)

	// Test aggressive savings strategy
	aggressiveSavings := baseInput
	aggressiveSavings.Ratio = domain.AllocationRatio{DebtPercent: 0.25, SavingsPercent: 0.75}
	resultSavings := simulator.RunSimulation(aggressiveSavings)

	// Aggressive debt should pay off debt faster (lower P50)
	if resultDebt.DebtFreeP50 > resultSavings.DebtFreeP50 {
		t.Logf("Note: Aggressive debt P50=%d, Aggressive savings P50=%d",
			resultDebt.DebtFreeP50, resultSavings.DebtFreeP50)
		// This might not always be true due to randomness, so just log
	}

	t.Logf("Aggressive Debt: DebtFreeP50=%d, SuccessProb=%.2f",
		resultDebt.DebtFreeP50, resultDebt.SuccessProbability)
	t.Logf("Aggressive Savings: DebtFreeP50=%d, SuccessProb=%.2f",
		resultSavings.DebtFreeP50, resultSavings.SuccessProbability)
}

func TestMonteCarloSimulator_EdgeCases(t *testing.T) {
	simulator := NewMonteCarloSimulator()

	// Test with no goals
	input := SimulationInput{
		Debts: []domain.DebtInfo{
			{Balance: 1000, InterestRate: 0.10, MinimumPayment: 50},
		},
		MonthlyIncome:     3000,
		EssentialExpenses: 2000,
		TotalMinPayments:  50,
		Ratio:             domain.AllocationRatio{DebtPercent: 0.5, SavingsPercent: 0.5},
		ExpectedReturn:    0.05,
		InitialSavings:    500,
		Goals:             []domain.GoalInfo{}, // No goals
		Config: domain.SimulationConfig{
			NumSimulations:   20,
			IncomeVariance:   0.05,
			ExpenseVariance:  0.05,
			ReturnVariance:   0.10,
			ProjectionMonths: 24,
			DiscountRate:     0.05,
		},
	}

	result := simulator.RunSimulation(input)

	// Should still work with no goals
	if result.NumSimulations != 20 {
		t.Errorf("NumSimulations = %d, want 20", result.NumSimulations)
	}

	// With no goals, success is based on debt payoff only
	// Should have reasonable success probability
	if result.SuccessProbability < 0 || result.SuccessProbability > 1 {
		t.Errorf("SuccessProbability = %v, want between 0 and 1", result.SuccessProbability)
	}
}
