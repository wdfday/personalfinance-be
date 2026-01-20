package ahp

import (
	"context"
	"personalfinancedss/internal/module/analytics/goal_prioritization/domain"
	"personalfinancedss/internal/module/analytics/goal_prioritization/dto"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createTestAHPInput(userID string, criteriaValue, altValue float64) *dto.AHPInput {
	return &dto.AHPInput{
		UserID: userID,
		Criteria: []domain.Criteria{
			{ID: "c1", Name: "Importance"},
			{ID: "c2", Name: "Urgency"},
		},
		CriteriaComparisons: []domain.PairwiseComparison{
			{ElementA: "c1", ElementB: "c2", Value: criteriaValue},
		},
		Alternatives: []domain.Alternative{
			{ID: "a1", Name: "Emergency Fund"},
			{ID: "a2", Name: "Vacation"},
		},
		AlternativeComparisons: map[string][]domain.PairwiseComparison{
			"c1": {{ElementA: "a1", ElementB: "a2", Value: altValue}},
			"c2": {{ElementA: "a1", ElementB: "a2", Value: altValue}},
		},
	}
}

func TestAHPModel_ExecuteGroupDecision_SingleDecisionMaker(t *testing.T) {
	model := NewAHPModel()
	ctx := context.Background()

	groupInput := &GroupDecisionInput{
		DecisionMakers: []DecisionMakerInput{
			{
				DecisionMakerID:   "dm1",
				DecisionMakerName: "Alice",
				Input:             createTestAHPInput("dm1", 3.0, 5.0),
			},
		},
	}

	result, err := model.ExecuteGroupDecision(ctx, groupInput)
	require.NoError(t, err)

	assert.NotNil(t, result.AggregatedResult)
	assert.Len(t, result.IndividualResults, 1)
	assert.Equal(t, 1.0, result.ConsensusMetrics.ConsensusIndex)
	assert.Equal(t, "high", result.ConsensusMetrics.ConsensusLevel)
}

func TestAHPModel_ExecuteGroupDecision_TwoDecisionMakers_Agreement(t *testing.T) {
	model := NewAHPModel()
	ctx := context.Background()

	// Two decision makers with similar judgments
	groupInput := &GroupDecisionInput{
		DecisionMakers: []DecisionMakerInput{
			{
				DecisionMakerID:   "dm1",
				DecisionMakerName: "Alice",
				Input:             createTestAHPInput("dm1", 3.0, 5.0),
			},
			{
				DecisionMakerID:   "dm2",
				DecisionMakerName: "Bob",
				Input:             createTestAHPInput("dm2", 3.0, 5.0), // Same judgments
			},
		},
		AggregationMethod: "geometric_mean",
	}

	result, err := model.ExecuteGroupDecision(ctx, groupInput)
	require.NoError(t, err)

	assert.NotNil(t, result.AggregatedResult)
	assert.Len(t, result.IndividualResults, 2)

	// High consensus expected
	assert.Equal(t, "high", result.ConsensusMetrics.ConsensusLevel)
	assert.GreaterOrEqual(t, result.ConsensusMetrics.ConsensusIndex, 0.8)

	t.Logf("Consensus Index: %.2f", result.ConsensusMetrics.ConsensusIndex)
	t.Logf("Criteria Consensus: %.2f", result.ConsensusMetrics.CriteriaConsensus)
	t.Logf("Ranking Consensus: %.2f", result.ConsensusMetrics.RankingConsensus)
}

func TestAHPModel_ExecuteGroupDecision_TwoDecisionMakers_Disagreement(t *testing.T) {
	model := NewAHPModel()
	ctx := context.Background()

	// Two decision makers with different judgments
	groupInput := &GroupDecisionInput{
		DecisionMakers: []DecisionMakerInput{
			{
				DecisionMakerID:   "dm1",
				DecisionMakerName: "Alice",
				Input:             createTestAHPInput("dm1", 7.0, 9.0), // Strong preference for c1 and a1
			},
			{
				DecisionMakerID:   "dm2",
				DecisionMakerName: "Bob",
				Input:             createTestAHPInput("dm2", 0.2, 0.2), // Opposite preference
			},
		},
		AggregationMethod: "geometric_mean",
	}

	result, err := model.ExecuteGroupDecision(ctx, groupInput)
	require.NoError(t, err)

	assert.NotNil(t, result.AggregatedResult)
	assert.Len(t, result.IndividualResults, 2)

	// Lower consensus expected due to disagreement
	t.Logf("Consensus Index: %.2f", result.ConsensusMetrics.ConsensusIndex)
	t.Logf("Consensus Level: %s", result.ConsensusMetrics.ConsensusLevel)
	t.Logf("Recommendation: %s", result.ConsensusMetrics.Recommendation)

	// Check disagreement analysis
	if len(result.DisagreementAnalysis) > 0 {
		t.Log("\nDisagreement Analysis:")
		for _, item := range result.DisagreementAnalysis {
			t.Logf("  %s (%s): variance=%.4f, range=[%.2f, %.2f]",
				item.ElementID, item.Type, item.Variance, item.MinValue, item.MaxValue)
		}
	}
}

func TestAHPModel_ExecuteGroupDecision_ThreeDecisionMakers(t *testing.T) {
	model := NewAHPModel()
	ctx := context.Background()

	groupInput := &GroupDecisionInput{
		DecisionMakers: []DecisionMakerInput{
			{
				DecisionMakerID:   "dm1",
				DecisionMakerName: "Alice",
				Input:             createTestAHPInput("dm1", 3.0, 5.0),
				Weight:            1.0,
			},
			{
				DecisionMakerID:   "dm2",
				DecisionMakerName: "Bob",
				Input:             createTestAHPInput("dm2", 2.0, 4.0),
				Weight:            1.0,
			},
			{
				DecisionMakerID:   "dm3",
				DecisionMakerName: "Charlie",
				Input:             createTestAHPInput("dm3", 4.0, 6.0),
				Weight:            1.0,
			},
		},
		AggregationMethod: "geometric_mean",
	}

	result, err := model.ExecuteGroupDecision(ctx, groupInput)
	require.NoError(t, err)

	assert.NotNil(t, result.AggregatedResult)
	assert.Len(t, result.IndividualResults, 3)

	// Log aggregated results
	t.Log("\n=== Group Decision Results ===")
	t.Logf("Aggregation Method: geometric_mean")
	t.Logf("\nAggregated Criteria Weights:")
	for id, weight := range result.AggregatedResult.CriteriaWeights {
		t.Logf("  %s: %.4f", id, weight)
	}

	t.Logf("\nAggregated Ranking:")
	for _, item := range result.AggregatedResult.Ranking {
		t.Logf("  #%d %s: %.4f", item.Rank, item.AlternativeName, item.Priority)
	}

	t.Logf("\nConsensus Metrics:")
	t.Logf("  Index: %.2f", result.ConsensusMetrics.ConsensusIndex)
	t.Logf("  Level: %s", result.ConsensusMetrics.ConsensusLevel)
}

func TestAHPModel_ExecuteGroupDecision_WeightedAggregation(t *testing.T) {
	model := NewAHPModel()
	ctx := context.Background()

	// Expert has higher weight
	groupInput := &GroupDecisionInput{
		DecisionMakers: []DecisionMakerInput{
			{
				DecisionMakerID:   "expert",
				DecisionMakerName: "Expert",
				Input:             createTestAHPInput("expert", 5.0, 7.0),
				Weight:            3.0, // Expert weight = 3
			},
			{
				DecisionMakerID:   "novice",
				DecisionMakerName: "Novice",
				Input:             createTestAHPInput("novice", 2.0, 3.0),
				Weight:            1.0, // Novice weight = 1
			},
		},
		AggregationMethod: "weighted",
	}

	result, err := model.ExecuteGroupDecision(ctx, groupInput)
	require.NoError(t, err)

	// Result should be closer to expert's judgment
	t.Logf("Expert weight: 3.0, Novice weight: 1.0")
	t.Logf("Aggregated criteria weights: %v", result.AggregatedResult.CriteriaWeights)

	// Expert preferred c1 more strongly, so aggregated c1 weight should be higher
	assert.Greater(t, result.AggregatedResult.CriteriaWeights["c1"],
		result.AggregatedResult.CriteriaWeights["c2"])
}

func TestAHPModel_ExecuteGroupDecision_EmptyInput(t *testing.T) {
	model := NewAHPModel()
	ctx := context.Background()

	groupInput := &GroupDecisionInput{
		DecisionMakers: []DecisionMakerInput{},
	}

	_, err := model.ExecuteGroupDecision(ctx, groupInput)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "at least one decision maker")
}

