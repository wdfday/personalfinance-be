package ahp

import (
	"context"
	"personalfinancedss/internal/module/analytics/goal_prioritization/domain"
	"personalfinancedss/internal/module/analytics/goal_prioritization/dto"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAHPModel_AnalyzeSensitivity(t *testing.T) {
	model := NewAHPModel()
	ctx := context.Background()

	// Create test input
	input := &dto.AHPInput{
		UserID: "test-user",
		Criteria: []domain.Criteria{
			{ID: "c1", Name: "Importance"},
			{ID: "c2", Name: "Urgency"},
		},
		CriteriaComparisons: []domain.PairwiseComparison{
			{ElementA: "c1", ElementB: "c2", Value: 3.0},
		},
		Alternatives: []domain.Alternative{
			{ID: "a1", Name: "Emergency Fund"},
			{ID: "a2", Name: "Vacation"},
		},
		AlternativeComparisons: map[string][]domain.PairwiseComparison{
			"c1": {{ElementA: "a1", ElementB: "a2", Value: 5.0}},
			"c2": {{ElementA: "a1", ElementB: "a2", Value: 3.0}},
		},
	}

	// Execute AHP first
	result, err := model.Execute(ctx, input)
	require.NoError(t, err)
	output := result.(*dto.AHPOutput)

	// Run sensitivity analysis
	sensitivity := model.AnalyzeSensitivity(input, output)

	// Verify criteria sensitivity
	assert.NotEmpty(t, sensitivity.CriteriaSensitivity)
	assert.Len(t, sensitivity.CriteriaSensitivity, 2)

	for _, item := range sensitivity.CriteriaSensitivity {
		assert.NotEmpty(t, item.CriterionID)
		assert.NotEmpty(t, item.SensitivityLevel)
		t.Logf("Criterion %s: sensitivity=%s, score=%.4f",
			item.CriterionName, item.SensitivityLevel, item.SensitivityScore)
	}

	// Verify ranking stability
	assert.NotEmpty(t, sensitivity.RankingStability.Recommendation)
	t.Logf("Ranking stability: score=%.2f, stable=%v",
		sensitivity.RankingStability.StabilityScore,
		sensitivity.RankingStability.IsStable)
}

func TestAHPModel_AnalyzeSensitivity_3x3(t *testing.T) {
	model := NewAHPModel()
	ctx := context.Background()

	input := &dto.AHPInput{
		UserID: "test-user",
		Criteria: []domain.Criteria{
			{ID: "impact", Name: "Financial Impact"},
			{ID: "urgency", Name: "Urgency"},
			{ID: "cost", Name: "Cost"},
		},
		CriteriaComparisons: []domain.PairwiseComparison{
			{ElementA: "impact", ElementB: "urgency", Value: 3.0},
			{ElementA: "impact", ElementB: "cost", Value: 5.0},
			{ElementA: "urgency", ElementB: "cost", Value: 2.0},
		},
		Alternatives: []domain.Alternative{
			{ID: "emergency", Name: "Emergency Fund"},
			{ID: "house", Name: "House Down Payment"},
			{ID: "retirement", Name: "Retirement"},
		},
		AlternativeComparisons: map[string][]domain.PairwiseComparison{
			"impact": {
				{ElementA: "emergency", ElementB: "house", Value: 2.0},
				{ElementA: "emergency", ElementB: "retirement", Value: 0.5},
				{ElementA: "house", ElementB: "retirement", Value: 0.333},
			},
			"urgency": {
				{ElementA: "emergency", ElementB: "house", Value: 5.0},
				{ElementA: "emergency", ElementB: "retirement", Value: 7.0},
				{ElementA: "house", ElementB: "retirement", Value: 2.0},
			},
			"cost": {
				{ElementA: "emergency", ElementB: "house", Value: 5.0},
				{ElementA: "emergency", ElementB: "retirement", Value: 7.0},
				{ElementA: "house", ElementB: "retirement", Value: 2.0},
			},
		},
	}

	result, err := model.Execute(ctx, input)
	require.NoError(t, err)
	output := result.(*dto.AHPOutput)

	sensitivity := model.AnalyzeSensitivity(input, output)

	// Should have 3 criteria sensitivity items
	assert.Len(t, sensitivity.CriteriaSensitivity, 3)

	// Log results
	t.Log("\n=== Sensitivity Analysis Results ===")
	t.Log("\nCriteria Sensitivity (sorted by impact):")
	for _, item := range sensitivity.CriteriaSensitivity {
		t.Logf("  %s: level=%s, score=%.4f, +10%%=%.4f, -10%%=%.4f",
			item.CriterionName, item.SensitivityLevel, item.SensitivityScore,
			item.ImpactIfIncrease, item.ImpactIfDecrease)
	}

	t.Logf("\nRanking Stability:")
	t.Logf("  Score: %.2f", sensitivity.RankingStability.StabilityScore)
	t.Logf("  Top Two Gap: %.4f", sensitivity.RankingStability.TopTwoGap)
	t.Logf("  Is Stable: %v", sensitivity.RankingStability.IsStable)
	t.Logf("  Recommendation: %s", sensitivity.RankingStability.Recommendation)

	if len(sensitivity.CriticalThresholds) > 0 {
		t.Log("\nCritical Thresholds:")
		for _, th := range sensitivity.CriticalThresholds {
			t.Logf("  %s: current=%.2f, threshold=%.2f (%s) -> %s",
				th.CriterionID, th.CurrentWeight, th.ThresholdWeight,
				th.ChangeDirection, th.AffectedRanking)
		}
	}
}

func TestRankingStability_HighGap(t *testing.T) {
	model := NewAHPModel()

	// Create output with large gap between top two
	output := &dto.AHPOutput{
		Ranking: []domain.RankItem{
			{AlternativeID: "a1", AlternativeName: "Top Choice", Priority: 0.70, Rank: 1},
			{AlternativeID: "a2", AlternativeName: "Second", Priority: 0.30, Rank: 2},
		},
	}

	stability := model.calculateRankingStability(output)

	assert.True(t, stability.IsStable)
	assert.Greater(t, stability.StabilityScore, 50.0)
	assert.Contains(t, stability.Recommendation, "stable")
}

func TestRankingStability_LowGap(t *testing.T) {
	model := NewAHPModel()

	// Create output with small gap between top two
	output := &dto.AHPOutput{
		Ranking: []domain.RankItem{
			{AlternativeID: "a1", AlternativeName: "Top Choice", Priority: 0.35, Rank: 1},
			{AlternativeID: "a2", AlternativeName: "Second", Priority: 0.33, Rank: 2},
		},
	}

	stability := model.calculateRankingStability(output)

	assert.False(t, stability.IsStable)
	assert.Contains(t, stability.Recommendation, "sensitive")
}
