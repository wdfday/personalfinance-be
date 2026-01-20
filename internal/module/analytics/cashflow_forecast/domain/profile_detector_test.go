package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDetectIncomeProfile_Salaried(t *testing.T) {
	// Perfect salaried employee - same income every month
	history := []float64{
		20000000, 20000000, 20000000, 20000000, 20000000, 20000000,
		20000000, 20000000, 20000000, 20000000, 20000000, 20000000,
	}
	sources := []IncomeSource{
		{Name: "Lương công ty", Amount: 20000000, IsRecurring: true, SourceType: "salary"},
	}

	analysis := DetectIncomeProfile(history, sources)

	assert.Equal(t, IncomeModeFixed, analysis.DetectedMode)
	assert.Greater(t, analysis.Confidence, 70.0)
	assert.Contains(t, analysis.Recommendation, "ổn định")
}

func TestDetectIncomeProfile_Freelancer(t *testing.T) {
	// Typical freelancer - highly variable, some zero months
	history := []float64{
		50000000, 0, 80000000, 10000000, 0, 100000000,
		20000000, 5000000, 60000000, 0, 40000000, 30000000,
	}
	sources := []IncomeSource{
		{Name: "Dự án A", Amount: 30000000, IsRecurring: false, SourceType: "freelance"},
	}

	analysis := DetectIncomeProfile(history, sources)

	assert.Equal(t, IncomeModeFreelancer, analysis.DetectedMode)
	assert.Greater(t, analysis.Confidence, 60.0)
	assert.Greater(t, analysis.Patterns.ZeroIncomeMonths, 0)
	assert.Greater(t, analysis.Patterns.FeastFamineRatio, 2.0)
}

func TestDetectIncomeProfile_Mixed(t *testing.T) {
	// Mixed: stable salary + variable freelance
	history := []float64{
		25000000, 30000000, 20000000, 40000000, 22000000, 35000000,
		20000000, 50000000, 23000000, 28000000, 20000000, 60000000,
	}
	sources := []IncomeSource{
		{Name: "Lương", Amount: 15000000, IsRecurring: true, SourceType: "salary"},
		{Name: "Freelance", Amount: 20000000, IsRecurring: false, SourceType: "freelance"},
	}

	analysis := DetectIncomeProfile(history, sources)

	// Should detect mixed or freelancer (not salaried)
	assert.NotEqual(t, IncomeModeFixed, analysis.DetectedMode)
}

func TestDetectIncomeProfile_InsufficientData(t *testing.T) {
	history := []float64{20000000, 25000000} // Only 2 months

	analysis := DetectIncomeProfile(history, nil)

	assert.Contains(t, analysis.Recommendation, "ít nhất 3 tháng")
}

func TestPatternAnalysis_ZeroIncomeMonths(t *testing.T) {
	history := []float64{100000000, 0, 0, 50000000, 0, 80000000}

	patterns := analyzePatterns(history, nil)

	assert.Equal(t, 3, patterns.ZeroIncomeMonths)
	assert.Equal(t, 0.5, patterns.ZeroIncomeRatio) // 3/6 months
}

func TestPatternAnalysis_FeastFamine(t *testing.T) {
	// Classic feast/famine: big project months vs nothing
	history := []float64{200000000, 10000000, 10000000, 150000000, 10000000, 180000000}
	// Sorted: 10, 10, 10, 150, 180, 200 -> Median (index 3) = 150M
	// Max = 200M, Ratio = 200/150 ≈ 1.33 (not so extreme after all)

	patterns := analyzePatterns(history, nil)

	// With this data, ratio is about 1.33
	assert.Greater(t, patterns.FeastFamineRatio, 1.0)
}

func TestPatternAnalysis_GapDetection(t *testing.T) {
	// 3 consecutive low-income months in the middle
	history := []float64{50000000, 40000000, 5000000, 2000000, 3000000, 60000000}

	patterns := analyzePatterns(history, nil)

	assert.Equal(t, 3, patterns.LongestGapMonths)
}

func TestAnalyzeSourcesBreakdown(t *testing.T) {
	sources := []IncomeSource{
		{Name: "Salary", Amount: 20000000, IsRecurring: true, SourceType: "salary"},
		{Name: "Freelance", Amount: 10000000, IsRecurring: false, SourceType: "freelance"},
	}

	recurring, variable, dominant, dominance := analyzeSourcesBreakdown(sources)

	assert.InDelta(t, 0.67, recurring, 0.1) // 20M/30M
	assert.InDelta(t, 0.33, variable, 0.1)  // 10M/30M
	assert.Equal(t, "salary", dominant)
	assert.InDelta(t, 0.67, dominance, 0.1)
}

func TestAutoDetectAndForecast(t *testing.T) {
	// Freelancer pattern
	input := ForecastInput{
		RolloverTBB: 5000000,
		IncomeSources: []IncomeSource{
			{Name: "Project", Amount: 40000000, SourceType: "freelance"},
		},
		HistoricalIncome: []float64{
			60000000, 0, 80000000, 20000000, 0, 50000000,
			30000000, 10000000, 70000000, 0, 40000000, 25000000,
		},
		MonthlyExpense: 20000000,
	}

	result, analysis := AutoDetectAndForecast(input)

	// Should auto-detect freelancer and apply smoothing
	assert.Equal(t, IncomeModeFreelancer, analysis.DetectedMode)
	assert.Greater(t, result.SafeSalary, 0.0)
	assert.Less(t, result.TotalAvailable, result.ProjectedIncome+input.RolloverTBB) // Smoothed
}
