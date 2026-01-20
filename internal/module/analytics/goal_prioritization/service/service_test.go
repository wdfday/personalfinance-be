package service

import (
	"context"
	"testing"

	"personalfinancedss/internal/module/analytics/goal_prioritization/domain"
	"personalfinancedss/internal/module/analytics/goal_prioritization/dto"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestNewService(t *testing.T) {
	logger := zap.NewNop()
	svc := NewService(logger)

	assert.NotNil(t, svc)
}

func TestService_ExecuteAHP_Success(t *testing.T) {
	logger := zap.NewNop()
	svc := NewService(logger)
	ctx := context.Background()

	input := &dto.AHPInput{
		UserID: "user123",
		Criteria: []domain.Criteria{
			{ID: "importance", Name: "Importance"},
			{ID: "urgency", Name: "Urgency"},
		},
		CriteriaComparisons: []domain.PairwiseComparison{
			{ElementA: "importance", ElementB: "urgency", Value: 3.0},
		},
		Alternatives: []domain.Alternative{
			{ID: "goal1", Name: "Emergency Fund"},
			{ID: "goal2", Name: "Vacation"},
		},
		AlternativeComparisons: map[string][]domain.PairwiseComparison{
			"importance": {
				{ElementA: "goal1", ElementB: "goal2", Value: 5.0},
			},
			"urgency": {
				{ElementA: "goal1", ElementB: "goal2", Value: 3.0},
			},
		},
	}

	t.Logf("\n=== INPUT ===")
	t.Logf("User ID: %s", input.UserID)
	t.Logf("Criteria: %d items", len(input.Criteria))
	for _, c := range input.Criteria {
		t.Logf("  - %s (%s)", c.Name, c.ID)
	}
	t.Logf("Alternatives: %d items", len(input.Alternatives))
	for _, a := range input.Alternatives {
		t.Logf("  - %s (%s)", a.Name, a.ID)
	}
	t.Logf("Criteria Comparisons:")
	for _, comp := range input.CriteriaComparisons {
		t.Logf("  - %s vs %s = %.1f", comp.ElementA, comp.ElementB, comp.Value)
	}

	output, err := svc.ExecuteAHP(ctx, input)

	require.NoError(t, err)
	assert.NotNil(t, output)

	t.Logf("\n=== OUTPUT ===")
	t.Logf("Consistency Ratio: %.4f (Consistent: %v)", output.ConsistencyRatio, output.IsConsistent)
	t.Logf("Criteria Weights:")
	for id, weight := range output.CriteriaWeights {
		t.Logf("  - %s: %.4f", id, weight)
	}
	t.Logf("Alternative Priorities:")
	for id, priority := range output.AlternativePriorities {
		t.Logf("  - %s: %.4f", id, priority)
	}
	t.Logf("Ranking:")
	for _, item := range output.Ranking {
		t.Logf("  #%d - %s (%.4f)", item.Rank, item.AlternativeName, item.Priority)
	}

	// Verify output structure
	assert.NotNil(t, output.AlternativePriorities)
	assert.NotNil(t, output.CriteriaWeights)
	assert.NotNil(t, output.LocalPriorities)
	assert.NotNil(t, output.Ranking)

	// Verify consistency
	assert.True(t, output.IsConsistent)
	assert.Less(t, output.ConsistencyRatio, 0.1)

	// Verify criteria weights sum to 1
	sum := 0.0
	for _, weight := range output.CriteriaWeights {
		sum += weight
	}
	assert.InDelta(t, 1.0, sum, 0.001)

	// Verify alternative priorities sum to 1
	sum = 0.0
	for _, priority := range output.AlternativePriorities {
		sum += priority
	}
	assert.InDelta(t, 1.0, sum, 0.001)

	// Verify ranking
	assert.Len(t, output.Ranking, 2)
	assert.Equal(t, 1, output.Ranking[0].Rank)
	assert.Equal(t, 2, output.Ranking[1].Rank)

	// goal1 should be ranked higher
	assert.Equal(t, "goal1", output.Ranking[0].AlternativeID)
}

func TestService_ExecuteAHP_ValidationError(t *testing.T) {
	logger := zap.NewNop()
	svc := NewService(logger)
	ctx := context.Background()

	tests := []struct {
		name  string
		input *dto.AHPInput
	}{
		{
			name: "Missing criteria",
			input: &dto.AHPInput{
				UserID:                 "user123",
				Criteria:               []domain.Criteria{{ID: "c1"}}, // Only 1
				Alternatives:           []domain.Alternative{{ID: "a1"}, {ID: "a2"}},
				CriteriaComparisons:    []domain.PairwiseComparison{},
				AlternativeComparisons: map[string][]domain.PairwiseComparison{},
			},
		},
		{
			name: "Missing alternatives",
			input: &dto.AHPInput{
				UserID:                 "user123",
				Criteria:               []domain.Criteria{{ID: "c1"}, {ID: "c2"}},
				Alternatives:           []domain.Alternative{{ID: "a1"}}, // Only 1
				CriteriaComparisons:    []domain.PairwiseComparison{},
				AlternativeComparisons: map[string][]domain.PairwiseComparison{},
			},
		},
		{
			name: "Incorrect number of comparisons",
			input: &dto.AHPInput{
				UserID:              "user123",
				Criteria:            []domain.Criteria{{ID: "c1"}, {ID: "c2"}},
				Alternatives:        []domain.Alternative{{ID: "a1"}, {ID: "a2"}},
				CriteriaComparisons: []domain.PairwiseComparison{
					// Missing comparison
				},
				AlternativeComparisons: map[string][]domain.PairwiseComparison{
					"c1": {{ElementA: "a1", ElementB: "a2", Value: 2.0}},
					"c2": {{ElementA: "a1", ElementB: "a2", Value: 3.0}},
				},
			},
		},
		{
			name: "Invalid comparison value",
			input: &dto.AHPInput{
				UserID:       "user123",
				Criteria:     []domain.Criteria{{ID: "c1"}, {ID: "c2"}},
				Alternatives: []domain.Alternative{{ID: "a1"}, {ID: "a2"}},
				CriteriaComparisons: []domain.PairwiseComparison{
					{ElementA: "c1", ElementB: "c2", Value: 15.0}, // Out of range
				},
				AlternativeComparisons: map[string][]domain.PairwiseComparison{
					"c1": {{ElementA: "a1", ElementB: "a2", Value: 2.0}},
					"c2": {{ElementA: "a1", ElementB: "a2", Value: 3.0}},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, err := svc.ExecuteAHP(ctx, tt.input)

			assert.Error(t, err)
			assert.Nil(t, output)
		})
	}
}

func TestService_ExecuteAHP_MultipleScenarios(t *testing.T) {
	logger := zap.NewNop()
	svc := NewService(logger)
	ctx := context.Background()

	t.Run("3 criteria, 3 alternatives", func(t *testing.T) {
		input := &dto.AHPInput{
			UserID: "user123",
			Criteria: []domain.Criteria{
				{ID: "c1", Name: "Impact"},
				{ID: "c2", Name: "Urgency"},
				{ID: "c3", Name: "Cost"},
			},
			CriteriaComparisons: []domain.PairwiseComparison{
				{ElementA: "c1", ElementB: "c2", Value: 3.0},
				{ElementA: "c1", ElementB: "c3", Value: 5.0},
				{ElementA: "c2", ElementB: "c3", Value: 2.0},
			},
			Alternatives: []domain.Alternative{
				{ID: "a1", Name: "Goal A"},
				{ID: "a2", Name: "Goal B"},
				{ID: "a3", Name: "Goal C"},
			},
			AlternativeComparisons: map[string][]domain.PairwiseComparison{
				"c1": {
					{ElementA: "a1", ElementB: "a2", Value: 2.0},
					{ElementA: "a1", ElementB: "a3", Value: 4.0},
					{ElementA: "a2", ElementB: "a3", Value: 2.0},
				},
				"c2": {
					{ElementA: "a1", ElementB: "a2", Value: 1.0},
					{ElementA: "a1", ElementB: "a3", Value: 3.0},
					{ElementA: "a2", ElementB: "a3", Value: 3.0},
				},
				"c3": {
					{ElementA: "a1", ElementB: "a2", Value: 0.5},
					{ElementA: "a1", ElementB: "a3", Value: 2.0},
					{ElementA: "a2", ElementB: "a3", Value: 4.0},
				},
			},
		}

		t.Logf("\n=== INPUT: 3x3 Scenario ===")
		t.Logf("Criteria: %v", []string{"Impact", "Urgency", "Cost"})
		t.Logf("Alternatives: %v", []string{"Goal A", "Goal B", "Goal C"})
		t.Logf("Criteria Comparisons: Impact>Urgency(3x), Impact>Cost(5x), Urgency>Cost(2x)")

		output, err := svc.ExecuteAHP(ctx, input)

		require.NoError(t, err)
		assert.NotNil(t, output)

		t.Logf("\n=== OUTPUT ===")
		t.Logf("Consistency Ratio: %.4f", output.ConsistencyRatio)
		t.Logf("Ranking: #1=%s(%.3f) #2=%s(%.3f) #3=%s(%.3f)",
			output.Ranking[0].AlternativeName, output.Ranking[0].Priority,
			output.Ranking[1].AlternativeName, output.Ranking[1].Priority,
			output.Ranking[2].AlternativeName, output.Ranking[2].Priority)

		// Verify all 3 criteria have weights
		assert.Len(t, output.CriteriaWeights, 3)

		// Verify all 3 alternatives have priorities
		assert.Len(t, output.AlternativePriorities, 3)

		// Verify local priorities for each criterion
		assert.Len(t, output.LocalPriorities, 3)
		for _, criterion := range input.Criteria {
			assert.Contains(t, output.LocalPriorities, criterion.ID)
			assert.Len(t, output.LocalPriorities[criterion.ID], 3)
		}

		// Verify ranking
		assert.Len(t, output.Ranking, 3)
		assert.Equal(t, 1, output.Ranking[0].Rank)
		assert.Equal(t, 2, output.Ranking[1].Rank)
		assert.Equal(t, 3, output.Ranking[2].Rank)
	})

	t.Run("Equal weights scenario", func(t *testing.T) {
		input := &dto.AHPInput{
			UserID: "user123",
			Criteria: []domain.Criteria{
				{ID: "c1", Name: "Criteria 1"},
				{ID: "c2", Name: "Criteria 2"},
			},
			CriteriaComparisons: []domain.PairwiseComparison{
				{ElementA: "c1", ElementB: "c2", Value: 1.0}, // Equal
			},
			Alternatives: []domain.Alternative{
				{ID: "a1", Name: "Alternative 1"},
				{ID: "a2", Name: "Alternative 2"},
			},
			AlternativeComparisons: map[string][]domain.PairwiseComparison{
				"c1": {
					{ElementA: "a1", ElementB: "a2", Value: 1.0}, // Equal
				},
				"c2": {
					{ElementA: "a1", ElementB: "a2", Value: 1.0}, // Equal
				},
			},
		}

		t.Logf("\n=== INPUT: Equal Weights ===")
		t.Logf("All comparisons = 1.0 (equal importance)")

		output, err := svc.ExecuteAHP(ctx, input)

		require.NoError(t, err)
		assert.NotNil(t, output)

		t.Logf("\n=== OUTPUT ===")
		t.Logf("Criteria Weights: c1=%.3f, c2=%.3f", output.CriteriaWeights["c1"], output.CriteriaWeights["c2"])
		t.Logf("Alternative Priorities: a1=%.3f, a2=%.3f", output.AlternativePriorities["a1"], output.AlternativePriorities["a2"])
		t.Logf("Consistency Ratio: %.4f (Perfect)", output.ConsistencyRatio)

		// All criteria should have equal weights
		assert.InDelta(t, 0.5, output.CriteriaWeights["c1"], 0.01)
		assert.InDelta(t, 0.5, output.CriteriaWeights["c2"], 0.01)

		// All alternatives should have equal priorities
		assert.InDelta(t, 0.5, output.AlternativePriorities["a1"], 0.01)
		assert.InDelta(t, 0.5, output.AlternativePriorities["a2"], 0.01)

		// CR should be perfect (0)
		assert.Equal(t, 0.0, output.ConsistencyRatio)
		assert.True(t, output.IsConsistent)
	})

	t.Run("Strong preference scenario", func(t *testing.T) {
		input := &dto.AHPInput{
			UserID: "user123",
			Criteria: []domain.Criteria{
				{ID: "c1", Name: "Very Important"},
				{ID: "c2", Name: "Less Important"},
			},
			CriteriaComparisons: []domain.PairwiseComparison{
				{ElementA: "c1", ElementB: "c2", Value: 9.0}, // Maximum preference
			},
			Alternatives: []domain.Alternative{
				{ID: "a1", Name: "Best Choice"},
				{ID: "a2", Name: "Poor Choice"},
			},
			AlternativeComparisons: map[string][]domain.PairwiseComparison{
				"c1": {
					{ElementA: "a1", ElementB: "a2", Value: 9.0}, // Maximum
				},
				"c2": {
					{ElementA: "a1", ElementB: "a2", Value: 9.0}, // Maximum
				},
			},
		}

		t.Logf("\n=== INPUT: Strong Preference ===")
		t.Logf("All comparisons = 9.0 (maximum preference)")
		t.Logf("c1 ('Very Important') >> c2 ('Less Important')")
		t.Logf("a1 ('Best Choice') >> a2 ('Poor Choice')")

		output, err := svc.ExecuteAHP(ctx, input)

		require.NoError(t, err)
		assert.NotNil(t, output)

		t.Logf("\n=== OUTPUT ===")
		t.Logf("Criteria Weights: c1=%.3f (%.0f%%), c2=%.3f (%.0f%%)",
			output.CriteriaWeights["c1"], output.CriteriaWeights["c1"]*100,
			output.CriteriaWeights["c2"], output.CriteriaWeights["c2"]*100)
		t.Logf("Alternative Priorities: a1=%.3f (%.0f%%), a2=%.3f (%.0f%%)",
			output.AlternativePriorities["a1"], output.AlternativePriorities["a1"]*100,
			output.AlternativePriorities["a2"], output.AlternativePriorities["a2"]*100)
		t.Logf("Winner: %s with %.1f%% priority", output.Ranking[0].AlternativeName, output.Ranking[0].Priority*100)

		// c1 should dominate
		assert.Greater(t, output.CriteriaWeights["c1"], 0.8)
		assert.Less(t, output.CriteriaWeights["c2"], 0.2)

		// a1 should strongly dominate
		assert.Greater(t, output.AlternativePriorities["a1"], 0.8)
		assert.Less(t, output.AlternativePriorities["a2"], 0.2)

		// a1 should be ranked first
		assert.Equal(t, "a1", output.Ranking[0].AlternativeID)
		assert.Equal(t, 1, output.Ranking[0].Rank)
	})
}

func TestService_ExecuteAHP_WithLogging(t *testing.T) {
	// Create a logger that captures output for testing
	config := zap.NewDevelopmentConfig()
	config.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	logger, err := config.Build()
	require.NoError(t, err)
	defer logger.Sync()

	svc := NewService(logger)
	ctx := context.Background()

	input := &dto.AHPInput{
		UserID: "user123",
		Criteria: []domain.Criteria{
			{ID: "c1", Name: "Criteria 1"},
			{ID: "c2", Name: "Criteria 2"},
		},
		CriteriaComparisons: []domain.PairwiseComparison{
			{ElementA: "c1", ElementB: "c2", Value: 2.0},
		},
		Alternatives: []domain.Alternative{
			{ID: "a1", Name: "Alternative 1"},
			{ID: "a2", Name: "Alternative 2"},
		},
		AlternativeComparisons: map[string][]domain.PairwiseComparison{
			"c1": {{ElementA: "a1", ElementB: "a2", Value: 3.0}},
			"c2": {{ElementA: "a1", ElementB: "a2", Value: 2.0}},
		},
	}

	output, err := svc.ExecuteAHP(ctx, input)

	require.NoError(t, err)
	assert.NotNil(t, output)

	// If we got here, logging didn't cause any panics or errors
	// More detailed logging tests would require capturing log output
}
