package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCalculateForecast_FixedMode(t *testing.T) {
	input := ForecastInput{
		Mode:        IncomeModeFixed,
		RolloverTBB: 5000000, // 5M rollover
		IncomeSources: []IncomeSource{
			{Name: "Salary", Amount: 20000000, IsRecurring: true, SourceType: "salary"},
		},
		MonthlyExpense: 15000000,
	}

	result := CalculateForecast(input)

	assert.Equal(t, float64(20000000), result.ProjectedIncome)
	assert.Equal(t, float64(20000000), result.FixedIncome)
	assert.Equal(t, float64(25000000), result.TotalAvailable) // Income + Rollover
	assert.Equal(t, float64(5000000), result.RolloverFromPrev)
}

func TestCalculateForecast_FreelancerMode(t *testing.T) {
	input := ForecastInput{
		Mode:        IncomeModeFreelancer,
		RolloverTBB: 0,
		IncomeSources: []IncomeSource{
			{Name: "Project A", Amount: 50000000, IsRecurring: false, SourceType: "freelance"},
		},
		HistoricalIncome: []float64{
			30000000, 20000000, 40000000, 10000000, 50000000, 25000000,
			35000000, 15000000, 45000000, 20000000, 30000000, 40000000,
		}, // 12 months, avg = 30M
		FreelancerConfig: &FreelancerConfig{
			ReservoirBalance:   100000000, // 100M in reservoir
			SafeSalaryMonths:   6,
			SmoothingFactor:    0.8,
			TaxWithholdingRate: 0.1,
		},
		MonthlyExpense: 20000000,
	}

	result := CalculateForecast(input)

	// Safe salary should be min(30M * 0.8, 100M / 6) = min(24M, 16.67M) = 16.67M
	assert.InDelta(t, 16666666.67, result.SafeSalary, 1)

	// Tax reserve = 50M * 10% = 5M
	assert.Equal(t, float64(5000000), result.TaxReserve)

	// Reservoir deposit = (50M - 5M) - 16.67M ≈ 28.33M
	assert.InDelta(t, 28333333.33, result.ReservoirDeposit, 1)

	// Total available = Safe Salary + Rollover = ~16.67M
	assert.InDelta(t, 16666666.67, result.TotalAvailable, 1)
}

func TestCalculateForecast_MixedMode(t *testing.T) {
	input := ForecastInput{
		Mode:        IncomeModeMixed,
		RolloverTBB: 2000000,
		IncomeSources: []IncomeSource{
			{Name: "Salary", Amount: 15000000, IsRecurring: true, SourceType: "salary"},
			{Name: "Freelance", Amount: 20000000, IsRecurring: false, SourceType: "freelance"},
		},
		HistoricalIncome: []float64{
			30000000, 25000000, 35000000, 28000000, 32000000, 27000000,
			33000000, 26000000, 34000000, 29000000, 31000000, 30000000,
		},
		FreelancerConfig: &FreelancerConfig{
			SmoothingFactor:    0.8,
			TaxWithholdingRate: 0.1,
		},
		MonthlyExpense: 20000000,
	}

	result := CalculateForecast(input)

	assert.Equal(t, float64(15000000), result.FixedIncome)
	assert.Equal(t, float64(20000000), result.VariableIncome)

	// Total = Fixed + Smoothed Variable + Rollover
	// Variable avg = (total_avg - fixed) ≈ 30M - 15M = 15M
	// Smoothed = 15M * 0.8 = 12M
	// Total = 15M + 12M + 2M = 29M
	assert.Greater(t, result.TotalAvailable, float64(25000000))
}

func TestCalculateStability_LowRisk(t *testing.T) {
	// Very stable income (salaried)
	history := []float64{
		20000000, 20000000, 20000000, 20000000, 20000000, 20000000,
		20000000, 20000000, 20000000, 20000000, 20000000, 20000000,
	}

	metrics := CalculateStability(history, 15000000)

	assert.Equal(t, float64(0), metrics.CoefficientOfVariation)
	assert.Equal(t, float64(100), metrics.StabilityScore)
	assert.Equal(t, "low", metrics.RiskLevel)
}

func TestCalculateStability_HighRisk(t *testing.T) {
	// Highly variable income (freelancer)
	history := []float64{
		50000000, 10000000, 80000000, 5000000, 40000000, 15000000,
		70000000, 8000000, 60000000, 12000000, 45000000, 20000000,
	}

	metrics := CalculateStability(history, 25000000)

	assert.Greater(t, metrics.CoefficientOfVariation, 0.5) // High variance
	assert.Less(t, metrics.StabilityScore, float64(50))    // Low stability
	assert.Contains(t, []string{"high", "critical"}, metrics.RiskLevel)
	assert.Greater(t, metrics.RecommendedVIB, float64(0))
}

func TestCalculateStability_InsufficientData(t *testing.T) {
	history := []float64{20000000, 25000000} // Only 2 months

	metrics := CalculateStability(history, 15000000)

	assert.Equal(t, "unknown", metrics.RiskLevel)
	assert.Equal(t, float64(50), metrics.StabilityScore)
}

func TestDefaultFreelancerConfig(t *testing.T) {
	config := DefaultFreelancerConfig()

	assert.Equal(t, 6, config.SafeSalaryMonths)
	assert.Equal(t, 0.8, config.SmoothingFactor)
	assert.Equal(t, 0.1, config.TaxWithholdingRate)
}
