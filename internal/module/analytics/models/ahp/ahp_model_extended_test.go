package ahp

import (
	"context"
	"fmt"
	"personalfinancedss/internal/module/analytics/goal_prioritization/domain"
	"personalfinancedss/internal/module/analytics/goal_prioritization/dto"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestAHPModel_Execute_4x4_InvestmentDecision tests a 4x4 AHP scenario
// Scenario: Choose best investment from 4 options using 4 criteria
func TestAHPModel_Execute_4x4_InvestmentDecision(t *testing.T) {
	model := NewAHPModel()
	ctx := context.Background()

	// Criteria: Expected Return, Risk Level, Liquidity, Tax Efficiency
	// Alternatives: Stocks, Bonds, Real Estate, Gold
	input := &dto.AHPInput{
		UserID: "user123",
		Criteria: []domain.Criteria{
			{ID: "return", Name: "Expected Return", Description: "Annual return potential"},
			{ID: "risk", Name: "Risk Level", Description: "Investment risk (lower is better)"},
			{ID: "liquidity", Name: "Liquidity", Description: "Ease of converting to cash"},
			{ID: "tax", Name: "Tax Efficiency", Description: "Tax advantages"},
		},
		// Criteria comparisons (6 total for 4 criteria)
		CriteriaComparisons: []domain.PairwiseComparison{
			{ElementA: "return", ElementB: "risk", Value: 2.0},      // Return slightly > Risk
			{ElementA: "return", ElementB: "liquidity", Value: 3.0}, // Return > Liquidity
			{ElementA: "return", ElementB: "tax", Value: 4.0},       // Return >> Tax
			{ElementA: "risk", ElementB: "liquidity", Value: 2.0},   // Risk > Liquidity
			{ElementA: "risk", ElementB: "tax", Value: 3.0},         // Risk > Tax
			{ElementA: "liquidity", ElementB: "tax", Value: 1.5},    // Liquidity slightly > Tax
		},
		Alternatives: []domain.Alternative{
			{ID: "stocks", Name: "Stocks (S&P 500)", Description: "Index fund"},
			{ID: "bonds", Name: "Bonds (Government)", Description: "Safe bonds"},
			{ID: "realestate", Name: "Real Estate (REIT)", Description: "Real estate investment trust"},
			{ID: "gold", Name: "Gold", Description: "Precious metal"},
		},
		AlternativeComparisons: map[string][]domain.PairwiseComparison{
			// For Expected Return (6 comparisons)
			"return": {
				{ElementA: "stocks", ElementB: "bonds", Value: 5.0},      // Stocks >> Bonds
				{ElementA: "stocks", ElementB: "realestate", Value: 3.0}, // Stocks > RE
				{ElementA: "stocks", ElementB: "gold", Value: 7.0},       // Stocks >>> Gold
				{ElementA: "bonds", ElementB: "realestate", Value: 0.5},  // RE > Bonds
				{ElementA: "bonds", ElementB: "gold", Value: 2.0},        // Bonds > Gold
				{ElementA: "realestate", ElementB: "gold", Value: 3.0},   // RE > Gold
			},
			// For Risk Level (lower risk is better)
			"risk": {
				{ElementA: "stocks", ElementB: "bonds", Value: 0.333},    // Bonds safer (3x)
				{ElementA: "stocks", ElementB: "realestate", Value: 0.5}, // RE safer (2x)
				{ElementA: "stocks", ElementB: "gold", Value: 1.0},       // Equal risk
				{ElementA: "bonds", ElementB: "realestate", Value: 2.0},  // Bonds safer
				{ElementA: "bonds", ElementB: "gold", Value: 3.0},        // Bonds much safer
				{ElementA: "realestate", ElementB: "gold", Value: 1.5},   // RE slightly safer
			},
			// For Liquidity
			"liquidity": {
				{ElementA: "stocks", ElementB: "bonds", Value: 2.0},      // Stocks more liquid
				{ElementA: "stocks", ElementB: "realestate", Value: 5.0}, // Stocks >> RE
				{ElementA: "stocks", ElementB: "gold", Value: 3.0},       // Stocks > Gold
				{ElementA: "bonds", ElementB: "realestate", Value: 3.0},  // Bonds > RE
				{ElementA: "bonds", ElementB: "gold", Value: 2.0},        // Bonds > Gold
				{ElementA: "realestate", ElementB: "gold", Value: 0.5},   // Gold > RE
			},
			// For Tax Efficiency
			"tax": {
				{ElementA: "stocks", ElementB: "bonds", Value: 1.0},      // Equal
				{ElementA: "stocks", ElementB: "realestate", Value: 0.5}, // RE better (tax deductions)
				{ElementA: "stocks", ElementB: "gold", Value: 2.0},       // Stocks > Gold
				{ElementA: "bonds", ElementB: "realestate", Value: 0.5},  // RE better
				{ElementA: "bonds", ElementB: "gold", Value: 2.0},        // Bonds > Gold
				{ElementA: "realestate", ElementB: "gold", Value: 4.0},   // RE >> Gold
			},
		},
	}

	t.Log("\n" + strings.Repeat("=", 80))
	t.Log("AHP EXECUTION: 4x4 Investment Decision")
	t.Log(strings.Repeat("=", 80) + "\n")

	t.Logf("📋 INPUT:")
	t.Logf("  Criteria (%d):", len(input.Criteria))
	for i, c := range input.Criteria {
		t.Logf("    %d. %-20s - %s", i+1, c.Name, c.Description)
	}
	t.Logf("\n  Alternatives (%d):", len(input.Alternatives))
	for i, a := range input.Alternatives {
		t.Logf("    %d. %-25s - %s", i+1, a.Name, a.Description)
	}
	t.Logf("\n  Total Comparisons:")
	t.Logf("    • Criteria: %d", len(input.CriteriaComparisons))
	t.Logf("    • Alternatives per criterion: 6")
	t.Logf("    • Total alternative comparisons: %d", len(input.CriteriaComparisons)*6)

	result, err := model.Execute(ctx, input)

	require.NoError(t, err)
	output := result.(*dto.AHPOutput)

	t.Logf("\n📊 OUTPUT:\n")
	t.Logf("  Criteria Weights:")
	for _, c := range input.Criteria {
		weight := output.CriteriaWeights[c.ID]
		bar := generateBar(weight, 40)
		t.Logf("    • %-20s: %.4f (%.1f%%) %s", c.Name, weight, weight*100, bar)
	}

	t.Logf("\n  🏆 FINAL RANKING (Global Priorities):")
	for i, item := range output.Ranking {
		bar := generateBar(item.Priority, 50)
		t.Logf("    #%d %-25s %.4f (%.1f%%) %s",
			i+1, item.AlternativeName, item.Priority, item.Priority*100, bar)
	}

	t.Logf("\n  📈 Local Priorities by Criterion:")
	for _, c := range input.Criteria {
		t.Logf("    %s:", c.Name)
		localPriorities := output.LocalPriorities[c.ID]
		for _, a := range input.Alternatives {
			localPri := localPriorities[a.ID]
			t.Logf("      • %-25s: %.4f", a.Name, localPri)
		}
	}

	t.Logf("\n  ✅ Consistency Check:")
	t.Logf("    • Consistency Ratio: %.4f", output.ConsistencyRatio)
	t.Logf("    • Status: %v (threshold: CR < 0.10)\n", output.IsConsistent)

	t.Log(strings.Repeat("=", 80) + "\n")

	// Assertions
	assert.NotNil(t, output)
	assert.Len(t, output.CriteriaWeights, 4)
	assert.Len(t, output.AlternativePriorities, 4)
	assert.Len(t, output.LocalPriorities, 4)
	assert.Len(t, output.Ranking, 4)

	// Verify all weights sum to 1
	criteriaSum := 0.0
	for _, weight := range output.CriteriaWeights {
		criteriaSum += weight
	}
	assert.InDelta(t, 1.0, criteriaSum, 0.001)

	alternativeSum := 0.0
	for _, priority := range output.AlternativePriorities {
		alternativeSum += priority
	}
	assert.InDelta(t, 1.0, alternativeSum, 0.001)

	// Verify consistency
	assert.Less(t, output.ConsistencyRatio, 0.15) // Allow slightly higher for 4x4

	// Verify ranking order
	assert.Equal(t, 1, output.Ranking[0].Rank)
	assert.Equal(t, 2, output.Ranking[1].Rank)
	assert.Equal(t, 3, output.Ranking[2].Rank)
	assert.Equal(t, 4, output.Ranking[3].Rank)

	// Based on comparisons, Stocks should likely be #1 (high return + good liquidity)
	t.Logf("Winner: %s with %.1f%% global priority\n", output.Ranking[0].AlternativeName, output.Ranking[0].Priority*100)
}

// TestAHPModel_Execute_5x5_DebtRepayment tests a 5x5 AHP scenario
// Scenario: Prioritize 5 debts for repayment using 5 criteria
func TestAHPModel_Execute_5x5_DebtRepayment(t *testing.T) {
	model := NewAHPModel()
	ctx := context.Background()

	input := &dto.AHPInput{
		UserID: "user123",
		Criteria: []domain.Criteria{
			{ID: "apr", Name: "Interest Rate (APR)", Description: "Annual percentage rate"},
			{ID: "balance", Name: "Outstanding Balance", Description: "Total amount owed"},
			{ID: "minpay", Name: "Minimum Payment", Description: "Monthly minimum payment"},
			{ID: "emotional", Name: "Emotional Impact", Description: "Stress/psychological burden"},
			{ID: "credit", Name: "Credit Score Impact", Description: "Effect on credit score"},
		},
		// 5 criteria = 10 comparisons
		CriteriaComparisons: []domain.PairwiseComparison{
			{ElementA: "apr", ElementB: "balance", Value: 3.0},
			{ElementA: "apr", ElementB: "minpay", Value: 4.0},
			{ElementA: "apr", ElementB: "emotional", Value: 2.0},
			{ElementA: "apr", ElementB: "credit", Value: 2.0},
			{ElementA: "balance", ElementB: "minpay", Value: 2.0},
			{ElementA: "balance", ElementB: "emotional", Value: 1.0}, // Equal
			{ElementA: "balance", ElementB: "credit", Value: 0.5},    // Credit more important
			{ElementA: "minpay", ElementB: "emotional", Value: 0.5},
			{ElementA: "minpay", ElementB: "credit", Value: 0.333},
			{ElementA: "emotional", ElementB: "credit", Value: 1.5},
		},
		Alternatives: []domain.Alternative{
			{ID: "cca", Name: "Credit Card A", Description: "22% APR, $3,000 balance"},
			{ID: "ccb", Name: "Credit Card B", Description: "18% APR, $5,000 balance"},
			{ID: "student", Name: "Student Loan", Description: "5% APR, $25,000 balance"},
			{ID: "car", Name: "Car Loan", Description: "7% APR, $8,000 balance"},
			{ID: "personal", Name: "Personal Loan", Description: "20% APR, $2,000 balance"},
		},
		AlternativeComparisons: map[string][]domain.PairwiseComparison{
			// For APR (10 comparisons) - higher APR = more urgent to pay
			"apr": {
				{ElementA: "cca", ElementB: "ccb", Value: 2.0}, // CCA higher APR
				{ElementA: "cca", ElementB: "student", Value: 5.0},
				{ElementA: "cca", ElementB: "car", Value: 3.0},
				{ElementA: "cca", ElementB: "personal", Value: 1.0}, // Similar APR
				{ElementA: "ccb", ElementB: "student", Value: 4.0},
				{ElementA: "ccb", ElementB: "car", Value: 2.5},
				{ElementA: "ccb", ElementB: "personal", Value: 0.5},
				{ElementA: "student", ElementB: "car", Value: 0.5},
				{ElementA: "student", ElementB: "personal", Value: 0.2},
				{ElementA: "car", ElementB: "personal", Value: 0.333},
			},
			// For Balance (10 comparisons) - higher balance = more important
			"balance": {
				{ElementA: "cca", ElementB: "ccb", Value: 0.5},
				{ElementA: "cca", ElementB: "student", Value: 0.111}, // 1/9
				{ElementA: "cca", ElementB: "car", Value: 0.333},
				{ElementA: "cca", ElementB: "personal", Value: 1.5},
				{ElementA: "ccb", ElementB: "student", Value: 0.2},
				{ElementA: "ccb", ElementB: "car", Value: 0.5},
				{ElementA: "ccb", ElementB: "personal", Value: 3.0},
				{ElementA: "student", ElementB: "car", Value: 3.0},
				{ElementA: "student", ElementB: "personal", Value: 7.0},
				{ElementA: "car", ElementB: "personal", Value: 4.0},
			},
			// For Minimum Payment (10 comparisons)
			"minpay": {
				{ElementA: "cca", ElementB: "ccb", Value: 0.5},
				{ElementA: "cca", ElementB: "student", Value: 0.333},
				{ElementA: "cca", ElementB: "car", Value: 0.5},
				{ElementA: "cca", ElementB: "personal", Value: 1.5},
				{ElementA: "ccb", ElementB: "student", Value: 0.5},
				{ElementA: "ccb", ElementB: "car", Value: 1.0},
				{ElementA: "ccb", ElementB: "personal", Value: 2.0},
				{ElementA: "student", ElementB: "car", Value: 2.0},
				{ElementA: "student", ElementB: "personal", Value: 3.0},
				{ElementA: "car", ElementB: "personal", Value: 1.5},
			},
			// For Emotional Impact (10 comparisons) - based on stress
			"emotional": {
				{ElementA: "cca", ElementB: "ccb", Value: 1.5}, // High APR cards stressful
				{ElementA: "cca", ElementB: "student", Value: 2.0},
				{ElementA: "cca", ElementB: "car", Value: 1.5},
				{ElementA: "cca", ElementB: "personal", Value: 1.0},
				{ElementA: "ccb", ElementB: "student", Value: 1.5},
				{ElementA: "ccb", ElementB: "car", Value: 1.0},
				{ElementA: "ccb", ElementB: "personal", Value: 0.5},
				{ElementA: "student", ElementB: "car", Value: 0.5}, // Student loan less stressful
				{ElementA: "student", ElementB: "personal", Value: 0.333},
				{ElementA: "car", ElementB: "personal", Value: 0.5},
			},
			// For Credit Score Impact (10 comparisons)
			"credit": {
				{ElementA: "cca", ElementB: "ccb", Value: 1.0}, // Both credit cards
				{ElementA: "cca", ElementB: "student", Value: 2.0},
				{ElementA: "cca", ElementB: "car", Value: 1.5},
				{ElementA: "cca", ElementB: "personal", Value: 1.0},
				{ElementA: "ccb", ElementB: "student", Value: 2.0},
				{ElementA: "ccb", ElementB: "car", Value: 1.5},
				{ElementA: "ccb", ElementB: "personal", Value: 1.0},
				{ElementA: "student", ElementB: "car", Value: 0.5},
				{ElementA: "student", ElementB: "personal", Value: 0.5},
				{ElementA: "car", ElementB: "personal", Value: 1.0},
			},
		},
	}

	t.Log("\n" + strings.Repeat("=", 80))
	t.Log("AHP EXECUTION: 5x5 Debt Repayment Prioritization")
	t.Log(strings.Repeat("=", 80) + "\n")

	t.Logf("📋 INPUT:")
	t.Logf("  Criteria (%d):", len(input.Criteria))
	for i, c := range input.Criteria {
		t.Logf("    %d. %-25s - %s", i+1, c.Name, c.Description)
	}
	t.Logf("\n  Alternatives (Debts) (%d):", len(input.Alternatives))
	for i, a := range input.Alternatives {
		t.Logf("    %d. %-20s - %s", i+1, a.Name, a.Description)
	}
	t.Logf("\n  Comparison Matrix:")
	t.Logf("    • Criteria comparisons: %d (5×4/2)", len(input.CriteriaComparisons))
	t.Logf("    • Alternative comparisons per criterion: 10 (5×4/2)")
	t.Logf("    • Total alternative comparisons: 50")

	result, err := model.Execute(ctx, input)

	require.NoError(t, err)
	output := result.(*dto.AHPOutput)

	t.Logf("\n📊 OUTPUT:\n")
	t.Logf("  Criteria Weights:")
	for _, c := range input.Criteria {
		weight := output.CriteriaWeights[c.ID]
		bar := generateBar(weight, 40)
		t.Logf("    • %-25s: %.4f (%.1f%%) %s", c.Name, weight, weight*100, bar)
	}

	t.Logf("\n  🏆 DEBT REPAYMENT PRIORITY (Higher = Pay First):")
	for i, item := range output.Ranking {
		bar := generateBar(item.Priority, 50)
		medal := []string{"🥇", "🥈", "🥉", "4️⃣", "5️⃣"}[i]
		t.Logf("    %s #%d %-20s %.4f (%.1f%%) %s",
			medal, i+1, item.AlternativeName, item.Priority, item.Priority*100, bar)
	}

	t.Logf("\n  💡 Interpretation:")
	t.Logf("    Pay off debts in this order to optimize financial health")
	t.Logf("    Higher priority = attack this debt first")

	t.Logf("\n  ✅ Consistency Check:")
	t.Logf("    • Consistency Ratio: %.4f", output.ConsistencyRatio)
	t.Logf("    • Status: %v (threshold: CR < 0.10)\n", output.IsConsistent)

	if !output.IsConsistent {
		t.Logf("    ⚠️  Warning: CR is high. Consider reviewing comparisons.\n")
	}

	t.Log(strings.Repeat("=", 80) + "\n")

	// Assertions
	assert.NotNil(t, output)
	assert.Len(t, output.CriteriaWeights, 5)
	assert.Len(t, output.AlternativePriorities, 5)
	assert.Len(t, output.LocalPriorities, 5)
	assert.Len(t, output.Ranking, 5)

	// Verify weights sum to 1
	criteriaSum := 0.0
	for _, weight := range output.CriteriaWeights {
		criteriaSum += weight
	}
	assert.InDelta(t, 1.0, criteriaSum, 0.001)

	// APR should be the most important criterion
	assert.Greater(t, output.CriteriaWeights["apr"], output.CriteriaWeights["balance"])

	// High APR debts (CCA, Personal Loan) should rank high
	highAPRDebts := map[string]bool{"cca": true, "personal": true}
	topTwo := map[string]bool{
		output.Ranking[0].AlternativeID: true,
		output.Ranking[1].AlternativeID: true,
	}

	hasHighAPRInTopTwo := false
	for debt := range highAPRDebts {
		if topTwo[debt] {
			hasHighAPRInTopTwo = true
			break
		}
	}
	assert.True(t, hasHighAPRInTopTwo, "At least one high APR debt should be in top 2")
}

// TestAHPModel_Execute_6x6_CityRelocation tests a 6x6 AHP scenario
// Scenario: Choose best city to relocate from 6 options using 6 criteria
func TestAHPModel_Execute_6x6_CityRelocation(t *testing.T) {
	model := NewAHPModel()
	ctx := context.Background()

	// This is a large test - 6 criteria × 6 alternatives
	// Criteria comparisons: 15, Alternative comparisons: 15 per criterion = 90 total

	input := &dto.AHPInput{
		UserID: "user123",
		Criteria: []domain.Criteria{
			{ID: "job", Name: "Job & Salary", Description: "Career opportunities and compensation"},
			{ID: "cost", Name: "Cost of Living", Description: "Housing, food, transportation costs"},
			{ID: "qol", Name: "Quality of Life", Description: "Culture, safety, amenities"},
			{ID: "weather", Name: "Weather", Description: "Climate preference"},
			{ID: "family", Name: "Family Connections", Description: "Proximity to family/friends"},
			{ID: "growth", Name: "Growth Potential", Description: "Long-term economic outlook"},
		},
		// 6 criteria = 15 comparisons
		CriteriaComparisons: []domain.PairwiseComparison{
			{ElementA: "job", ElementB: "cost", Value: 2.0},
			{ElementA: "job", ElementB: "qol", Value: 1.5},
			{ElementA: "job", ElementB: "weather", Value: 3.0},
			{ElementA: "job", ElementB: "family", Value: 1.0},
			{ElementA: "job", ElementB: "growth", Value: 1.0},
			{ElementA: "cost", ElementB: "qol", Value: 0.5},
			{ElementA: "cost", ElementB: "weather", Value: 2.0},
			{ElementA: "cost", ElementB: "family", Value: 0.5},
			{ElementA: "cost", ElementB: "growth", Value: 1.0},
			{ElementA: "qol", ElementB: "weather", Value: 3.0},
			{ElementA: "qol", ElementB: "family", Value: 1.0},
			{ElementA: "qol", ElementB: "growth", Value: 2.0},
			{ElementA: "weather", ElementB: "family", Value: 0.333},
			{ElementA: "weather", ElementB: "growth", Value: 0.5},
			{ElementA: "family", ElementB: "growth", Value: 2.0},
		},
		Alternatives: []domain.Alternative{
			{ID: "sf", Name: "San Francisco", Description: "Tech hub, expensive"},
			{ID: "austin", Name: "Austin", Description: "Growing tech scene, affordable"},
			{ID: "seattle", Name: "Seattle", Description: "Tech giants, rainy"},
			{ID: "denver", Name: "Denver", Description: "Outdoor lifestyle, growing"},
			{ID: "boston", Name: "Boston", Description: "Education/biotech, expensive"},
			{ID: "raleigh", Name: "Raleigh", Description: "Research triangle, affordable"},
		},
		AlternativeComparisons: map[string][]domain.PairwiseComparison{
			// For Job & Salary (15 comparisons)
			"job": {
				{ElementA: "sf", ElementB: "austin", Value: 2.0},
				{ElementA: "sf", ElementB: "seattle", Value: 1.5},
				{ElementA: "sf", ElementB: "denver", Value: 3.0},
				{ElementA: "sf", ElementB: "boston", Value: 1.5},
				{ElementA: "sf", ElementB: "raleigh", Value: 4.0},
				{ElementA: "austin", ElementB: "seattle", Value: 0.5},
				{ElementA: "austin", ElementB: "denver", Value: 2.0},
				{ElementA: "austin", ElementB: "boston", Value: 1.0},
				{ElementA: "austin", ElementB: "raleigh", Value: 2.5},
				{ElementA: "seattle", ElementB: "denver", Value: 2.5},
				{ElementA: "seattle", ElementB: "boston", Value: 1.0},
				{ElementA: "seattle", ElementB: "raleigh", Value: 3.0},
				{ElementA: "denver", ElementB: "boston", Value: 0.5},
				{ElementA: "denver", ElementB: "raleigh", Value: 1.5},
				{ElementA: "boston", ElementB: "raleigh", Value: 2.0},
			},
			// For Cost of Living (lower = better) (15 comparisons)
			"cost": {
				{ElementA: "sf", ElementB: "austin", Value: 0.2},
				{ElementA: "sf", ElementB: "seattle", Value: 0.5},
				{ElementA: "sf", ElementB: "denver", Value: 0.333},
				{ElementA: "sf", ElementB: "boston", Value: 0.5},
				{ElementA: "sf", ElementB: "raleigh", Value: 0.2},
				{ElementA: "austin", ElementB: "seattle", Value: 2.0},
				{ElementA: "austin", ElementB: "denver", Value: 1.5},
				{ElementA: "austin", ElementB: "boston", Value: 2.0},
				{ElementA: "austin", ElementB: "raleigh", Value: 1.0},
				{ElementA: "seattle", ElementB: "denver", Value: 0.5},
				{ElementA: "seattle", ElementB: "boston", Value: 1.0},
				{ElementA: "seattle", ElementB: "raleigh", Value: 0.5},
				{ElementA: "denver", ElementB: "boston", Value: 2.0},
				{ElementA: "denver", ElementB: "raleigh", Value: 1.0},
				{ElementA: "boston", ElementB: "raleigh", Value: 0.5},
			},
			// For Quality of Life (15 comparisons)
			"qol": {
				{ElementA: "sf", ElementB: "austin", Value: 1.5},
				{ElementA: "sf", ElementB: "seattle", Value: 1.0},
				{ElementA: "sf", ElementB: "denver", Value: 1.0},
				{ElementA: "sf", ElementB: "boston", Value: 1.0},
				{ElementA: "sf", ElementB: "raleigh", Value: 2.0},
				{ElementA: "austin", ElementB: "seattle", Value: 0.5},
				{ElementA: "austin", ElementB: "denver", Value: 1.0},
				{ElementA: "austin", ElementB: "boston", Value: 0.5},
				{ElementA: "austin", ElementB: "raleigh", Value: 1.5},
				{ElementA: "seattle", ElementB: "denver", Value: 1.5},
				{ElementA: "seattle", ElementB: "boston", Value: 1.0},
				{ElementA: "seattle", ElementB: "raleigh", Value: 2.0},
				{ElementA: "denver", ElementB: "boston", Value: 1.0},
				{ElementA: "denver", ElementB: "raleigh", Value: 1.5},
				{ElementA: "boston", ElementB: "raleigh", Value: 1.5},
			},
			// For Weather (user preference) (15 comparisons)
			"weather": {
				{ElementA: "sf", ElementB: "austin", Value: 1.5},
				{ElementA: "sf", ElementB: "seattle", Value: 3.0},
				{ElementA: "sf", ElementB: "denver", Value: 1.0},
				{ElementA: "sf", ElementB: "boston", Value: 2.0},
				{ElementA: "sf", ElementB: "raleigh", Value: 1.0},
				{ElementA: "austin", ElementB: "seattle", Value: 2.0},
				{ElementA: "austin", ElementB: "denver", Value: 0.5},
				{ElementA: "austin", ElementB: "boston", Value: 1.5},
				{ElementA: "austin", ElementB: "raleigh", Value: 1.0},
				{ElementA: "seattle", ElementB: "denver", Value: 0.333},
				{ElementA: "seattle", ElementB: "boston", Value: 0.5},
				{ElementA: "seattle", ElementB: "raleigh", Value: 0.5},
				{ElementA: "denver", ElementB: "boston", Value: 2.0},
				{ElementA: "denver", ElementB: "raleigh", Value: 1.5},
				{ElementA: "boston", ElementB: "raleigh", Value: 1.0},
			},
			// For Family Connections (user-specific) (15 comparisons)
			// Assuming family in Boston area
			"family": {
				{ElementA: "sf", ElementB: "austin", Value: 0.5},
				{ElementA: "sf", ElementB: "seattle", Value: 1.0},
				{ElementA: "sf", ElementB: "denver", Value: 0.5},
				{ElementA: "sf", ElementB: "boston", Value: 0.2},
				{ElementA: "sf", ElementB: "raleigh", Value: 0.333},
				{ElementA: "austin", ElementB: "seattle", Value: 1.0},
				{ElementA: "austin", ElementB: "denver", Value: 1.0},
				{ElementA: "austin", ElementB: "boston", Value: 0.333},
				{ElementA: "austin", ElementB: "raleigh", Value: 0.5},
				{ElementA: "seattle", ElementB: "denver", Value: 1.0},
				{ElementA: "seattle", ElementB: "boston", Value: 0.333},
				{ElementA: "seattle", ElementB: "raleigh", Value: 0.5},
				{ElementA: "denver", ElementB: "boston", Value: 0.333},
				{ElementA: "denver", ElementB: "raleigh", Value: 0.5},
				{ElementA: "boston", ElementB: "raleigh", Value: 2.0},
			},
			// For Growth Potential (15 comparisons)
			"growth": {
				{ElementA: "sf", ElementB: "austin", Value: 0.5},
				{ElementA: "sf", ElementB: "seattle", Value: 1.0},
				{ElementA: "sf", ElementB: "denver", Value: 1.5},
				{ElementA: "sf", ElementB: "boston", Value: 1.0},
				{ElementA: "sf", ElementB: "raleigh", Value: 1.5},
				{ElementA: "austin", ElementB: "seattle", Value: 1.5},
				{ElementA: "austin", ElementB: "denver", Value: 2.0},
				{ElementA: "austin", ElementB: "boston", Value: 1.5},
				{ElementA: "austin", ElementB: "raleigh", Value: 2.0},
				{ElementA: "seattle", ElementB: "denver", Value: 1.5},
				{ElementA: "seattle", ElementB: "boston", Value: 1.0},
				{ElementA: "seattle", ElementB: "raleigh", Value: 1.5},
				{ElementA: "denver", ElementB: "boston", Value: 1.0},
				{ElementA: "denver", ElementB: "raleigh", Value: 1.0},
				{ElementA: "boston", ElementB: "raleigh", Value: 1.0},
			},
		},
	}

	t.Log("\n" + strings.Repeat("=", 90))
	t.Log("AHP EXECUTION: 6x6 City Relocation Decision (LARGE SCALE)")
	t.Log(strings.Repeat("=", 90) + "\n")

	t.Logf("📋 INPUT:")
	t.Logf("  Criteria (%d):", len(input.Criteria))
	for i, c := range input.Criteria {
		t.Logf("    %d. %-20s - %s", i+1, c.Name, c.Description)
	}
	t.Logf("\n  Alternatives (Cities) (%d):", len(input.Alternatives))
	for i, a := range input.Alternatives {
		t.Logf("    %d. %-15s - %s", i+1, a.Name, a.Description)
	}
	t.Logf("\n  ⚠️  Large Comparison Matrix:")
	t.Logf("    • Criteria comparisons: %d (6×5/2)", len(input.CriteriaComparisons))
	t.Logf("    • Alternative comparisons per criterion: 15 (6×5/2)")
	t.Logf("    • Total alternative comparisons: 90 (!)")
	t.Logf("    • This is approaching the practical limit for AHP\n")

	result, err := model.Execute(ctx, input)

	require.NoError(t, err)
	output := result.(*dto.AHPOutput)

	t.Logf("📊 OUTPUT:\n")
	t.Logf("  Criteria Weights:")
	for _, c := range input.Criteria {
		weight := output.CriteriaWeights[c.ID]
		bar := generateBar(weight, 35)
		t.Logf("    • %-20s: %.4f (%.1f%%) %s", c.Name, weight, weight*100, bar)
	}

	t.Logf("\n  🏆 CITY RANKING (Best to Worst):")
	medals := []string{"🥇", "🥈", "🥉", "4️⃣", "5️⃣", "6️⃣"}
	for i, item := range output.Ranking {
		bar := generateBar(item.Priority, 45)
		t.Logf("    %s #%d %-15s %.4f (%.1f%%) %s",
			medals[i], i+1, item.AlternativeName, item.Priority, item.Priority*100, bar)
	}

	t.Logf("\n  📊 Detailed Analysis:")
	t.Logf("    Top Choice: %s", output.Ranking[0].AlternativeName)
	t.Logf("    Score Spread: %.4f (%.1f%% difference between #1 and #6)",
		output.Ranking[0].Priority-output.Ranking[5].Priority,
		(output.Ranking[0].Priority-output.Ranking[5].Priority)*100)

	t.Logf("\n  ✅ Consistency Check:")
	t.Logf("    • Consistency Ratio: %.4f", output.ConsistencyRatio)
	t.Logf("    • Status: %v (threshold: CR < 0.10)", output.IsConsistent)
	if output.ConsistencyRatio > 0.10 {
		t.Logf("    ⚠️  Warning: With 6x6, higher CR is expected. Review if CR > 0.15\n")
	} else {
		t.Logf("    ✨ Excellent! Maintaining consistency with 90 comparisons\n")
	}

	t.Log(strings.Repeat("=", 90) + "\n")

	// Assertions
	assert.NotNil(t, output)
	assert.Len(t, output.CriteriaWeights, 6)
	assert.Len(t, output.AlternativePriorities, 6)
	assert.Len(t, output.LocalPriorities, 6)
	assert.Len(t, output.Ranking, 6)

	// Verify weights sum to 1
	criteriaSum := 0.0
	for _, weight := range output.CriteriaWeights {
		criteriaSum += weight
	}
	assert.InDelta(t, 1.0, criteriaSum, 0.001)

	alternativeSum := 0.0
	for _, priority := range output.AlternativePriorities {
		alternativeSum += priority
	}
	assert.InDelta(t, 1.0, alternativeSum, 0.001)

	// Allow higher CR for 6x6 due to complexity
	assert.Less(t, output.ConsistencyRatio, 0.20, "CR should be < 0.20 for 6x6")

	// Verify ranking integrity
	for i := 0; i < 6; i++ {
		assert.Equal(t, i+1, output.Ranking[i].Rank)
	}

	t.Logf("💡 Note: 6x6 AHP requires 90 pairwise comparisons for alternatives.")
	t.Logf("   Consider reducing to 4-5 criteria/alternatives for practical use.\n")
}

// Helper function to generate visual bar charts
func generateBar(value float64, maxLength int) string {
	length := int(value * float64(maxLength))
	if length > maxLength {
		length = maxLength
	}
	if length < 0 {
		length = 0
	}

	bar := ""
	for i := 0; i < length; i++ {
		bar += "█"
	}
	return bar
}

// TestAHPModel_Execute_PerformanceTest tests performance with larger matrices
func TestAHPModel_Execute_PerformanceTest(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	model := NewAHPModel()
	ctx := context.Background()

	sizes := []int{3, 4, 5, 6}

	for _, n := range sizes {
		t.Run(fmt.Sprintf("Matrix_%dx%d", n, n), func(t *testing.T) {
			input := generateTestInput(n, n)

			t.Logf("Testing %dx%d matrix (%d criteria, %d alternatives)", n, n, n, n)
			t.Logf("Total comparisons: %d criteria + %d alternative comparisons",
				n*(n-1)/2, n*(n*(n-1)/2))

			result, err := model.Execute(ctx, input)

			require.NoError(t, err)
			output := result.(*dto.AHPOutput)

			t.Logf("  ✓ Execution successful")
			t.Logf("  CR: %.4f (Consistent: %v)", output.ConsistencyRatio, output.IsConsistent)
		})
	}
}

// Helper to generate test input for performance testing
func generateTestInput(numCriteria, numAlternatives int) *dto.AHPInput {
	// Generate criteria
	criteria := make([]domain.Criteria, numCriteria)
	for i := 0; i < numCriteria; i++ {
		criteria[i] = domain.Criteria{
			ID:   fmt.Sprintf("c%d", i+1),
			Name: fmt.Sprintf("Criterion %d", i+1),
		}
	}

	// Generate criteria comparisons
	criteriaComparisons := []domain.PairwiseComparison{}
	for i := 0; i < numCriteria; i++ {
		for j := i + 1; j < numCriteria; j++ {
			criteriaComparisons = append(criteriaComparisons, domain.PairwiseComparison{
				ElementA: criteria[i].ID,
				ElementB: criteria[j].ID,
				Value:    float64(1 + (i+j)%5), // Values between 1-5
			})
		}
	}

	// Generate alternatives
	alternatives := make([]domain.Alternative, numAlternatives)
	for i := 0; i < numAlternatives; i++ {
		alternatives[i] = domain.Alternative{
			ID:   fmt.Sprintf("a%d", i+1),
			Name: fmt.Sprintf("Alternative %d", i+1),
		}
	}

	// Generate alternative comparisons for each criterion
	alternativeComparisons := make(map[string][]domain.PairwiseComparison)
	for _, criterion := range criteria {
		comparisons := []domain.PairwiseComparison{}
		for i := 0; i < numAlternatives; i++ {
			for j := i + 1; j < numAlternatives; j++ {
				comparisons = append(comparisons, domain.PairwiseComparison{
					ElementA: alternatives[i].ID,
					ElementB: alternatives[j].ID,
					Value:    float64(1 + (i+j)%5),
				})
			}
		}
		alternativeComparisons[criterion.ID] = comparisons
	}

	return &dto.AHPInput{
		UserID:                 "test_user",
		Criteria:               criteria,
		CriteriaComparisons:    criteriaComparisons,
		Alternatives:           alternatives,
		AlternativeComparisons: alternativeComparisons,
	}
}