func TestHelperFunctions(t *testing.T) {
	t.Run("geometricMean", func(t *testing.T) {
		result := geometricMean([]float64{2, 8})
		assert.InDelta(t, 4.0, result, 0.01) // sqrt(2*8) = 4

		result = geometricMean([]float64{1, 1, 1})
		assert.InDelta(t, 1.0, result, 0.01)
	})

	t.Run("arithmeticMean", func(t *testing.T) {
		result := arithmeticMean([]float64{2, 4, 6})
		assert.InDelta(t, 4.0, result, 0.01)
	})

	t.Run("weightedMean", func(t *testing.T) {
		result := weightedMean([]float64{10, 20}, []float64{1, 3})
		// (10*1 + 20*3) / (1+3) = 70/4 = 17.5
		assert.InDelta(t, 17.5, result, 0.01)
	})

	t.Run("calculateVariance", func(t *testing.T) {
		result := calculateVariance([]float64{2, 4, 6})
		// mean = 4, variance = ((2-4)^2 + (4-4)^2 + (6-4)^2) / 3 = 8/3 ≈ 2.67
		assert.InDelta(t, 2.67, result, 0.1)
	})

	t.Run("minMax", func(t *testing.T) {
		minVal, maxVal := minMax([]float64{5, 2, 8, 1, 9})
		assert.Equal(t, 1.0, minVal)
		assert.Equal(t, 9.0, maxVal)
	})
}
