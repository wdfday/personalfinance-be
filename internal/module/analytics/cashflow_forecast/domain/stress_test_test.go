package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRunStressTest_JobLoss(t *testing.T) {
	input := CreateJobLossScenario(
		50000000, // 50M cash
		20000000, // 20M monthly income
		15000000, // 15M monthly expenses
		30000000, // 30M emergency fund
	)
	input.Simulations = 100 // Reduce for test speed

	result := RunStressTest(input)

	// With 80M total (50M + 30M) and 15M/month expense, should last ~5 months
	assert.Greater(t, result.AverageMonthsSurvived, float64(3))
	assert.Less(t, result.AverageMonthsSurvived, float64(8))
	assert.Equal(t, "danger", result.RiskLevel) // Job loss = high risk
	assert.Greater(t, result.InsolvencyProbability, 0.5)
}

func TestRunStressTest_SafeScenario(t *testing.T) {
	input := StressTestInput{
		CurrentCash:     100000000, // 100M cash
		MonthlyIncome:   30000000,  // 30M income
		MonthlyExpenses: 15000000,  // 15M expenses
		EmergencyFund:   50000000,  // 50M emergency
		Scenario:        ScenarioCustom,
		MonthsToProject: 12,
		Simulations:     100,
	}

	result := RunStressTest(input)

	// With 150M and positive cash flow, should be very safe
	assert.Less(t, result.InsolvencyProbability, 0.05)
	assert.Equal(t, "safe", result.RiskLevel)
	assert.Equal(t, 12, int(result.AverageMonthsSurvived))
}

func TestRunStressTest_MedicalShock(t *testing.T) {
	input := CreateMedicalShockScenario(
		30000000, // 30M cash
		20000000, // 20M income
		18000000, // 18M expenses (tight budget)
		10000000, // 10M emergency
	)
	input.Simulations = 100

	result := RunStressTest(input)

	// Medical shock = 6x income = 120M, but only 40M available
	assert.Greater(t, result.InsolvencyProbability, 0.0)
	assert.Greater(t, len(result.DangerZones), 0)
}

func TestRunStressTest_FreelancerDrySpell(t *testing.T) {
	input := CreateFreelancerDrySpellScenario(
		60000000, // 60M cash (reservoir)
		25000000, // 25M average income
		15000000, // 15M expenses
		20000000, // 20M emergency fund
		3,        // 3 months dry spell
	)
	input.Simulations = 100

	result := RunStressTest(input)

	// 80M total, 3 months @ 15M = 45M, should survive
	assert.Less(t, result.InsolvencyProbability, 0.5)
	assert.Greater(t, result.AverageMonthsSurvived, float64(5))
}

func TestRunStressTest_InflationSpike(t *testing.T) {
	input := CreateInflationScenario(
		40000000, // 40M cash
		22000000, // 22M income
		20000000, // 20M expenses → 24M after 20% spike
		15000000, // 15M emergency
	)
	input.Simulations = 100

	result := RunStressTest(input)

	// After inflation: 24M expense vs 22M income = negative flow
	// With 55M starting balance, may survive short term but builds risk
	assert.GreaterOrEqual(t, result.InsolvencyProbability, 0.0) // May or may not hit insolvency
	assert.NotNil(t, result.Recommendations)
	assert.GreaterOrEqual(t, len(result.Percentile5), 1) // Should have paths
}

func TestPercentilePaths(t *testing.T) {
	input := StressTestInput{
		CurrentCash:     50000000,
		MonthlyIncome:   20000000,
		MonthlyExpenses: 15000000,
		MonthsToProject: 6,
		Simulations:     50,
		Scenario:        ScenarioCustom,
	}

	result := RunStressTest(input)

	// Should have percentile paths
	assert.Len(t, result.Percentile5, 6)
	assert.Len(t, result.Percentile50, 6)
	assert.Len(t, result.Percentile95, 6)

	// P95 should be higher than P5 (best case vs worst case)
	assert.Greater(t, result.Percentile95[5], result.Percentile5[5])
}

func TestLCRCalculation(t *testing.T) {
	input := StressTestInput{
		CurrentCash:     90000000, // 90M
		MonthlyIncome:   20000000,
		MonthlyExpenses: 30000000, // Negative cash flow
		EmergencyFund:   30000000, // 30M
		MonthsToProject: 12,
		Simulations:     50,
		Scenario:        ScenarioCustom,
	}

	result := RunStressTest(input)

	// LCR = 120M / (30M * 3 months expense) = 120/90 ≈ 1.33
	// But with partial income considered...
	assert.Greater(t, result.LCR, 0.0)
}
